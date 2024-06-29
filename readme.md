[![Go Reference](https://pkg.go.dev/badge/github.com/cdillond/gdf.svg)](https://pkg.go.dev/github.com/cdillond/gdf)

Package gdf defines an interface for generating PDFs. It hews closely to the basics of the PDF 2.0 specification and implements a relatively low-level API on which higher-level abstractions can be built. **This project is a work in progress.**

gdf should be sufficient for most basic English-language PDF use cases. It avoids complexity by purposefully ignoring some aspects of the PDF specification. These omissions may be critical if you need accessibility features or support for multiple languages.

## Document Structure
A PDF document is a graph of objects. With the exception of the root `PDF` object, which has no parent, each object, depending on its type, can have 1 or more parents and 0 or more children. The `PDF` object can be thought of as a "context" in the sense of, e.g., a Cairo `cairo_t` or MuPDF `fz_context`. That is, any object derived from a `PDF` or one of its children should not be referenced by objects belonging to a different `PDF`. Operations on objects within a `PDF` context are not concurrency safe.

The general flow of a gdf PDF document generation program goes:
1. Create a `PDF` struct.
2. Load a `Font`. 
3. Create a new `Page`.
4. Draw text/graphics to the `Page`'s `ContentStream`.
5. Append the `Page` to the `PDF`.
6. Write the `PDF` to an output.         

## Graphics
Understanding the PDF coordinate system can go a long way to simplifying the use of this package.

Every item is drawn, according to its type, at the origin of its coordinate space - either user space, text space, glyph space (mostly irrelevant), image space, form space, or pattern space (unsupported). Each space has its origin at the *lower left* corner of the page and increases up and to the right. The coordinate space is then transformed by one or more affine matrices, always including the current transformation matrix, and rendered onto the page's "device space." Text space, for instance, is transformed first by the current text matrix and then by the current transformation matrix.

Transformation matrices are defined by 6 parameters representing the translation, scale, and shear of the X and Y coordinates of a point transformed by the given matrix. Each matrix includes an implicit third column of `[0, 0, 1]`. Because the space of an object can be scaled or rotated, the effect of certain operations may be difficult to intuit. For example, if the Current Transformation Matrix were `[[1 0 0][2 0 0][0 0 1]]`, to draw a line from `(10, 10)` to `(250, 250)` in device space, you would need to first move the path cursor to `(10, 5)` in user space, and then draw and stroke a path to `(250, 125)`, since the Current Transformation Matrix would scale the y-coordinates of the original space by two. This could be achieved through the following code:
```go
    pdf := gdf.NewPDF() // create a new PDF instance
    page := gdf.NewPage(gdf.A4, gdf.NoMargins) // start a new page
    cs := page.ContentStream() // content is drawn to a page's content stream
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
    pdf.WriteTo(f) // write the PDF to out.pdf

```
## Units
The default basic unit for a PDF document is the point, defined as 1/72 of an inch. However, text can be measured in terms of both points and unscaled font units. The font size (in points) indicates the number of points per side of a glyph's em square. PDF fonts always contain 1000 font units per em square, so a conversion from font units to points can be obtained by calculating `fontSize*numFontUnits/1000`, or by using the `FUToPt` or `PtToFU` functions. The `CharSpace` and `WordSpace` elements of a `ContentStream`'s `TextState` are defined in font units.

## Raster Images
In general, raster images displayed within a PDF document can be thought of as having two parts: a header, containing information about the image's size and encoding characteristics, and a byte slice representing the image's RGB/Gray/CMYK pixels in scanline order. (Alpha channel values must be encoded in a separate grayscale image.) Lossless compression filters can be applied to the byte slice to reduce its size, but this is can be costly. Where possible, it is best to store images as pre-compressed XImage objects. As a notable exception, most JPEG images can be embedded in a PDF without the need to decode and re-encode them.

## Fonts and Text Encoding
There are many ways a font can exist in a PDF file, but gdf allows for just one. In it's current form, gdf supports only TrueType/OpenType/WOFF typefaces with *nonsymbolic* characters. To render any text to a page, you must load a supported font using either the `LoadTrueType` function or the `LoadTrueTypeFile` function. Despite their names, these functions can also be used for OpenType and WOFF fonts. In PDF documents, the font used to render a piece of text determines the character encoding of that text. That is, PDF documents do not have a necessarily uniform character encoding; instead a PDF document can be a patchwork of different, even custom encodings, each of which must be specified on a per-font basis. All text written to a PDF file by gdf is encoded using the Windows-1252 ("WinAnsiEncoding") code page. This covers nearly all English-language use cases, but it is, of course, less than ideal, and hopefully, temporary. Users should be aware that any text that contains characters not included in the Windows-1252 character set will not be rendered as intended.

The PDF 2.0 spec requires fonts to be embedded in any PDF file that uses them. Font subsetting can help avoid bloated output file sizes and is strongly recommended. Subsetting functions can be set on a per-font basis. By default, gdf uses the `font.TTFSubset` function to subset embedded fonts, but this has known issues with WOFF fonts. If the usage of CGO is acceptable for your application, the `font.HBSubsetC` function is best. The `font.HBSubset` function, which can be used as a replacement, is usually preferable to `font.TTFSubset`, but it requires the user to install the [HarfBuzz hb-subset tool](https://github.com/harfbuzz/harfbuzz/tree/main) and it won't work on Windows (though it should be easy enough to tweak the source code so that it does).

## Text Formatting
While text can be drawn directly to a `ContentStream` by calling methods like `ContentStream.ShowString()`, the `text.Controller` type implements line-breaking and text-shaping algorithms, and simplifies text formatting by offering an easier to use API.

## Annotations and AcroForms
Annotations are objects, rendered by the PDF viewer on a page, that are not part of the `Page`'s `ContentStream`. gdf supports two kinds of annotation: `TextAnnot`s and `Widget`s. To a far greater extent than the graphics objects controlled by the `ContentStream`, the visual appearance of an annotation depends on the PDF viewing software.

`Widget` annotations are the visual representations of an AcroForm field, and must be paired with an `AcroField` object. gdf supports only a subset of AcroForm capabilities. Whereas the PDF specification describes AcroForms as similar to HTML forms, which are intended to be "submitted" and to trigger an action on submission, the facilities provided by gdf allow only for the user to manipulate the `Widget`'s state without submitting the form and/or triggering an action.

## Roadmap
1. ~~Provide support for embedding JPEG and PNG images.~~
2. ~~Write a tool for converting SVGs to XObjects.~~ (In progress.)
3. Improve the text formatting interface.
4. More stuff I haven't thought of.