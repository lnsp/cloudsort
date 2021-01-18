package worker

type ExecutionQueue struct {
	token chan struct{}
}

func NewExecutionQueue() *ExecutionQueue {
	bw := &ExecutionQueue{
		token: make(chan struct{}, 1),
	}
	bw.token <- struct{}{}
	return bw
}

func (q *ExecutionQueue) Run(f func()) {
	<-q.token
	f()
	q.token <- struct{}{}
}
