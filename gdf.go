package gdf

import (
	"io"
	"log"
)

type PDF struct {
	catalog catalog // root object
	objects []obj
	n       int   // byte offset
	xref    []int // maps reference numbers [index] to the corresponding object's byte offset
}

func NewPDF() *PDF {
	pdf := new(PDF)
	pdf.catalog.Pages = new(Pages)
	return pdf
}

func buildPDFTree(pdf *PDF) {
	includeObj(pdf, obj((&pdf.catalog)))
	includeChildren(pdf, obj(&(pdf.catalog)))
}

func WritePDF(pdf *PDF, w io.Writer) error {
	buildPDFTree(pdf)

	if err := writeHeader(pdf, w); err != nil {
		return err
	}
	if err := writeObjects(pdf, w); err != nil {
		return err
	}
	if err := writeXref(pdf, w); err != nil {
		return err
	}
	if err := writeTrailer(pdf, w); err != nil {
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
				err := fnt.subset()
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
		includeObj(pdf, child)
		includeChildren(pdf, child)
	}
}
