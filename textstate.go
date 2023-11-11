package gdf

import (
	"fmt"
)

type TextState struct {
	*Font
	FontSize  float64 // points
	CharSpace float64 // points
	WordSpace float64 // points
	Scale     float64 // percent of normal width
	Leading   float64 // points
	Rise      float64 // points
	RenderMode
}

type RenderMode uint

const (
	TR_FILL RenderMode = iota
	TR_STROKE
	TR_FILL_STROKE
	TR_INVISIBLE
	TR_FILL_ADD_PATH
	TR_STROKE_ADD_PATH
	TR_FILL_STROKE_ADD_PATH
	TR_ADD_PATH
)

// sets the character spacing
func (c *ContentStream) Tc(f float64) {
	c.CharSpace = f
	fmt.Fprintf(c.buf, "%f Tc\n", f)
}

// sets the word spacing
func (c *ContentStream) Tw(f float64) {
	c.WordSpace = f
	fmt.Fprintf(c.buf, "%f Tw\n", f)
}

// sets the scale
func (c *ContentStream) Tz(f float64) {
	c.Scale = f
	fmt.Fprintf(c.buf, "%f Tz\n", f)
}

// sets the leading
func (c *ContentStream) TL(f float64) {
	c.Leading = f
	fmt.Fprintf(c.buf, "%f TL\n", f)
}

// sets the font and font size
func (c *ContentStream) Tf(size float64, font *Font) {
	var i int
	for ; i < len(c.Parent.Fonts); i++ {
		if c.Parent.Fonts[i] == font {
			break
		}
	}
	if i == len(c.Parent.Fonts) {
		c.Parent.Fonts = append(c.Parent.Fonts, font)
	}
	c.FontSize = size
	c.Font = font
	fmt.Fprintf(c.buf, "/F%d %f Tf\n", i, size)
}

// sets the text rendering mode
func (c *ContentStream) Tr(r RenderMode) {
	c.RenderMode = r
	fmt.Fprintf(c.buf, "%d Tr\n", r)
}

// sets the text rise
func (c *ContentStream) Ts(f float64) {
	c.Rise = f
	fmt.Fprintf(c.buf, "%f Ts\n", f)
}
