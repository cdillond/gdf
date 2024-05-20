package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cdillond/gdf"
)

func unmarshalXML(n *node, r io.Reader, defs map[string]*node) {
	if n == nil {
		return
	}
	dec := xml.NewDecoder(r)
	tok, err := dec.Token()
	for {
		// loop until the outermost svg element is found; fragments are not accepted.
		if v, ok := tok.(xml.StartElement); ok {
			if v.Name.Local == "svg" {
				break
			}
		}
		tok, err = dec.Token()
	}
	for ; err == nil; tok, err = dec.Token() {
		switch v := tok.(type) {
		case xml.StartElement:
			next := new(node)
			next.parent = n
			err := unmarshalAttrs(next, v.Attr)
			if err != nil {
				fmt.Println(err.Error())
			}
			next.transforms = append(next.transforms, next.parent.transforms...)
			if next.self.transform != nil {
				next.transforms = append(next.transforms, next.self.transform...)
			}
			next.k = toKind(v.Name.Local)
			if next.k == styleKind {
				tok, _ = dec.Token()
				if styleStr, ok := tok.(xml.CharData); ok {
					unmarshalCSS(string(styleStr))
				}
			}
			if next.self.id != nil {
				defs[*next.self.id] = next
			}
			next.inherited = merge(n.inherited, n.self)
			next.parent = n
			n.children = append(n.children, next)
			n = next
		case xml.EndElement:
			n = n.parent
		}
	}
}

// Most Attrs can be parsed easily, but "d" and "points" attributes are more complex.
func unmarshalAttrs(n *node, x []xml.Attr) error {
	for _, v := range x {
		if v.Value == "" {
			continue
		}
		switch v.Name.Local {
		case "class":
			a, ok := classes[v.Value]
			if ok {
				n.self = merge(a, n.self)
			}
		case "clip-rule":
			if v.Value == "evenodd" {
				cr := gdf.EvenOdd
				n.self.clipRule = &cr
			} else {
				cr := gdf.NonZero
				n.self.clipRule = &cr
			}
		case "color":
			c, ok := parseColor(v.Value)
			if ok {
				n.self.color = &c
			}
		case "cx":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.cx = &f64
		case "cy":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.cy = &f64
		case "d":
			rawCmds, err := run([]byte(v.Value), lexSVGPathOp, parseDAttr)
			n.tmpCmds = rawCmds
			if err != nil {
				return err
			}
		case "dx":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.dx = &f64
		case "dy":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.dy = &f64
		case "fill":
			c, ok := parseColor(v.Value)
			if ok {
				if n.k == maskKind {
					n.self.maskFill = &c
				} else {
					n.self.fill = &c
				}
			}
		case "fillRule":
			if v.Value == "evenodd" {
				fr := gdf.EvenOdd
				n.self.fillRule = &fr
			} else {
				fr := gdf.NonZero
				n.self.fillRule = &fr
			}
		case "height":
			f64, _ := parseAbsoluteLength(v.Value)
			n.self.height = &f64
		case "href":
			n.self.href = &v.Value
		case "id":
			n.self.id = &v.Value
		case "mask":
			s := parseMaskURL(v.Value)
			n.self.mask = &s
		case "points":
			rawCmds, err := run([]byte(v.Value), lexPolyNum, parsePointsAttr)
			if err != nil {
				return err
			}
			n.tmpCmds = rawCmds
		case "r":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.r = &f64
		case "rx":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.rx = &f64
		case "ry":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.ry = &f64
		case "stroke":
			c, ok := parseColor(v.Value)
			if ok {
				n.self.stroke = &c
			}
		case "strokeLineCap":
			n.self.strokeLineCap = new(gdf.LineCap)
		case "strokeLineJoin":
			n.self.strokeLineJoin = new(gdf.LineJoin)
		case "strokeWidth":
			f64, _ := strconv.ParseFloat(v.Value, 64)
			n.self.strokeWidth = &f64
		//case "style":

		case "transform":
			n.self.transform = parseTransform(v.Value)
		case "viewBox":
			s := strings.Fields(v.Value)
			if len(s) != 4 {
				continue
			}
			var nums [4]float64
			for i := range s {
				f64, _ := strconv.ParseFloat(s[i], 64)
				nums[i] = f64
			}
			n.self.viewBox = &nums
		case "width":
			f64, _ := parseAbsoluteLength(v.Value)
			n.self.width = &f64
		case "x":
			f64, _ := parseAbsoluteLength(v.Value)
			n.self.x = &f64
		case "y":
			f64, _ := parseAbsoluteLength(v.Value)
			n.self.y = &f64
		}
	}
	return nil
}
