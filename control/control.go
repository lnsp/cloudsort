package control

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lnsp/cloudsort/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Control struct {
	Address string

	mu         sync.Mutex
	workers    map[string]*Worker
	currentJob *Job

	pb.UnimplementedControlServer
}

type Worker struct {
	Address string

	// mu protects the fields below.
	mu            sync.Mutex
	lastHeartbeat time.Time
	activeTasks   map[string]*pb.Task
}

func NewWorker(addr string) *Worker {
	return &Worker{
		Address:     addr,
		activeTasks: make(map[string]*pb.Task),
	}
}

func (w *Worker) Heartbeat(heartbeat time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.lastHeartbeat = heartbeat
}

func (w *Worker) Assign(task *pb.Task) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.activeTasks[task.Name] = task
}

func (w *Worker) Drop(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.activeTasks, name)
}

type Event struct {
	Progress  float32
	Message   string
	Timestamp time.Time
}

func (ctl *Control) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	ctl.mu.Lock()
	worker, ok := ctl.workers[req.Address]
	if !ok {
		worker = NewWorker(req.Address)
		ctl.workers[req.Address] = worker
		zap.S().With("worker", req.Address).Info("Worker joined cluster")
	}
	job := ctl.currentJob
	ctl.mu.Unlock()

	// Update current worker stats
	worker.Heartbeat(time.Now())

	taskAvailable := false
	if job != nil {
		taskAvailable = job.HasScheduledTask(worker)
	}

	// Check if task is available
	return &pb.HeartbeatResponse{
		TaskAvailable: taskAvailable,
	}, nil
}

func (ctl *Control) ReportTask(ctx context.Context, req *pb.ReportTaskRequest) (*pb.ReportTaskResponse, error) {
	ctl.mu.Lock()
	job := ctl.currentJob
	ctl.mu.Unlock()

	job.UpdateState(req.Name, req.State, req.Data)
	job.events <- Event{
		Message:   fmt.Sprintf("%s=%s (%v)", req.Name, req.State.String(), req.Data),
		Timestamp: time.Now(),
	}

	return &pb.ReportTaskResponse{}, nil
}

func (ctl *Control) dropCurrentJob() error {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	close(ctl.currentJob.events)
	ctl.currentJob = nil
	return nil
}

func (ctl *Control) updateCurrentJob(name string, creds *pb.S3Credentials, workerPool []*Worker) (*Job, error) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	if ctl.currentJob != nil {
		return nil, status.Errorf(codes.AlreadyExists, "another job is already active")
	} else if len(ctl.workers) < 1 {
		return nil, status.Errorf(codes.Unavailable, "no worker available")
	}

	ctl.currentJob = NewJob(name, workerPool, creds)
	return ctl.currentJob, nil
}

func (ctl *Control) ActiveWorkerPool() []*Worker {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	pool := make([]*Worker, len(ctl.workers))
	i := 0
	for _, w := range ctl.workers {
		pool[i] = w
		i++
	}
	return pool
}

func (ctl *Control) SubmitJob(req *pb.SubmitJobRequest, resp pb.Control_SubmitJobServer) error {
	job, err := ctl.updateCurrentJob(req.Name, req.Creds, ctl.ActiveWorkerPool())
	if err != nil {
		return err
	}
	zap.S().Infof("Received new job")
	// Start job handout
	go func() {
		job.events <- Event{
			Message:   "Job scheduled",
			Timestamp: time.Now(),
		}
		if err := job.Initialize(); err != nil {
			job.events <- Event{
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Prepare job: %v", err),
			}
			ctl.dropCurrentJob()
		}
	}()
	// Close event channel after job done
	for evt := range job.events {
		ts, _ := ptypes.TimestampProto(evt.Timestamp)
		resp.Send(&pb.SubmitJobResponse{
			Event: &pb.Event{
				Progress:  evt.Progress,
				Message:   evt.Message,
				Timestamp: ts,
			},
		})
	}
	return nil
}

func (ctl *Control) PullTask(ctx context.Context, req *pb.PullTaskRequest) (*pb.PullTaskResponse, error) {
	ctl.mu.Lock()
	worker := ctl.workers[req.Address]
	job := ctl.currentJob
	ctl.mu.Unlock()

	task, err := job.ScheduleTask(worker)
	if err == ErrJobFinished {
		ctl.dropCurrentJob()
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "schedule task: %v", err)
	}

	if task == nil {
		return nil, status.Errorf(codes.NotFound, "no task available")
	}
	zap.S().With("task", task.Name, "worker", worker.Address).Debug("Scheduled task to worker")
	return &pb.PullTaskResponse{
		Task: task,
	}, nil
}

func New(addr string) (*Control, error) {
	ctl := &Control{
		Address: addr,
		workers: make(map[string]*Worker),
	}
	zap.S().With("address", addr).Info("Created control server")
	return ctl, nil
}
