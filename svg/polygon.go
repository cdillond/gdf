package svg

import (
	"strconv"
)

func parsePolygonPoints(in chan token) []svgCmd {
	var coords []float64
	tok, ok := <-in
	for ok {
		switch tok.typ {
		case bad:
		case num:
			f, err := strconv.ParseFloat(string(tok.text), 64)
			if err != nil {
				panic(err)
			}
			coords = append(coords, f)
		}
		tok, ok = <-in
	}
	cmds := make([]svgCmd, len(coords)/2)
	for i, j := 0, 0; i < len(cmds); i, j = i+1, j+2 {
		cmds[i].op = lAbs
		cmds[i].args = []float64{coords[j], coords[j+1]}
	}
	cmds[0].op = mAbs
	cmds = append(cmds, svgCmd{op: zAbs})
	return cmds
}
