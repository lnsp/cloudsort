package sort

import (
	"bufio"
	"bytes"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"runtime"
	"sort"

	"github.com/jfcg/sorty"
	"github.com/lnsp/cloudsort/structs"
	"go.uber.org/zap"
)

type ChunkReader struct {
	buf    *bytes.Buffer
	stream chan []byte
}

func (r *ChunkReader) Read(b []byte) (int, error) {
	if r.buf.Len() == 0 {
		chunk := <-r.stream
		r.buf = bytes.NewBuffer(chunk)
	}
	return r.buf.Read(b)
}

func NewChunkReader(stream chan []byte) *ChunkReader {
	return &ChunkReader{
		buf:    new(bytes.Buffer),
		stream: stream,
	}
}

func BytesCompare10U16(a, b []byte) bool {
	x := binary.BigEndian.Uint64(a)
	y := binary.BigEndian.Uint64(b)
	return x < y
}

func BytesCompare10(a, b []byte) bool {
	x := binary.BigEndian.Uint64(a)
	y := binary.BigEndian.Uint64(b)
	if x < y {
		return true
	} else if x == y {
		xp := binary.BigEndian.Uint16(a[8:])
		yp := binary.BigEndian.Uint16(b[8:])
		if xp < yp {
			return true
		}
	}
	return false
}

func BytesCompare10Equ(a, b []byte) bool {
	x := binary.BigEndian.Uint64(a)
	y := binary.BigEndian.Uint64(b)
	if x < y {
		return true
	} else if x == y {
		xp := binary.BigEndian.Uint16(a[8:])
		yp := binary.BigEndian.Uint16(b[8:])
		if xp <= yp {
			return true
		}
	}
	return false
}

func ByKey(values []byte) func(a, b interface{}) bool {
	return func(a, b interface{}) bool {
		if BytesCompare10(values[a.(int)*structs.EntrySize:structs.KeySize+a.(int)*structs.EntrySize], values[b.(int)*structs.EntrySize:structs.KeySize+b.(int)*structs.EntrySize]) {
			return true
		}
		return false
	}
}

type RunWriter struct {
	Memory int
	Base   string
	Runs   chan string

	index   int
	written int
	file    *os.File
	writer  *bufio.Writer
}

func NewRunWriter(memory int, base string) (*RunWriter, error) {
	w := &RunWriter{
		Memory: memory,
		Base:   base,
		Runs:   make(chan string),
	}
	if err := w.start(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RunWriter) start() error {
	f, err := os.Create(fmt.Sprintf("%s-%d", w.Base, w.index))
	if err != nil {
		return err
	}
	w.writer = bufio.NewWriter(f)
	w.file = f
	w.written = 0
	return nil
}

func (w *RunWriter) next() error {
	err := w.writer.Flush()
	if err != nil {
		return err
	}
	err = w.file.Close()
	if err != nil {
		return err
	}
	go func(n string) { w.Runs <- n }(w.file.Name())
	w.index++
	return w.start()
}

func (w *RunWriter) Write(b []byte) (int, error) {
	total := 0
	for w.written+len(b) >= w.Memory {
		d := w.Memory - w.written
		if _, err := w.writer.Write(b[:d]); err != nil {
			return 0, err
		}
		if err := w.next(); err != nil {
			return 0, err
		}
		b = b[d:]
		total += d
	}
	_, err := w.writer.Write(b)
	if err != nil {
		return 0, err
	}
	w.written += len(b)
	return total + len(b), nil
}

func (w *RunWriter) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	if err := w.file.Close(); err != nil {
		return err
	}
	if w.written > 0 {
		w.Runs <- w.file.Name()
	} else {
		os.Remove(w.file.Name())
	}
	close(w.Runs)
	return nil
}

type MergeHeapItem struct {
	Reader io.Reader
	Buffer []byte
	Offset int
}

func (m *MergeHeapItem) Value() []byte {
	return m.Buffer[m.Offset : m.Offset+structs.EntrySize]
}

type MergeHeap []*MergeHeapItem

func (h MergeHeap) Len() int { return len(h) }
func (h MergeHeap) Less(i, j int) bool {
	return BytesCompare10(h[i].Value()[:structs.KeySize], h[j].Value()[:structs.KeySize])
}
func (h MergeHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *MergeHeap) Push(x interface{}) { *h = append(*h, x.(*MergeHeapItem)) }
func (h *MergeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

type Collecter interface {
	Collect(value []byte) error
	FlushAndClose() error
}

type PipeWriterCollecter struct {
	PipeWr *io.PipeWriter
}

func (c *PipeWriterCollecter) Collect(value []byte) error {
	_, err := c.PipeWr.Write(value)
	return err
}

func (c *PipeWriterCollecter) FlushAndClose() error {
	return c.PipeWr.Close()
}

type BufWriterCollecter struct {
	Writer *bufio.Writer
}

func (c *BufWriterCollecter) Collect(value []byte) error {
	_, err := c.Writer.Write(value)
	return err
}

func (c *BufWriterCollecter) FlushAndClose() error {
	return c.Writer.Flush()
}

type Marker struct {
	Value []byte
	Index int
}

const markerInterval = 100000

func MergeRuns(items MergeHeap, collecter Collecter) ([]Marker, error) {
	heapBuffer := make([]byte, len(items)*structs.EntrySize)
	for i := range items {
		items[i].Offset = i * structs.EntrySize
		items[i].Buffer = heapBuffer
	}
	// Try to fill heap values
	{
		i, j := 0, 0
		for i < len(items) {
			if _, err := io.ReadFull(items[i].Reader, items[i].Value()); err == nil {
				items[j] = items[i]
				j++
			} else {
				zap.S().Debugf("Dropped tmp file on heap init", i)
			}
			i++
		}
		items = items[:j]
		heap.Init(&items)
	}
	var (
		index   int
		markers []Marker
	)
	// Now we can write back into our original input file
	for items.Len() > 0 {
		// Pop next min value
		next := heap.Pop(&items).(*MergeHeapItem)
		if err := collecter.Collect(next.Value()); err != nil {
			return nil, fmt.Errorf("collect value: %w", err)
		}
		// Check if we've reached the marker
		if index%markerInterval == 0 {
			value := make([]byte, structs.EntrySize)
			copy(value, next.Value())
			markers = append(markers, Marker{
				Value: value,
				Index: index,
			})
		}
		// Check if reader has next value
		if _, err := io.ReadFull(next.Reader, next.Value()); err == nil {
			heap.Push(&items, next)
		}
		index++
	}
	if err := collecter.FlushAndClose(); err != nil {
		return nil, fmt.Errorf("flush collecter: %w", err)
	}
	// Fill rest of marked indices with max index
	return markers, nil
}

func SortSingleFileUniform(path, output string, rangeStart []byte, rangeEnd []byte) error {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var (
		numValues     = len(buffer) / structs.EntrySize
		numBuckets    = (numValues + 10000 - 1) / 10000
		numBucketsInt = new(big.Int).SetInt64(int64(numBuckets))
		rangeStartInt = new(big.Int).SetBytes(rangeStart)
		rangeSizeInt  = new(big.Int).Sub(new(big.Int).SetBytes(rangeEnd), rangeStartInt)
		rangePos      = new(big.Int)
	)
	buckets := make([][]int, numBuckets)
	for i := 0; i < numValues; i++ {
		rangePos.SetBytes(buffer[i*structs.EntrySize:i*structs.EntrySize+structs.KeySize]).
			Sub(rangePos, rangeStartInt).
			Mul(rangePos, numBucketsInt).
			Div(rangePos, rangeSizeInt)
		j := int(rangePos.Int64())
		buckets[j] = append(buckets[j], i)
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for b := 0; b < numBuckets; b++ {
		sort.Slice(buckets[b], func(i, k int) bool {
			return BytesCompare10(buffer[buckets[b][i]*structs.EntrySize:buckets[b][i]*structs.EntrySize+structs.KeySize],
				buffer[buckets[b][k]*structs.EntrySize:buckets[b][k]*structs.EntrySize+structs.KeySize])
		})
		// write out bucket
		for j := 0; j < len(buckets[b]); j++ {
			if _, err := writer.Write(buffer[buckets[b][j]*structs.EntrySize : (buckets[b][j]+1)*structs.EntrySize]); err != nil {
				return err
			}
		}
	}
	return nil
}

func SortSingleFileSwapped(path, output string) error {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	indices := make([]int, len(buffer)/structs.EntrySize)
	for i := range indices {
		indices[i] = i
	}
	lsw := func(i, k, r, s int) bool {
		if BytesCompare10(buffer[i*structs.EntrySize:i*structs.EntrySize+structs.KeySize],
			buffer[k*structs.EntrySize:k*structs.EntrySize+structs.KeySize]) {
			if r != s {
				for j := 0; j < structs.EntrySize; j++ {
					buffer[i*structs.EntrySize+j], buffer[k*structs.EntrySize+j] = buffer[k*structs.EntrySize+j], buffer[i*structs.EntrySize+j]
				}
			}
			return true
		}
		return false
	}
	sorty.Sort(len(buffer)/structs.EntrySize, lsw)
	// Write back
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	if _, err := io.Copy(writer, bytes.NewBuffer(buffer)); err != nil {
		return err
	}
	return nil
}

func SortChunkAndWrite(buffer []byte, path string, markerInterval int, rwBufSize int) ([]Marker, error) {
	// Sort chunk in-memory
	indices := make([]int, len(buffer)/structs.EntrySize)
	for i := range indices {
		indices[i] = i
	}
	lsw := func(i, k, r, s int) bool {
		left := buffer[indices[i]*structs.EntrySize : indices[i]*structs.EntrySize+structs.KeySize]
		right := buffer[indices[k]*structs.EntrySize : indices[k]*structs.EntrySize+structs.KeySize]
		if BytesCompare10(left, right) {
			if r != s {
				indices[r], indices[s] = indices[s], indices[r]
			}
			return true
		}
		return false
	}
	sorty.Sort(len(indices), lsw)
	// Write out and generate markers
	var markers []Marker
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	writer := bufio.NewWriterSize(file, rwBufSize)
	defer writer.Flush()
	for i := 0; i < len(indices); i++ {
		j := indices[i]
		// Append to markers
		if i%markerInterval == 0 {
			key := make([]byte, structs.KeySize)
			copy(key, buffer[j*structs.EntrySize:j*structs.EntrySize+structs.KeySize])
			markers = append(markers, Marker{
				Value: key,
				Index: i,
			})
		}
		// Append to file
		if _, err := writer.Write(buffer[j*structs.EntrySize : structs.EntrySize*(j+1)]); err != nil {
			return nil, err
		}
	}
	return markers, nil
}

func SortInMemory(buffer []byte) {
	indices := make([]int, len(buffer)/structs.EntrySize)
	for i := range indices {
		indices[i] = i
	}
	lsw := func(i, k, r, s int) bool {
		left := buffer[indices[i]*structs.EntrySize : indices[i]*structs.EntrySize+structs.KeySize]
		right := buffer[indices[k]*structs.EntrySize : indices[k]*structs.EntrySize+structs.KeySize]
		if BytesCompare10(left, right) {
			if r != s {
				indices[r], indices[s] = indices[s], indices[r]
			}
			return true
		}
		return false
	}
	sorty.Sort(len(indices), lsw)
}

/*
func WriteSplits(buffer []byte, keyRanges []structs.KeyRange, prefix string) ([]string, error) {
	var (
		splits  []string
		current *os.File
	)
	for {

	}
}
*/

func SortSingleBuffer(buffer []byte, outputPath string) error {
	indices := make([]int, len(buffer)/structs.EntrySize)
	for i := range indices {
		indices[i] = i
	}
	lsw := func(i, k, r, s int) bool {
		left := buffer[indices[i]*structs.EntrySize : indices[i]*structs.EntrySize+structs.KeySize]
		right := buffer[indices[k]*structs.EntrySize : indices[k]*structs.EntrySize+structs.KeySize]
		if BytesCompare10(left, right) {
			if r != s {
				indices[r], indices[s] = indices[s], indices[r]
			}
			return true
		}
		return false
	}
	sorty.Sort(len(indices), lsw)
	// Write back
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for i := 0; i < len(indices); i++ {
		if _, err := writer.Write(buffer[indices[i]*structs.EntrySize : (indices[i]+1)*structs.EntrySize]); err != nil {
			return err
		}
	}
	return nil
}

func SortSingleFile(inputPath, outputPath string) error {
	buffer, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return err
	}
	return SortSingleBuffer(buffer, outputPath)
}

func CreateSortRuns(file string, maxMemory int) ([]string, error) {
	input, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer input.Close()
	// Load file in maxMemory chunks
	zap.S().Debugf("Allocate external sort buffer of size %d", maxMemory)
	buffer := make([]byte, maxMemory)
	runs := make([]string, 0, 1)
	for run := 0; ; run++ {
		// Performe GC to free up mem
		runtime.GC()
		// Read into buffer
		n := 0
		for n != maxMemory {
			m, err := input.Read(buffer[n:])
			n += m
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("read input: %w", err)
			}
			zap.S().Debugf("Read chunk of %d into buffer at offset %d", m, n)
		}
		if n == 0 {
			break
		}
		// Sort indices via TimSort
		indices := make([]int, n/structs.EntrySize)
		for i := range indices {
			indices[i] = i
		}
		lsw := func(i, k, r, s int) bool {
			left := buffer[indices[i]*structs.EntrySize : indices[i]*structs.EntrySize+structs.KeySize]
			right := buffer[indices[k]*structs.EntrySize : indices[k]*structs.EntrySize+structs.KeySize]
			if BytesCompare10(left, right) {
				if r != s {
					indices[r], indices[s] = indices[s], indices[r]
				}
				return true
			}
			return false
		}
		sorty.Sort(len(indices), lsw)
		// Write buffer in sorted order back to tmpFile
		runFile, err := os.Create(fmt.Sprintf("%s-%d", file, run))
		if err != nil {
			return nil, fmt.Errorf("create run: %w", err)
		}
		runs = append(runs, runFile.Name())
		zap.S().Debugf("Allocated tmp file at %s", runFile.Name())
		runWriter := bufio.NewWriter(runFile)
		for i := 0; i < len(indices); i++ {
			if _, err := runWriter.Write(buffer[indices[i]*structs.EntrySize : (indices[i]+1)*structs.EntrySize]); err != nil {
				runFile.Close()
				return nil, fmt.Errorf("write tmp file: %w", err)
			}
		}
		runWriter.Flush()
		runFile.Close()
	}
	return runs, nil
}
