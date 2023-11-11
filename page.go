package gdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Page struct {
	refnum   int
	MediaBox Rect
	CropBox  Rect // "the rectangle of user space corresponding to the visible area of the intended output medium (display window or printed page)"
	Content  []*ContentStream
	ResourceDict
	Margins
}
type ResourceDict struct {
	Fonts     []*Font
	ExtGState []*ExtGState
	/*
		TODO:
		ColorSpace
		Pattern
		Shading
		XObject
		Properties

	*/
}

func NewPage(pageSize Rect, margins Margins) Page {
	return Page{MediaBox: pageSize, CropBox: pageSize, Margins: margins}
} // e.g. page := NewPage(pdf.DefaultPageSize)
func AppendPage(pdf *PDF, page *Page) {
	pdf.catalog.Pages.P = append(pdf.catalog.Pages.P, page)
}
func InsertPage(pdf *PDF, page *Page, i int) error {
	if i < 0 || i > len(pdf.catalog.Pages.P) {
		return errors.New("out of bounds")
	}
	if i == len(pdf.catalog.Pages.P) {
		pdf.catalog.Pages.P = append(pdf.catalog.Pages.P, page)
		return nil
	}
	dst := make([]*Page, len(pdf.catalog.Pages.P)+1)
	copy(dst, pdf.catalog.Pages.P[:i])
	dst[i] = page
	copy(dst[i+1:], pdf.catalog.Pages.P[i:])
	pdf.catalog.Pages.P = dst
	return nil
}
func ReplacePage(pdf *PDF, page *Page, i int) error {
	if i < 0 || i >= len(pdf.catalog.Pages.P) {
		return errors.New("out of bounds")
	}
	pdf.catalog.Pages.P[i] = page
	return nil
}

// glyph space
// text space (defined by text matrix)
// image space (predefined
// form space (for form XObjects)
// pattern space (patern matrix)

func (p *Page) setRef(i int) { p.refnum = i }
func (p *Page) refNum() int  { return p.refnum }
func (p *Page) children() []Obj {
	out := make([]Obj, 0, len(p.ResourceDict.Fonts)+len(p.Content)) //+len(p.ResourceDict.XObject))
	for i := range p.ResourceDict.Fonts {
		out = append(out, p.ResourceDict.Fonts[i])
	}
	for i := range p.ResourceDict.ExtGState {
		out = append(out, p.ResourceDict.ExtGState[i])
	}
	//for i := range p.ResourceDict.XObject {
	//	obj := Obj(p.ResourceDict.XObject[i])
	//	out = append(out, &obj)
	//}
	for i := range p.Content {
		out = append(out, p.Content[i])
	}
	return out
}

func (p *Page) encode(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [%f %f %f %f]\n/CropBox [%f %f %f %f]\n/Contents %d 0 R\n/Resources %s>>\n",
		p.MediaBox.LLX, p.MediaBox.LLY, p.MediaBox.URX, p.MediaBox.URY, p.CropBox.LLX, p.CropBox.LLY, p.CropBox.URX, p.CropBox.URY, p.Content[0].RefNum(), p.ResourceDict.String())
}

func (p *Page) NewContentStream() *ContentStream {
	cs := new(ContentStream)
	p.Content = append(p.Content, cs)
	cs.Parent = p
	cs.buf = new(bytes.Buffer)
	cs.GS = NewGS()
	return cs
}
