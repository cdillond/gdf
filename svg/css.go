package svg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cdillond/gdf"
)

const px float64 = gdf.In / 96.0

func parseNumPct(s string) (float64, error) {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, '%')
	if i > 0 {
		f, err := strconv.ParseFloat(s[:i], 64)
		return f / 100., err
	}
	return strconv.ParseFloat(s, 64)
}

func parseAbsoluteLength(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return 0, fmt.Errorf("invalid length")
	}
	if s[len(s)-1] == '%' {
		return 0, fmt.Errorf("relative length")
	}

	var adj float64
	if len(s) > 2 {
		switch s[len(s)-2:] {
		case "cm":
			adj = gdf.Cm
		case "mm":
			adj = gdf.Mm
		case "pc":
			adj = gdf.Pica
		case "px":
			adj = px
		default:
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

func (s *style) parseCSS(src string) {
	src = strings.TrimSpace(src)
	src = strings.Trim(src, ";")
	pairs := strings.Split(src, ";")
	for _, pair := range pairs {
		before, after, ok := strings.Cut(pair, ":")
		if !ok {
			fmt.Println("not ok", pair)
			continue
		}
		before = strings.TrimSpace(before)
		after = strings.TrimSpace(after)
		switch before {
		case "fill":
			s.fill, s.fillOpacity = parseColor(after)
		case "fill-opacity":
			fo, err := parseNumPct(after)
			if err == nil {
				s.fillOpacity.val = fo
				s.fillOpacity.isSet = true
			}
		case "stroke":
			s.stroke, s.strokeOpacity = parseColor(after)
		case "stroke-width":
			s.swidth, _ = parseAbsoluteLength(after)
		case "stroke-opacity":
			so, err := parseNumPct(after)
			if err == nil {
				s.strokeOpacity.val = so
				s.strokeOpacity.isSet = true
			}
		case "fill-rule":
			if after == "evenodd" {
				s.fr = eo
			} else {
				s.fr = nz
			}
		}
	}

}
