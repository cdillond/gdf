package gdf

import (
	"errors"
	"math"
	"os"
)

// A FillRule represents an algorithm for determining whether a particular point is interior to a path. See ISO 32000-2:2020 sections
// 8.5.3.3.2 and 8.5.3.3.3 for further details. NonZero is the default, but EvenOdd can produce results that are easier to intuit.
type FillRule bool

const (
	NonZero FillRule = false
	EvenOdd FillRule = true
)

// MoveTo begins a new path starting at (x, y); m.
func (c *ContentStream) MoveTo(x, y float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	c.buf = cmdf(c.buf, op_m, x, y)
}

// LineTo appends a straight line from the current point to (x, y), which becomes the new current point; l.
// If the PathState is not Building, no action is taken.
func (c *ContentStream) LineTo(x, y float64) {
	if c.PathState == Building {
		c.CurPt = Point{x, y}
		c.buf = cmdf(c.buf, op_l, x, y)
	}
}

// DrawLine is a convenience function that strokes a line from Point{x1,y1} to Point{x2,y2}.
func (c *ContentStream) DrawLine(x1, y1, x2, y2 float64) {
	c.MoveTo(x1, y1)
	c.LineTo(x2, y2)
	c.Stroke()
}

// CubicBezier1 appends a cubic Bézier curve to the current path; c. The curve extends
// from the current point to (x3, y3) using using (x1, y1) and (x2, y2) as the Bézier control points.
// If the PathState is not Building, no action is taken.
func (c *ContentStream) CubicBezier1(x1, y1, x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.CurPt = Point{x3, y3}
		c.buf = cmdf(c.buf, op_c, x1, y1, x2, y2, x3, y3)
	}
}

// CubicBezier2 appends a cubic Bézier curve to the current path; v. The curve extends from
// the current point to (x3, y3), using the current point and (x2, y2) as the Bézier control points.
// If the PathState is not Building, no action is taken.
func (c *ContentStream) CubicBezier2(x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.CurPt = Point{x3, y3}
		c.buf = cmdf(c.buf, op_v, x2, y2, x3, y3)
	}
}

// CubicBezier3 appends a cubic Bézier curve to the current path; y. The curve extends
// from the current point to (x3, y3), using (x1, y1) and (x3, y3) as the Bézier control points. If the PathState is
// not Building, no action is taken.
func (c *ContentStream) CubicBezier3(x1, y1, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.CurPt = Point{x3, y3}
		c.buf = cmdf(c.buf, op_y, x1, y1, x3, y3)
	}
}

func join(e1, e2 error) error {
	if e1 == nil {
		return e2
	}
	return errors.Join(e1, e2)
}

func Q() (err error) {
	f, err := os.Open("hm..")
	if err != nil {
		return err
	}

	defer func() {
		err = join(err, f.Close())
	}()

	return err
}

// Circle begins a new path and appends a circle of radius r with a center point of (x, y) to the path.
// The new current point becomes (x + r, y).
func (c *ContentStream) Circle(x, y, r float64) {
	c.MoveTo(x+r, y)
	c.CurPt = Point{x + r, y}

	// reference: https://pomax.github.io/bezierinfo/#circles_cubic
	pointAx := r * 1.
	pointAy := r * 0.265216489839544

	pointBx := r * 0.8946431596345822
	pointBy := r * 0.5195704027385128

	endX := r * 0.70710678118654752
	endY := endX

	// This forms the upper-right quadrant of the circle.
	c.buf = cmdf(c.buf, op_c,
		x+pointAx, y+pointAy,
		x+pointBx, y+pointBy,
		x+endX, y+endY)

	c.buf = cmdf(c.buf, op_c,
		x+pointBy, y+pointBx,
		x+pointAy, y+pointAx,
		x, y+r)

	// Upper-left quadrant
	c.buf = cmdf(c.buf, op_c,
		x-pointAy, y+pointAx,
		x-pointBy, y+pointBx,
		x-endX, y+endY)

	c.buf = cmdf(c.buf, op_c,
		x-pointBx, y+pointBy,
		x-pointAx, y+pointAy,
		x-r, y)

	// Lower-left quadrant
	c.buf = cmdf(c.buf, op_c,
		x-pointAx, y-pointAy,
		x-pointBx, y-pointBy,
		x-endX, y-endY)

	c.buf = cmdf(c.buf, op_c,
		x-pointBy, y-pointBx,
		x-pointAy, y-pointAx,
		x, y-r)

	// Lower-right quadrant
	c.buf = cmdf(c.buf, op_c,
		x+pointAy, y-pointAx,
		x+pointBy, y-pointBx,
		x+endX, y-endY)

	c.buf = cmdf(c.buf, op_c,
		x+pointBx, y-pointBy,
		x+pointAx, y-pointAy,
		x+r, y)
}

func (c *ContentStream) CArc(x, y, r, theta float64) {

	// if theta is ?, break it up...
	var total, a float64

	for i := 0; i < 4 && total < theta; i++ {
		a = min(90*Deg, theta-total)
		total += a
		k := (4.0 / 3.0) * math.Tan(a/4)

		pointAX := r
		pointAY := r * k

		pointBX := r * (math.Cos(a) + k*math.Sin(a))
		pointBY := r * (math.Sin(a) - k*math.Cos(a))

		endX := r * math.Cos(a)
		endY := r * math.Sin(a)

		switch i {
		case 0:
			c.MoveTo(x+r, y)
			c.CubicBezier1(pointAX+x, pointAY+y, pointBX+x, pointBY+y, endX+x, endY+y)
		case 1:
			c.MoveTo(x, y+r)
			c.CubicBezier1(pointAX+x, pointAY+y, pointBX+x, pointBY+y, endX+x, endY+y)
		case 2:
			c.MoveTo(x-r, y)
			c.CubicBezier1(pointAX+x, pointAY+y, pointBX+x, pointBY+y, endX+x, endY+y)
		case 3:
			c.MoveTo(x, y-r)
			c.CubicBezier1(pointAX+x, pointAY+y, pointBX+x, pointBY+y, endX+x, endY+y)
		}

	}

}

func (c *ContentStream) Ellipse(x, y, rx, ry float64) {
	// rx is 1
	//c.QSave()
	c.Concat(ScaleBy(1, ry/rx))
	c.Circle(x, y, rx)
	//c.QRestore()
}

// ClosePath closes the current path; h.
func (c *ContentStream) ClosePath() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = Building
		c.buf = append(c.buf, op_h...)
	}
}

// Stroke strokes the path; S.
func (c *ContentStream) Stroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_S...)
	}
}

// ClosePathStroke closes and strokes the path; s.
func (c *ContentStream) ClosePathStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_s...)
	}

}

// Fill fills the path.
func (c *ContentStream) Fill(wr FillRule) {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		paintOp := op_f
		if wr {
			paintOp = op_f_X
		}
		c.buf = append(c.buf, paintOp...)
	}
}

// FillStroke fills and then strokes the path.
func (c *ContentStream) FillStroke(wr FillRule) {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		paintOp := op_B
		if wr {
			paintOp = op_B_X
		}
		c.buf = append(c.buf, paintOp...)
	}
}

// ClosePathFillStroke closes the path, fills the path, and then strokes the path; b.
func (c *ContentStream) ClosePathFillStroke(wr FillRule) {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		paintOp := op_b
		if wr {
			paintOp = op_b_X
		}
		c.buf = append(c.buf, paintOp...)
	}
}

// EndPath ends the current path. It is used primarily to apply changes to the current clipping path; n.
func (c *ContentStream) EndPath() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_n...)
	}
}

// Re appends a rectangle, of width w and height h, starting at the point (x, y), to the current path; re.
func (c *ContentStream) Re(x, y, w, h float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	c.buf = cmdf(c.buf, op_re, x, y, w, h)
}

// Re2 appends r to the current path; it is intended as a possibly more convenient version of Re.
func (c *ContentStream) Re2(r Rect) {
	c.Re(r.LLX, r.LLY, r.URX-r.LLX, r.URY-r.LLY)
}

// Clip clips the path. It may appear after the last path construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) Clip(wr FillRule) {
	switch c.PathState {
	case Building:
		c.PathState = Clipping
		clipOp := op_W
		if wr {
			clipOp = op_W_X
		}
		c.buf = append(c.buf, clipOp...)
	}
}
