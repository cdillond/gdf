//go:build hbsubsetc
#include <harfbuzz/hb-subset.h>
#include <harfbuzz/hb.h>
#include <string.h>

#if HB_VERSION_ATLEAST(2, 9, 0)
#else
#error "only harfbuzz versions 2.9.0 and above are supported"
#endif

int subset(const unsigned char *src, unsigned int src_len, uint32_t uni_chars[],
           int num_chars, unsigned char *out);
