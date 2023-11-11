package gdf

/*
Package gpdf defines an interface for generating PDFs. It hews closely to the basics of the
PDF 2.0 specification and implements a relatively low-level API that higher-level abstractions can
be built on.

gpdf should be sufficient for most basic English-language PDF use cases. It avoids complexity
by purposefully ignoring some of the usages that either 1) offer broader language and accessibility support,
or 2) make use of rarely-used, specialized features.

The remaining complexity lies in the fonts and understanding the coordinate system.

Every item is drawn, according to its type, at the origin of its coordinate space. The
coordinate space is then transformed by one or more matrices, always including the
Current Transformation Matrix, and rendered onto the page's "user space." Text space is
transformed first by the current text matrix and then by the current transformation matrix.
The PDF specification also mentions glyph space, image space, form space, and pattern space,
but they are not relevant to this package.

Transformation matrices are defined by 6 parameters representing the translation,
scale, and shear of the X and Y coordinates of a point transformed by the given matrix.
Because the space of an object can be scaled or rotated, the effect of certain operations
may be difficult to determine. For instance, drawing a line from (10, 10) to (10, 20)
in the default graphics space moves the cursor from (10, 10) to (10, 40) in
user space if the Current Transformation Matrix is [1 0 0][2 0 0][0 0 1]; i.e.,
if it scales the y-coordinates of the original space by 2.

Additionally, text is defined both in terms of points (1/72 of an inch) and unscaled font units.
The default basic unit for a PDF document is the point. The font size (in points)
indicates the number of points per side of a glyph's em square. PDF fonts always
contain 1000 font units per em square. The font unit to point conversion therefore
depends on the text's current font size. The Tc (character spacing) and Tw (word spacing)
elements of a PDF's Text State are defined in terms of font units, not points.

A final source of complexity is fonts. The original PDF specification made use of
Type1 fonts built in to all PDF rendering software. Built-in Type1 fonts are now
deprecated, but their legacy remains. There are multiple ways to embed a font in
a PDF, but they all must specify a text encoding. A page's current font and its
associated objects therefore defines both the appearance of the document's glyphs
and the encoding of the source text. The default encoding is ASCII, not UTF-8 :(.
Writing systems that include more than 256 code points must be, to some extent,
user-defined. Luckily for English speakers, it is relatively easy to use Windows-1252
("WinAnsiEncoding") encoded text or, less commonly Mac OS Roman ("MacRomanEncoding") text.

Other encodings can be accomodated, but they require more work.


// glyph space
// text space (defined by text matrix)
// image space (predefined
// form space (for form XObjects)
// pattern space (patern matrix)

*/
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
