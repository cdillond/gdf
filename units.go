package gdf

import "math"

const (
	DEG float64 = math.Pi / 180 // 1 degree in radians
	CM  float64 = 72 * 1 / 2.54 // 1 centimeter in points
	IN  float64 = 72            // 1 inch in points
	MM  float64 = CM / 10       // 1 millimeter in points
)

var (
	A4        = Rect{0, 0, 210 * MM, 297 * MM}
	US_LETTER = Rect{0, 0, 8.5 * IN, 11 * IN}
	US_LEGAL  = Rect{0, 0, 8.5 * IN, 14 * IN}
)

var (
	HALF_INCH_MARGINS = Margins{.5 * IN, .5 * IN, .5 * IN, .5 * IN}
	ONE_INCH_MARGINS  = Margins{IN, IN, IN, IN}
	ONE_CM_MARGINS    = Margins{CM, CM, CM, CM}
)

// Converts n font units to points given a font size in points. For PDFs, ppem is always 1000.
func FUToPt(n, fontSize float64) float64 { return n * fontSize / 1000 }

// Converts n points to font units given a font size in points. For PDFs, ppem is always 1000.
func PtToFU(n, fontSize float64) float64 { return n * 1000 / fontSize }

func Bounds(r Rect, m Margins) Rect {
	return Rect{
		LLX: r.LLX + m.Left,
		LLY: r.LLY + m.Bottom,
		URX: r.URX - m.Right,
		URY: r.URY - m.Top,
	}
}

type Rect struct {
	LLX, LLY, URX, URY float64
}

func (r Rect) Height() float64 { return r.URY - r.LLY }
func (r Rect) Width() float64  { return r.URX - r.LLX }

// returns the length of a rectangle's diagonal
func Diagonal(r Rect) float64 {
	a := r.URX - r.LLX
	b := r.URY - r.LLY
	return math.Sqrt(a*a + b*b)
}

// returns the angle formed by the lower side of the rectangle and the lower-left to upper-right diagonal, and that angle's complement
func Angles(r Rect) (float64, float64) {
	a := math.Atan((r.URY - r.LLY) / (r.URX - r.LLX))
	return a, 90 - a
}

type Margins struct {
	Left, Right, Top, Bottom float64
}
