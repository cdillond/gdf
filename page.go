package gdf

import (
	"errors"
	"io"
)

type Page struct {
	refnum   int
	MediaBox Rect
	CropBox  Rect           // "the rectangle of user space corresponding to the visible area of the intended output medium (display window or printed page)"
	Content  *ContentStream // This will be only a single content stream so that we don't need to worry about the effects of concatenating multiple streams.
	Margins
	parent *pages
}

// A resourceDict holds references to resources used by the content stream. When the PDF is built, this is promoted to part of the Page object dictionary.
// It is easier to group the resourceDict with the ContentStream, since doing so allows for Form XObjects to include resources of their own.
type resourceDict struct {
	Fonts []*Font // *Fonts
	XObjs []*XObject

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
	if len(r.Fonts) == 0 && len(r.XObjs) == 0 {
		return []byte("<<>>")
	}

	fields := make([]field, 0, 2)
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
	if len(r.XObjs) > 0 {
		xfields := make([]field, len(r.XObjs))
		for i := range r.XObjs {
			xfields[i] = field{"/P" + itoa(i), iref(r.XObjs[i])}
		}
		fields = append(fields, field{
			"/XObject", subdict(128, xfields),
		})
	}
	return subdict(256, fields)
}

// NewPage returns a new Page object with the specified size and margins.
func NewPage(pageSize Rect, margins Margins) Page {
	p := Page{MediaBox: pageSize, CropBox: pageSize, Margins: margins}
	p.Content = p.newContentStream()
	return p
}

// Appends page to p.
func (p *PDF) AppendPage(page *Page) {
	p.catalog.Pages.P = append(p.catalog.Pages.P, page)
	page.parent = p.catalog.Pages
}

// Inserts page at index i of the PDF's internal page structure.
func (p *PDF) InsertPage(page *Page, i int) error {
	if i < 0 || i > len(p.catalog.Pages.P) {
		return errors.New("out of bounds")
	}
	if i == len(p.catalog.Pages.P) {
		p.catalog.Pages.P = append(p.catalog.Pages.P, page)
		return nil
	}
	dst := make([]*Page, len(p.catalog.Pages.P)+1)
	copy(dst, p.catalog.Pages.P[:i])
	dst[i] = page
	copy(dst[i+1:], p.catalog.Pages.P[i:])
	p.catalog.Pages.P = dst
	return nil
}

// ReplacePage replaces the page at index i of the PDF's internal page structure with page.
func (p *PDF) ReplacePage(page *Page, i int) error {
	if i < 0 || i >= len(p.catalog.Pages.P) {
		return errors.New("out of bounds")
	}
	p.catalog.Pages.P[i] = page
	return nil
}

func (p *Page) mark(i int) { p.refnum = i }
func (p *Page) id() int    { return p.refnum }
func (p *Page) children() []obj {
	out := make([]obj, 0, len(p.Content.resources.Fonts)+len(p.Content.resources.XObjs)+1+(len(p.Content.resources.Widgets)))
	for i := range p.Content.resources.Fonts {
		out = append(out, p.Content.resources.Fonts[i])
	}
	for i := range p.Content.resources.XObjs {
		out = append(out, p.Content.resources.XObjs[i])
	}
	for i := range p.Content.resources.TextAnnots {
		out = append(out, p.Content.resources.TextAnnots[i])
	}
	for i := range p.Content.resources.Widgets {
		out = append(out, p.Content.resources.Widgets[i])
	}
	return append(out, p.Content)
}

func (p *Page) encode(w io.Writer) (int, error) {
	var fields []field

	if len(p.Content.resources.Widgets) > 0 {
		a := make([]string, 0, len(p.Content.resources.Widgets)+len(p.Content.resources.TextAnnots))
		for _, an := range p.Content.resources.TextAnnots {
			a = append(a, iref(an))
			//a[i] = iref(p.C.resources.Annots[i].id())
		}
		for _, an := range p.Content.resources.Widgets {
			a = append(a, iref(an))
			//a[i] = iref(p.C.resources.Annots[i].id())
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
		{"/Contents", iref(p.Content)},
		{"/Resources", p.Content.resources.bytes()},
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
	p.Content.resources.TextAnnots = append(p.Content.resources.TextAnnots, t)
}
