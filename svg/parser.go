package svg

import (
	"strconv"
)

func parseDAttr(in chan token) ([]svgCmd, error) {
	var cmds []svgCmd
	tmp := new(svgCmd)
	tok, ok := <-in
	for ok {
		switch tok.typ {
		case bad:
			// ignore for now
		case num:
			f, err := strconv.ParseFloat(string(tok.text), 64)
			if err != nil {
				return nil, err
			}
			tmp.args = append(tmp.args, f)
		case op:
			if len(tmp.args) > 0 {
				cmds = append(cmds, *tmp)
				tmp.args = nil
				val := parseCmd(tok.text)
				if val != repeat {
					tmp.op = val
				}
				if tmp.op == zAbs || tmp.op == zRel {
					cmds = append(cmds, *tmp)
				}
			} else {
				tmp.op = parseCmd(tok.text)
			}
		}
		tok, ok = <-in
	}
	if len(tmp.args) > 0 {
		cmds = append(cmds, *tmp)
	}
	return cmds, nil
}

func parsePointsAttr(in chan token) ([]svgCmd, error) {
	var coords []float64
	tok, ok := <-in
	for ok {
		if tok.typ == num {
			f, err := strconv.ParseFloat(string(tok.text), 64)
			if err != nil {
				return nil, err
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
	return cmds, nil
}
