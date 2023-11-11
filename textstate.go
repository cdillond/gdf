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

// Sets the content stream's character spacing.
func (c *ContentStream) Tc(f float64) {
	c.CharSpace = f
	fmt.Fprintf(c.buf, "%f Tc\n", f)
}

// Sets the content stream's word spacing.
func (c *ContentStream) Tw(f float64) {
	c.WordSpace = f
	fmt.Fprintf(c.buf, "%f Tw\n", f)
}

// Sets the content stream's horizontal text scaling percentage. The default value is 100.0 percent.
func (c *ContentStream) Tz(f float64) {
	c.Scale = f
	fmt.Fprintf(c.buf, "%f Tz\n", f)
}

// Sets the content stream's text leading (line height).
func (c *ContentStream) TL(f float64) {
	c.Leading = f
	fmt.Fprintf(c.buf, "%f TL\n", f)
}

// Sets the current font size and font.
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

// Sets the current text rendering mode.
func (c *ContentStream) Tr(r RenderMode) {
	c.RenderMode = r
	fmt.Fprintf(c.buf, "%d Tr\n", r)
}

// Sets the current text rise.
func (c *ContentStream) Ts(f float64) {
	c.Rise = f
	fmt.Fprintf(c.buf, "%f Ts\n", f)
}
