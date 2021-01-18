package control

import (
	"sync"

	"github.com/lnsp/cloudsort/pb"
)

type FCFSSchedule struct {
	// mu protects the fields below.
	mu         sync.Mutex
	head, tail *FCFSScheduleItem
}

func (tq *FCFSSchedule) Enqueue(task *pb.Task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	// Check if queue is empty
	item := &FCFSScheduleItem{Task: task}
	if tq.head == nil {
		tq.head, tq.tail = item, item
		return
	}
	tq.tail, tq.tail.Next = item, item
}

func (tq *FCFSSchedule) Dequeue() *pb.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if tq.head == nil {
		return nil
	}
	task := tq.head.Task
	tq.head = tq.head.Next
	return task
}

func (tq *FCFSSchedule) Empty() bool {
	return tq.head == nil
}

func (tq *FCFSSchedule) Peek() *pb.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if tq.head == nil {
		return nil
	}
	return tq.head.Task
}

type FCFSScheduleItem struct {
	Task *pb.Task
	Next *FCFSScheduleItem
}

type OnceSchedule struct {
	// mu protects the fields below.
	mu        sync.Mutex
	tasks     []*pb.Task
	scheduled []bool
	count     int
}

func NewOnceSchedule(poolSize int) *OnceSchedule {
	return &OnceSchedule{
		tasks:     make([]*pb.Task, poolSize),
		scheduled: make([]bool, poolSize),
	}
}

func (mq *OnceSchedule) Assign(i int, t *pb.Task) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.tasks[i] = t
}

func (mq *OnceSchedule) Index(i int) *pb.Task {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.tasks[i]
}

func (mq *OnceSchedule) Schedule(i int) bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.scheduled[i] {
		return false
	}
	mq.scheduled[i] = true
	mq.count++
	return mq.count == len(mq.scheduled)
}

func (mq *OnceSchedule) AllScheduled() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.count == len(mq.scheduled)
}
