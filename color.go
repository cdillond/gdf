package gdf

type ColorSpace uint

const (
	DeviceGray ColorSpace = iota
	DeviceRGB
	DeviceCMYK
	PatternCS
	badColorSpace
)

var (
	colorSpaces = [...]string{"/DeviceGray", "/DeviceRGB", "/DeviceCMYK", "/Pattern"}
	_           = (int8(badColorSpace) - int8(len(colorSpaces))) << 8
)

func (c ColorSpace) isValid() bool { return c < badColorSpace }
func (c ColorSpace) String() string {
	if c.isValid() {
		return colorSpaces[c]
	}
	return ""
}

type Color interface {
	color() []float64
}

// Sets c's stroking color (SColor) to cl and sets its SColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColorStroke(cl Color) {
	switch v := cl.(type) {
	case GColor:
		c.NColorSpace = DeviceGray
		c.buf = cmdf(c.buf, op_G, float64(v))
	case RGBColor:
		c.NColorSpace = DeviceRGB
		c.buf = cmdf(c.buf, op_RG, v.R, v.G, v.B)
	case CMYKColor:
		c.NColorSpace = DeviceCMYK
		c.buf = cmdf(c.buf, op_K, v.C, v.M, v.Y, v.K)
	default:
		return
	}
	c.SColor = cl
}

// Sets c's nonstroking color (NColor) to cl and sets its NColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColor(cl Color) {
	switch v := cl.(type) {
	case GColor:
		c.NColorSpace = DeviceGray
		c.buf = cmdf(c.buf, op_g, float64(v))
	case RGBColor:
		c.NColorSpace = DeviceRGB
		c.buf = cmdf(c.buf, op_rg, v.R, v.G, v.B)
	case CMYKColor:
		c.NColorSpace = DeviceCMYK
		c.buf = cmdf(c.buf, op_k, v.C, v.M, v.Y, v.K)
	default:
		return
	}
	c.NColor = cl
}

// Grayscale color; must be in [0,1].
type GColor float64

const (
	Black GColor = 0
	Gray  GColor = .5
	White GColor = 1
)

func (g GColor) color() []float64 { return []float64{float64(g)} }

// RGB Color; R,G, and B must be in [0,1].
type RGBColor struct {
	R, G, B float64
}

func (r RGBColor) color() []float64 { return []float64{r.R, r.G, r.B} }

var (
	Red   = RGBColor{1, 0, 0}
	Green = RGBColor{0, 1, 0}
	Blue  = RGBColor{0, 0, 1}
)

// CMYK Color; C, M, Y, and K must be in [0,1].
type CMYKColor struct {
	C, M, Y, K float64
}

func (c CMYKColor) color() []float64 { return []float64{c.C, c.M, c.Y, c.K} }

var (
	Cyan      = CMYKColor{1, 0, 0, 0}
	Magenta   = CMYKColor{0, 1, 0, 0}
	Yellow    = CMYKColor{0, 0, 1, 0}
	CMYKBlack = CMYKColor{0, 0, 0, 1}
)
