package control

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/lnsp/cloudsort/control/workgraph"
	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/storage"
	"github.com/lnsp/cloudsort/structs"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type Job struct {
	log         *zap.SugaredLogger
	name        string
	credentials *pb.S3Credentials
	events      chan Event
	pool        []*Worker

	// mu protects the fields below.
	mu        sync.Mutex
	workgraph *workgraph.WorkGraph
	tasks     map[string]int
	peers     []*pb.Peer
	size      int64
}

// indexOfWorker assumes that WorkerPool is read-only.
func (j *Job) indexOfWorker(address string) int {
	return sort.Search(len(j.pool), func(i int) bool { return j.pool[i].Address >= address })
}

func (j *Job) determineObjectSize() error {
	// First, we need to know how big the file is
	client, err := storage.GetS3Client(j.credentials)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	stat, err := client.StatObject(ctx, j.credentials.BucketId, j.credentials.ObjectKey, minio.StatObjectOptions{})
	if err != nil {
		return err
	}
	j.size = stat.Size / structs.EntrySize
	return nil
}

func (j *Job) Initialize() error {
	// Determine key ranges
	keyRanges := GetKeyRanges(len(j.pool))
	// Assign key ranges to peers
	for i := range keyRanges {
		j.peers[i] = &pb.Peer{
			KeyRangeStart: keyRanges[i].Start,
			KeyRangeEnd:   keyRanges[i].End,
		}
	}

	// Determine object size
	if err := j.determineObjectSize(); err != nil {
		return err
	}

	// Determine number of splits
	j.log.With("size", j.size, "poolsize", len(j.pool)).Info("Initializing job schedule")
	// We can now get the real file size

	var (
		chunkSplits = int64(len(j.pool))
		chunkSize   = j.size / chunkSplits
	)

	// Generate tasks for sort
	for i := int64(0); i < chunkSplits; i++ {
		chunkBegin := chunkSize * i
		chunkEnd := chunkSize*(i+1) - 1
		if i == chunkSplits-1 {
			chunkEnd = j.size - 1
		}

		task := &pb.Task{
			Name: fmt.Sprintf("sort-%d", i),
			Job:  j.name,
			Type: pb.TaskType_SORT,
			Details: &pb.Task_Sort{
				Sort: &pb.SortTask{
					Credentials: j.credentials,
					RangeStart:  chunkBegin,
					RangeEnd:    chunkEnd,
				},
			},
		}
		j.tasks[task.Name] = j.workgraph.AddItem(&workgraph.Item{
			Task:   task,
			Worker: j.pool[i].Address,
		})
		j.workgraph.AddRelation(j.tasks[task.Name], workgraph.Done, workgraph.Source)
	}

	// Generate tasks for shuffleRecv
	for i := int64(0); i < chunkSplits; i++ {
		task := &pb.Task{
			Name: fmt.Sprintf("shufflerecv-%d", i),
			Job:  j.name,
			Type: pb.TaskType_SHUFFLE_RECV,
			Details: &pb.Task_ShuffleRecv{
				ShuffleRecv: &pb.ShuffleRecvTask{
					NumberOfPeers: chunkSplits,
				},
			},
		}
		j.tasks[task.Name] = j.workgraph.AddItem(&workgraph.Item{
			Task:   task,
			Worker: j.pool[i].Address,
		})
		for k := int64(0); k < chunkSplits; k++ {
			j.workgraph.AddRelation(j.tasks[task.Name], workgraph.Done, j.tasks[fmt.Sprintf("sort-%d", k)])
		}
	}

	// Generate tasks for shuffleSend
	for i := int64(0); i < chunkSplits; i++ {
		task := &pb.Task{
			Name: fmt.Sprintf("shufflesend-%d", i),
			Job:  j.name,
			Type: pb.TaskType_SHUFFLE_SEND,
			Details: &pb.Task_ShuffleSend{
				ShuffleSend: &pb.ShuffleSendTask{
					Peers: j.peers,
				},
			},
		}
		j.tasks[task.Name] = j.workgraph.AddItem(&workgraph.Item{
			Task:   task,
			Worker: j.pool[i].Address,
		})
		for k := int64(0); k < chunkSplits; k++ {
			j.workgraph.AddRelation(j.tasks[task.Name], workgraph.InProgress, j.tasks[fmt.Sprintf("shufflerecv-%d", k)])
		}
	}

	// Generate tasks for upload
	for i := int64(0); i < chunkSplits; i++ {
		task := &pb.Task{
			Name: fmt.Sprintf("upload-%d", i),
			Job:  j.name,
			Type: pb.TaskType_UPLOAD,
			Details: &pb.Task_Upload{
				Upload: &pb.UploadTask{
					Credentials: j.credentials,
				},
			},
		}
		j.tasks[task.Name] = j.workgraph.AddItem(&workgraph.Item{
			Task:   task,
			Worker: j.pool[i].Address,
		})
		j.workgraph.AddRelation(j.tasks[task.Name], workgraph.Done, j.tasks[fmt.Sprintf("shufflerecv-%d", i)])
		j.workgraph.AddRelation(workgraph.Sink, workgraph.Done, j.tasks[task.Name])
	}

	return nil
}

var ErrUnknownState = errors.New("unknown state change")

func (j *Job) UpdateState(name string, state pb.TaskState, payload *pb.TaskData) {
	j.mu.Lock()
	defer j.mu.Unlock()

	task := j.tasks[name]
	worker := j.workgraph.Items[task].Worker
	index := j.indexOfWorker(worker)

	// In case of payload data, handle that
	if payload != nil {
		switch p := payload.Properties.(type) {
		case *pb.TaskData_ShuffleRecvAddr:
			j.peers[index].Address = p.ShuffleRecvAddr
		}
	}

	// Update state of task
	var err error
	switch state {
	case pb.TaskState_IN_PROGRESS:
		err = j.workgraph.BeginItem(task)
	case pb.TaskState_FAILED:
		err = j.workgraph.FailItem(task)
	case pb.TaskState_DONE:
		err = j.workgraph.FinishItem(task)
	case pb.TaskState_ACCEPTED:
	default:
		err = ErrUnknownState
	}

	l := j.log.With("state", state, "name", name)
	if err != nil {
		l.With("error", err).Warn("Failed update state")
	} else {
		l.Debug("Update state")
	}
}

var ErrJobFinished = errors.New("job is finished")

func (j *Job) ScheduleTask(worker *Worker) (*pb.Task, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	t, err := j.workgraph.Schedule(worker.Address)
	if err == workgraph.ErrNoItems {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else if t.ID == workgraph.Sink {
		return nil, ErrJobFinished
	}
	return t.Task, nil
}

// HasScheduledTask returns true if there is a task assigned to the worker.
func (j *Job) HasScheduledTask(worker *Worker) bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	_, err := j.workgraph.CanSchedule(worker.Address)
	return err == nil
}

func NewJob(name string, pool []*Worker, s3creds *pb.S3Credentials) *Job {
	// Sort worker pool by address, so we can do binary search later on
	sort.Slice(pool, func(i, j int) bool { return pool[i].Address < pool[j].Address })
	return &Job{
		name:        name,
		credentials: s3creds,
		events:      make(chan Event),
		pool:        pool,
		log:         zap.S(),

		workgraph: workgraph.New(),
		tasks:     make(map[string]int),
		peers:     make([]*pb.Peer, len(pool)),
	}
}
