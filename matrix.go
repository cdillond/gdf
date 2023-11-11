package gdf

import "math"

type Matrix struct {
	a float64 // X scale
	b float64 // X shear
	c float64 // Y shear
	d float64 // Y scale
	e float64 // X offset
	f float64 // Y offset
}

// Empty/unitialized transformation matrices can result in undefined behavior.
func (m *Matrix) IsEmpty() bool {
	return m.a == 0 && m.a == m.b && m.b == m.c && m.c == m.d && m.d == m.e && m.e == m.f
}
func (m *Matrix) SetIdentity() {
	m.a = 1
	m.b = 0
	m.c = 0
	m.d = 1
	m.e = 0
	m.f = 0
}
func NewMatrix() Matrix { return Matrix{1, 0, 0, 1, 0, 0} }

func SetMatrix(a, b, c, d, e, f float64) Matrix { return Matrix{a, b, c, d, e, f} }

// order of operations: translate, rotate, scale or skew

func Translate(m Matrix, dx, dy float64) Matrix {
	return Mul(m, Matrix{1, 0, 0, 1, dx, dy})
	// [1 0 0 1 tx ty]
	/* Translations shall be specified as [ 1 0 0 1 t x t y], where tx and ty shall be the distances to translate
	the origin of the coordinate system in the horizontal and vertical dimensions, respectively. */
}

func ScaleBy(m Matrix, sx, sy float64) Matrix {
	// [sx 0 0 sy 0 0]
	/*
		Scaling shall be obtained by [ sx 0 0 s y 0 0]. This scales the coordinates so that 1 unit in the
		horizontal and vertical dimensions of the new coordinate system is the same size as sx and sy
		units, respectively, in the previous coordinate system.
	*/
	return Mul(m, Matrix{sx, 0, 0, sy, 0, 0})
}

func Rotate(m Matrix, theta float64) Matrix {
	// [rc rs -rs rc 0 0]
	/*
		Rotations shall be produced by [rc rs -rs rc 0 0], where rc = cos(q ) and rs = sin(q) which has the
		effect of rotating the coordinate system axes by an angle q counter clockwise.
	*/
	return Mul(m, Matrix{math.Cos(theta), math.Sin(theta), -math.Sin(theta), math.Cos(theta), 0, 0})
}

func Skew(m Matrix, xTheta, yTheta float64) Matrix {
	// [1 wx wy 1 0 0]
	/*
		Skew shall be specified by [1 wx wy 1 0 0], where wx = tan(a) and wy = tan(b) which skews the x
		axis by an angle a and the y axis by an angle b.
	*/
	return Mul(m, Matrix{1, math.Tan(xTheta), math.Tan(yTheta), 1, 0, 0})
}

// [X Y 1]
type Point struct {
	X, Y float64
}

// implicitly [X Y 1] * [[a b 0][c d 0][e f 1]]
func Transform(p Point, m Matrix) Point {
	return Point{
		X: p.X*m.a + p.Y*m.c + m.e,
		Y: p.X*m.b + p.Y*m.d + m.f,
	}
}

// returns the coordinates of the vertices of R transformed by m in the order LL, UL, LR, UR
func TransformRect(r Rect, m Matrix) (Point, Point, Point, Point) {
	LL := Point{X: r.LLX, Y: r.LLY}
	UL := Point{X: r.LLX, Y: r.URY}
	LR := Point{X: r.URX, Y: r.LLY}
	UR := Point{X: r.URX, Y: r.URY}
	return Transform(LL, m), Transform(UL, m), Transform(LR, m), Transform(UR, m)
}

func Mul(m1, m2 Matrix) Matrix {
	C00 := m1.a*m2.a + m1.b*m2.c + 0*m2.e
	C01 := m1.a*m2.b + m1.b*m2.d + 0*m2.f
	//C02 := tm1.a*0 + tm1.b*0 + 0 * 1 = 0

	C10 := m1.c*m2.a + m1.d*m2.c + 0*m2.e
	C11 := m1.c*m2.b + m1.d*m2.d + 0*m2.f
	//C12 := tm1.c*0 + tm1.d*0 + 0*1 = 0

	C20 := m1.e*m2.a + m1.f*m2.c + 1*m2.e
	C21 := m1.e*m2.b + m1.f*m2.d + 1*m2.f
	//C22 := tm1.e*0 + tm1.f*0 + 1*1 = 1
	return Matrix{a: C00, b: C01, c: C10, d: C11, e: C20, f: C21}
}

func (m Matrix) All() [6]float64 {
	return [6]float64{m.a, m.b, m.c, m.d, m.e, m.f}
}
