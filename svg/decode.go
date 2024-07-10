package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/cdillond/gdf"
)

type fr uint

const (
	unset fr = iota
	nz
	eo
)

func (f fr) toFR() gdf.FillRule {
	if f < eo {
		return gdf.NonZero
	}
	return gdf.EvenOdd
}

type cl struct {
	isSet, isNone bool
	gdf.RGBColor
}

func tf(x, y, h float64, m1 gdf.Matrix) (float64, float64) {
	pt := gdf.Transform(gdf.Point{x, y}, m1)
	return pt.X, h - pt.Y
}

type category uint

const (
	CAT_SVG = iota
	CAT_STRUCTRUAL
	CAT_GRAPHICAL
	CAT_UNDEFINED
)

type color [3]float64

var (
	COL_BLACK = color{1, 1, 1}
	COL_WHITE = color{0, 0, 0}
	COL_NONE  = color{-1, -1, -1}
)

type fillRule bool

const (
	FR_NZ fillRule = false
	FR_EO fillRule = true
)

type transform [3][2]float64

type opacity struct {
	isSet bool
	val   float64
}

type style struct {
	gdf.Matrix
	fr
	fill, stroke  cl
	fillOpacity   opacity
	strokeOpacity opacity
	swidth        float64
	id, use       string
	xOff, yOff    float64
}

// If the element is a structural element, it returns a non-nil slice of child elements.
// If the element is a graphical element, it returns a nil slice of child elements.
type element interface {
	Category() category
	Draw(*gdf.ContentStream, float64)
	Children() []element
	AddChild(element)
	Style() style
	Inherit(style) element // apply parent style
}

func ParseStyleID(attrs []xml.Attr) (out style, id string) {
	out.Matrix = gdf.NewMatrix()
	for _, a := range attrs {
		switch a.Name.Local {
		case "transform":
			ts := parseTransform(a.Value)
			for i := range ts {
				out.Matrix = gdf.Mul(out.Matrix, ts[i])
			}
		case "style":
			out.parseCSS(a.Value)
		case "id":
			id = a.Value
		case "fill":
			out.fill, out.fillOpacity = parseColor(a.Value)
		case "fill-rule":
			if a.Name.Local == "evenodd" {
				out.fr = eo
			} else if a.Name.Local == "nonzero" {
				out.fr = nz
			}
		case "fill-opacity":
			fo, err := parseNumPct(a.Value)
			if err == nil {
				out.fillOpacity.val = fo
				out.fillOpacity.isSet = true
			}
		case "stroke":
			out.stroke, out.strokeOpacity = parseColor(a.Value)
		case "stroke-width":
			sw, _ := strconv.ParseFloat(a.Value, 64)
			out.swidth = sw
		case "stroke-opacity":
			so, err := parseNumPct(a.Value)
			if err == nil {
				out.strokeOpacity.val = so
				out.strokeOpacity.isSet = true
			}
		default:
		}
	}
	return out, id
}

func Paint(cs *gdf.ContentStream, closePath bool, s style) {
	if closePath {
		if !s.stroke.isNone && s.stroke.isSet {
			if s.fill.isNone {
				cs.ClosePathStroke()
				return
			}
			cs.ClosePathFillStroke(s.fr.toFR())
			return
		}
		if !s.fill.isNone {
			cs.Fill(s.fr.toFR())
			return
		}
		cs.ClosePath()
	} else {
		// I think technically filling operations are also implicitly close operations, but whatever...
		if !s.stroke.isNone && s.stroke.isSet {
			if s.fill.isNone {
				cs.Stroke()
				return
			}
			cs.FillStroke(s.fr.toFR())
			return
		}
		if !s.fill.isNone {
			cs.Fill(s.fr.toFR())
			return
		}
		cs.EndPath()
	}
}

func Render(root element, cs *gdf.ContentStream, h float64) {
	//fmt.Println(reflect.TypeOf(root), root)
	switch root.Category() {
	case CAT_STRUCTRUAL:
		for _, child := range root.Children() {
			child = child.Inherit(root.Style())
			Render(child, cs, h)
		}
	case CAT_GRAPHICAL:
		root.Draw(cs, h)
	}
}

var defmap = make(map[string]element)

// Decode reads from r, which contains the SVG source data, and returns a gdf.XContent representation of the SVG and an error.
func Decode(r io.Reader) (gdf.XContent, error) {
	root := new(node)
	dec := xml.NewDecoder(r)
	xc := gdf.NewXContent(nil, gdf.Rect{})
	var h, w float64
	tok, err := dec.Token()
	for ; err == nil; tok, err = dec.Token() {
		switch val := tok.(type) {
		case xml.StartElement:
			style, id := ParseStyleID(val.Attr)
			switch val.Name.Local {
			case "svg":
				sv := new(svg)
				sv.Matrix = gdf.NewMatrix()
				vb, empty := [4]float64{}, [4]float64{}
				var H, W float64
				for _, a := range val.Attr {
					if a.Name.Local == "viewBox" {
						fmt.Sscanf(a.Value, "%f %f %f %f", &vb[0], &vb[1], &vb[2], &vb[3])
						h = vb[3] - vb[1]
						w = vb[2] - vb[0]
					} else if a.Name.Local == "height" {
						H, _ = parseAbsoluteLength(a.Value)
					} else if a.Name.Local == "width" {
						W, _ = parseAbsoluteLength(a.Value)
					}
				}
				if vb == empty {
					h = H
					w = W
				}
				root.element = sv

			case "defs":
				next := new(node)
				defs := new(defs)
				defs.Parse(val.Attr)
				defs.style = style
				next.element = defs
				next.parent = root
				root.element.AddChild(next.element)
				root = next
			case "g":
				next := new(node)
				gp := new(group)
				gp.Parse(val.Attr)
				gp.style = style
				next.element = gp
				next.parent = root
				if gp.style.id != "" {
					defmap[gp.style.id] = next
				}
				root.element.AddChild(next.element)
				root = next
			case "symbol":
				next := new(node)
				sym := new(symbol)
				sym.Parse(val.Attr)
				sym.style = style
				next.element = sym
				next.parent = root
				if sym.style.id != "" {
					defmap[sym.style.id] = next
				}
				root.element.AddChild(next.element)
				root = next
			case "use":
				next := new(node)
				use := new(use)
				use.Parse(val.Attr)
				use.style = style
				next.element = use
				next.parent = root
				if use.style.id != "" {
					defmap[use.style.id] = next
				}
				root.element.AddChild(next.element)
				root = next
			case "circle":
				c := new(circle)
				c.Parse(val.Attr)
				c.style = style
				if id != "" {
					fmt.Println("hey", id)
					defmap[id] = c
					fmt.Println(defmap)
				}
				root.element.AddChild(c)
			case "ellipse":
				e := new(ellipse)
				e.Parse(val.Attr)
				e.style = style
				if id != "" {
					defmap[id] = e
				}
				root.element.AddChild(e)
			case "line":
				l := new(line)
				l.Parse(val.Attr)
				l.style = style
				if id != "" {
					defmap[id] = l
				}
				root.element.AddChild(l)
			case "path":
				p := new(path)
				p.Parse(val.Attr)
				p.style = style
				if id != "" {
					defmap[id] = p
				}
				root.element.AddChild(p)
			case "polygon":
				p := new(polygon)
				p.Parse(val.Attr)
				p.style = style
				if id != "" {
					defmap[id] = p
				}
				root.element.AddChild(p)
			case "polyline":
				p := new(polyline)
				p.Parse(val.Attr)
				p.style = style
				if p.style.id != "" {
					defmap[p.style.id] = p
				}
				root.element.AddChild(p)
			case "rect":
				r := new(rect)
				r.Parse(val.Attr)
				r.style = style
				if r.style.id != "" {
					defmap[r.style.id] = r
				}
				root.element.AddChild(r)
			default:
				//fmt.Println(val.Name)
			}

		case xml.EndElement:
			switch val.Name.Local {
			case "svg", "defs", "g", "symbol", "use":
				if root.parent == nil {
					break
				}
				root = root.parent
			}
		}
	}
	Render(root.element, &xc.ContentStream, h)
	xc.BBox = gdf.Rect{0, 0, w, h}

	if err != io.EOF {
		return *xc, err
	}
	return *xc, nil
}
