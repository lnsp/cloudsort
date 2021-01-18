package worker

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/sort"
	"github.com/lnsp/cloudsort/storage"
	"github.com/minio/minio-go/v7"
)

type UploadTask struct {
	*BasicTask
}

var _ Task = (*UploadTask)(nil)

func (t *UploadTask) Abort() {
	// TODO: Unimplemented
}

func (t *UploadTask) Run() error {
	// Open merge items
	var mergeHeap sort.MergeHeap
	for _, fpath := range t.Context.shuffleRecvFiles {
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
	// Create upload pipeline with buffered writer
	pipeRd, pipeWr, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("open merge pipe: %w", err)
	}
	pipeWriteBuffer := bufio.NewWriter(pipeWr)
	pipeReadBuffer := bufio.NewReader(pipeRd)
	// Start merge
	go func() {
		sort.MergeRuns(mergeHeap, &sort.BufWriterCollecter{Writer: pipeWriteBuffer})
		pipeWriteBuffer.Flush()
		pipeWr.Close()
	}()
	// And upload to S3
	credentials := t.Spec.Details.(*pb.Task_Upload).Upload.Credentials
	s3, err := storage.GetS3Client(credentials)
	if err != nil {
		return fmt.Errorf("get S3 client: %w", err)
	}
	if _, err := s3.PutObject(context.Background(), credentials.BucketId, t.Name(), pipeReadBuffer, -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("upload to S3: %w", err)
	}
	// Kill of context
	t.Context.Kill()
	return nil
}
