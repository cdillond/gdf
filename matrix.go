package gdf

import (
	"fmt"
	"math"
)

/*
Represents the first 2 columns of a 3-column affine transformation matrix, whose third column is always [0,0,1].
Matrices are used to represent transformations applied to an object's coordinate space by the PDF viewer. A ContentStream's
Current Transformation Matrix (CTM, c.GS.Matrix) and text matrix (c.TextObj.Matrix) are the only two matrices within a PDF that gdf allows users to directly
manipulate. CTMs can be altered by calling c.Concat(m). Text matrices are implicitly altered by any text showing operation, and
can be explicitly set through calls to c.SetTextMatrix(m). Whereas Concat sets the CTM to the matrix product of m and the CTM,
SetTextMatrix replaces the current text matrix with m.
*/
type Matrix struct {
	A float64 // X scale
	B float64 // X shear
	C float64 // Y shear
	D float64 // Y scale
	E float64 // X offset
	F float64 // Y offset
}

// Returns a new instance of the Identity Matrix. This function should be used instead of Matrix{} or *new(Matrix),
// since empty matrices can result in undefined behavior. When combining transformations, the PDF spec's recommended
// order of operations is translate, rotate, scale or skew.
func NewMatrix() Matrix { return Matrix{1, 0, 0, 1, 0, 0} }

// Returns a Matrix that represents the translation of a coordinate space by dX and dY.
func Translate(dX, dY float64) Matrix {
	// [1 0 0 1 tx ty]
	/* Translations shall be specified as [ 1 0 0 1 t x t y], where tx and ty shall be the distances to translate
	the origin of the coordinate system in the horizontal and vertical dimensions, respectively. */
	return Matrix{1, 0, 0, 1, dX, dY}
}

// Returns a Matrix that represents the scaling of a coordinate space by scaleX and scaleY.
func ScaleBy(scaleX, scaleY float64) Matrix {
	// [sx 0 0 sy 0 0]
	/*
		Scaling shall be obtained by [ sx 0 0 s y 0 0]. This scales the coordinates so that 1 unit in the
		horizontal and vertical dimensions of the new coordinate system is the same size as sx and sy
		units, respectively, in the previous coordinate system.
	*/
	return Matrix{scaleX, 0, 0, scaleY, 0, 0}
}

// Returns a Matrix that represents the rotation of a coordinate space counter-clockwise about the origin by theta.
func Rotate(theta float64) Matrix {
	// [rc rs -rs rc 0 0]
	/*
		Rotations shall be produced by [rc rs -rs rc 0 0], where rc = cos(q ) and rs = sin(q) which has the
		effect of rotating the coordinate system axes by an angle q counter clockwise.
	*/
	return Matrix{math.Cos(theta), math.Sin(theta), -math.Sin(theta), math.Cos(theta), 0, 0}
}

// Returns a Matrix that represents the transformation of a coordinate space by skewing its x axis by xTheta and its y axis by yTheta.
func Skew(xTheta, yTheta float64) Matrix {
	// [1 wx wy 1 0 0]
	/*
		Skew shall be specified by [1 wx wy 1 0 0], where wx = tan(a) and wy = tan(b) which skews the x
		axis by an angle a and the y axis by an angle b.
	*/
	return Matrix{1, math.Tan(xTheta), math.Tan(yTheta), 1, 0, 0}
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

// Returns the coordinates of the vertices of R transformed by m in the order LL, UL, LR, UR. The returned points do not necessarily form a valid Rect.
func TransformRect(r Rect, m Matrix) (Point, Point, Point, Point) {
	LL := Point{X: r.LLX, Y: r.LLY}
	UL := Point{X: r.LLX, Y: r.URY}
	LR := Point{X: r.URX, Y: r.LLY}
	UR := Point{X: r.URX, Y: r.URY}
	return Transform(LL, m), Transform(UL, m), Transform(LR, m), Transform(UR, m)
}

// Returns the matrix product of m1 and m2. Note: this operation is not commutative.
func Mul(m1, m2 Matrix) Matrix {
	C00 := m1.A*m2.A + m1.B*m2.C //+ 0*m2.E
	C01 := m1.A*m2.B + m1.B*m2.D //+ 0*m2.F
	//C02 := tm1.a*0 + tm1.b*0 + 0 * 1 = 0

	C10 := m1.C*m2.A + m1.D*m2.C //+ 0*m2.E
	C11 := m1.C*m2.B + m1.D*m2.D //+ 0*m2.F
	//C12 := tm1.c*0 + tm1.d*0 + 0*1 = 0

	C20 := m1.E*m2.A + m1.F*m2.C + 1*m2.E
	C21 := m1.E*m2.B + m1.F*m2.D + 1*m2.F
	//C22 := tm1.e*0 + tm1.f*0 + 1*1 = 1
	return Matrix{A: C00, B: C01, C: C10, D: C11, E: C20, F: C21}
}

// Returns the inverse of m and an error. If m has no inverse, an empty Matrix and a non-nil error are returned.
func (m Matrix) Inverse() (Matrix, error) {
	det := m.A*m.D - m.C*m.B
	if det == 0 {
		return *new(Matrix), fmt.Errorf("Matrix m has no inverse")
	}
	return Matrix{
		A: +m.D / det,
		B: -m.B / det,

		C: -m.C / det,
		D: +m.A / det,

		E: +(m.C*m.F - m.E*m.D) / det,
		F: -(m.A*m.F - m.E*m.B) / det,
	}, nil
}

// Returns true if all editable fields of the matrix are set to 0. Empty matrices should be
// avoided, as they can result in undefined behavior.
//func (m *Matrix) IsEmpty() bool {
//	return m.A == 0 && m.A == m.B && m.B == m.C && m.C == m.D && m.D == m.E && m.E == m.F
//}

// Sets m equal to the Identity Matrix, Matrix{A:1,B:0,C:0,D:1,E:0,F:0}. (Values for the third
// column are always implicitly [0,0,1]).
//func (m *Matrix) SetIdentity() {
//	m.A = 1
//	m.B = 0
//	m.C = 0
//	m.D = 1
//	m.E = 0
//	m.F = 0
//}
