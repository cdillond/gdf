//go:build cgo && hbsubsetc

package subset

/*
#cgo LDFLAGS: -lharfbuzz -lharfbuzz-subset
#include <stdio.h>
#include <string.h>
#include <harfbuzz/hb.h>
#include <harfbuzz/hb-subset.h>

int subset(const unsigned char *src, unsigned int src_len, uint32_t uni_chars[], int num_chars, unsigned char *out)
{
    unsigned int out_len = 0;
	hb_blob_t *data = hb_blob_create_or_fail(src, src_len, HB_MEMORY_MODE_READONLY, NULL, NULL);
    if (data == NULL)
        return out_len;

    hb_face_t *face = hb_face_create(data, 0);

    hb_subset_input_t *input = hb_subset_input_create_or_fail();
    if (input == NULL)
        goto destroy_face;

    hb_set_t *charset = hb_subset_input_unicode_set(input);
    for (int i = 0; i < num_chars; i++)
        hb_set_add(charset, uni_chars[i]);

    hb_subset_input_set_flags(input, HB_SUBSET_FLAGS_RETAIN_GIDS);

    hb_face_t *sub_face = hb_subset_or_fail(face, input);
    if (sub_face == NULL)
        goto destroy_input;

    hb_blob_t *sub_blob = hb_face_reference_blob(sub_face);
    const char *out_data = hb_blob_get_data_writable(sub_blob, &out_len);
	memcpy(out, out_data, out_len);
	hb_blob_destroy(sub_blob);
destroy_input:
    hb_subset_input_destroy(input);
destroy_face:
    hb_face_destroy(face);

    return out_len;
}
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

	written := int(C.subset(
		(*C.uchar)(srcData),
		C.uint(uint(len(src))),
		(*C.uint)(charsetData),
		C.int(len(charset_u32)),
		(*C.uchar)(outData)))

	if written < 1 {
		return nil, fmt.Errorf("error subsetting font")
	}
	b = unsafe.Slice(outData, written)
	return b, nil
}
