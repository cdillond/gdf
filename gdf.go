package gdf

import (
	"io"
	"log"

	"github.com/cdillond/gdf/ttf"
)

type PDF struct {
	catalog catalog // root object
	objects []obj
	n       int   // byte offset
	xref    []int // maps reference numbers [index] to the corresponding object's byte offset
	info    *InfoDict
}

func NewPDF() *PDF {
	pdf := new(PDF)
	pdf.catalog.Pages = new(pages)
	pdf.catalog.Acroform = new(acroform)
	return pdf
}

func buildPDFTree(pdf *PDF) {
	// first we need to check for any widget annotations so that we can resolve the acroform fields.
	/*for _, p := range pdf.catalog.Pages.P {
		p.parent = pdf.catalog.Pages
		for _, widget := range p.C.resources.Widgets {
			if pdf.catalog.Acroform == nil {
				pdf.catalog.Acroform = new(acroform)
			}
			pdf.catalog.Acroform.acrofields = append(pdf.catalog.Acroform.acrofields, widget.acrofield)
			widget.acrofield.parent = pdf.catalog.Acroform
			widget.acrofield.Child = widget
		}
	}*/

	includeObj(pdf, obj((&pdf.catalog)))
	includeChildren(pdf, obj(&(pdf.catalog)))
	if pdf.info != nil {
		includeObj(pdf, obj(pdf.info))
	}
}

// Builds the PDF and writes it to w. Attempting to write an empty PDF, i.e., one without any content streams, causes a panic.
func (p *PDF) WriteTo(w io.Writer) (int64, error) {
	buildPDFTree(p)
	if err := writeHeader(p, w); err != nil {
		return int64(p.n), err
	}
	if err := writeObjects(p, w); err != nil {
		return int64(p.n), err
	}
	if err := writeXref(p, w); err != nil {
		return int64(p.n), err
	}
	if err := writeTrailer(p, w); err != nil {
		return int64(p.n), err
	}
	return int64(p.n), nil
}

func includeObj(pdf *PDF, o obj) {
	if o.id() == 0 { // has not been set yet
		pdf.objects = append(pdf.objects, o)
		o.mark(len(pdf.objects))
	}
}

func includeChildren(pdf *PDF, o obj) {
	for _, child := range o.children() {
		// finalize fonts
		if fnt, ok := child.(*Font); ok {
			calculateWidths(fnt)
			if !fnt.noSubset {
				fnt.noSubset = true
				tmp := make(map[rune]struct{}, len(fnt.charset))
				for key := range fnt.charset {
					tmp[key] = struct{}{}
				}
				b, err := ttf.Subset(fnt.Font, fnt.srcb, tmp)
				if err != nil {
					log.Println(err.Error())
				} else {
					//fnt.source.buf.Write(b)
					fnt.source.buf = b
				}
			}
		}
		includeObj(pdf, child)
		includeChildren(pdf, child)
	}
}
