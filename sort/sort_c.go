package sort

/*
#include <stdint.h>
static char Cmp10(char* a, char* b) {
uint64_t ax = __builtin_bswap64(*(uint64_t*)(a));
uint64_t bx = __builtin_bswap64(*(uint64_t*)(b));
if (ax < bx) return 1;
else if (ax > bx) return 0;
uint16_t as = __builtin_bswap16(*(uint16_t*)(a+8));
uint16_t bs = __builtin_bswap16(*(uint16_t*)(b+8));
return as < bs ? 1 : 0;
}
*/
import "C"
import "unsafe"

func Cmp10(a []byte, b []byte) bool {
	r := C.Cmp10((*C.char)(unsafe.Pointer(&a[0])), (*C.char)(unsafe.Pointer(&b[0])))
	return r != 0
}
