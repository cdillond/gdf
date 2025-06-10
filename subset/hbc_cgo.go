//go:build cgo && hbsubsetc

package subset

/*
#cgo CFLAGS: -fno-strict-aliasing -g
#cgo LDFLAGS: -lharfbuzz -lharfbuzz-subset
#cgo nocallback subset
#cgo noescape subset
#include "hbc.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// HBSubsetC calls functions in libharfbuzz and libharfbuzz-subset via CGo and returns the source bytes of a font containing only the
// characters included in the cutset. In order for this function to work, CGo must be enabled, HarfBuzz v>=2.9.0 must be installed on
// your system, and `hbsubsetc` must be passed to the Go compiler as a build tag.
func HBSubsetC(src []byte, cutset map[rune]struct{}) ([]byte, error) {
	// convert runes to uint32_t chars readable by hb-subset
	charset_u32 := make([]uint32, len(cutset))
	for char := range cutset {
		charset_u32 = append(charset_u32, uint32(char))
	}
	// allocate at least as much as the current file size
	b := make([]byte, 0, len(src))

	srcData := unsafe.SliceData(src)
	charsetData := unsafe.SliceData(charset_u32)
	outData := unsafe.SliceData(b)

	written := int(C.subset((*C.uchar)(srcData), C.uint(uint(len(src))), (*C.uint)(charsetData), C.int(len(charset_u32)), (*C.uchar)(outData)))

	if written < 1 {
		return nil, fmt.Errorf("error subsetting font")
	}

	return b[:written], nil
}
