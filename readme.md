Package gdf defines an interface for generating PDFs. It hews closely to the basics of the
PDF 2.0 specification and implements a relatively low-level API on which higher-level abstractions can
be built. **This project is a work in progress.**

gdf should be sufficient for most basic English-language PDF use cases. It avoids complexity
by purposefully ignoring some aspects of the PDF specification. These omissions may be
critical if you need accessibility features or support for multiple languages.

Understanding the PDF coordinate system can go a long way to simplifying the use of this package.

Every item is drawn, according to its type, at the origin of its coordinate space. The
coordinate space is then transformed by one or more affine matrices, always including the
current transformation matrix, and rendered onto the page's "user space," which has its origin
at the *lower left* corner of the page. Text space, for instance, is transformed first by the current text matrix and then by the current transformation matrix. The PDF specification also mentions glyph space, image space, form space, and pattern space, but they are not relevant to this package in its present form.

Transformation matrices are defined by 6 parameters representing the translation,
scale, and shear of the X and Y coordinates of a point transformed by the given matrix.
Each matrix includes an implicit third column of [0, 0, 1]. Because the space of an object
can be scaled or rotated, the effect of certain operations may be difficult to intuit. For example,
if the Current Transformation Matrix were [[1 0 0][2 0 0][0 0 1]], to draw a line from (10, 10) to (250, 250) in "user space," you would need to first move the path cursor to (10, 5) in the default graphics space, and then draw and stroke a path to (250, 125), since the Current Transformation Matrix would scale the y-coordinates of the original space by two. This could be achieved through the following code:
```go
    pdf := gdf.NewPDF() // create a new PDF instance
    page := gdf.NewPage(gdf.A4, gdf.NoMargins) // start a new page
    cs := page.NewContentStream() // content is drawn to a page's content stream
    cs.Concat(gdf.ScaleBy(1, 2)) // concatenate an affine matrix representing a 2*y scaling to the Current Transformation Matrix (by default the identity matrix)
    cs.MoveTo(10, 5) // start a new path at (10, 5); this will be (10, 10) on the page
    cs.LineTo(250, 125) // create a line to (250, 125), which will be (250, 250) on the page
    cs.Stroke() // stroke the line so that it appears on the page
    pdf.AppendPage(&page) // add the page to the current PDF document
    f, err := os.Create("out.pdf")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    pdf.Write(f) // write the PDF to out.pdf

```

The default basic unit for a PDF document is the point, defined as 1/72 of an inch.
However, text can be measured in terms of both points and unscaled font units.
The font size (in points) indicates the number of points per side of a glyph's em square. PDF fonts always contain 1000 font units per em square, so a conversion from font units to points can be
obtained by calculating fontSize*numFontUnits/1000. The CharSpace and WordSpace
elements of a PDF's TextState are defined in font units.

Fonts are the last source of complexity that will be mentioned here. The PDF specification allows for several different types of font. gdf supports only TrueType/OpenType fonts. Unlike in many document formats, a PDF font determines the encoding of the text rendered in that font. Attending to all of the many font types, font table formats, encodings, etc. can quickly become tedious and overwhelming (not to mention difficult to debug without knowledge of non-Latin scripts). It is gdf's aim to be lightweight and simple, rather than comprehensive. gdf therefore only supports Windows-1252 ("WinAnsiEncoding") character encodings. gdf takes care of encoding the UTF-8 strings accepted as input to its functions, but users should be aware that any text that contains characters not included in the Windows-1252 character set will not be rendered as intended.