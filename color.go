package gdf

import "fmt"

type ColorSpace Name

const (
	DeviceGray ColorSpace = "DeviceGray"
	DeviceRGB  ColorSpace = "DeviceRGB"
	DeviceCMYK ColorSpace = "DeviceCMYK"
	Pattern    ColorSpace = "Pattern"
)

type Color interface {
	color() // this is required here to prevent user-defined color types, but it isn't actually used
}

// Sets c's stroking color (SColor) to cl and sets its SColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColorStroke(cl Color) {
	c.SColor = cl
	switch v := cl.(type) {
	case GColor:
		c.SColorSpace = DeviceGray
		fmt.Fprintf(c.buf, "%f G\n", v)
	case RGBColor:
		c.SColorSpace = DeviceRGB
		fmt.Fprintf(c.buf, "%f %f %f RG\n", v.R, v.G, v.B)
	case CMYKColor:
		c.SColorSpace = DeviceCMYK
		fmt.Fprintf(c.buf, "%f %f %f %f K\n", v.C, v.M, v.Y, v.K)
	}
}

// Sets c's non-stroking color (NColor) to cl and sets its NColorSpace to cl's ColorSpace.
func (c *ContentStream) SetColor(cl Color) {
	c.NColor = cl
	switch v := cl.(type) {
	case GColor:
		c.NColorSpace = DeviceGray
		fmt.Fprintf(c.buf, "%f g\n", v)
	case RGBColor:
		c.NColorSpace = DeviceRGB
		fmt.Fprintf(c.buf, "%f %f %f rg\n", v.R, v.G, v.B)
	case CMYKColor:
		c.NColorSpace = DeviceCMYK
		fmt.Fprintf(c.buf, "%f %f %f %f k\n", v.C, v.M, v.Y, v.K)
	}
}

// Grayscale color; must be in [0,1].
type GColor float64

const (
	Black GColor = 0
	Gray  GColor = .5
	White GColor = 1
)

func (g GColor) color() {}

// RGB Color; R,G, and B must be in [0,1].
type RGBColor struct {
	R, G, B float64
}

func (r RGBColor) color() {}

var (
	Red   = RGBColor{1, 0, 0}
	Green = RGBColor{0, 1, 0}
	Blue  = RGBColor{0, 0, 1}
)

// CMYK Color; C, M, Y, and K must be in [0,1].
type CMYKColor struct {
	C, M, Y, K float64
}

func (c CMYKColor) color() {}

var (
	Cyan      = CMYKColor{1, 0, 0, 0}
	Magenta   = CMYKColor{0, 1, 0, 0}
	Yellow    = CMYKColor{0, 0, 1, 0}
	CMYKBlack = CMYKColor{0, 0, 0, 1}
)
