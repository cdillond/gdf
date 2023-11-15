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
}

func NewPDF() *PDF {
	pdf := new(PDF)
	pdf.catalog.Pages = new(pages)
	return pdf
}

func buildPDFTree(pdf *PDF) {
	includeObj(pdf, obj((&pdf.catalog)))
	includeChildren(pdf, obj(&(pdf.catalog)))
}

// Builds the PDF and writes it to w.
func (p *PDF) Write(w io.Writer) error {
	buildPDFTree(p)

	if err := writeHeader(p, w); err != nil {
		return err
	}
	if err := writeObjects(p, w); err != nil {
		return err
	}
	if err := writeXref(p, w); err != nil {
		return err
	}
	if err := writeTrailer(p, w); err != nil {
		return err
	}
	return nil
}

func includeObj(pdf *PDF, o obj) {
	if o.refNum() == 0 { // has not been set yet
		pdf.objects = append(pdf.objects, o)
		o.setRef(len(pdf.objects))
	}
}

func includeChildren(pdf *PDF, o obj) {

	for _, child := range o.children() {
		// finalize fonts
		if fnt, ok := child.(*Font); ok {
			calculateWidths(fnt)
			if !fnt.noSubset {
				fnt.noSubset = true
				tmp := make(map[rune]struct{}, len(fnt.Charset))
				for key := range fnt.Charset {
					tmp[key] = struct{}{}
				}
				b, err := ttf.Subset(fnt.Font, fnt.srcb, tmp)
				if err != nil {
					log.Println(err.Error())
				} else {
					fnt.source.buf.Write(b)
				}
			}
		} /*
			if tfnt, ok := child.(*Type0Font); ok {
				tfnt.Subset()
				tfnt.GUnicode()
				CalculateWidths2(tfnt.DescendantFonts)
			}*/
		includeObj(pdf, child)
		includeChildren(pdf, child)
	}
}
