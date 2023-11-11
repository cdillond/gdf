Package gdf defines an interface for generating PDFs. It hews closely to the basics of the
PDF 2.0 specification and implements a relatively low-level API on which higher-level abstractions can
be built.

gdf should be sufficient for most basic English-language PDF use cases. It avoids complexity
by purposefully ignoring some aspects of the PDF specification. These omissions may be
critical if you need accessibility features or support for multiple languages.

Understanding the PDF coordinate system can go a long way to simplifying the use of this package.

Every item is drawn, according to its type, at the origin of its coordinate space. The
coordinate space is then transformed by one or more affine matrices, always including the
current transformation matrix, and rendered onto the page's "user space." Text space is
transformed first by the current text matrix and then by the current transformation matrix.
The PDF specification also mentions glyph space, image space, form space, and pattern space,
but they are not relevant to this package in its present form.

Transformation matrices are defined by 6 parameters representing the translation,
scale, and shear of the X and Y coordinates of a point transformed by the given matrix.
Because the space of an object can be scaled or rotated, the effect of certain operations
may be difficult to intuit. For instance, drawing a line from (10, 10) to (10, 20)
in the default graphics space moves the cursor from (10, 10) to (10, 40) in
user space if the Current Transformation Matrix is [1 0 0][2 0 0][0 0 1]; i.e.,
if it scales the y-coordinates of the original space by 2.

The default basic unit for a PDF document is the point, defined as 1/72 of an inch.
However, text can be measured in terms of both points and unscaled font units.
The font size (in points) indicates the number of points per side of a glyph's em square. PDF fonts always
contain 1000 font units per em square, so a conversion from font units to points can be
obtained by calculating fontSize*numFontUnits/1000. The Tc (character spacing) and Tw (word spacing)
elements of a PDF's Text State are defined in font units.

A final source of complexity is fonts. The original PDF specification made use of
Type1 fonts built in to all PDF rendering software. Built-in Type1 fonts are now
deprecated, but their legacy remains. There are multiple ways to embed a font in
a PDF, but they all must specify a text encoding. A page's current font and its
associated objects therefore define both the appearance of the document's glyphs
and the encoding of the source text. The default encoding is ASCII, not UTF-8 :(.
Windows-1252 ("WinAnsiEncoding") is recommended for English-language documents.