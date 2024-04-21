Package svg provides extremely limited, experimental facilities for rendering SVG images to PDFs. In addition to the obvious constraints (such as lack of animation), this package does not implement several important SVG features. Here's a rundown of some of them:
1. There is limited support for CSS properties.
2. There is no support for text elements; unplanned.
3. Elliptical Arc Curve (`A` and `a`) path commands are not currently supported, but they may be soon.
4. Likewise, the present solution for displaying ellipse elements needs substantial improvement.
5. Mask elements and transparency/opacity-related attributes can't yet be represented.
   
Some of these limitations preclude the use of this package, but it can work with a surprising number of basic SVG images, and simple manual adjustments to the SVG's source text can often fix rendering issues.
