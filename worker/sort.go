package worker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/sort"
	"github.com/lnsp/cloudsort/storage"
	"github.com/lnsp/cloudsort/structs"
	"github.com/minio/minio-go/v7"
)

type SortTask struct {
	*BasicTask
}

func (t *SortTask) Abort() {
	// TODO: unimplemented
}

func (t *SortTask) fetchAndSort() ([]string, error) {
	// Determine byte range
	spec := t.Spec.Details.(*pb.Task_Sort).Sort
	bytesStart := spec.RangeStart * structs.EntrySize
	bytesEnd := spec.RangeEnd*structs.EntrySize + structs.EntrySize - 1
	opts := minio.GetObjectOptions{}
	opts.Set("Range", fmt.Sprintf("bytes=%d-%d", bytesStart, bytesEnd))

	t.Log.With("start", bytesStart, "end", bytesEnd).Debug("Fetching data from S3")
	s3, err := storage.GetS3Client(spec.Credentials)
	if err != nil {
		return nil, fmt.Errorf("get S3 client: %w", err)
	}
	// Get S3 object
	object, err := s3.GetObject(context.Background(), spec.Credentials.BucketId, spec.Credentials.ObjectKey, opts)
	if err != nil {
		return nil, fmt.Errorf("download from S3: %w", err)
	}
	defer object.Close()
	objectReader := bufio.NewReader(object)
	// Write runs to disk
	t.Context.Lock()
	sortRunSize := t.Context.sortRunSize
	t.Context.Unlock()

	runWriter, err := sort.NewRunWriter(sortRunSize, t.Name())
	if err != nil {
		return nil, fmt.Errorf("create run writer: %w", err)
	}
	// Consume runs and sort them
	done := make(chan []string, 1)
	defer close(done)
	go func() {
		runs := make([]string, 0)
		for run := range runWriter.Runs {
			// Sort run
			if err := sort.SortSingleFile(run, run); err != nil {
				t.Log.With("error", err, "run", run).Error("sort single file")
			}
			runs = append(runs, run)
		}
		done <- runs
	}()
	if _, err := objectReader.WriteTo(runWriter); err != nil {
		return nil, fmt.Errorf("write object runs: %w", err)
	}
	if err := runWriter.Close(); err != nil {
		return nil, fmt.Errorf("flush and close run writer: %w", err)
	}
	return <-done, nil
}

func (t *SortTask) mergeLocal(runs []string) error {
	runtime.GC()
	// Merge runs back into single file
	var mergeHeap sort.MergeHeap
	for _, fpath := range runs {
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer os.Remove(fpath)
		defer f.Close()
		mergeHeap = append(mergeHeap, &sort.MergeHeapItem{
			Reader: bufio.NewReader(f),
		})
	}
	sortRunFile := t.Name() + "-sorted"
	// Open destination file
	mergeDestination, err := os.Create(sortRunFile)
	if err != nil {
		return fmt.Errorf("create sort file: %w", err)
	}
	defer mergeDestination.Close()
	mergeDestinationWriter := bufio.NewWriter(mergeDestination)
	defer mergeDestinationWriter.Flush()
	// Merge into destination and generate markers
	markers, err := sort.MergeRuns(mergeHeap, &sort.BufWriterCollecter{Writer: mergeDestinationWriter})
	if err != nil {
		return fmt.Errorf("merge sort runs: %w", err)
	}

	t.Context.Lock()
	t.Context.sortRunFile = sortRunFile
	t.Context.sortRunMarkers = markers
	t.Context.Unlock()
	return nil
}

func (t *SortTask) Run() error {
	runs, err := t.fetchAndSort()
	if err != nil {
		return err
	}
	return t.mergeLocal(runs)
}
