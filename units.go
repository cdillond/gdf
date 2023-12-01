package gdf

import "math"

const (
	Deg  float64 = math.Pi / 180 // 1 degree in radians
	In   float64 = 72            // 1 inch in points
	Pica float64 = In / 6        // 1 PostScript pica in points
	Cm   float64 = In / 2.54     // 1 centimeter in points
	Mm   float64 = Cm / 10       // 1 millimeter in points
)

// A Rect represents a rectangular area of a coordinate space, where (LLX, LLY) is the lower left vertex
// and (URX, URY) is the upper right vertex.
type Rect struct {
	LLX, LLY, URX, URY float64
}

var (
	A5        = Rect{0, 0, 148 * Mm, 210 * Mm}
	A4        = Rect{0, 0, 210 * Mm, 297 * Mm}
	A3        = Rect{0, 0, 297 * Mm, 420 * Mm}
	USLetter  = Rect{0, 0, 8.5 * In, 11 * In}
	USLegal   = Rect{0, 0, 8.5 * In, 14 * In}
	USTabloid = Rect{0, 0, 11 * In, 17 * In}
)

// Each field in a Margins struct represents an interior offset from the corresponding edge of a Rect.
type Margins struct {
	Left, Right, Top, Bottom float64
}

var (
	NoMargins = Margins{}
	HalfInch  = Margins{.5 * In, .5 * In, .5 * In, .5 * In}
	OneInch   = Margins{In, In, In, In}
	OneCm     = Margins{Cm, Cm, Cm, Cm}
	FivePt    = Margins{5, 5, 5, 5}
)

// Converts n font units to points given a font size in points. For PDFs, ppem is always 1000.
func FUToPt(n, fontSize float64) float64 { return n * fontSize / 1000 }

// Converts n points to font units given a font size in points. For PDFs, ppem is always 1000.
func PtToFU(n, fontSize float64) float64 { return n * 1000 / fontSize }

// Returns the Rect that results from applying m to r without checking whether that Rect is valid.
func (r Rect) Bounds(m Margins) Rect {
	return Rect{
		LLX: r.LLX + m.Left,
		LLY: r.LLY + m.Bottom,
		URX: r.URX - m.Right,
		URY: r.URY - m.Top,
	}
}

func (r Rect) Height() float64 { return r.URY - r.LLY }
func (r Rect) Width() float64  { return r.URX - r.LLX }

// Returns the length of r's diagonal.
func (r Rect) Diagonal() float64 {
	a := r.URX - r.LLX
	b := r.URY - r.LLY
	return math.Sqrt(a*a + b*b)
}

// Returns the angle, in radians, formed by the lower side of the rectangle and the lower-left to upper-right diagonal, and that angle's complement.
func (r Rect) Angles() (float64, float64) {
	a := math.Atan((r.URY - r.LLY) / (r.URX - r.LLX))
	return a, 90 - a
}

// Returns the rectangle that results from swapping r's x and y values.
func (r Rect) Landscape() Rect {
	return Rect{
		LLX: r.LLY,
		LLY: r.LLX,
		URX: r.URY,
		URY: r.URX,
	}
}
