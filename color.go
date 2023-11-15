package gdf

import "fmt"

type ColorSpace Name

const (
	DEVICE_GRAY ColorSpace = "DeviceGray"
	DEVICE_RGB  ColorSpace = "DeviceRGB"
	DEVICE_CMYK ColorSpace = "DeviceCMYK"
	PATTERN     ColorSpace = "Pattern"
)

type Color interface {
	color()
}

// Sets c's stroking color (SColor) to cl and sets its SColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColorStroke(cl Color) {
	c.SColor = cl
	switch v := cl.(type) {
	case GColor:
		c.SColorSpace = DEVICE_GRAY
		fmt.Fprintf(c.buf, "%f G\n", v)
	case RGBColor:
		c.SColorSpace = DEVICE_RGB
		fmt.Fprintf(c.buf, "%f %f %f RG\n", v.R, v.G, v.B)
	case CMYKColor:
		c.SColorSpace = DEVICE_CMYK
		fmt.Fprintf(c.buf, "%f %f %f %f K\n", v.C, v.M, v.Y, v.K)
	}
}

// Sets c's non-stroking color (NColor) to cl and sets its NColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColor(cl Color) {
	c.NColor = cl
	switch v := cl.(type) {
	case GColor:
		c.NColorSpace = DEVICE_GRAY
		fmt.Fprintf(c.buf, "%f g\n", v)
	case RGBColor:
		c.NColorSpace = DEVICE_RGB
		fmt.Fprintf(c.buf, "%f %f %f rg\n", v.R, v.G, v.B)
	case CMYKColor:
		c.NColorSpace = DEVICE_CMYK
		fmt.Fprintf(c.buf, "%f %f %f %f k\n", v.C, v.M, v.Y, v.K)
	}
}

// Grayscale color; must be in [0,1].
type GColor float64

const (
	BLACK GColor = 0
	WHITE GColor = 1
	GRAY  GColor = .5
)

func (g GColor) color() {}

// RGB Color; R,G, and B must be in [0,1].
type RGBColor struct {
	R, G, B float64
}

func (r RGBColor) color() {}

var (
	RED   = RGBColor{1, 0, 0}
	GREEN = RGBColor{0, 1, 0}
	BLUE  = RGBColor{0, 0, 1}
)

// CMYK Color; C, M, Y, and K must be in [0,1].
type CMYKColor struct {
	C, M, Y, K float64
}

func (c CMYKColor) color() {}

var (
	CYAN       = CMYKColor{1, 0, 0, 0}
	MAGENTA    = CMYKColor{0, 1, 0, 0}
	YELLOW     = CMYKColor{0, 0, 1, 0}
	CMYK_BLACK = CMYKColor{0, 0, 0, 1}
)
