package sort

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/jfcg/sorty"
	"github.com/lnsp/cloudsort/structs"
	"github.com/psilva261/timsort"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBucketSort(t *testing.T) {
	assert.NoError(t, exec.Command("gensort", "100000", "test-file").Run())
	defer os.Remove("test-file")
	assert.NoError(t, SortSingleFileUniform("test-file", "test-file", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255}))
	assert.NoError(t, exec.Command("valsort", "-q", "test-file").Run())
}

func TestSingleSort(t *testing.T) {
	assert.NoError(t, exec.Command("gensort", "100000", "test-file").Run())
	defer os.Remove("test-file")
	assert.NoError(t, SortSingleFile("test-file", "test-file"))
	assert.NoError(t, exec.Command("valsort", "-q", "test-file").Run())
}

type rndreader struct {
}

func (rndreader) Read(b []byte) (int, error) {
	return rand.Read(b)
}

func BenchmarkSortThroughput(b *testing.B) {
	assert.NoError(b, exec.Command("gensort", "10000000", "test-file").Run())
	defer os.Remove("test-file")
	defer os.Remove("test-file-out")

	b.Run("sort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assert.NoError(b, SortSingleFile("test-file", "test-file-out"))
		}
	})
	b.Run("sort-swapped", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assert.NoError(b, SortSingleFileSwapped("test-file", "test-file-out"))
		}
	})
	b.Run("sort-uniform", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assert.NoError(b, SortSingleFileUniform("test-file", "test-file-out", structs.KeyRangeStart, structs.KeyRangeEnd))
		}
	})
}

func TestRunWriter(t *testing.T) {
	var runs []string
	wr, err := NewRunWriter(4000, "t")
	assert.NoError(t, err)
	done := make(chan struct{})
	go func() {
		for r := range wr.Runs {
			runs = append(runs, r)
		}
		close(done)
	}()
	_, err = io.CopyN(wr, rndreader{}, 22500)
	assert.NoError(t, err)
	assert.NoError(t, wr.Close())
	<-done
	assert.Equal(t, []string{"t-0", "t-1", "t-2", "t-3", "t-4", "t-5"}, runs)
}

func TestCgoBytesCompare(t *testing.T) {
	x := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	//y := []byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	assert.True(t, !Cmp10(x, x))
}

var bbindices []int

func BenchmarkBytesSwap(b *testing.B) {
	const N int = 10000
	data := make([]byte, N*structs.EntrySize)
	rand.Read(data)
	b.Run("noswap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bbindices = make([]int, N)
			for i := range bbindices {
				bbindices[i] = i
			}
			sort.Slice(bbindices, func(i, j int) bool {
				return BytesCompare10(
					data[bbindices[i]*structs.EntrySize:bbindices[i]*structs.EntrySize+structs.KeySize],
					data[bbindices[j]*structs.EntrySize:bbindices[j]*structs.EntrySize+structs.KeySize])
			})
			runtime.KeepAlive(bbindices)
		}
	})
}

func TestBytesCompare(t *testing.T) {
	const N int = 100
	rand.Seed(time.Now().Unix())
	x, y := make([]byte, 10), make([]byte, 10)
	for i := 0; i < N; i++ {
		rand.Read(x)
		rand.Read(y)
		assert.Equal(t, bytes.Compare(x, y) < 0, BytesCompare10(x, y), "compared %s < %s", hex.EncodeToString(x), hex.EncodeToString(y))
		assert.Equal(t, bytes.Compare(x, x) < 0, BytesCompare10(x, x), "compared %s < %s", hex.EncodeToString(x), hex.EncodeToString(y))
		assert.Equal(t, bytes.Compare(y, x) < 0, BytesCompare10(y, x), "compared %s > %s", hex.EncodeToString(x), hex.EncodeToString(y))
		assert.Equal(t, bytes.Compare(x, y) < 0, Cmp10(x, y))
	}
}

func BenchmarkSorty(b *testing.B) {
	const N = 10_000_000
	rand.Seed(time.Now().Unix())
	buffer := make([]byte, N*structs.EntrySize)
	rand.Read(buffer)
	b.Run("sorty", func(b *testing.B) {
		b.ReportAllocs()
		for t := 0; t < b.N; t++ {
			indices := make([]int, N)
			for i := range indices {
				indices[i] = i
			}
			lsw := func(i, k, r, s int) bool {
				if BytesCompare10(buffer[indices[i]*structs.EntrySize:indices[i]*structs.EntrySize+structs.KeySize],
					buffer[indices[k]*structs.EntrySize:indices[k]*structs.EntrySize+structs.KeySize]) {
					if r != s {
						indices[r], indices[s] = indices[s], indices[r]
					}
					return true
				}
				return false
			}
			sorty.Sort(len(indices), lsw)
		}
	})
	b.Run("sorty-u16", func(b *testing.B) {
		b.ReportAllocs()
		for t := 0; t < b.N; t++ {
			indices := make([]interface{}, N)
			for i := range indices {
				indices[i] = i
			}
			timsort.Sort(indices, func(i, j interface{}) bool {
				return BytesCompare10(buffer[i.(int)*structs.EntrySize:i.(int)*structs.EntrySize+structs.KeySize],
					buffer[j.(int)*structs.EntrySize:i.(int)*structs.EntrySize+structs.KeySize])
			})
		}
	})
	// reset indices
	b.Run("sorty-basic", func(b *testing.B) {
		b.ReportAllocs()
		for t := 0; t < b.N; t++ {
			indices := make([]int, N)
			for i := range indices {
				indices[i] = i
			}
			lsw := func(i, k, r, s int) bool {
				left := buffer[indices[i]*structs.EntrySize : indices[i]*structs.EntrySize+structs.KeySize]
				right := buffer[indices[k]*structs.EntrySize : indices[k]*structs.EntrySize+structs.KeySize]
				if bytes.Compare(left, right) < 0 {
					if r != s {
						indices[r], indices[s] = indices[s], indices[r]
					}
					return true
				}
				return false
			}
			sorty.Sort(len(indices), lsw)
		}
	})
	b.Run("sorty-aligned", func(b *testing.B) {
		// Generate index
		for t := 0; t < b.N; t++ {
			indices := make([]uint64, N*2)
			for i := 0; i < N; i++ {
				indices[2*i] = binary.BigEndian.Uint64(buffer[i*structs.EntrySize:])
				indices[2*i+1] = uint64(binary.BigEndian.Uint16(buffer[i*structs.EntrySize+8:])) | (uint64(i) << 32)
			}
			lsw := func(i, k, r, s int) bool {
				if indices[2*i] < indices[2*k] && (indices[2*i+1]&0xffff) < (indices[2*k+1]&0xffff) {
					if r != s {
						indices[2*i], indices[2*k] = indices[2*k], indices[2*i]
						indices[2*i+1], indices[2*k+1] = indices[2*k+1], indices[2*i+1]
					}
					return true
				}
				return false
			}
			sorty.Sort(N, lsw)
		}
	})
	b.Run("sorty-strings", func(b *testing.B) {
		for t := 0; t < b.N; t++ {
			indices := make([]int, N)
			for i := range indices {
				indices[i] = i
			}
			lsw := func(i, k, r, s int) bool {
				var jhdr, khdr reflect.StringHeader
				base := *(*reflect.SliceHeader)(unsafe.Pointer(&buffer))
				jhdr.Len = structs.KeySize
				khdr.Len = structs.KeySize
				jhdr.Data = base.Data + uintptr(indices[i]*structs.EntrySize)
				khdr.Data = base.Data + uintptr(indices[k]*structs.EntrySize)
				s1 := *(*string)(unsafe.Pointer(&jhdr))
				s2 := *(*string)(unsafe.Pointer(&khdr))
				if s1 < s2 {
					if indices[r] != indices[s] {
						indices[r], indices[s] = indices[s], indices[r]
					}
					return true
				}
				return false
			}
			sorty.Sort(len(indices), lsw)
		}
	})
	b.Run("sorty-swapped", func(b *testing.B) {
		b.ReportAllocs()
		for t := 0; t < b.N; t++ {
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
			sorty.Sort(N, lsw)
		}
	})
	// check that indices are sorted
}

type randreader struct {
}

func (randreader) Read(p []byte) (int, error) {
	return rand.Read(p)
}

func BenchmarkMergeToDisk(b *testing.B) {
	const N int64 = 1_000_000
	const K int = 10

	values := make([][]byte, K)
	for i := range values {
		values[i] = make([]byte, N*structs.EntrySize)
		rand.Read(values[i])
	}
	for i := 0; i < b.N; i++ {
		// reset buffer pointer
		heap := make(MergeHeap, K)
		for i := range heap {
			heap[i] = &MergeHeapItem{
				Reader: bytes.NewBuffer(values[i]),
			}
		}
		// create tmp output
		output, err := os.CreateTemp("", "merge")
		assert.NoError(b, err)
		// create pipe to output
		pipeRd, pipeWr, err := os.Pipe()
		assert.NoError(b, err)
		bufferedPipeRd := bufio.NewReader(pipeRd)
		go io.Copy(output, bufferedPipeRd)

		_, err = MergeRuns(heap, &BufWriterCollecter{bufio.NewWriter(pipeWr)})
		assert.NoError(b, err)
		output.Close()
		os.Remove(output.Name())
	}
}

func BenchmarkBytesCompare(b *testing.B) {
	const N int = 100000
	type Pair struct {
		I, J int
	}
	values := make([]byte, N*structs.EntrySize)
	pairs := make([]Pair, N)
	for i := range pairs {
		pairs[i] = Pair{I: rand.Intn(N), J: rand.Intn(N)}
	}
	rand.Seed(time.Now().Unix())
	rand.Read(values)
	total := 0
	b.Run("compare with bytes.Compare", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c := 0
			for _, p := range pairs {
				if bytes.Compare(values[p.I*structs.EntrySize:p.I*structs.EntrySize+structs.KeySize], values[p.J*structs.EntrySize:p.J*structs.EntrySize+structs.KeySize]) < 0 {
					c++
				}
			}
			total = c
		}
	})
	b.Run("compare with strings", func(b *testing.B) {
		b.ReportAllocs()
		for t := 0; t < b.N; t++ {
			s := 0
			for _, p := range pairs {
				var jhdr, khdr reflect.StringHeader
				base := *(*reflect.SliceHeader)(unsafe.Pointer(&values))
				jhdr.Len = structs.KeySize
				khdr.Len = structs.KeySize
				jhdr.Data = base.Data + uintptr(p.I*structs.EntrySize)
				khdr.Data = base.Data + uintptr(p.J*structs.EntrySize)
				s1 := *(*string)(unsafe.Pointer(&jhdr))
				s2 := *(*string)(unsafe.Pointer(&khdr))
				if s1 < s2 {
					s++
				}
			}
			assert.Equal(b, total, s)
		}
	})
	b.Run("compare with uint64/16", func(b *testing.B) {
		for t := 0; t < b.N; t++ {
			r := 0
			for _, p := range pairs {
				if BytesCompare10(values[p.I*structs.EntrySize:p.I*structs.EntrySize+structs.KeySize], values[p.J*structs.EntrySize:p.J*structs.EntrySize+structs.KeySize]) {
					r++
				}
			}
			assert.Equal(b, total, r)
		}
	})
	b.Run("compare with cgo", func(b *testing.B) {
		for t := 0; t < b.N; t++ {
			s := 0
			for _, p := range pairs {
				if Cmp10(values[p.I*structs.EntrySize:p.I*structs.EntrySize+structs.KeySize], values[p.J*structs.EntrySize:p.J*structs.EntrySize+structs.KeySize]) {
					s++
				}
			}
			assert.Equal(b, total, s)
		}
	})
	//assert.Equal(b, t, s)
}

func TestSortFile(t *testing.T) {
	zl, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(zl)
	// generate tmp file at the beginning
	assert.NoError(t, exec.Command("gensort", "100000000", "test-file").Run())
	defer os.Remove("test-file")
	//defer os.Remove("test-file")
	// sort file
	const maxMem = 2000000000
	runs, err := CreateSortRuns("test-file", maxMem)
	assert.NoError(t, err)
	runtime.GC()
	// create merge heap
	heapItems := make([]*MergeHeapItem, len(runs))
	for i := range runs {
		runFile, err := os.Open(runs[i])
		assert.NoError(t, err)
		heapItems[i] = &MergeHeapItem{Reader: bufio.NewReader(runFile)}
		defer os.Remove(runs[i])
	}
	targetFile, err := os.Create("test-file")
	assert.NoError(t, err)
	heapCollecter := &BufWriterCollecter{Writer: bufio.NewWriter(targetFile)}
	_, err = MergeRuns(heapItems, heapCollecter)
	assert.NoError(t, err)
	assert.NoError(t, targetFile.Close())
	// run validation
	assert.NoError(t, exec.Command("valsort", "-q", "test-file").Run())
}
