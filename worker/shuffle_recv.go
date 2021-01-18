package worker

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/lnsp/cloudsort/pb"
	"go.uber.org/zap"
)

type ShuffleRecvTask struct {
	*BasicTask
}

var _ Task = (*ShuffleRecvTask)(nil)

func (t *ShuffleRecvTask) Abort() {}

func (t *ShuffleRecvTask) openListener() (net.Listener, error) {
	t.Context.Lock()
	shuffleRecvHost := t.Context.shuffleRecvHost
	t.Context.Unlock()

	listener, err := net.Listen("tcp", shuffleRecvHost+":0")
	if err != nil {
		return nil, err
	}
	// Report listener address
	t.Report(pb.TaskState_IN_PROGRESS, &pb.TaskData{
		Properties: &pb.TaskData_ShuffleRecvAddr{
			ShuffleRecvAddr: listener.Addr().String(),
		},
	})
	return listener, nil
}

func (t *ShuffleRecvTask) collectToDisk(listener net.Listener) error {
	spec := t.Spec.Details.(*pb.Task_ShuffleRecv).ShuffleRecv
	// Allocate series of temporary files
	errs := make(chan error, spec.NumberOfPeers)
	defer close(errs)
	for i := int64(0); i < spec.NumberOfPeers; i++ {
		i := i
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		connReader := bufio.NewReader(conn)
		t.Log.Debugf("Accepted merge conn %d from %s", i, conn.RemoteAddr())
		// Generate path for merge conn
		tmpPath := fmt.Sprintf("%s-%d", t.Name(), i)

		t.Context.Lock()
		t.Context.shuffleRecvFiles = append(t.Context.shuffleRecvFiles, tmpPath)
		t.Context.Unlock()

		// Copy everything to local disk
		tmpFile, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		tmpFileWriter := bufio.NewWriter(tmpFile)
		go func() {
			defer conn.Close()
			defer tmpFile.Close()
			defer tmpFileWriter.Flush()
			n, err := io.Copy(tmpFileWriter, connReader)
			errs <- err
			t.Log.Debugf("Finished merge conn %d from %s (%d bytes)", i, conn.RemoteAddr(), n)
		}()
	}
	// Await finish of all shuffles
	var err error
	for i := 0; i < int(spec.NumberOfPeers); i++ {
		if e := <-errs; e != nil {
			err = e
			zap.S().Errorf("Error during shuffle, got %s", e)
		} else {
			zap.S().Debugf("Got shuffle %d/%d (%.2f%%)", i, spec.NumberOfPeers, float32(i)/float32(spec.NumberOfPeers)*100.0)
		}
	}
	if err != nil {
		return err
	}
	zap.S().Debugf("Successfully collected %d runs", spec.NumberOfPeers)
	return nil
}

func (t *ShuffleRecvTask) Run() error {
	// Step 1: Open listener
	listener, err := t.openListener()
	if err != nil {
		return err
	}
	// Step 2: Await N connections
	if err := t.collectToDisk(listener); err != nil {
		return err
	}
	return nil
}
