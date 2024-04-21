package svg

import (
	"github.com/cdillond/gdf"
)

type pdfPathCmd struct {
	op   pdfPathOp
	args []gdf.Point
}

type pdfPathOp uint

const (
	moveTo pdfPathOp = iota
	lineTo
	curveTo
	closePath
	circle
	ellipse
	badpathOp
)

func (p pdfPathOp) isValid() bool {
	return p < badpathOp
}

func svgPathOp2pdfPathOp(p svgPathOp) pdfPathOp {
	switch p {
	case mAbs, mRel:
		return moveTo
	case lAbs, lRel, hAbs, hRel, vAbs, vRel:
		return lineTo
	case zAbs, zRel:
		return closePath
	case cAbs, cRel, sAbs, sRel, qAbs, qRel, tAbs, tRel:
		return curveTo
	}
	return badpathOp
}

// Converts a Quadratic Bezier Curve to a Cubic Bezier Curve.
func quadraticToCubic(start, P1, dst gdf.Point) [4]gdf.Point {
	// https://fontforge.org/docs/techref/bezier.html#converting-truetype-to-postscript
	var out [4]gdf.Point
	out[0] = start
	out[3] = dst

	// control point 1
	out[1].X = 2. / 3. * (P1.X - start.X)
	out[1].Y = 2. / 3. * (P1.Y - start.Y)
	out[1].X += P1.X
	out[1].Y += P1.Y

	// control point 2
	out[2].X = 2. / 3. * (P1.X - dst.X)
	out[2].Y = 2. / 3. * (P1.Y - dst.Y)
	out[2].X += dst.X
	out[2].Y += dst.Y

	return out
}

func resolvePathCmds(data []svgCmd) []pdfPathCmd {
	var cur gdf.Point
	var ctrlPt gdf.Point
	var out []pdfPathCmd

	for _, rcmd := range data {
		switch rcmd.op {
		case zAbs, zRel:
			out = append(out, pdfPathCmd{op: svgPathOp2pdfPathOp(rcmd.op)})
		case hAbs, hRel, vAbs, vRel:
			switch rcmd.op {
			case hAbs:
				cur.X = rcmd.args[0]
			case hRel:
				cur.X += rcmd.args[0]
			case vAbs:
				cur.Y = rcmd.args[0]
			case vRel:
				cur.Y += rcmd.args[0]
			}
			out = append(out, pdfPathCmd{
				op:   svgPathOp2pdfPathOp(rcmd.op),
				args: []gdf.Point{cur},
			})
		case mAbs, lAbs, mRel, lRel:
			for n := 0; n < len(rcmd.args); n += 2 {
				switch rcmd.op {
				case mAbs, lAbs:
					cur = gdf.Point{X: rcmd.args[n+0], Y: rcmd.args[n+1]}
					if n > 0 {
						rcmd.op = lAbs
					}
				case mRel, lRel:
					cur.X += rcmd.args[n+0]
					cur.Y += rcmd.args[n+1]
					if n > 0 {
						rcmd.op = lRel
					}
				}
				out = append(out, pdfPathCmd{
					op:   svgPathOp2pdfPathOp(rcmd.op),
					args: []gdf.Point{cur},
				})
			}
		case cAbs, cRel, sAbs, sRel:
			for n := 0; n < len(rcmd.args); n += 6 {
				var pts []gdf.Point
				if ctrlPt.X == 0 && ctrlPt.Y == 0 {
					ctrlPt = cur
				}
				switch rcmd.op {
				case sAbs:
					ctrlPt.X = cur.X + (cur.X - ctrlPt.X)
					ctrlPt.Y = cur.Y + (cur.Y - ctrlPt.Y)
					pts = []gdf.Point{
						ctrlPt,
						{X: rcmd.args[n+0], Y: rcmd.args[n+1]},
						{X: rcmd.args[n+2], Y: rcmd.args[n+3]},
					}
					ctrlPt = gdf.Point{X: rcmd.args[n+0], Y: rcmd.args[n+1]}
					cur = gdf.Point{X: rcmd.args[n+2], Y: rcmd.args[n+3]}
				case sRel:
					ctrlPt.X = cur.X + (cur.X - ctrlPt.X)
					ctrlPt.Y = cur.Y + (cur.Y - ctrlPt.Y)
					pts = []gdf.Point{
						ctrlPt,
						{X: cur.X + rcmd.args[n+0], Y: cur.Y + rcmd.args[n+1]},
						{X: cur.X + rcmd.args[n+2], Y: cur.Y + rcmd.args[n+3]},
					}
					ctrlPt = gdf.Point{X: cur.X + rcmd.args[n+0], Y: cur.Y + rcmd.args[n+1]}
					cur = gdf.Point{X: cur.X + rcmd.args[n+2], Y: cur.Y + rcmd.args[n+3]}

				case cAbs:
					pts = []gdf.Point{
						{X: rcmd.args[n+0], Y: rcmd.args[n+1]},
						{X: rcmd.args[n+2], Y: rcmd.args[n+3]},
						{X: rcmd.args[n+4], Y: rcmd.args[n+5]},
					}
					ctrlPt = gdf.Point{X: rcmd.args[n+2], Y: rcmd.args[n+3]}
					cur = gdf.Point{X: rcmd.args[n+4], Y: rcmd.args[n+5]}
				case cRel:
					pts = []gdf.Point{
						{X: cur.X + rcmd.args[n+0], Y: cur.Y + rcmd.args[n+1]},
						{X: cur.X + rcmd.args[n+2], Y: cur.Y + rcmd.args[n+3]},
						{X: cur.X + rcmd.args[n+4], Y: cur.Y + rcmd.args[n+5]},
					}
					ctrlPt = gdf.Point{X: cur.X + rcmd.args[n+2], Y: cur.Y + rcmd.args[n+3]}
					cur = gdf.Point{X: cur.X + rcmd.args[n+4], Y: cur.Y + rcmd.args[n+5]}
				}
				out = append(out, pdfPathCmd{
					op:   svgPathOp2pdfPathOp(rcmd.op),
					args: pts,
				})
			}
		case tAbs, tRel, qAbs, qRel:
			var endPt gdf.Point
			switch rcmd.op {
			case tAbs:
				ctrlPt.X += (cur.X - ctrlPt.X)
				ctrlPt.Y += (cur.Y - ctrlPt.Y)
				endPt.X = rcmd.args[0]
				endPt.Y = rcmd.args[1]
			case tRel:
				ctrlPt.X += (cur.X - ctrlPt.X)
				ctrlPt.Y += (cur.Y - ctrlPt.Y)
				endPt.X = cur.X + rcmd.args[0]
				endPt.Y = cur.Y + rcmd.args[1]
			case qAbs:
				ctrlPt.X = rcmd.args[0]
				ctrlPt.Y = rcmd.args[1]
				endPt.X = rcmd.args[2]
				endPt.Y = rcmd.args[3]
			case qRel:
				ctrlPt.X = cur.X + rcmd.args[0]
				ctrlPt.Y = cur.Y + rcmd.args[1]
				endPt.X = cur.X + rcmd.args[2]
				endPt.Y = cur.Y + rcmd.args[3]
			}
			cubic := quadraticToCubic(cur, ctrlPt, endPt)
			cur = endPt
			out = append(out, pdfPathCmd{
				op:   svgPathOp2pdfPathOp(rcmd.op),
				args: cubic[:],
			})
		}
	}
	return out
}
