package gdf

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/klauspost/compress/zlib"
)

// FlateCompress compresses x's Data using the Flate filter, and updates the XImage to indicate that the filter has been applied.
// Most images (except JPEGs) are compressed using this filter automatically when the output PDF file is written. However, completing
// the process ahead of time by calling FlateCompress and then re-using a pre-decoded and pre-compressed XImage can be significantly more efficient.
func (x *XImage) FlateCompress() error {
	x.RawDataLen = len(x.Data)
	buf := new(bytes.Buffer)
	buf.Grow(x.RawDataLen)
	zw := zlib.NewWriter(buf)
	if _, err := zw.Write(x.Data); err != nil {
		return err
	}
	zw.Close()
	x.Data = buf.Bytes()
	x.AppliedFilter = Flate

	if x.Alpha != nil {
		x.Alpha.RawDataLen = len(x.Alpha.Data)
		buf = new(bytes.Buffer)
		buf.Grow(x.Alpha.RawDataLen)
		zw = zlib.NewWriter(buf)
		if _, err := zw.Write(x.Alpha.Data); err != nil {
			return err
		}
		zw.Close()
		x.Alpha.Data = buf.Bytes()
		x.Alpha.AppliedFilter = Flate
	}
	return nil
}

func (x XImage) SaveTo(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(x)
}

func LoadXImage(r io.Reader) (XImage, error) {
	dec := gob.NewDecoder(r)
	var v XImage
	return v, dec.Decode(&v)
}
