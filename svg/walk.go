package svg

import (
	"io"

	"github.com/cdillond/gdf"
)

type svgRoot struct {
	n              *node
	xContent       gdf.XContent
	Height, Width  float64
	HScale, WScale float64
	defs           map[string]*node
}

func Decode(r io.Reader) (gdf.XContent, error) {
	out := svgRoot{
		n:        new(node),
		xContent: *gdf.NewXContent(nil, gdf.Rect{}),
		HScale:   1,
		WScale:   1,
		defs:     make(map[string]*node),
	}
	out.xContent.Filter = gdf.NoFilter
	unmarshalXML(out.n, r, out.defs)
	out.n = out.n.children[0]
	var h, w float64
	if out.n.self.height != nil {
		h = *out.n.self.height
	}
	if out.n.self.width != nil {
		w = *out.n.self.width
	}
	if out.n.self.viewBox != nil {
		out.Height = px * (out.n.self.viewBox[3] - out.n.self.viewBox[1])
		out.Width = px * (out.n.self.viewBox[2] - out.n.self.viewBox[0])
	} else {
		out.Height = h
		out.Width = w
	}
	if out.Height != 0 && h != 0 {
		out.HScale = out.Height / h
	}
	if out.WScale != 0 && w != 0 {
		out.WScale = out.Width / w
	}
	out.xContent.BBox.URY = out.Height
	out.xContent.BBox.URX = out.Width
	walk(out.n, &out.xContent, out.HScale, out.WScale, out.defs)
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
		WriteCmd(&x.ContentStream, c)
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
