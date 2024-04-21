package svg

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/cdillond/gdf"
)

const px float64 = gdf.In / 96.0

var adjY = gdf.Rotate(180 * gdf.Deg)

type kind uint

const (
	svgKind kind = iota
	aKind
	circleKind
	ellipseKind
	defsKind
	gKind
	lineKind
	pathKind
	polygonKind
	rectKind
	useKind
	styleKind
	maskKind
	badKind
)

var kinds = [...]string{"svg", "a", "circle", "ellipse", "defs", "g", "line", "path", "polygon", "rect", "use", "style", "mask"}

func isValid(k kind) bool {
	return k < badKind
}

func (k kind) String() string {
	if isValid(k) {
		return kinds[k]
	}
	return "invalid"
}

func toKind(s string) kind {
	for i := range kinds {
		if kinds[i] == s {
			return kind(i)
		}
	}
	return badKind
}

type node struct {
	k          kind
	parent     *node
	children   []*node
	self       attributes   // attributes defined for the node
	inherited  attributes   // attributes inherited from ancestor nodes
	transforms []gdf.Matrix // transforms are applied for each node
	tmpCmds    []svgCmd
}

// Pointers or pointer-types are used for all fields here in order to differentiate between
// fields with zero values and fields that have not been set.
type attributes struct {
	color          *gdf.RGBColor // fill and stroke color if neither are set
	clipRule       *gdf.FillRule
	cx, cy         *float64      // center point
	d              []pdfPathCmd  // path
	dx, dy         *float64      // x and y offset
	fill           *gdf.RGBColor // fill color
	maskFill       *gdf.RGBColor
	fillRule       *gdf.FillRule
	height         *float64
	href           *string
	id             *string
	mask           *string
	points         *string
	r              *float64 // radius
	rx, ry         *float64
	stroke         *gdf.RGBColor // stroke color
	strokeLineCap  *gdf.LineCap
	strokeLineJoin *gdf.LineJoin
	strokeWidth    *float64
	transform      []gdf.Matrix
	viewBox        *[4]float64
	width          *float64
}

var classes = map[string]attributes{}

func unmarshalCSS(s string) {
	buf := newBuffer([]byte(s))

	// get class name
	//start:
	for _, ok := buf.peek(); ok; _, ok = buf.peek() {
		buf.skipWSP()
		var a attributes
		_, ok = buf.peek()
		if !ok {
			return
		}
		buf.skip()
		cname := buf.upTo('{')
		buf.skip()
		attrs := bytes.Split(buf.upTo('}'), []byte(";"))
		for _, attr := range attrs {
			before, after, ok := strings.Cut(string(attr), ":")
			if !ok {
				continue
			}
			before = strings.TrimSpace(before)
			after = strings.TrimSpace(after)

			switch before {
			case "color":
				c, ok := parseColor(after)
				if ok {
					a.color = &c
				}
			case "clipRule":
				if after == "evenodd" {
					cr := gdf.EvenOdd
					a.clipRule = &cr
				} else {
					cr := gdf.NonZero
					a.clipRule = &cr
				}
			case "cx":
				f64, _ := strconv.ParseFloat(after, 64)
				a.cx = &f64
			case "cy":
				f64, _ := strconv.ParseFloat(after, 64)
				a.cy = &f64
			case "dx":
				f64, _ := strconv.ParseFloat(after, 64)
				a.dx = &f64
			case "dy":
				f64, _ := strconv.ParseFloat(after, 64)
				a.dy = &f64
			case "fill": // fill color
				c, ok := parseColor(string(after))
				if ok {
					a.fill = &c
				}
			case "fillRule":
				if after == "evenodd" {
					fr := gdf.EvenOdd
					a.fillRule = &fr
				} else {
					fr := gdf.NonZero
					a.fillRule = &fr
				}
			case "height":
				f64, _ := parseAbsoluteLength(after)
				a.height = &f64
			case "mask":
				s := parseMaskURL(after)
				a.mask = &s
			case "r": // radius
				f64, _ := strconv.ParseFloat(after, 64)
				a.r = &f64
			case "rx": // radius
				f64, _ := strconv.ParseFloat(after, 64)
				a.rx = &f64
			case "ry": // radius
				f64, _ := strconv.ParseFloat(after, 64)
				a.ry = &f64
			case "stroke":
				c, ok := parseColor(string(after))
				if ok {
					a.stroke = &c
				}
			//case "strokeLineCap":
			//case "strokeLineJoin":
			case "strokeWidth":
				f64, _ := strconv.ParseFloat(after, 64)
				a.strokeWidth = &f64
			case "transform":
				a.transform = parseTransform(after)
			case "viewBox":
				s := strings.Fields(after)
				if len(s) != 4 {
					continue
				}
				var nums [4]float64
				for i := range s {
					f64, _ := strconv.ParseFloat(s[i], 64)
					nums[i] = f64
				}
				a.viewBox = &nums
			case "width":
				f64, _ := parseAbsoluteLength(after)
				a.width = &f64
			}
		}
		classes[string(cname)] = a
		buf.skipWSP()
		buf.skip()
	}
	//if _, ok = buf.peek(); ok {
	//	goto start
	//}

}

func merge(parent, child attributes) attributes {
	return attributes{
		color:    inherit(parent.color, child.color),
		clipRule: inherit(parent.clipRule, child.clipRule),
		cx:       inherit(parent.cx, child.cx),
		cy:       inherit(parent.cy, child.cy),
		d:        child.d,
		dx:       inherit(parent.dx, child.dx),
		dy:       inherit(parent.dy, child.dy),
		fill:     inherit(parent.fill, child.fill),
		fillRule: inherit(parent.fillRule, child.fillRule),
		height:   inherit(parent.height, child.height),
		href:     child.href,
		//mask:           inherit(parent.mask, child.mask),
		points:         child.points,
		r:              inherit(parent.r, child.r),
		rx:             inherit(parent.rx, child.rx),
		ry:             inherit(parent.ry, child.ry),
		stroke:         inherit(parent.stroke, child.stroke),
		strokeLineCap:  inherit(parent.strokeLineCap, child.strokeLineCap),
		strokeLineJoin: inherit(parent.strokeLineJoin, child.strokeLineJoin),
		strokeWidth:    inherit(parent.strokeWidth, child.strokeWidth),
		transform:      inheritSlice(parent.transform, child.transform),
		viewBox:        inherit(parent.viewBox, child.viewBox),
	}
}

// returns b if b is not nil; returns a otherwise
func inherit[T any](parent, child *T) *T {
	if child != nil {
		return child
	}
	return parent
}

func inheritSlice(parent, child []gdf.Matrix) []gdf.Matrix {
	if child != nil {
		return child
	}
	return parent
}
