package worker

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/lnsp/cloudsort/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Worker struct {
	TCPHost     string
	GRPCAddress string
	Control     pb.ControlClient
	Memory      int

	heartbeat *time.Timer

	// mu protects the fields below.
	mu       sync.Mutex
	tasks    map[string]Task
	contexts map[string]*JobContext
}

const (
	heartbeatInterval     = 5 * time.Second
	heartbeatFetchBackoff = 500 * time.Millisecond
)

func (wk *Worker) sendHeartbeat() {
	defer func() {
		wk.mu.Lock()
		wk.heartbeat.Reset(heartbeatInterval)
		wk.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), heartbeatInterval)
	defer cancel()

	resp, err := wk.Control.Heartbeat(ctx, &pb.HeartbeatRequest{
		Address: wk.GRPCAddress,
	})
	if err != nil {
		zap.S().Error("Heartbeat failed")
		return
	} else if !resp.TaskAvailable {
		return
	}
	go wk.fetchTask()
}

func (wk *Worker) retrieveContext(job string) *JobContext {
	wk.mu.Lock()
	defer wk.mu.Unlock()

	_, ok := wk.contexts[job]
	if !ok {
		zap.S().With("job", job).Debug("Create new context")
		wk.contexts[job] = NewJobContext(wk)
	}
	return wk.contexts[job]
}

func (wk *Worker) releaseContext(job string) {
	wk.mu.Lock()
	defer wk.mu.Unlock()

	delete(wk.contexts, job)
}

func (wk *Worker) fetchTask() {
	zap.S().Info("Detected available task")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	specs, err := wk.Control.PullTask(ctx, &pb.PullTaskRequest{
		Address: wk.GRPCAddress,
	})
	if err != nil {
		zap.S().Error("Fetch task:", err)
		return
	}
	reporter := &StatefulReporter{
		Reporter: func(state pb.TaskState, data *pb.TaskData) {
			if _, err := wk.Control.ReportTask(context.TODO(), &pb.ReportTaskRequest{
				Name:  specs.Task.Name,
				State: state,
				Data:  data,
			}); err != nil {
				zap.S().Errorf("Report task %s with state %s: %s", specs.Task.Name, state, err)
			}
		},
	}
	var task Task
	jobContext := wk.retrieveContext(specs.Task.Job)
	base := &BasicTask{
		StatefulReporter: reporter,
		Log:              zap.S().With("task", specs.Task.Name, "job", specs.Task.Job),
		Spec:             specs.Task,
		Context:          jobContext,
	}
	switch specs.Task.Type {
	case pb.TaskType_SORT:
		task = &SortTask{base}
	case pb.TaskType_SAMPLE:
		// TODO
	case pb.TaskType_SHUFFLE_SEND:
		task = &ShuffleSendTask{base}
	case pb.TaskType_SHUFFLE_RECV:
		task = &ShuffleRecvTask{base}
	case pb.TaskType_UPLOAD:
		task = &UploadTask{base}
	default:
		zap.S().Errorf("Received task with unknown type %s", specs.Task.Type)
		return
	}

	wk.bindTask(task)
	// Start running task
	go func() {
		wk.runTask(task)
		// Free up resources
		wk.unbindTask(task)
		if !jobContext.Alive() {
			wk.releaseContext(specs.Task.Job)
		}
	}()
}

func (wk *Worker) bindTask(task Task) {
	wk.mu.Lock()
	defer wk.mu.Unlock()

	zap.S().Debugf("Binding task %s to worker", task.Name())
	wk.tasks[task.Name()] = task
}

func (wk *Worker) unbindTask(task Task) {
	wk.mu.Lock()
	defer wk.mu.Unlock()

	delete(wk.tasks, task.Name())
}

func (wk *Worker) runTask(task Task) {
	defer runtime.GC()

	task.Report(pb.TaskState_ACCEPTED, nil)
	if err := task.Run(); err != nil {
		task.Report(pb.TaskState_FAILED, nil)
		return
	}
	task.Report(pb.TaskState_DONE, nil)
}

func (wk *Worker) Wait() error {
	c := make(chan error)
	// TODO: Make use of this channel somehow
	return <-c
}

func New(workerAddr, controlAddr, workerHost string, totalMemory int) (*Worker, error) {
	grpcClient, err := grpc.Dial(controlAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	cc := pb.NewControlClient(grpcClient)
	// Announce to join cluster
	wk := &Worker{
		TCPHost:     workerHost,
		GRPCAddress: workerAddr,
		Control:     cc,
		Memory:      totalMemory,

		tasks:    make(map[string]Task),
		contexts: make(map[string]*JobContext),
	}
	// Setup heartbeat timer
	wk.mu.Lock()
	wk.heartbeat = time.AfterFunc(heartbeatInterval, wk.sendHeartbeat)
	wk.mu.Unlock()
	// Log that worker has been created
	zap.S().Debugf("Created worker server listening on %s", workerAddr)
	return wk, nil
}
