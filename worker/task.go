package worker

import (
	"sync"

	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/sort"
	"go.uber.org/zap"
)

type StatefulReporter struct {
	Reporter func(s pb.TaskState, metadata *pb.TaskData)

	// mu protects the fields below.
	mu    sync.Mutex
	state pb.TaskState
}

func (t *StatefulReporter) Report(state pb.TaskState, data *pb.TaskData) {
	t.mu.Lock()
	t.state = state
	t.mu.Unlock()

	t.Reporter(state, data)
}

func (t *StatefulReporter) State() pb.TaskState {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.state
}

type Task interface {
	Report(pb.TaskState, *pb.TaskData)
	Name() string
	Run() error
	State() pb.TaskState
	Abort()
}

type BasicTask struct {
	*StatefulReporter

	Log     *zap.SugaredLogger
	Spec    *pb.Task
	Context *JobContext
}

func (t *BasicTask) Name() string {
	return t.Spec.Name
}

type JobContext struct {
	sync.Mutex

	// sortRunSize describes the size of each run during sort.
	// typically set to the max memory of the worker.
	sortRunSize    int
	sortRunFile    string
	sortRunMarkers []sort.Marker

	// shuffleRecvFiles contains all files collected by shuffleRecvTask.
	shuffleRecvFiles []string
	shuffleRecvHost  string

	// mu protects the fields below.
	done bool
}

func (ctx *JobContext) Alive() bool {
	ctx.Lock()
	defer ctx.Unlock()

	return !ctx.done
}

func (ctx *JobContext) Kill() {
	ctx.Lock()
	defer ctx.Unlock()
	ctx.done = true
}

func NewJobContext(wk *Worker) *JobContext {
	return &JobContext{
		sortRunSize:     int(wk.Memory),
		shuffleRecvHost: wk.TCPHost,
	}
}
