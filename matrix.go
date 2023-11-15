package gdf

import "math"

// Represents the first 2 columns of a 3-column affine transformation matrix, whose third column is always [0,0,1].
type Matrix struct {
	A float64 // X scale
	B float64 // X shear
	C float64 // Y shear
	D float64 // Y scale
	E float64 // X offset
	F float64 // Y offset
}

// Empty/unitialized transformation matrices can result in undefined behavior.
func (m *Matrix) IsEmpty() bool {
	return m.A == 0 && m.A == m.B && m.B == m.C && m.C == m.D && m.D == m.E && m.E == m.F
}

// Sets m equal to the Identity Matrix, Matrix{A:1,B:0,C:0,D:1,E:0,F:0}. (Values for the third
// column are always implicitly [0,0,1]).
func (m *Matrix) SetIdentity() {
	m.A = 1
	m.B = 0
	m.C = 0
	m.D = 1
	m.E = 0
	m.F = 0
}

// Returns a new instance of the Identity Matrix. This function should be used instead of Matrix{} or *new(Matrix),
// since empty matrices can result in undefined behavior.
func NewMatrix() Matrix { return Matrix{1, 0, 0, 1, 0, 0} }

// Returns a Matrix that represents the translation of a coordinate space by dX and dY.
func Translate(dX, dY float64) Matrix {
	return Mul(NewMatrix(), Matrix{1, 0, 0, 1, dX, dY})
	// [1 0 0 1 tx ty]
	/* Translations shall be specified as [ 1 0 0 1 t x t y], where tx and ty shall be the distances to translate
	the origin of the coordinate system in the horizontal and vertical dimensions, respectively. */
}

// Returns a Matrix that represents the scaling of a coordinate space by scaleX and scaleY.
func ScaleBy(scaleX, scaleY float64) Matrix {
	// [sx 0 0 sy 0 0]
	/*
		Scaling shall be obtained by [ sx 0 0 s y 0 0]. This scales the coordinates so that 1 unit in the
		horizontal and vertical dimensions of the new coordinate system is the same size as sx and sy
		units, respectively, in the previous coordinate system.
	*/
	return Mul(NewMatrix(), Matrix{scaleX, 0, 0, scaleY, 0, 0})
}

// Returns a Matrix that represents the rotation of a coordinate space counter-clockwise about the origin by theta.
func Rotate(theta float64) Matrix {
	// [rc rs -rs rc 0 0]
	/*
		Rotations shall be produced by [rc rs -rs rc 0 0], where rc = cos(q ) and rs = sin(q) which has the
		effect of rotating the coordinate system axes by an angle q counter clockwise.
	*/
	return Mul(NewMatrix(), Matrix{math.Cos(theta), math.Sin(theta), -math.Sin(theta), math.Cos(theta), 0, 0})
}

// Returns a Matrix that represents the transformation of a coordinate space by skewing its x axis by xTheta and its y axis by yTheta.
func Skew(xTheta, yTheta float64) Matrix {
	// [1 wx wy 1 0 0]
	/*
		Skew shall be specified by [1 wx wy 1 0 0], where wx = tan(a) and wy = tan(b) which skews the x
		axis by an angle a and the y axis by an angle b.
	*/
	return Mul(NewMatrix(), Matrix{1, math.Tan(xTheta), math.Tan(yTheta), 1, 0, 0})
}

// Represents a Point. All Points implicitly include a Z coordinate of 1.
type Point struct {
	X, Y float64
}

// Returns the Point resulting from the transformation of p by m.
func Transform(p Point, m Matrix) Point {
	return Point{
		X: p.X*m.A + p.Y*m.C + m.E,
		Y: p.X*m.B + p.Y*m.D + m.F,
	}
}

// Returns the coordinates of the vertices of R transformed by m in the order LL, UL, LR, UR.
func TransformRect(r Rect, m Matrix) (Point, Point, Point, Point) {
	LL := Point{X: r.LLX, Y: r.LLY}
	UL := Point{X: r.LLX, Y: r.URY}
	LR := Point{X: r.URX, Y: r.LLY}
	UR := Point{X: r.URX, Y: r.URY}
	return Transform(LL, m), Transform(UL, m), Transform(LR, m), Transform(UR, m)
}

// Returns the matrix product of m1 and m2.
func Mul(m1, m2 Matrix) Matrix {
	C00 := m1.A*m2.A + m1.B*m2.C + 0*m2.E
	C01 := m1.A*m2.B + m1.B*m2.D + 0*m2.F
	//C02 := tm1.a*0 + tm1.b*0 + 0 * 1 = 0

	C10 := m1.C*m2.A + m1.D*m2.C + 0*m2.E
	C11 := m1.C*m2.B + m1.D*m2.D + 0*m2.F
	//C12 := tm1.c*0 + tm1.d*0 + 0*1 = 0

	C20 := m1.E*m2.A + m1.F*m2.C + 1*m2.E
	C21 := m1.E*m2.B + m1.F*m2.D + 1*m2.F
	//C22 := tm1.e*0 + tm1.f*0 + 1*1 = 1
	return Matrix{A: C00, B: C01, C: C10, D: C11, E: C20, F: C21}
}

// order of operations: translate, rotate, scale or skew
