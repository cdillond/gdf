package gdf

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Filter uint

const (
	NoFilter Filter = iota
	Flate
)

func flateCompress(w io.Writer, in *bytes.Buffer) (int, error) {
	zw := zlib.NewWriter(w)
	n, err := in.WriteTo(zw)
	if err != nil {
		return 0, err
	}
	return int(n), zw.Flush()
}
