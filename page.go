package gdf

import (
	"errors"
	"io"
)

type Page struct {
	c        *ContentStream
	MediaBox Rect
	CropBox  Rect // "the rectangle of user space corresponding to the visible area of the intended output medium (display window or printed page)"
	Margins
	refnum int
	parent *pages
}

// ContentStream returns a pointer to p's ContentStream.
func (p Page) ContentStream() *ContentStream {
	if p.c == nil {
		p.c = p.newContentStream()
	}
	return p.c
}

// A resourceDict holds references to resources used by the content stream. When the PDF is built, this is promoted to part of the Page object dictionary.
// It is easier to group the resourceDict with the ContentStream, since doing so allows for Form XObjects to include resources of their own.
type resourceDict struct {
	Fonts []*Font

	ExtGState []*extGS
	Images    []*Image
	XForms    []*XContent

	Widgets    []*Widget
	TextAnnots []*TextAnnot

	/*
		TODO:
		ExtGState []*ExtGState
		ColorSpace
		Pattern
		Shading
		XObject
		Properties

	*/
}

func (r resourceDict) bytes() []byte {
	if len(r.Fonts)+len(r.XForms)+len(r.Images) == 0 {
		return []byte("<<>>")
	}

	fields := make([]field, 0, 3)
	// Font Subdict
	if len(r.Fonts) > 0 {
		ffields := make([]field, len(r.Fonts))
		for i := range r.Fonts {
			ffields[i] = field{"/F" + itoa(i), iref(r.Fonts[i])}
		}
		fields = append(fields, field{
			"/Font", subdict(128, ffields),
		})
	}

	if len(r.XForms) > 0 || len(r.Images) > 0 {
		xfields := make([]field, 0, len(r.XForms)+len(r.Images))
		for i := range r.XForms {
			xfields = append(xfields, field{"/P" + itoa(i), iref(r.XForms[i])})
		}
		for i := range r.Images {
			xfields = append(xfields, field{"/Im" + itoa(i), iref(r.Images[i])})
		}
		fields = append(fields, field{
			"/XObject", subdict(128, xfields),
		})
	}

	// ExtGState Subdict
	if len(r.ExtGState) > 0 {
		efields := make([]field, len(r.ExtGState))
		for i := range r.ExtGState {
			efields[i] = field{"/GS" + itoa(i), iref(r.ExtGState[i])}
		}
		fields = append(fields, field{
			"/ExtGState", subdict(128, efields),
		})
	}
	return subdict(256, fields)
}

// NewPage returns a new Page object with the specified size and margins.
func NewPage(pageSize Rect, margins Margins) Page {
	p := Page{MediaBox: pageSize, CropBox: pageSize, Margins: margins}
	p.c = p.newContentStream()
	return p
}

// Appends page to p.
func (p *PDF) AppendPage(page *Page) {
	p.catalog.pages.P = append(p.catalog.pages.P, page)
	page.parent = p.catalog.pages
}

// Inserts page at index i of the PDF's internal page structure.
func (p *PDF) InsertPage(page *Page, i int) error {
	if i < 0 || i > len(p.catalog.pages.P) {
		return errors.New("out of bounds")
	}
	if i == len(p.catalog.pages.P) {
		p.catalog.pages.P = append(p.catalog.pages.P, page)
		return nil
	}
	dst := make([]*Page, len(p.catalog.pages.P)+1)
	copy(dst, p.catalog.pages.P[:i])
	dst[i] = page
	copy(dst[i+1:], p.catalog.pages.P[i:])
	p.catalog.pages.P = dst
	return nil
}

// ReplacePage replaces the page at index i of the PDF's internal page structure with page.
func (p *PDF) ReplacePage(page *Page, i int) error {
	if i < 0 || i >= len(p.catalog.pages.P) {
		return errors.New("out of bounds")
	}
	p.catalog.pages.P[i] = page
	return nil
}

func (p *Page) mark(i int) { p.refnum = i }
func (p *Page) id() int    { return p.refnum }
func (p *Page) children() []obj {
	if p.c == nil {
		p.c = p.newContentStream()
	}
	out := make([]obj, 0, len(p.c.resources.Fonts)+len(p.c.resources.XForms)+len(p.c.resources.Images)+len(p.c.resources.ExtGState)+1+(len(p.c.resources.Widgets)))
	for i := range p.c.resources.Fonts {
		out = append(out, p.c.resources.Fonts[i])
	}
	//for i := range p.C.resources.XObjs {
	//	out = append(out, p.C.resources.XObjs[i])
	//}
	for i := range p.c.resources.XForms {
		out = append(out, p.c.resources.XForms[i])
	}
	for i := range p.c.resources.Images {
		out = append(out, p.c.resources.Images[i])
	}
	for i := range p.c.resources.ExtGState {
		out = append(out, p.c.resources.ExtGState[i])
	}
	for i := range p.c.resources.TextAnnots {
		out = append(out, p.c.resources.TextAnnots[i])
	}
	for i := range p.c.resources.Widgets {
		out = append(out, p.c.resources.Widgets[i])
	}
	return append(out, p.c)
}

func (p *Page) encode(w io.Writer) (int, error) {
	var fields []field

	if len(p.c.resources.Widgets) > 0 {
		a := make([]string, 0, len(p.c.resources.Widgets)+len(p.c.resources.TextAnnots))
		for _, an := range p.c.resources.TextAnnots {
			a = append(a, iref(an))
		}
		for _, an := range p.c.resources.Widgets {
			a = append(a, iref(an))
		}
		fields = append(fields, field{
			"/Annots", a,
		})
	}

	return w.Write(dict(512, append([]field{
		{"/Type", "/Page"},
		{"/Parent", iref(p.parent)},
		{"/MediaBox", p.MediaBox},
		{"/CropBox", p.CropBox},
		{"/Contents", iref(p.c)},
		{"/Resources", p.c.resources.bytes()},
	}, fields...)))
}

func (p *Page) newContentStream() *ContentStream {
	cs := new(ContentStream)
	cs.buf = make([]byte, 0, 4096)
	cs.GS = newGS()
	cs.Filter = Flate
	return cs
}

// Annotate draws the TextAnnot t to the area of p described by r.
func (p *Page) Annotate(t *TextAnnot, r Rect) {
	t.rect = r
	p.c.resources.TextAnnots = append(p.c.resources.TextAnnots, t)
}
