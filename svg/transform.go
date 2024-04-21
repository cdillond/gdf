package svg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cdillond/gdf"
)

func parseTransform(s string) []gdf.Matrix {
	var out []gdf.Matrix
	buf := newBuffer([]byte(s))
	for buf.n < len(buf.b) {
		s := string(buf.upTo('('))
		switch s {
		case "matrix":
			// TODO
		case "translate":
			buf.skip()
			buf.skipWSP()
			s = string(buf.upTo2(',', '\x20'))
			dx, _ := strconv.ParseFloat(s, 64)
			buf.skip()
			buf.skipWSP()
			dy, _ := strconv.ParseFloat(string(buf.upTo(')')), 64)
			out = append(out, gdf.Translate(dx*px, dy*px))
		case "skewX":
			s = string(buf.upTo(')'))
			sx, _ := strconv.ParseFloat(s, 64)
			out = append(out, gdf.Skew(sx*gdf.Deg, 0))
		case "skewY":
			s = string(buf.upTo(')'))
			sy, _ := strconv.ParseFloat(s, 64)
			out = append(out, gdf.Skew(0, sy*gdf.Deg))
		case "scale":
			buf.skip()
			buf.skipWSP()
			s = string(buf.upTo2(',', '\x20'))
			scaleX, _ := strconv.ParseFloat(s, 64)
			buf.skip()
			buf.skipWSP()
			scaleY, _ := strconv.ParseFloat(string(buf.upTo(')')), 64)
			out = append(out, gdf.ScaleBy(scaleX, scaleY))
		case "rotate":
			// TODO
		}
		buf.skip()
		buf.skipWSP()
	}
	return out
}

func parseAbsoluteLength(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return 0, fmt.Errorf("invalid length")
	}
	if s[len(s)-1] == '%' {
		return 0, fmt.Errorf("relative length")
	}

	adj := px
	if len(s) > 2 {
		switch s[len(s)-2:] {
		case "cm":
			adj = gdf.Cm
		case "mm":
			adj = gdf.Mm
		case "pc":
			adj = gdf.Pica
		case "pt":
			adj = 1.0
		}
	}

	// skip non-numerical trailing chars
	i := len(s) - 1
	for i > -1 && (s[i] < '0' || s[i] > '9') {
		i--
	}
	s = s[:i+1]

	length, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return length * adj, nil
}
