package svg

import (
	"encoding/xml"
	"strconv"

	"github.com/cdillond/gdf"
)

type circle struct {
	cx, cy string
	r      string
	style
}

func (c circle) AddChild(e element) {}
func (c circle) Category() category { return CAT_GRAPHICAL }
func (c circle) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if c.fillOpacity.isSet || c.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if c.fillOpacity.isSet {
		cs.SetAlphaConst(c.fillOpacity.val, false)
	}
	if c.strokeOpacity.isSet {
		cs.SetAlphaConst(c.strokeOpacity.val, true)
	}
	if !c.fill.isNone {
		if c.fill.RGBColor != cs.NColor {
			cs.SetColor(c.fill.RGBColor)
		}
	}
	if !c.stroke.isNone {
		if c.stroke.RGBColor != cs.SColor {
			cs.SetColorStroke(c.stroke.RGBColor)
		}
		if c.swidth != cs.LineWidth {
			if c.swidth == 0 {
				c.swidth = 1
			}
			cs.SetLineWidth(c.swidth)
		}
	}
	cx, _ := strconv.ParseFloat(c.cx, 64)
	cy, _ := strconv.ParseFloat(c.cy, 64)
	cx, cy = tf(cx+c.xOff, cy+c.yOff, h, c.Matrix)
	r, _ := strconv.ParseFloat(c.r, 64)
	cs.Circle(cx, cy, r)
	Paint(cs, true, c.style)
}
func (c circle) Children() []element { return nil }
func (c circle) Style() style        { return c.style }
func (c circle) Inherit(s style) element {
	// inherit properties
	if !c.fill.isSet {
		c.fill = s.fill
	}
	if !c.stroke.isSet {
		c.stroke = s.stroke
	}
	if !c.fillOpacity.isSet {
		c.fillOpacity = s.fillOpacity
	}
	if !c.strokeOpacity.isSet {
		c.strokeOpacity = s.strokeOpacity
	}
	if c.fr == unset {
		c.fr = s.fr
	}
	if c.swidth == 0 {
		c.swidth = s.swidth
	}

	// inherit transformation
	c.Matrix = gdf.Mul(c.Matrix, s.Matrix)
	return &c
}
func (c *circle) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "cx":
			c.cx = a.Value
		case "cy":
			c.cy = a.Value
		case "r":
			c.r = a.Value
		}
	}
}

type ellipse struct {
	cx, cy string
	rx, ry string
	style
}

func (el ellipse) AddChild(e element) {}
func (e ellipse) Category() category  { return CAT_GRAPHICAL }
func (e ellipse) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if e.fillOpacity.isSet || e.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if e.fillOpacity.isSet {
		cs.SetAlphaConst(e.fillOpacity.val, false)
	}
	if e.strokeOpacity.isSet {
		cs.SetAlphaConst(e.strokeOpacity.val, true)
	}
	if !e.fill.isNone {
		if e.fill.RGBColor != cs.NColor {
			cs.SetColor(e.fill.RGBColor)
		}
	}
	if !e.stroke.isNone {
		if e.stroke.RGBColor != cs.SColor {
			// will work correctly with gdf v >= 0.1.10
			cs.SetColorStroke(e.stroke.RGBColor)
		}
	}
	if e.swidth != cs.LineWidth {
		if e.swidth == 0 {
			e.swidth = 1
		}
		cs.SetLineWidth(e.swidth)
	}

	cx, _ := strconv.ParseFloat(e.cx, 64)
	cy, _ := strconv.ParseFloat(e.cy, 64)
	rx, _ := strconv.ParseFloat(e.rx, 64)
	ry, _ := strconv.ParseFloat(e.ry, 64)
	cx, cy = tf(cx+e.xOff, cy+e.yOff, h, e.Matrix)

	cs.Ellipse(cx, cy, rx, ry)
	Paint(cs, true, e.style)
}
func (e ellipse) Children() []element { return nil }
func (e ellipse) Style() style        { return e.style }
func (e ellipse) Inherit(s style) element {
	// inherit properties
	if !e.fill.isSet {
		e.fill = s.fill
	}
	if !e.stroke.isSet {
		e.stroke = s.stroke
	}
	if !e.fillOpacity.isSet {
		e.fillOpacity = s.fillOpacity
	}
	if !e.strokeOpacity.isSet {
		e.strokeOpacity = s.strokeOpacity
	}
	if e.fr == unset {
		e.fr = s.fr
	}
	if e.swidth == 0 {
		e.swidth = s.swidth
	}

	// inherit transformation
	e.Matrix = gdf.Mul(e.Matrix, s.Matrix)
	return &e
}
func (e *ellipse) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "cx":
			e.cx = a.Value
		case "cy":
			e.cy = a.Value
		case "rx":
			e.rx = a.Value
		case "ry":
			e.ry = a.Value
		}
	}
}

type line struct {
	x1, y1 string
	x2, y2 string
	style
}

func (l line) AddChild(e element) {}
func (l line) Category() category { return CAT_GRAPHICAL }
func (l line) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if l.fillOpacity.isSet || l.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if l.fillOpacity.isSet {
		cs.SetAlphaConst(l.fillOpacity.val, false)
	}
	if l.strokeOpacity.isSet {
		cs.SetAlphaConst(l.strokeOpacity.val, true)
	}
	if !l.stroke.isNone {
		if l.stroke.RGBColor != cs.SColor {
			// will work correctly with gdf v >= 0.1.10
			cs.SetColorStroke(l.stroke.RGBColor)
		}
	}
	if l.swidth != cs.LineWidth {
		if l.swidth == 0 {
			l.swidth = 1
		}
		cs.SetLineWidth(l.swidth)
	}
	x1, _ := strconv.ParseFloat(l.x1, 64)
	y1, _ := strconv.ParseFloat(l.y1, 64)
	x2, _ := strconv.ParseFloat(l.x2, 64)
	y2, _ := strconv.ParseFloat(l.y2, 64)

	x1, y1 = tf(x1+l.xOff, y1+l.yOff, h, l.Matrix)
	x2, y2 = tf(x2+l.xOff, y2+l.yOff, h, l.Matrix)
	cs.MoveTo(x1, y1)
	cs.LineTo(x2, y2)
	Paint(cs, false, l.style)
}
func (l line) Children() []element { return nil }
func (l line) Style() style        { return l.style }
func (l line) Inherit(s style) element {
	// inherit properties
	if !l.stroke.isSet {
		l.stroke = s.stroke
	}
	if l.swidth == 0 {
		l.swidth = s.swidth
	}
	if !l.strokeOpacity.isSet {
		l.strokeOpacity = s.strokeOpacity
	}
	l.style.fill.isNone = true
	// inherit transformation
	l.Matrix = gdf.Mul(l.Matrix, s.Matrix)
	return &l
}
func (l *line) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "x1":
			l.x1 = a.Value
		case "y1":
			l.y1 = a.Value
		case "x2":
			l.x2 = a.Value
		case "y2":
			l.y2 = a.Value
		}
	}
}

type path struct {
	d string
	style
}

func (p path) AddChild(e element) {}
func (p path) Category() category { return CAT_GRAPHICAL }
func (p path) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if p.fillOpacity.isSet || p.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if p.fillOpacity.isSet {
		cs.SetAlphaConst(p.fillOpacity.val, false)
	}
	if p.strokeOpacity.isSet {
		cs.SetAlphaConst(p.strokeOpacity.val, true)
	}
	if !p.fill.isNone {
		if p.fill.RGBColor != cs.NColor {
			cs.SetColor(p.fill.RGBColor)
		}
	}
	if !p.stroke.isNone {
		if p.stroke.isSet && p.stroke.RGBColor != cs.SColor {
			// will work correctly with gdf v >= 0.1.10
			cs.SetColorStroke(p.stroke.RGBColor)
		}
	}
	if p.swidth != cs.LineWidth {
		if p.swidth == 0 {
			p.swidth = 1
		}
		cs.SetLineWidth(p.swidth)
	}
	parsePath(cs, p.style, p.d, h, p.Matrix)

}
func (p path) Children() []element { return nil }
func (p path) Style() style        { return p.style }
func (p path) Inherit(s style) element {

	// inherit properties
	if !p.fill.isSet {
		p.fill = s.fill
	}
	if !p.stroke.isSet {
		p.stroke = s.stroke
	}
	if !p.fillOpacity.isSet {
		p.fillOpacity = s.fillOpacity
	}
	if !p.strokeOpacity.isSet {
		p.strokeOpacity = s.strokeOpacity
	}
	if p.fr == unset {
		p.fr = s.fr
	}
	if p.swidth == 0 {
		p.swidth = s.swidth
	}

	// inherit transformation
	p.Matrix = gdf.Mul(p.Matrix, s.Matrix)
	return &p
}
func (p *path) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		if a.Name.Local == "d" {
			p.d = a.Value
		}
	}
}

type polygon struct {
	points string
	style
}

func (p polygon) AddChild(e element) {}
func (p polygon) Category() category { return CAT_GRAPHICAL }
func (p polygon) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if p.fillOpacity.isSet || p.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if p.fillOpacity.isSet {
		cs.SetAlphaConst(p.fillOpacity.val, false)
	}
	if p.strokeOpacity.isSet {
		cs.SetAlphaConst(p.strokeOpacity.val, true)
	}
	if !p.fill.isNone {
		if p.fill.RGBColor != cs.NColor {
			cs.SetColor(p.fill.RGBColor)
		}
	}
	if !p.stroke.isNone {
		if p.stroke.RGBColor != cs.SColor {
			// will work correctly with gdf v >= 0.1.10
			cs.SetColorStroke(p.stroke.RGBColor)
		}
	}

	if p.swidth != cs.LineWidth {
		if p.swidth == 0 {
			p.swidth = 1
		}
		cs.SetLineWidth(p.swidth)
	}
	var b buf
	b.b = []byte(p.points)
	var points []float64
	for fs := b.ConsumeNumber(); fs != ""; fs = b.ConsumeNumber() {
		f64, err := strconv.ParseFloat(fs, 64)
		if err != nil {
			break
		}
		points = append(points, f64)
	}
	if len(points) > 2 && len(points)%2 == 0 {
		x, y := tf(points[0]+p.xOff, points[1]+p.yOff, h, p.Matrix)
		cs.MoveTo(x, y)
		for i := 2; i < len(points); i += 2 {
			x, y = tf(points[i], points[i+1], h, p.Matrix)
			cs.LineTo(x, y)
		}
	}
	Paint(cs, true, p.style)

}
func (p polygon) Children() []element { return nil }
func (p polygon) Style() style        { return p.style }
func (p polygon) Inherit(s style) element {
	// inherit properties
	if !p.fill.isSet {
		p.fill = s.fill
	}
	if !p.stroke.isSet {
		p.stroke = s.stroke
	}
	if !p.fillOpacity.isSet {
		p.fillOpacity = s.fillOpacity
	}
	if !p.strokeOpacity.isSet {
		p.strokeOpacity = s.strokeOpacity
	}
	if p.fr == unset {
		p.fr = s.fr
	}
	if p.swidth == 0 {
		p.swidth = s.swidth
	}
	// inherit transformation
	p.Matrix = gdf.Mul(p.Matrix, s.Matrix)
	return &p
}
func (p *polygon) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		if a.Name.Local == "points" {
			p.points = a.Value
		}
	}
}

type polyline struct {
	points string
	style
}

func (p polyline) AddChild(e element) {}
func (p polyline) Category() category { return CAT_GRAPHICAL }
func (p polyline) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if p.fillOpacity.isSet || p.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if p.fillOpacity.isSet {
		cs.SetAlphaConst(p.fillOpacity.val, false)
	}
	if p.strokeOpacity.isSet {
		cs.SetAlphaConst(p.strokeOpacity.val, true)
	}
	if !p.fill.isNone {
		if p.fill.RGBColor != cs.NColor {
			cs.SetColor(p.fill.RGBColor)
		}
	}
	if !p.stroke.isNone {
		if p.stroke.RGBColor != cs.SColor {
			// will work correctly with gdf v >= 0.1.10
			cs.SetColorStroke(p.stroke.RGBColor)
		}
	}
	if p.swidth != cs.LineWidth {
		if p.swidth == 0 {
			p.swidth = 1
		}
		cs.SetLineWidth(p.swidth)
	}
	// now we actually have to do the parsing...
	// TO DO //
	var b buf
	b.b = []byte(p.points)
	var points []float64
	for fs := b.ConsumeNumber(); fs != ""; fs = b.ConsumeNumber() {
		f64, err := strconv.ParseFloat(fs, 64)
		if err != nil {
			break
		}
		points = append(points, f64)
	}
	if len(points) > 2 && len(points)%2 == 0 {
		x, y := tf(points[0]+p.xOff, points[1]+p.yOff, h, p.Matrix)
		cs.MoveTo(x, y)
		for i := 2; i < len(points); i += 2 {
			x, y = tf(points[i], points[i+1], h, p.Matrix)
			cs.LineTo(x, y)
		}
	}
	Paint(cs, false, p.style)
}
func (p polyline) Children() []element { return nil }
func (p polyline) Style() style        { return p.style }
func (p polyline) Inherit(s style) element {
	// inherit properties
	if !p.fill.isSet {
		p.fill = s.fill
	}
	if !p.stroke.isSet {
		p.stroke = s.stroke
	}
	if !p.stroke.isSet {
		p.stroke = s.stroke
	}
	if !p.fillOpacity.isSet {
		p.fillOpacity = s.fillOpacity
	}
	if p.fr == unset {
		p.fr = s.fr
	}
	if p.swidth == 0 {
		p.swidth = s.swidth
	}

	// inherit transformation
	p.Matrix = gdf.Mul(p.Matrix, s.Matrix)
	return &p
}
func (p *polyline) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		if a.Name.Local == "points" {
			p.points = a.Value
		}
	}
}

type rect struct {
	x, y          string
	width, height string
	// rx, ry     unsupported
	style
}

func (r rect) AddChild(e element) {}
func (r rect) Category() category { return CAT_GRAPHICAL }
func (r rect) Draw(cs *gdf.ContentStream, h float64) {
	// check opacity...
	if r.fillOpacity.isSet || r.strokeOpacity.isSet {
		cs.QSave()
		defer cs.QRestore()
	}
	if r.fillOpacity.isSet {
		cs.SetAlphaConst(r.fillOpacity.val, false)
	}
	if r.strokeOpacity.isSet {
		cs.SetAlphaConst(r.strokeOpacity.val, true)
	}
	// set values
	if !r.fill.isNone {
		if r.fill.RGBColor != cs.NColor {
			cs.SetColor(r.fill.RGBColor)
		}
	}
	if !r.stroke.isNone {
		if r.stroke.RGBColor != cs.SColor {
			cs.SetColorStroke(r.stroke.RGBColor)
		}
	}
	if r.swidth != cs.LineWidth {
		if r.swidth == 0 {
			r.swidth = 1
		}
		cs.SetLineWidth(r.swidth)
	}
	// svg rectangles are drawn top down rather than bottom up,
	// even when considering the difference in y coordinates between pdf and svg.
	// draw content
	x, _ := strconv.ParseFloat(r.x, 64)
	y, _ := strconv.ParseFloat(r.y, 64)
	width, _ := strconv.ParseFloat(r.width, 64)
	height, _ := strconv.ParseFloat(r.height, 64)
	y += height
	x, y = tf(x+r.xOff, y+r.yOff, h, r.Matrix)
	cs.Re(x, y, width, height)
	Paint(cs, false, r.style) // path is already closed

}
func (r rect) Children() []element { return nil }
func (r rect) Style() style        { return r.style }
func (r rect) Inherit(s style) element {
	// inherit properties
	if !r.fill.isSet {
		r.fill = s.fill
	}
	if !r.stroke.isSet {
		r.stroke = s.stroke
	}
	if !r.stroke.isSet {
		r.stroke = s.stroke
	}
	if !r.fillOpacity.isSet {
		r.fillOpacity = s.fillOpacity
	}
	if r.fr == unset {
		r.fr = s.fr
	}
	if r.swidth == 0 {
		r.swidth = s.swidth
	}
	// inherit transformation
	r.Matrix = gdf.Mul(r.Matrix, s.Matrix)
	return &r
}
func (r *rect) Parse(attrs []xml.Attr) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "x":
			r.x = a.Value
		case "y":
			r.y = a.Value
		case "width":
			r.width = a.Value
		case "height":
			r.height = a.Value
		}
	}
}
