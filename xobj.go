package gdf

import (
	"io"
)

type xType bool

const (
	XForm  xType = false
	XImage xType = true
)

func (x xType) String() string {
	if x {
		return "/Image"
	}
	return "/Form"
}

// An XObject is a content stream that can be reused. It can either represent an image or an arbitrary sequence of objects, i.e. a "form". (Form XObjects are unrelated to AcroForms and XFA Forms; the nomenclature is confusing.)
type XObject struct {
	ContentStream
	BBox Rect
	xType
}

// NewXObjFromBytes returns an XObject with b as the raw byte source. Use with caution.
func NewXObjFromBytes(b []byte, t xType, BBox Rect) *XObject {
	x := NewXObj(t, BBox)
	x.buf = append(x.buf, b...)
	return x
}

// NewXObj returns an XObject of type t.
func NewXObj(t xType, BBox Rect) *XObject {
	x := &XObject{
		BBox:  BBox,
		xType: t,
	}
	x.buf = make([]byte, 0, 2048)
	x.GS = newGS()
	x.Filter = Flate
	return x
}

func (x *XObject) mark(i int) { x.refnum = i }
func (x *XObject) id() int    { return x.refnum }
func (x *XObject) children() []obj {
	out := make([]obj, 0, len(x.resources.Fonts)+len(x.resources.XObjs))
	for i := range x.resources.Fonts {
		out = append(out, x.resources.Fonts[i])
	}
	for i := range x.resources.XObjs {
		out = append(out, x.resources.XObjs[i])
	}
	return out
} // not a pretty way to do this, but it works for now.

func (x *XObject) encode(w io.Writer) (int, error) {
	x.stream.extras = []field{
		{"/Type", "/XObject"},
		{"/Subtype", x.xType.String()},
		{"/BBox", x.BBox},
		{"/Resources", x.resources.bytes()},
	}
	return x.stream.encode(w)
}
