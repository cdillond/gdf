package gdf

import "io"

type extGS struct {
	fields []field
	refnum int
}

func (e *extGS) mark(i int)      { e.refnum = i }
func (e *extGS) id() int         { return e.refnum }
func (e *extGS) children() []obj { return nil }
func (e *extGS) encode(w io.Writer) (int, error) {
	return w.Write(dict(256, e.fields))
}
