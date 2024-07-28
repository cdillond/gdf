package gdf

import (
	"io"
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
	pdf.catalog.pages = new(pages)
	pdf.catalog.acroform = new(acroform)
	return pdf
}

func buildPDFTree(pdf *PDF) error {
	includeObj(pdf, &pdf.catalog)
	if err := includeChildren(pdf, &pdf.catalog); err != nil {
		return err
	}
	if pdf.info != nil {
		includeObj(pdf, pdf.info)
	}
	return nil
}

// Builds the PDF and writes it to w.
func (p *PDF) WriteTo(w io.Writer) (int64, error) {
	if err := buildPDFTree(p); err != nil {
		return 0, err
	}
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

func includeChildren(pdf *PDF, o obj) error {
	for _, child := range o.children() {
		// finalize fonts
		if f, ok := child.(*Font); ok {
			calculateWidths(f)

			tmp := make(map[rune]struct{}, len(f.charset))
			for key := range f.charset {
				tmp[key] = struct{}{}
			}
			if f.Subsetter == nil {
				f.source.buf = f.srcb
			} else {
				f.Subsetter.Init(f.SFNT, f.srcb, f.srcPath)
				b, err := f.Subsetter.Subset(tmp)
				if err != nil {
					return err
				} else {
					f.source.buf = b
				}
			}
		}
		includeObj(pdf, child)
		if err := includeChildren(pdf, child); err != nil {
			return err
		}
	}
	return nil
}
