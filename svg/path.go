package svg

import (
	"strconv"

	"github.com/cdillond/gdf"
)

var cmds = [...]byte{'A', 'a', 'C', 'c', 'H', 'h', 'L', 'l', 'M', 'm', 'Q', 'q', 'S', 's', 'T', 't', 'V', 'v', 'Z', 'z'}

func isCmd(c byte) bool {
	for i := 0; i < len(cmds); i++ {
		if c == cmds[i] {
			return true
		}
	}
	return false
}

func pf(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parsePath(cs *gdf.ContentStream, style style, s string, h float64, m gdf.Matrix) {
	// let's begin with a simple path...
	var buf buf
	buf.b = []byte(s)
	if !style.stroke.isSet {
		style.stroke.isNone = true
	}
	var cur gdf.Point
	cur.X = style.xOff
	cur.Y = style.yOff
	var c, op byte
	var lastOp byte      // for S, s, T, and t cmds
	var lastCP gdf.Point // for S, s, T, and t cmds
	ok := true
	for {
		buf.skipWSPComma()
		c, ok = buf.peek()
		if !ok {
			break
		}
		if isCmd(c) {
			buf.skip()
			op = c
		}
		switch op {
		case 'A', 'a':
			rxs, rys := buf.ConsumeNumber(), buf.ConsumeNumber()
			angles := buf.ConsumeNumber()
			buf.skipWSPComma()
			isLong, _ := buf.next()
			buf.skipWSPComma()
			isClockwise, _ := buf.next()
			xs, ys := buf.ConsumeNumber(), buf.ConsumeNumber()

			rx := pf(rxs)
			ry := pf(rys)
			angle := pf(angles)
			x := pf(xs)
			y := pf(ys)

			if op == 'a' {
				x += cur.X
				y += cur.Y
			} else {
				x += style.xOff
				y += style.yOff
			}

			a := gdf.SVGArcParams{
				X1:          cur.X,
				Y1:          cur.Y,
				Rx:          rx,
				Ry:          ry,
				Phi:         angle * gdf.Deg,
				IsLong:      isLong == '1',
				IsClockwise: isClockwise == '1',
				X2:          x,
				Y2:          y,
			}

			cur.X, cur.Y = x, y
			cs.SVGArc(a, h, m)
			//cp := center(ep, h, m)
			//cs.ArcSVG(cp.cx, cp.cy, cp.rx, cp.ry, cp.theta, cp.delta, cp.phi, math.Pi/4., h, m)

		case 'S', 's':
			x2s, y2s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x3s, y3s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x2 := pf(x2s)
			y2 := pf(y2s)
			x3 := pf(x3s)
			y3 := pf(y3s)

			var x1, y1 float64
			if lastOp == 'C' || lastOp == 'c' || lastOp == 'S' || lastOp == 's' {
				x1 = 2*cur.X - lastCP.X
				y1 = 2*cur.Y - lastCP.Y
			} else {
				x1, y1 = cur.X, cur.Y
			}
			if op == 's' {
				x2 += cur.X
				x3 += cur.X
				y2 += cur.Y
				y3 += cur.Y
			} else {
				x2 += style.xOff
				x3 += style.xOff
				y2 += style.yOff
				y3 += style.yOff
			}
			cur.X, cur.Y = x3, y3
			x1, y1 = tf(x1, y1, h, m)
			x2, y2 = tf(x2, y2, h, m)
			x3, y3 = tf(x3, y3, h, m)
			cs.CubicBezier1(x1, y1, x2, y2, x3, y3)

		case 'C', 'c':
			x1s, y1s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x2s, y2s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x3s, y3s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x1 := pf(x1s)
			y1 := pf(y1s)
			x2 := pf(x2s)
			y2 := pf(y2s)
			x3 := pf(x3s)
			y3 := pf(y3s)

			if op == 'c' {
				x1 += cur.X
				x2 += cur.X
				x3 += cur.X
				y1 += cur.Y
				y2 += cur.Y
				y3 += cur.Y
			} else {
				x1 += style.xOff
				x2 += style.xOff
				x3 += style.xOff
				y1 += style.yOff
				y2 += style.yOff
				y3 += style.yOff
			}
			lastCP.X = x2
			lastCP.Y = y2
			cur.X, cur.Y = x3, y3
			x1, y1 = tf(x1, y1, h, m)
			x2, y2 = tf(x2, y2, h, m)
			x3, y3 = tf(x3, y3, h, m)
			cs.CubicBezier1(x1, y1, x2, y2, x3, y3)
		case 'H', 'h':
			xs := buf.ConsumeNumber()
			x := pf(xs)
			if op == 'h' {
				x += cur.X
			} else {
				x += style.xOff
			}
			xf, y := tf(x, cur.Y, h, m)
			cs.LineTo(xf, y)
			cur.X = x
		case 'V', 'v':
			ys := buf.ConsumeNumber()
			y := pf(ys)
			if op == 'v' {
				y += cur.Y
			} else {
				y += style.yOff
			}
			x, yf := tf(cur.X, y, h, m)
			cs.LineTo(x, yf)
			cur.Y = y
		case 'T', 't':
			x3s, y3s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x3 := pf(x3s)
			y3 := pf(y3s)

			var x1, y1 float64
			if lastOp == 'Q' || lastOp == 'q' || lastOp == 'T' || lastOp == 't' {
				x1 = 2*cur.X - lastCP.X
				y1 = 2*cur.Y - lastCP.Y
			} else {
				x1, y1 = cur.X, cur.Y
			}

			if op == 't' {
				x3 += cur.X
				y3 += cur.Y
			} else {
				x3 += style.xOff
				y3 += style.yOff
			}
			lastCP.X = x1
			lastCP.Y = y1

			tcurx, tcury := tf(cur.X, cur.Y, h, m)
			tx1, ty1 := tf(x1, y1, h, m)
			tx3, ty3 := tf(x3, y3, h, m)
			cubic := quadraticToCubic(gdf.Point{tcurx, tcury}, gdf.Point{tx1, ty1}, gdf.Point{tx3, ty3})
			cur.X, cur.Y = x3, y3
			cs.CubicBezier1(cubic[1].X, cubic[1].Y, cubic[2].X, cubic[2].Y, tx3, ty3)

		case 'Q', 'q':
			x1s, y1s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x3s, y3s := buf.ConsumeNumber(), buf.ConsumeNumber()
			x1 := pf(x1s)
			y1 := pf(y1s)
			x3 := pf(x3s)
			y3 := pf(y3s)

			if op == 'q' {
				x1 += cur.X
				x3 += cur.X
				y1 += cur.Y
				y3 += cur.Y
			} else {
				x1 += style.xOff
				x3 += style.xOff
				y1 += style.yOff
				y3 += style.yOff
			}
			lastCP.X = x1
			lastCP.Y = y1

			tcurx, tcury := tf(cur.X, cur.Y, h, m)
			tx1, ty1 := tf(x1, y1, h, m)
			tx3, ty3 := tf(x3, y3, h, m)
			cubic := quadraticToCubic(gdf.Point{tcurx, tcury}, gdf.Point{tx1, ty1}, gdf.Point{tx3, ty3})
			cur.X, cur.Y = x3, y3
			cs.CubicBezier1(cubic[1].X, cubic[1].Y, cubic[2].X, cubic[2].Y, tx3, ty3)

		case 'M', 'm':
			xs, ys := buf.ConsumeNumber(), buf.ConsumeNumber()
			x := pf(xs)
			y := pf(ys)
			if op == 'm' {
				x += cur.X
				y += cur.Y
			} else {
				x += style.xOff
				y += style.yOff
			}
			cur.X, cur.Y = x, y
			x, y = tf(x, y, h, m)
			cs.MoveTo(x, y)
		case 'L', 'l':
			xs, ys := buf.ConsumeNumber(), buf.ConsumeNumber()
			x := pf(xs)
			y := pf(ys)
			if op == 'l' {
				x += cur.X
				y += cur.Y
			} else {
				x += style.xOff
				y += style.yOff
			}
			cur.X, cur.Y = x, y
			x, y = tf(x, y, h, m)
			cs.LineTo(x, y)
		case 'Z', 'z':
			cs.ClosePath()
		}
		lastOp = op
	}
	Paint(cs, false, style)
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
	out[1].X += start.X
	out[1].Y += start.Y

	// control point 2
	out[2].X = 2. / 3. * (P1.X - dst.X)
	out[2].Y = 2. / 3. * (P1.Y - dst.Y)
	out[2].X += dst.X
	out[2].Y += dst.Y

	return out
}
