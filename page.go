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
	Margins
	resourceDict
}
type resourceDict struct {
	Fonts     []*Font // *Fonts
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
}

// Appends page to p.
func (p *PDF) AppendPage(page *Page) {
	p.catalog.Pages.P = append(p.catalog.Pages.P, page)
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

// Replaces the page at index i of the PDF's internal page structure with page.
func (p *PDF) ReplacePage(page *Page, i int) error {
	if i < 0 || i >= len(p.catalog.Pages.P) {
		return errors.New("out of bounds")
	}
	p.catalog.Pages.P[i] = page
	return nil
}

// glyph space
// text space (defined by text matrix)
// image space (predefined
// form space (for form XObjects)
// pattern space (patern matrix)

func (p *Page) setRef(i int) { p.refnum = i }
func (p *Page) refNum() int  { return p.refnum }
func (p *Page) children() []obj {
	out := make([]obj, 0, len(p.resourceDict.Fonts)+len(p.Content)) //+len(p.ResourceDict.XObject))
	for i := range p.resourceDict.Fonts {
		out = append(out, p.resourceDict.Fonts[i])
	}
	for i := range p.resourceDict.ExtGState {
		out = append(out, p.resourceDict.ExtGState[i])
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
		p.MediaBox.LLX, p.MediaBox.LLY, p.MediaBox.URX, p.MediaBox.URY, p.CropBox.LLX, p.CropBox.LLY, p.CropBox.URX, p.CropBox.URY, p.Content[0].refNum(), p.resourceDict.String())
}

func (p *Page) NewContentStream() *ContentStream {
	cs := new(ContentStream)
	p.Content = append(p.Content, cs)
	cs.Parent = p
	cs.buf = new(bytes.Buffer)
	cs.GS = NewGS()
	return cs
}
