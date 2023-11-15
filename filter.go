package gdf

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Filter uint

const (
	FILTER_NONE Filter = iota
	FILTER_LZW
	FILTER_FLATE
)

func flateCompress(w io.Writer, in *bytes.Buffer) (int, error) {
	zw := zlib.NewWriter(w)
	n, err := in.WriteTo(zw)
	if err != nil {
		return 0, err
	}
	return int(n), zw.Flush()
}
