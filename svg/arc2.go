package svg

import (
	"math"
)

type endParams struct {
	x1, y1    float64 // implicit x and y coordinates of the start point
	rx, ry    float64 // size of the x and y radii
	phi       float64 // x-axis rotation angle
	largeFlag float64 // determines the size of the arc angle
	sweepFlag float64 // determines the drawing direction
	x2, y2    float64 // x and y coordinates of the end point
}

type centerParams struct {
	rx, ry float64 // sizes of the x and y radii
	cx, cy float64 // x and y coordinates of the center point
	phi    float64 // x-axis rotation angle
	theta  float64 // angle of the arc's start point
	delta  float64 // angle of difference between the arc's start point and end point
}

func circle2(theta float64) (x, y float64) {
	return math.Cos(theta), math.Sin(theta)
}

func translate(x1, y1, offsetX, offsetY float64) (x, y float64) {
	return x1 + offsetX, y1 + offsetY
}
func scale(x1, y1, scaleX, scaleY float64) (x, y float64) {
	return x1 * scaleX, y1 * scaleY
}
func rotate(x1, y1, phi float64) (x, y float64) {
	sinPhi, cosPhi := math.Sincos(phi)
	return x1 * cosPhi, y1 * sinPhi
}

func (c centerParams) transform(x1, y1 float64) (x, y float64) {
	//x, y = rotate(x1, y1, c.phi)

	x, y = scale(x1, y1, c.rx, c.ry)
	x, y = translate(x, y, c.cx, c.cy)
	return x, y
}
func (c centerParams) ellipse(theta float64) (x, y float64) {
	x, y = circle2(theta)
	return c.transform(x, y)
}

type segment struct {
	startX, startY float64
	x1, y1         float64 // control point 1
	endX, endY     float64
	x2, y2         float64 // control point 2
}

func fit(alpha float64) segment {
	k := 0.551784777779014
	if alpha < math.Pi/2 {
		k = (4.0 / 3.0) * (math.Tan(alpha / 4.0))
	}
	sinAlpha, cosAlpha := math.Sincos(alpha)
	return segment{
		startX: 1.0,
		startY: 0.0,
		x1:     1.0,
		y1:     k,
		endX:   cosAlpha,
		endY:   sinAlpha,
		x2:     (cosAlpha + k*sinAlpha),
		y2:     (sinAlpha - k*cosAlpha),
	}
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
	return math.Acos(clamp(dot(u, v)/(mag(u)*mag(v)), -1, 1))
}

// center returns the center parameterization parameters of the elliptic arc described by the arguments. x1 and y1 are the coordinates of the current point.
func center(ep endParams) centerParams {
	var out centerParams
	// the p, s, and t suffixes are used here to denote 'prime', 'squared', and 'temporary' respectively.

	// ensure rx and ry are positive.
	rx, ry := math.Abs(ep.rx), math.Abs(ep.ry)
	// pre-calculate sin(phi) and cos(phi); phi must previously have been converted to radians.
	out.phi = ep.phi
	sinPhi, cosPhi := math.Sincos(ep.phi)

	// Step 1: Compute (x1′, y1′)
	xt := (ep.x1 - ep.x2) / 2.0
	yt := (ep.y1 - ep.y2) / 2.0
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
	num := rxs*rys - rxs*y1ps - rys*x1ps
	den := rxs*y1ps + rys*x1ps
	coeff := math.Sqrt(num / den)
	if ep.largeFlag == ep.sweepFlag {
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
	out.cx = vt.i + ((ep.x1 + ep.x2) / 2.0)
	out.cy = vt.j + ((ep.y1 + ep.y2) / 2.0)

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
	delta := math.Abs(angle(vt, u))
	delta = math.Mod(delta, 2*math.Pi)
	if ep.sweepFlag == 0 {
		delta = -delta
	}
	out.delta = delta
	return out
}
