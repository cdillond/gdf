package gdf

import "math"

const (
	DEG float64 = math.Pi / 180 // 1 degree in radians
	CM  float64 = 72 / 2.54     // 1 centimeter in points
	IN  float64 = 72            // 1 inch in points
	MM  float64 = CM / 10       // 1 millimeter in points
)

type Rect struct {
	LLX, LLY, URX, URY float64
}

var (
	A5         = Rect{0, 0, 148 * MM, 210 * MM}
	A4         = Rect{0, 0, 210 * MM, 297 * MM}
	A3         = Rect{0, 0, 297 * MM, 420 * MM}
	US_LETTER  = Rect{0, 0, 8.5 * IN, 11 * IN}
	US_LEGAL   = Rect{0, 0, 8.5 * IN, 14 * IN}
	US_TABLOID = Rect{0, 0, 11 * IN, 17 * IN}

	A5_LS         = Rect{0, 0, 210 * MM, 148 * MM}
	A4_LS         = Rect{0, 0, 297 * MM, 210 * MM}
	A3_LS         = Rect{0, 0, 420 * MM, 297 * MM}
	US_LETTER_LS  = Rect{0, 0, 11 * IN, 8.5 * IN}
	US_LEGAL_LS   = Rect{0, 0, 14 * IN, 8.5 * IN}
	US_TABLOID_LS = Rect{0, 0, 17 * IN, 11 * IN}
)

type Margins struct {
	Left, Right, Top, Bottom float64
}

var (
	MARGINS_NONE    = Margins{0, 0, 0, 0}
	MARGINS_HALF_IN = Margins{.5 * IN, .5 * IN, .5 * IN, .5 * IN}
	MARGINS_1_IN    = Margins{IN, IN, IN, IN}
	MARGINS_1_CM    = Margins{CM, CM, CM, CM}
	MARGINS_5_PT    = Margins{5, 5, 5, 5}
)

// Converts n font units to points given a font size in points. For PDFs, ppem is always 1000.
func FUToPt(n, fontSize float64) float64 { return n * fontSize / 1000 }

// Converts n points to font units given a font size in points. For PDFs, ppem is always 1000.
func PtToFU(n, fontSize float64) float64 { return n * 1000 / fontSize }

// Returns the Rect that results from applying m to r.
func Bounds(r Rect, m Margins) Rect {
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
