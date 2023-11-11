package gdf

import (
	"io"
	"log"
)

type PDF struct {
	catalog Catalog // root object
	objects []Obj
	n       int   // byte offset
	xref    []int // maps reference numbers [index] to the corresponding object's byte offset
}

func NewPDF() *PDF {
	pdf := new(PDF)
	pdf.catalog.Pages = new(Pages)
	return pdf
}

func buildPDFTree(pdf *PDF) {
	includeObj(pdf, Obj((&pdf.catalog)))
	includeChildren(pdf, Obj(&(pdf.catalog)))
}

func WritePDF(pdf *PDF, w io.Writer) error {
	buildPDFTree(pdf)
	// finalize fonts
	for _, obj := range pdf.objects {
		if fnt, ok := obj.(*Font); ok {
			CalculateWidths(fnt)
			if !fnt.noSubset {
				err := fnt.subset()
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}
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

func includeObj(pdf *PDF, obj Obj) {
	if obj.refNum() == 0 { // has not been set yet
		pdf.objects = append(pdf.objects, obj)
		obj.setRef(len(pdf.objects))
	}
}

func includeChildren(pdf *PDF, obj Obj) {
	for _, child := range obj.children() {
		includeObj(pdf, child)
		includeChildren(pdf, child)
	}
}
