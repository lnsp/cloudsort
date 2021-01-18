package structs

const EntrySize = 100
const KeySize = 10

var KeyRangeStart = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var KeyRangeEnd = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

type KeyRange struct {
	Start, End []byte
}
