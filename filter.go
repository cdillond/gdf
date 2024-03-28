package gdf

import (

	//"compress/zlib"
	"io"

	"github.com/klauspost/compress/zlib"
)

// A Filter is a compression algorithm that can compress internal PDF data.
type Filter uint

const (
	Flate Filter = iota
	NoFilter
	invalidFilter
)

var filters = [...]string{"/FlateDecode", ""}

func (f Filter) isValid() bool  { return f < invalidFilter }
func (f Filter) String() string { return filters[oneif(!f.isValid())] }

// n records the number of bytes written to w; the io.Writer.Write(p) method returns the number of bytes from p consumed by the writer.
// This is needed to determine the length of the encoded portion of a compressed resource stream.
type cwriter struct {
	w io.Writer
	n int
}

func (c *cwriter) Write(p []byte) (int, error) {
	t, err := c.w.Write(p)
	c.n += t
	return t, err
}

// flateCompress returns the number of (compressed) bytes written to w, not the number of bytes written from src.
func flateCompress(w io.Writer, src []byte) (int, error) {
	c := &cwriter{
		w: w,
	}
	zw := zlib.NewWriter(c)
	_, err := zw.Write(src)
	if err != nil {
		return c.n, err
	}
	return c.n, zw.Close()
}
