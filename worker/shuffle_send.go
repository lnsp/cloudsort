package worker

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	gsort "sort"

	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/sort"
	"github.com/lnsp/cloudsort/structs"
)

type ShuffleSendTask struct {
	*BasicTask
}

var _ Task = (*ShuffleSendTask)(nil)

func (t *ShuffleSendTask) Abort() {

}

func (t *ShuffleSendTask) shuffleToPeer(peer *pb.Peer, index int) error {
	// Open chunk file
	t.Context.Lock()
	sortRunFile := t.Context.sortRunFile
	t.Context.Unlock()

	chunk, err := os.Open(sortRunFile)
	if err != nil {
		return err
	}
	defer os.Remove(chunk.Name())
	defer chunk.Close()
	// Find start markers and seek ahead
	if _, err := chunk.Seek(int64(index)*structs.EntrySize, io.SeekStart); err != nil {
		return err
	}
	// Connect to peer
	var conn net.Conn
	for retry := 0; conn == nil; retry++ {
		c, err := net.Dial("tcp", peer.Address)
		if err != nil {
			t.Log.Errorf("Dial peer %s: %s", peer.Address, err)
			time.Sleep(time.Second)
		} else {
			conn = c
		}
	}
	defer conn.Close()
	connWriter := bufio.NewWriter(conn)
	defer connWriter.Flush()
	// Skip ahead until finding valid key
	entry := make([]byte, structs.EntrySize)
	skipSmall := true
	chunkReader := bufio.NewReader(chunk)
	for {
		if _, err := io.ReadFull(chunkReader, entry); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		// Check if key is in peer range
		if skipSmall && sort.BytesCompare10(entry[:structs.KeySize], peer.KeyRangeStart) {
			continue
		} else if sort.BytesCompare10(peer.KeyRangeEnd, entry[:structs.KeySize]) {
			return nil
		}
		skipSmall = false
		// Send entry over to peer
		if _, err := connWriter.Write(entry); err != nil {
			return err
		}
	}
}

func (t *ShuffleSendTask) shuffleToPeers(markers []sort.Marker) error {
	spec := t.Spec.Details.(*pb.Task_ShuffleSend).ShuffleSend
	// Sort peers by key-dest
	numPeers := len(spec.Peers)
	gsort.Slice(spec.Peers, func(i, j int) bool {
		return sort.BytesCompare10(spec.Peers[i].KeyRangeStart, spec.Peers[j].KeyRangeStart)
	})
	// Find matching markers for each peer
	startIndices := make([]int, numPeers)
	for i, peer := range spec.Peers {
		n := gsort.Search(len(markers)-1, func(j int) bool {
			return !sort.BytesCompare10Equ(markers[j+1].Value, peer.KeyRangeStart)
		})
		startIndices[i] = markers[n].Index
		t.Log.Debugf("Determined start index %d for shuffle peer %d", markers[n].Index, i)
	}
	// Shuffle in parallel
	errs := make(chan error, numPeers)
	defer close(errs)
	for i, peer := range spec.Peers {
		i, peer := i, peer
		go func() {
			errs <- t.shuffleToPeer(peer, startIndices[i])
		}()
	}
	var err error
	// Collect errors
	for i := 0; i < numPeers; i++ {
		if e := <-errs; e != nil {
			err = e
		}
	}
	if err != nil {
		return fmt.Errorf("shuffle to peer: %w", err)
	}
	return nil
}

func (t *ShuffleSendTask) Run() error {
	t.Context.Lock()
	sortRunMarkers := t.Context.sortRunMarkers
	t.Context.Unlock()

	if err := t.shuffleToPeers(sortRunMarkers); err != nil {
		return err
	}
	return nil
}
