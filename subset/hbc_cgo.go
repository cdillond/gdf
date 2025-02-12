//go:build cgo && hbsubsetc

package subset

/*
#cgo CFLAGS: -fno-strict-aliasing -g
#cgo LDFLAGS: -lharfbuzz -lharfbuzz-subset
#cgo nocallback subset
#cgo noescape subset
#include <stdio.h>
#include <string.h>
#include <harfbuzz/hb.h>
#include <harfbuzz/hb-subset.h>

#if HB_VERSION_ATLEAST(2, 9, 0)
#else
#error "only harfbuzz versions 2.9.0 and above are supported"
#endif

int subset(const unsigned char *src, unsigned int src_len, uint32_t uni_chars[], int num_chars, unsigned char *out)
{
    hb_blob_t *src_blob, *dst_blob;
    hb_face_t *src_face, *dst_face;
    hb_subset_input_t *input;
    hb_set_t *charset;
    unsigned const char *dst_data;
    unsigned int dst_len = 0;

    src_blob = hb_blob_create_or_fail(src, src_len, HB_MEMORY_MODE_READONLY, NULL, NULL);
    if (src_blob == NULL)
        return dst_len;

#if HB_VERSION_ATLEAST(10, 1, 0)
   	src_face = hb_face_create_or_fail(src_blob, 0);
    if (src_face == NULL)
        goto destroy_src_blob;
#else
    src_face = hb_face_create(src_blob, 0);
#endif

    input = hb_subset_input_create_or_fail();
    if (input == NULL)
        goto destroy_src_face;

    hb_subset_input_set_flags(input, HB_SUBSET_FLAGS_RETAIN_GIDS);

    charset = hb_subset_input_unicode_set(input);

    for (int i = 0; i < num_chars; i++)
        hb_set_add(charset, uni_chars[i]);

    dst_face = hb_subset_or_fail(src_face, input);
    if (dst_face == NULL)
        goto destroy_input;

    dst_blob = hb_face_reference_blob(dst_face);

    dst_data = hb_blob_get_data(dst_blob, &dst_len);

    if (dst_len > src_len)
        dst_len = 0;

    if (dst_len == 0)
        goto destroy_dst;

    memcpy(out, dst_data, dst_len);

destroy_dst:
    hb_blob_destroy(dst_blob);
    hb_face_destroy(dst_face);

destroy_input:
    hb_subset_input_destroy(input);

destroy_src_face:
    hb_face_destroy(src_face);

#if HB_VERSION_ATLEAST(10, 1, 0)
destroy_src_blob:
#endif
    hb_blob_destroy(src_blob);

    return dst_len;
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

	written := int(C.subset((*C.uchar)(srcData), C.uint(uint(len(src))), (*C.uint)(charsetData), C.int(len(charset_u32)), (*C.uchar)(outData)))

	if written < 1 {
		return nil, fmt.Errorf("error subsetting font")
	}

	return b[:written], nil
}
