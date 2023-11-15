package gdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type resourceStream struct {
	Filter
	Length int
	buf    *bytes.Buffer
	refnum int
	ot     bool
}

func (r *resourceStream) setRef(i int)    { r.refnum = i }
func (r *resourceStream) refNum() int     { return r.refnum }
func (r *resourceStream) children() []obj { return []obj{} }
func (r *resourceStream) encode(w io.Writer) (int, error) {
	var n int
	switch r.Filter {
	case FILTER_FLATE:
		encbuf := new(bytes.Buffer)
		l1 := r.buf.Len()
		_, err := flateCompress(encbuf, r.buf)
		if err != nil {
			return 0, err
		}
		var t int
		if !r.ot {
			t, err = fmt.Fprintf(w, "<<\n/Filter /FlateDecode\n/Length1 %d\n/Length %d\n>>\nstream\n", l1, encbuf.Len())
			if err != nil {
				return t, err
			}
		} else {
			t, err = fmt.Fprintf(w, "<<\n/Filter /FlateDecode\n/Length1 %d\n/Length %d\n/Subtype /OpenType\n>>\nstream\n", l1, encbuf.Len())
			if err != nil {
				return t, err
			}
		}

		encbuf.Write([]byte("\nendstream\n"))
		t2, err := encbuf.WriteTo(w)
		if err != nil {
			return t + int(t2), err
		}
		return t + int(t2), err
	default:
		t, err := fmt.Fprintf(w, "<<\n/Length %d\n>>\nstream\n", r.buf.Len())
		if err != nil {
			return t, err
		}
		n += t
		t2, err := r.buf.WriteTo(w)
		n += int(t2)
		if err != nil {
			return n, err
		}
	}
	t, err := w.Write([]byte("\nendstream\n"))
	if err != nil {
		return n + t, err
	}
	return n + t, nil
}

func (r resourceDict) String() string {
	bldr := new(strings.Builder)
	bldr.WriteString("<<\n")
	if len(r.Fonts) > 0 {
		bldr.WriteString("/Font <<\n")
		for i, f := range r.Fonts {
			fmt.Fprintf(bldr, "/F%d %d 0 R\n", i, f.refNum())
		}
		bldr.WriteString(">>\n")
	}
	if len(r.ExtGState) > 0 {
		bldr.WriteString("/ExtGState <<\n")
		for i, e := range r.ExtGState {
			fmt.Fprintf(bldr, "/GS%d %d 0 R\n", i, e.refNum())
		}
		bldr.WriteString(">>\n")
	}

	//if len(r.XObject) != 0 {
	//	bldr.WriteString("/XObject <<\n")
	//	for i, x := range r.XObject {
	//		fmt.Fprintf(bldr, "/Im%d %d 0 R\n", i, x.RefNum())
	//	}
	//	bldr.WriteString(">>\n")
	//}
	bldr.WriteString(">>\n")
	return bldr.String()
}
