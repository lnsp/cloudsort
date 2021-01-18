package workgraph

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/lnsp/cloudsort/pb"
)

const (
	Source = 0
	Sink   = 1
)

type State int32

func (s State) Matches(t State) bool {
	switch s {
	case Waiting:
		return true
	case Scheduled:
		switch t {
		case Scheduled, InProgress, Done:
			return true
		}
	case InProgress:
		switch t {
		case InProgress, Done:
			return true
		}
	case Failed:
		switch t {
		case Failed:
			return true
		}
	case Done:
		switch t {
		case Done:
			return true
		}
	}
	return false
}

const (
	Waiting State = iota + 1
	Scheduled
	InProgress
	Failed
	Done
)

func (s State) String() string {
	switch s {
	case Waiting:
		return "waiting"
	case Scheduled:
		return "scheduled"
	case InProgress:
		return "started"
	case Failed:
		return "failed"
	case Done:
		return "done"
	}
	return "unknown"
}

type Item struct {
	ID     int
	Task   *pb.Task
	Worker string

	State     State
	DependsOn []Relation
}

func (item *Item) Ready() bool {
	for _, rel := range item.DependsOn {
		if !rel.State.Matches(rel.Target.State) {
			return false
		}
	}
	return true
}

type Relation struct {
	State  State
	Target *Item
}

type WorkGraph struct {
	Items []*Item
}

func New() *WorkGraph {
	source := &Item{
		ID:    0,
		State: Done,
	}
	sink := &Item{
		ID:        1,
		State:     Waiting,
		DependsOn: []Relation{{Done, source}},
	}
	return &WorkGraph{
		Items: []*Item{source, sink},
	}
}

func (wg *WorkGraph) AddItem(item *Item) int {
	item.ID = len(wg.Items)
	item.State = Waiting
	wg.Items = append(wg.Items, item)
	return item.ID
}

func (wg *WorkGraph) AddRelation(item int, state State, dependency int) {
	wg.Items[item].DependsOn = append(wg.Items[item].DependsOn, Relation{state, wg.Items[dependency]})
}

var ErrBadState = errors.New("state transition not allowed")

func (wg *WorkGraph) BeginItem(item int) error {
	if wg.Items[item].State != Scheduled {
		return ErrBadState
	}
	wg.Items[item].State = InProgress
	return nil
}

func (wg *WorkGraph) FailItem(item int) error {
	switch wg.Items[item].State {
	default:
		return ErrBadState
	case Scheduled, InProgress:
	}
	wg.Items[item].State = Failed
	return nil
}

func (wg *WorkGraph) FinishItem(item int) error {
	switch wg.Items[item].State {
	default:
		return ErrBadState
	case Scheduled, InProgress:
	}
	wg.Items[item].State = Done
	return nil
}

var ErrNoItems = errors.New("no items found")

func (wg *WorkGraph) CanSchedule(worker string) (*Item, error) {
	for _, item := range wg.Items {
		// The item might be processing already
		if item.State != Waiting {
			continue
		}
		// The item may be assigned to another worker
		if item.Worker != worker && len(item.Worker) != 0 {
			continue
		}
		// The item may not have all its dependencies fulfilled
		if !item.Ready() {
			continue
		}
		// Change the state to Scheduled
		return item, nil
	}
	return nil, ErrNoItems
}

func (wg *WorkGraph) Schedule(worker string) (*Item, error) {
	item, err := wg.CanSchedule(worker)
	if err != nil {
		return nil, err
	}
	item.State = Scheduled
	item.Worker = worker
	return item, nil
}

func (wg *WorkGraph) Graphviz() []byte {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, "digraph G {")

	// Spill out items
	for _, item := range wg.Items {
		fmt.Fprintln(buf, item.ID, ";")
		for _, rel := range item.DependsOn {
			fmt.Fprintln(buf, rel.Target.ID, "->", item.ID, "[label=", rel.State.String(), "]", ";")
		}
	}

	fmt.Fprintln(buf, "}")
	return buf.Bytes()
}
