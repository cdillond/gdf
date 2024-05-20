package gdf

import (
	"math"
)

func unitCircle(theta float64) (p Point) {
	p.Y, p.X = math.Sincos(theta)
	return p
}

func circleDxDy(theta float64) (dx, dy float64) {
	dx, dy = math.Sincos(theta)
	return -dx, dy
}

// Affine transformations of Points

func (p *Point) scale(x, y float64) *Point {
	p.X *= x
	p.Y *= y
	return p
}

func (p *Point) translate(dx, dy float64) *Point {
	p.X += dx
	p.Y += dy
	return p
}

func (p *Point) rotate(theta float64) *Point {
	sin, cos := math.Sincos(theta)
	x, y := p.X, p.Y
	p.X = x*cos - y*sin
	p.Y = x*sin + y*cos
	return p
}

// Returns the distance between a unit circle segment end point and its Bézier control point for a given theta.
func kappa(theta float64) float64 { return 4. / 3. * math.Tan(theta/4.) }

// basic arc returns the arc from theta to theta+delta iff delta <= pi.
func basicArc(theta, delta float64) (p1, q1, q2, p2 Point) {
	p1 = unitCircle(0)
	p2 = unitCircle(delta)

	k := kappa(delta)

	dx, dy := circleDxDy(0)
	q1.X = p1.X - dx*k
	q1.Y = p1.Y + dy*k

	dx, dy = circleDxDy(delta)
	q2.X = p2.X - dx*k
	q2.Y = p2.Y - dy*k

	p1.rotate(theta)
	q1.rotate(theta)
	q2.rotate(theta)
	p2.rotate(theta)
	return p1, q1, q2, p2
}

// Arc draws an elliptic arc. If e is an ellipse defined by the parametric
// function y(theta) = cy + ry*sin(theta) and x(theta) = cx + rx*cos(theta),
// then the arc is the segment of the ellipse that begins at e(theta) and
// ends at e(theta+delta). If delta is negative, the arc is drawn clockwise.
func (c *ContentStream) Arc(cx, cy, rx, ry, theta, delta float64) {
	c.Arc2(cx, cy, rx, ry, theta, delta, 0, math.Pi/2.)
}

// Circle begins a new path and appends a circle of radius r with a center point of (cx, cy) to the path.
// The new current point becomes (cx + r, cy).
func (c *ContentStream) Circle(cx, cy, r float64) {
	c.Arc2(cx, cy, r, r, 0, 2*math.Pi, 0, math.Pi/4.)
}

// Ellipse begins a new path and appends an ellipse with a center point of (cx, cy), an x-radius of rx, and
// a y-radius of ry to the path.
func (c *ContentStream) Ellipse(cx, cy, rx, ry float64) {
	c.Arc2(cx, cy, rx, ry, 0, 2*math.Pi, 0, math.Pi/4.)
}

// Arc2 draws an elliptic arc. It is similar to Arc, but includes additional parameters. phi indicates
// an angle relative to the x-axis that the arc is rotated about its center point. step is the maximum
// size in radians of each segment of the arc. This value must be greater than 0 and less than or equal
// to pi. Lower values increase the accuracy of the arc, but can cause significant performance degradations.
// In practice, values of pi/2 and pi/4 are reasonable upper- and lower-bounds. The Arc method uses
// pi/2 by default. Note: the step argument has no effect on arcs with deltas that have an absolute value
// less than the value of step.
func (c *ContentStream) Arc2(cx, cy, rx, ry, theta, delta, phi, step float64) {
	var tau = min(math.Abs(step), math.Pi)
	if tau <= 0 {
		tau = math.Pi / 8.
	}
	var p1, q1, q2, p2 Point
	if delta > 0 {
		for beta := 0.0; beta < delta; beta += tau {
			p1, q1, q2, p2 = basicArc(theta, min(tau, delta-beta))
			p1.scale(rx, ry).rotate(phi).translate(cx, cy)
			q1.scale(rx, ry).rotate(phi).translate(cx, cy)
			q2.scale(rx, ry).rotate(phi).translate(cx, cy)
			p2.scale(rx, ry).rotate(phi).translate(cx, cy)
			if beta == 0 {
				c.MoveTo(p1.X, p1.Y)
			}
			c.CubicBezier1(q1.X, q1.Y, q2.X, q2.Y, p2.X, p2.Y)
			theta += tau
		}

	} else if delta < 0 {
		for beta := 0.0; beta > delta; beta -= tau {
			p1, q1, q2, p2 := basicArc(theta, max(-tau, delta-beta))
			p1.scale(rx, ry).rotate(phi).translate(cx, cy)
			q1.scale(rx, ry).rotate(phi).translate(cx, cy)
			q2.scale(rx, ry).rotate(phi).translate(cx, cy)
			p2.scale(rx, ry).rotate(phi).translate(cx, cy)
			if beta == 0 {
				c.MoveTo(p1.X, p1.Y)
			}
			c.CubicBezier1(q1.X, q1.Y, q2.X, q2.Y, p2.X, p2.Y)
			theta -= tau
		}
	}

}
