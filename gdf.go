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

func BuildPDFTree(pdf *PDF) {
	IncludeObj(pdf, Obj((&pdf.catalog)))
	IncludeChildren(pdf, Obj(&(pdf.catalog)))
}

func WritePDF(pdf *PDF, w io.Writer) error {
	BuildPDFTree(pdf)
	// finalize fonts
	for _, obj := range pdf.objects {
		if fnt, ok := obj.(*Font); ok {
			CalculateWidths(fnt)
			err := fnt.Subset()
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
	if err := WriteHeader(pdf, w); err != nil {
		return err
	}
	if err := WriteObjects(pdf, w); err != nil {
		return err
	}
	if err := WriteXref(pdf, w); err != nil {
		return err
	}
	if err := WriteTrailer(pdf, w); err != nil {
		return err
	}
	return nil
}

func IncludeObj(pdf *PDF, obj Obj) {
	if obj.RefNum() == 0 { // has not been set yet
		pdf.objects = append(pdf.objects, obj)
		obj.SetRef(len(pdf.objects))
	}
}

func IncludeChildren(pdf *PDF, obj Obj) {
	for _, child := range obj.Children() {
		IncludeObj(pdf, child)
		IncludeChildren(pdf, child)
	}
}
