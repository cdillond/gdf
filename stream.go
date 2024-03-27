package gdf

import (
	"bytes"
	"fmt"
	"io"
)

var (
	// start of stream token
	sos = []byte("stream\n")
	// end of stream token
	eos = []byte("\nendstream\n")
)

// A stream represents a PDF data stream.
type stream struct {
	Filter
	buf    []byte
	cLen   *lenObj // Records the length of the compressed resource stream.
	refnum int
	extras []field
}

func (s *stream) mark(i int) { s.refnum = i }
func (s *stream) id() int    { return s.refnum }
func (s *stream) children() []obj {
	if s.Filter == Flate && len(s.buf) > 2048 && s.cLen == nil {
		s.cLen = new(lenObj)
		return []obj{s.cLen}
	}
	return nil
}
func (s *stream) encode(w io.Writer) (int, error) {
	if !(s.Filter.isValid()) {
		return 0, fmt.Errorf("invalid compression filter %d", s.Filter)
	}
	// It's not worth it to compress data that's already small.
	if len(s.buf) < 1024 {
		s.Filter = NoFilter
	}

	// 7.3.8.2: "There may be an additional EOL marker, preceding endstream, that is not included in the count and is not logically part of the stream data."
	// It is simpler just to remove trailing \n or \r bytes.
	i := len(s.buf) - 1
	for ; i > -1; i-- {
		if !isEOL(s.buf[i]) {
			break
		}
	}
	s.buf = s.buf[:i+1]
	dlen := len(s.buf) // take the len now because once part of s.buf has been written to an io.Writer, there's no guarantee it will remain unchanged.

	if s.Filter == Flate && s.cLen != nil {
		// For longer streams, writes are made directly to w. The compressed length is recorded as an indirectly-referenced object.
		n, err := w.Write(append(dict(512, append(
			[]field{
				{"/Filter", "/FlateDecode"},
				{"/Length1", dlen},
				{"/Length", iref(s.cLen.id())},
			}, s.extras...)),
			sos...))
		if err != nil {
			return n, err
		}
		t, err := flateCompress(w, s.buf)
		if err != nil {
			return n + t, err
		}
		n += t
		s.cLen.Length = t
		t, err = w.Write(eos)
		return n + t, err

	} else if s.Filter == Flate {
		// For shorter streams, the content is compressed to a buffer, which is then written to w.
		encbuf := new(bytes.Buffer)
		encbuf.Grow(dlen + len(eos))
		t, err := flateCompress(encbuf, s.buf)
		if err != nil {
			return 0, err
		}
		n, err := w.Write(append(dict(512, append(
			[]field{
				{"/Filter", "/FlateDecode"},
				{"/Length1", dlen},
				{"/Length", t},
			}, s.extras...)),
			sos...))
		if err != nil {
			return n, err
		}
		encbuf.Write(eos)
		t64, err := encbuf.WriteTo(w)
		return n + int(t64), err
	}

	// Uncompressed streams are written directly to w.
	n, err := w.Write(append(dict(256, append(
		[]field{
			{"/Length", dlen},
		}, s.extras...)),
		sos...))
	if err != nil {
		return n, err
	}
	s.buf = append(s.buf, eos...)
	t, err := w.Write(s.buf)
	return n + t, err
}

type lenObj struct {
	Length int
	refnum int
}

func (l *lenObj) mark(i int)      { l.refnum = i }
func (l *lenObj) id() int         { return l.refnum }
func (l *lenObj) children() []obj { return nil }
func (l *lenObj) encode(w io.Writer) (int, error) {
	return w.Write(append(itob(l.Length), '\n'))
}
