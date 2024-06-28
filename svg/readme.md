Package svg provides very limited, experimental facilities for rendering SVG images to PDFs. In addition to the obvious constraints (e.g., lack of animation), this package does not implement several important SVG features. Here's a rundown of some of them:
1. There is limited support for CSS properties.
2. Support for SVG text elements is unplanned.
3. Elliptical Arc Curve (`A` and `a`) path commands may be improperly rendered.
4. Mask elements and transparency/opacity-related attributes are not supported.
   
These limitations preclude the use of this package for certain applications, but it can work with a surprising number of basic SVG images. Running `rsvg-convert` with the `-f svg` option on the input SVG prior to its inclusion in the PDF is **highly** recommended. Making simple manual adjustments to the SVG's source text can also often fix rendering issues.
