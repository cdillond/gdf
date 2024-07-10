package gdf

import (
	"math"
)

func equiv(p1, p2 Point) bool {
	p1.X = math.Round(p1.X*1_000) / 1_000
	p1.Y = math.Round(p1.Y*1_000) / 1_000
	p2.X = math.Round(p2.X*1_000) / 1_000
	p2.Y = math.Round(p2.Y*1_000) / 1_000
	return p1 == p2
}

// SVGArcParams represents the "endpoint parameterization" of an SVG arc subpath.
type SVGArcParams struct {
	X1, Y1      float64 // x and y coordinates of the start point in SVG user space
	Rx, Ry      float64 // size of the x and y radii
	Phi         float64 // x-axis rotation angle in radians
	IsLong      bool    // determines the size of the arc angle
	IsClockwise bool    // determines the drawing direction
	X2, Y2      float64 // x and y coordinates of the end point in SVG user space
}

// SVGArc draws an arc, represented in the SVG "endpoint parameterization" form, to c. The h and m
// parameters are the SVG's height and transformation matrix. This function should be avoided by most users.
func (c *ContentStream) SVGArc(s SVGArcParams, h float64, m Matrix) {
	cp := center(s, h, m)

	const step = math.Pi / 4.
	var tau = min(math.Abs(step), math.Pi)
	if tau <= 0 {
		tau = math.Pi / 8.
	}
	var p1, q1, q2, p2 Point
	if cp.delta > 0 {
		for beta := 0.0; beta < cp.delta; beta += tau {
			p1, q1, q2, p2 = basicArc(cp.theta, min(tau, cp.delta-beta))
			p1.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			q1.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			q2.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			p2.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			if beta == 0 && !equiv(c.CurPt, p1) {
				c.MoveTo(p1.X, p1.Y)
			}
			c.CubicBezier1(q1.X, q1.Y, q2.X, q2.Y, p2.X, p2.Y)
			cp.theta += tau
		}

	} else if cp.delta < 0 {
		for beta := 0.0; beta > cp.delta; beta -= tau {
			p1, q1, q2, p2 := basicArc(cp.theta, max(-tau, cp.delta-beta))
			p1.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			q1.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			q2.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			p2.scale(cp.rx, cp.ry).rotate(cp.phi).translate(cp.cx, cp.cy).svgpdf(h, m)
			if beta == 0 && !equiv(c.CurPt, p1) {
				c.MoveTo(p1.X, p1.Y)
			}
			c.CubicBezier1(q1.X, q1.Y, q2.X, q2.Y, p2.X, p2.Y)
			cp.theta -= tau
		}
	}

}

type centerParams struct {
	rx, ry float64 // sizes of the x and y radii
	cx, cy float64 // x and y coordinates of the center point
	phi    float64 // x-axis rotation angle
	theta  float64 // angle of the arc's start point
	delta  float64 // angle of difference between the arc's start point and end point
}

type vec struct {
	i, j float64
}

// dot returns the dot product of u and v, both 1x2 vectors.
func dot(u, v vec) float64 {
	return u.i*v.i + u.j*v.j
}

func lmul(m [2]vec, v vec) vec {
	return vec{
		i: dot(vec{i: m[0].i, j: m[1].i}, v),
		j: dot(vec{i: m[0].j, j: m[1].j}, v),
	}
}

// mag returns the magnitude of v.
func mag(v vec) float64 {
	return math.Hypot(v.i, v.j)
}

// clamp constrains f to a value within the interval [minF, maxF].
func clamp(f, minF, maxF float64) float64 {
	return min(maxF, max(minF, f))
}

// angle returns the angle, in radians, between u and v.
func angle(u, v vec) float64 {
	sign := 1.0
	if u.i*v.j-u.j*v.i < 0 {
		sign = -sign
	}
	return sign * math.Acos(clamp(dot(u, v)/(mag(u)*mag(v)), -1, 1))
}

func (p *Point) svgpdf(h float64, m Matrix) *Point {
	*p = Transform(*p, m)
	p.Y = h - p.Y
	return p
}

// center returns the center parameterization parameters of the elliptic arc described by the arguments. x1 and y1 are the coordinates of the current point.
func center(a SVGArcParams, h float64, m Matrix) centerParams {
	const twoPi = 2 * math.Pi

	var out centerParams
	// the p, s, and t suffixes are used here to denote 'prime', 'squared', and 'temporary' respectively.

	// ensure rx and ry are positive.
	rx, ry := math.Abs(a.Rx), math.Abs(a.Ry)
	// pre-calculate sin(phi) and cos(phi); phi must previously have been converted to radians.
	out.phi = a.Phi
	sinPhi, cosPhi := math.Sincos(a.Phi)

	// Step 1: Compute (x1′, y1′)
	xt := (a.X1 - a.X2) / 2.0
	yt := (a.Y1 - a.Y2) / 2.0
	mt := [2]vec{
		{i: cosPhi, j: -sinPhi},
		{i: sinPhi, j: cosPhi},
	}
	vt := lmul(mt, vec{i: xt, j: yt})
	x1p, y1p := vt.i, vt.j
	x1ps := x1p * x1p
	y1ps := y1p * y1p

	// F6.6: ensure rx and ry are large enough; if not, scale them.
	lambda := x1ps/(rx*rx) + y1ps/(ry*ry)
	if lambda > 1.0 {
		rx *= math.Sqrt(lambda)
		ry *= math.Sqrt(lambda)
	}
	rxs := rx * rx
	rys := ry * ry
	out.rx = rx
	out.ry = ry

	// Step 2: Compute (cx′, cy′)
	num := math.Abs(rxs*rys - rxs*y1ps - rys*x1ps)
	den := math.Abs(rxs*y1ps + rys*x1ps)
	coeff := math.Sqrt(num / den)
	if a.IsLong == a.IsClockwise {
		coeff = -coeff
	}
	vt = vec{i: rx * y1p / ry, j: -1.0 * ry * x1p / rx}
	cxp := coeff * vt.i
	cyp := coeff * vt.j

	// Step 3: Compute (cx, cy)
	vt = vec{i: cxp, j: cyp}
	mt = [2]vec{
		{i: cosPhi, j: sinPhi},
		{i: -sinPhi, j: cosPhi},
	}
	vt = lmul(mt, vt)
	out.cx = vt.i + ((a.X1 + a.X2) / 2.0)
	out.cy = vt.j + ((a.Y1 + a.Y2) / 2.0)

	// Step 4: Compute theta and delta
	u := vec{i: 1, j: 0}
	vt = vec{
		i: (x1p - cxp) / rx,
		j: (y1p - cyp) / ry,
	}
	out.theta = angle(u, vt)

	u = vec{
		i: (-x1p - cxp) / rx,
		j: (-y1p - cyp) / ry,
	}

	delta := math.Mod(math.Abs(angle(vt, u)), 2*math.Pi)
	if delta < math.Pi && a.IsLong {
		delta = (twoPi) - delta
	} else if delta > math.Pi && (!a.IsLong) {
		delta = (twoPi) - delta
	}
	if !a.IsClockwise {
		delta = -delta
	}
	out.delta = delta
	return out
}
