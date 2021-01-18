package control

import (
	"math/big"

	"github.com/lnsp/cloudsort/structs"
)

func GetKeyRanges(n int) []structs.KeyRange {
	totalBytes := make([]byte, structs.KeySize)
	for i := range totalBytes {
		totalBytes[i] = 255
	}
	var total big.Int
	total.SetBytes(totalBytes)
	var partial big.Int
	partial.Div(&total, big.NewInt(int64(n)))
	var one big.Int
	one.SetInt64(1)

	ranges := make([]structs.KeyRange, n)
	next := big.Int{}
	for i := 0; i < n; i++ {
		ranges[i].Start = next.FillBytes(make([]byte, structs.KeySize))
		next.Add(&next, &partial)
		if i < n-1 {
			ranges[i].End = next.FillBytes(make([]byte, structs.KeySize))
		} else {
			ranges[i].End = total.FillBytes(make([]byte, structs.KeySize))
		}
		next.Add(&next, &one)
	}
	return ranges
}
