package workgraph_test

import (
	"fmt"
	"testing"

	"github.com/lnsp/cloudsort/control/workgraph"
	"github.com/stretchr/testify/assert"
)

func TestBasicGraph(t *testing.T) {
	wg := workgraph.New()

	// Should receive sink item
	item, err := wg.Schedule("worker1")
	if assert.NoError(t, err) {
		assert.Equal(t, workgraph.Sink, item.ID)
		assert.Equal(t, workgraph.Scheduled, item.State)
		assert.Equal(t, "worker1", item.Worker)
	}

	// Should throw error
	_, err = wg.Schedule("worker1")
	assert.Error(t, err)

	// Finish item, should throw error
	assert.NoError(t, wg.BeginItem(item.ID))
	assert.NoError(t, wg.FinishItem(item.ID))

	// Should throw error
	item, err = wg.Schedule("worker1")
	if assert.Error(t, err) {
		assert.Nil(t, item)
	}

	fmt.Println(string(wg.Graphviz()))
}

func buildCloudsortGraph() (*workgraph.WorkGraph, []int, []int, []int, []int) {
	wg := workgraph.New()

	// Add one sort task per worker
	sort := []int{
		wg.AddItem(&workgraph.Item{Worker: "w1"}),
		wg.AddItem(&workgraph.Item{Worker: "w2"}),
		wg.AddItem(&workgraph.Item{Worker: "w3"}),
	}

	// Wait for DAG to start
	wg.AddRelation(sort[0], workgraph.Done, workgraph.Source)
	wg.AddRelation(sort[1], workgraph.Done, workgraph.Source)
	wg.AddRelation(sort[2], workgraph.Done, workgraph.Source)

	// Add one merge recv task per worker
	mergeRecv := []int{
		wg.AddItem(&workgraph.Item{Worker: "w1"}),
		wg.AddItem(&workgraph.Item{Worker: "w2"}),
		wg.AddItem(&workgraph.Item{Worker: "w3"}),
	}

	// Wait for sort tasks to finish
	wg.AddRelation(mergeRecv[0], workgraph.Done, sort[0])
	wg.AddRelation(mergeRecv[0], workgraph.Done, sort[1])
	wg.AddRelation(mergeRecv[0], workgraph.Done, sort[2])

	wg.AddRelation(mergeRecv[1], workgraph.Done, sort[0])
	wg.AddRelation(mergeRecv[1], workgraph.Done, sort[1])
	wg.AddRelation(mergeRecv[1], workgraph.Done, sort[2])

	wg.AddRelation(mergeRecv[2], workgraph.Done, sort[0])
	wg.AddRelation(mergeRecv[2], workgraph.Done, sort[1])
	wg.AddRelation(mergeRecv[2], workgraph.Done, sort[2])

	// Add one merge send per worker
	mergeSend := []int{
		wg.AddItem(&workgraph.Item{Worker: "w1"}),
		wg.AddItem(&workgraph.Item{Worker: "w2"}),
		wg.AddItem(&workgraph.Item{Worker: "w3"}),
	}

	// Wait for merge recv tasks to open up ports
	wg.AddRelation(mergeSend[0], workgraph.InProgress, mergeRecv[0])
	wg.AddRelation(mergeSend[0], workgraph.InProgress, mergeRecv[1])
	wg.AddRelation(mergeSend[0], workgraph.InProgress, mergeRecv[2])

	wg.AddRelation(mergeSend[1], workgraph.InProgress, mergeRecv[0])
	wg.AddRelation(mergeSend[1], workgraph.InProgress, mergeRecv[1])
	wg.AddRelation(mergeSend[1], workgraph.InProgress, mergeRecv[2])

	wg.AddRelation(mergeSend[2], workgraph.InProgress, mergeRecv[0])
	wg.AddRelation(mergeSend[2], workgraph.InProgress, mergeRecv[1])
	wg.AddRelation(mergeSend[2], workgraph.InProgress, mergeRecv[2])

	// Add one upload per worker
	upload := []int{
		wg.AddItem(&workgraph.Item{Worker: "w1"}),
		wg.AddItem(&workgraph.Item{Worker: "w2"}),
		wg.AddItem(&workgraph.Item{Worker: "w3"}),
	}

	// Wait for merge_recv to finish
	wg.AddRelation(upload[0], workgraph.Done, mergeRecv[0])
	wg.AddRelation(upload[1], workgraph.Done, mergeRecv[1])
	wg.AddRelation(upload[2], workgraph.Done, mergeRecv[2])

	// Wait for uploads to finish
	wg.AddRelation(workgraph.Sink, workgraph.Done, upload[0])
	wg.AddRelation(workgraph.Sink, workgraph.Done, upload[1])
	wg.AddRelation(workgraph.Sink, workgraph.Done, upload[2])

	return wg, sort, mergeRecv, mergeSend, upload
}

func TestRelationScheduled(t *testing.T) {
	wg := workgraph.New()

	item1 := wg.AddItem(&workgraph.Item{})
	item2 := wg.AddItem(&workgraph.Item{})

	// source -(done)> item1 -(scheduled)> item2 -(done)> sink
	wg.AddRelation(item1, workgraph.Done, workgraph.Source)
	wg.AddRelation(item2, workgraph.Scheduled, item1)
	wg.AddRelation(workgraph.Sink, workgraph.Done, item2)

	item, err := wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, item1, item.ID)
	}

	item, err = wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, item2, item.ID)
	}

	_, err = wg.Schedule("")
	assert.Error(t, err)

	assert.NoError(t, wg.FinishItem(item2))
	item, err = wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, workgraph.Sink, item.ID)
	}
}

func TestRelationScheduledButFailed(t *testing.T) {
	wg := workgraph.New()

	item1 := wg.AddItem(&workgraph.Item{})

	// source -(done)> item1 -(scheduled)> item2 -(done)> sink
	wg.AddRelation(item1, workgraph.Done, workgraph.Source)
	wg.AddRelation(workgraph.Sink, workgraph.Scheduled, item1)

	item, err := wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, item1, item.ID)
	}

	assert.NoError(t, wg.FailItem(item1))
	_, err = wg.Schedule("")
	assert.Error(t, err)
}

func TestRelationScheduledButFinished(t *testing.T) {
	wg := workgraph.New()

	item1 := wg.AddItem(&workgraph.Item{})

	// source -(done)> item1 -(scheduled)> item2 -(done)> sink
	wg.AddRelation(item1, workgraph.Done, workgraph.Source)
	wg.AddRelation(workgraph.Sink, workgraph.Done, item1)

	item, err := wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, item1, item.ID)
	}

	assert.NoError(t, wg.FinishItem(item1))
	item, err = wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, workgraph.Sink, item.ID)
	}
}

func TestRelationDoneButFailed(t *testing.T) {
	wg := workgraph.New()

	item1 := wg.AddItem(&workgraph.Item{})

	wg.AddRelation(item1, workgraph.Done, workgraph.Source)
	wg.AddRelation(workgraph.Sink, workgraph.Done, item1)

	item, err := wg.Schedule("")
	if assert.NoError(t, err) {
		assert.Equal(t, item1, item.ID)
	}

	assert.NoError(t, wg.FailItem(item1))
	_, err = wg.Schedule("")
	assert.Error(t, err)
}

func TestCloudsortSchedule(t *testing.T) {
	wg, sort, mergeRecv, _, _ := buildCloudsortGraph()

	sort1, err := wg.Schedule("w1")
	assert.NoError(t, err)
	assert.Same(t, wg.Items[sort[0]], sort1)

	_, err = wg.Schedule("w1")
	assert.Error(t, err)

	sort2, err := wg.Schedule("w2")
	assert.NoError(t, err)
	assert.Same(t, wg.Items[sort[1]], sort2)

	assert.NoError(t, wg.FinishItem(sort2.ID))

	sort3, err := wg.Schedule("w3")
	assert.NoError(t, err)
	assert.Same(t, wg.Items[sort[2]], sort3)

	assert.NoError(t, wg.FinishItem(sort3.ID))

	_, err = wg.Schedule("w1")
	assert.Error(t, err)

	assert.NoError(t, wg.FinishItem(sort1.ID))

	mergeRecv1, err := wg.Schedule("w1")
	assert.NoError(t, err)
	assert.Same(t, wg.Items[mergeRecv[0]], mergeRecv1)
}

func TestCloudsortRoundRobin(t *testing.T) {
	wg, _, _, _, _ := buildCloudsortGraph()

	schedule := []int{}
	ws := []string{"w1", "w2", "w3"}

	for i := 0; ; i++ {
		item, err := wg.Schedule(ws[i%len(ws)])
		assert.NoError(t, err)
		schedule = append(schedule, item.ID)
		assert.NoError(t, wg.FinishItem(item.ID))

		if item.ID == workgraph.Sink {
			break
		}
	}

	assert.Equal(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1}, schedule)
}
