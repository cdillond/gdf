package svg

import (
	"encoding/xml"

	"github.com/cdillond/gdf"
)

type node struct {
	element
	parent *node
}

type svg struct {
	children []element
	style
}

func (s *svg) AddChild(e element) { s.children = append(s.children, e) }

func (s svg) Category() category                    { return CAT_STRUCTRUAL }
func (s svg) Draw(cs *gdf.ContentStream, h float64) {}
func (s svg) Children() []element                   { return s.children }
func (s svg) Style() style                          { return s.style }
func (sv svg) Inherit(s style) element {
	// inherit properties
	if !sv.fill.isSet {
		sv.fill = s.fill
	}
	if !sv.stroke.isSet {
		sv.stroke = s.stroke
	}
	if !sv.stroke.isSet {
		sv.stroke = s.stroke
	}
	if !sv.fillOpacity.isSet {
		sv.fillOpacity = s.fillOpacity
	}
	if sv.fr == unset {
		sv.fr = s.fr
	}
	if sv.swidth == 0 {
		sv.swidth = s.swidth
	}
	// inherit transformation
	sv.Matrix = gdf.Mul(sv.Matrix, s.Matrix)
	return &sv
}
func (s *svg) Parse(attrs []xml.Attr) {}

type defs struct {
	children []element
	style
}

func (d *defs) AddChild(e element)                   { d.children = append(d.children, e) }
func (d defs) Category() category                    { return CAT_UNDEFINED }
func (d defs) Draw(cs *gdf.ContentStream, h float64) {}
func (d defs) Children() []element                   { return d.children }
func (d defs) Style() style                          { return d.style }
func (d defs) Inherit(s style) element {
	// inherit properties
	if !d.fill.isSet {
		d.fill = s.fill
	}
	if !d.stroke.isSet {
		d.stroke = s.stroke
	}
	if !d.stroke.isSet {
		d.stroke = s.stroke
	}
	if !d.fillOpacity.isSet {
		d.fillOpacity = s.fillOpacity
	}
	if d.fr == unset {
		d.fr = s.fr
	}
	if d.swidth == 0 {
		d.swidth = s.swidth
	}
	// inherit transformation
	d.Matrix = gdf.Mul(d.Matrix, s.Matrix)
	return &d
}
func (d *defs) Parse(attrs []xml.Attr) {}

type group struct {
	children []element
	use      string
	style
}

func (g *group) AddChild(e element) { g.children = append(g.children, e) }

func (g group) Category() category                    { return CAT_STRUCTRUAL }
func (g group) Draw(cs *gdf.ContentStream, h float64) {}
func (g group) Children() []element                   { return g.children }
func (g group) Style() style                          { return g.style }
func (g group) Inherit(s style) element {
	// inherit properties
	if !g.fill.isSet {
		g.fill = s.fill
	}
	if !g.stroke.isSet {
		g.stroke = s.stroke
	}
	if !g.stroke.isSet {
		g.stroke = s.stroke
	}
	if !g.fillOpacity.isSet {
		g.fillOpacity = s.fillOpacity
	}
	if g.fr == unset {
		g.fr = s.fr
	}
	if g.swidth == 0 {
		g.swidth = s.swidth
	}
	// inherit transformation
	g.Matrix = gdf.Mul(s.Matrix, g.Matrix)
	return &g
}
func (g *group) Parse(attrs []xml.Attr) {}

type symbol struct {
	children []element
	use      string
	style
}

func (s *symbol) AddChild(e element) { s.children = append(s.children, e) }

func (s symbol) Category() category                    { return CAT_STRUCTRUAL }
func (s symbol) Draw(cs *gdf.ContentStream, h float64) {}
func (s symbol) Children() []element                   { return s.children }
func (s symbol) Style() style                          { return s.style }
func (sy symbol) Inherit(s style) element {
	// inherit properties
	if !sy.fill.isSet {
		sy.fill = s.fill
	}
	if !sy.stroke.isSet {
		sy.stroke = s.stroke
	}
	if !sy.stroke.isSet {
		sy.stroke = s.stroke
	}
	if !sy.fillOpacity.isSet {
		sy.fillOpacity = s.fillOpacity
	}
	if sy.fr == unset {
		sy.fr = s.fr
	}
	if sy.swidth == 0 {
		sy.swidth = s.swidth
	}

	// inherit transformation
	sy.Matrix = gdf.Mul(sy.Matrix, s.Matrix)
	return &sy
}
func (s *symbol) Parse(attrs []xml.Attr) {}

type use struct {
	children []element
	x, y     string
	href     string
	style
}

func (u *use) AddChild(e element) { u.children = append(u.children, e) }

func (u use) Category() category                    { return CAT_STRUCTRUAL }
func (u use) Draw(cs *gdf.ContentStream, h float64) {}
func (u use) Children() []element {
	var e element
	if u.href != "" {
		e = defmap[u.href[1:]]
	}
	if e == nil {
		return u.children
	}
	switch v := e.(type) {
	case *circle:
		c := *v
		c.style.xOff = pf(u.x)
		c.style.yOff = pf(u.y)
		return append(u.children, c)
	case *ellipse:
		r := *v
		r.style.xOff = pf(u.x)
		r.style.yOff = pf(u.y)
		return append(u.children, r)
	case *line:
		l := *v
		l.style.xOff = pf(u.x)
		l.style.yOff = pf(u.y)
		return append(u.children, l)
	default:
		return u.children
	}

	return u.children
}
func (u use) Style() style { return u.style }
func (u use) Inherit(s style) element {
	if !u.fill.isSet {
		u.fill = s.fill
	}
	if !u.stroke.isSet {
		u.stroke = s.stroke
	}
	if !u.stroke.isSet {
		u.stroke = s.stroke
	}
	if !u.fillOpacity.isSet {
		u.fillOpacity = s.fillOpacity
	}
	if u.fr == unset {
		u.fr = s.fr
	}
	if u.swidth == 0 {
		u.swidth = s.swidth
	}
	// inherit transformation
	u.Matrix = gdf.Mul(u.Matrix, s.Matrix)
	return &u
}
func (u *use) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "x":
			u.x = a.Value
		case "y":
			u.y = a.Value
		case "href":
			u.href = a.Value
		}
	}
}
