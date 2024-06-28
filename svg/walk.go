package svg

import (
	"io"

	"github.com/cdillond/gdf"
)

type svgRoot struct {
	n             *node
	xContent      gdf.XContent
	Height, Width float64
	ViewBox       gdf.Rect
	defs          map[string]*node
}

func abs(f float64) float64 {
	if f < 0 {
		f = -f
	}
	return f
}

type node_type uint

const (
	svg_type node_type = iota
	use_type
	path_type
	rect_type
	polygon_type
	circle_type
	ellipse_type
	moveto_type
	lineto_type
	curveto_type
	close_path_type
)

type svg2 struct {
	children  []svg2 // this gives us an idea of the number
	transform int
	node_type
}

func walk2(s svg2) {
	for i := range s.children {
		walk2(s.children[i])
	}
}

/*
SVGs may specify a height and width as well as a "viewBox". If no viewBox is given, it is assumed
to be coincident with the height and width. That is, the coordinates begin (at the top left of the page)
at (0, 0), and end at the bottom right of the page at (w, h). Any graphics drawn outside of this box
are not rendered. If both a viewBox and height and width parameters are provided, then the height
and width are interpreted instead as scale values. The values of an SVG with a height and width of 5px
but a viewBox of 0 0 10 10 are scaled down by 2.

OK: let's redo that and make it simpler. The height and width are the values against which relative lengths
are resolved. The viewBox determines the BoundingBox and the initial clip path. If no viewBox is provided,
the height and width are used, starting at 0, 0. If no height and width are provided, the viewBox is used.
It is in error for neither to be provided. The viewBox dimensions are always in pixels.
*/

func Decode(r io.Reader) (gdf.XContent, error) {
	out := svgRoot{
		n:        new(node),
		xContent: *gdf.NewXContent(nil, gdf.Rect{}),
		defs:     make(map[string]*node),
	}
	out.xContent.Filter = gdf.NoFilter // for now...
	// the SVG gets unmarshalled into a tree of structs
	unmarshalXML(out.n, r, out.defs)
	out.n = out.n.children[0]

	var h, w float64
	if out.n.self.height != nil {
		h = *out.n.self.height
		out.Height = h
	}
	if out.n.self.width != nil {
		w = *out.n.self.width
		out.Width = w
	}
	if out.n.self.viewBox != nil {
		out.ViewBox = gdf.Rect{
			LLX: out.n.self.viewBox[0],
			LLY: out.n.self.viewBox[1],
			URX: out.n.self.viewBox[2],
			URY: out.n.self.viewBox[3],
		}
	} else {
		out.ViewBox = gdf.NewRect(gdf.Point{0, 0}, gdf.Point{out.Width, out.Height})
	}
	if out.n.self.width == nil && out.n.self.height == nil {
		out.Width = abs(out.ViewBox.URX - out.ViewBox.LLX)
		out.Height = abs(out.ViewBox.URY - out.ViewBox.LLY)
	}

	out.xContent.BBox.URY = out.Height
	out.xContent.BBox.URX = out.Width

	walk(out.n, &out.xContent, 1, 1, out.defs)
	return out.xContent, nil
}

func walk(n *node, x *gdf.XContent, hScale, wScale float64, defs map[string]*node) {
	if n == nil || n.k == defsKind || n.k == badKind {
		return
	}

	style := merge(n.inherited, n.self)

	if style.fill != nil && *style.fill != badColor {
		if *style.fill == rgbBlack || n.inherited.fill == nil || *n.inherited.fill != *style.fill {
			x.SetColor(*style.fill)
		}
	}
	if style.stroke != nil && *style.stroke != badColor {
		if n.inherited.stroke == nil || *n.inherited.stroke != *style.stroke {
			x.SetColorStroke(*style.stroke)
		}
	}
	if style.strokeWidth != nil {
		x.SetLineWidth(*style.strokeWidth)
	}

	if n.k == useKind {
		if len(*n.self.href) > 0 {
			t1, ok := defs[(*n.self.href)[1:]]
			if ok {
				t2 := *t1
				t2.self = merge(style, t2.self)
				t2.transforms = append(t2.transforms, n.transforms...)
				walk(&t2, x, hScale, wScale, defs)
			}
			return
		}
	}

	if n.k == maskKind {
		x.QSave()
		//if n.self.maskFill != nil && *n.self.maskFill == rgbWhite {
		//fmt.Println(n.children)
		//x.SetAlphaConst(1)
		//}
		for _, c := range n.children {
			walk(c, x, hScale, wScale, defs)
		}
		x.QRestore()
		return
	}

	if n.self.mask != nil {
		t1, ok := defs[(*n.self.mask)]
		if ok {
			t2 := *t1
			t2.self = merge(style, t2.self)
			t2.transforms = append(t2.transforms, n.transforms...)
			walk(&t2, x, hScale, wScale, defs)
		}
		return
	}

	//if n.k != useKind {
	cmds := resolvePathCmds(n.tmpCmds)
	if n.k == circleKind {
		cmds = append(cmds, pdfPathCmd{
			op:   circle,
			args: []gdf.Point{{X: *n.self.cx, Y: *n.self.cy}, {X: *n.self.r, Y: 0}},
		})
	}
	if n.k == rectKind {
		cmds = append(cmds, pdfPathCmd{
			op:   rect,
			args: []gdf.Point{{X: *n.self.x, Y: *n.self.y}, {X: *n.self.width, Y: *n.self.height}},
		})
	}
	if n.k == ellipseKind {
		x.QSave()
		cmds = append(cmds, pdfPathCmd{
			op:   ellipse,
			args: []gdf.Point{{X: *n.self.cx, Y: *n.self.cy}, {X: *n.self.rx, Y: *n.self.ry}},
		})
	}
	for _, c := range cmds {
		for i := range c.args {
			c.args[i].X *= px
			c.args[i].Y *= px
			for j := range n.transforms {
				c.args[i] = gdf.Transform(c.args[i], n.transforms[j])
			}
			c.args[i].Y *= hScale
			c.args[i].Y = hScale*x.BBox.URY - c.args[i].Y
			c.args[i].X *= wScale
		}
		writeCmd(&x.ContentStream, c)
	}

	if len(cmds) > 0 {
		/*if style.clipRule != nil {
			x.Clip(*style.clipRule)
		}*/
		//if style.fill != nil {
		//	fmt.Println(*style.fill)
		//}
		if style.fill != nil && *style.fill != badColor && style.stroke != nil && *style.stroke != badColor {
			if style.fillRule != nil {
				x.FillStroke(*style.fillRule)
			} else {
				x.FillStroke(gdf.NonZero)
			}
		} else if style.fill != nil && *style.fill != badColor {
			if style.fillRule != nil {
				x.Fill(*style.fillRule)
			} else {
				x.Fill(gdf.NonZero)
			}
		} else if style.stroke != nil && *style.stroke != badColor {
			x.Stroke()
		}
	}

	if n.k == ellipseKind {
		x.QRestore()
	}

	//	}

	for _, c := range n.children {
		walk(c, x, hScale, wScale, defs)
	}

}
