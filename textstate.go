package gdf

import (
	"fmt"
)

type TextState struct {
	*Font
	FontSize  float64 // points
	CharSpace float64 // points
	WordSpace float64 // points
	HScale    float64 // horizontal scale, as a percentage of the normal width
	Leading   float64 // points
	Rise      float64 // points
	RenderMode
}

type RenderMode uint

const (
	FillText RenderMode = iota
	StrokeText
	FillStrokeText
	Invisible
	FillTextAddPath
	StrokeTextAddPath
	FillStrokeTextAddPath
	AddTextPath
)

// Sets the content stream's character spacing (c.TextState.CharSpace) f.
func (c *ContentStream) SetCharSpace(f float64) {
	c.CharSpace = f
	fmt.Fprintf(c.buf, "%f Tc\n", f)
}

// Sets the content stream's word spacing (c.TextState.WordSpace) to f.
func (c *ContentStream) SetWordSpace(f float64) {
	c.WordSpace = f
	fmt.Fprintf(c.buf, "%f Tw\n", f)
}

// Sets the content stream's horizontal text scaling percentage (c.TextState.Scale) to f. The default value is 100.0 percent.
func (c *ContentStream) SetHScale(f float64) {
	c.HScale = f
	fmt.Fprintf(c.buf, "%f Tz\n", f)
}

// Sets the content stream's text leading/line height (c.TextState.Leading) to f.
func (c *ContentStream) SetLeading(f float64) {
	c.Leading = f
	fmt.Fprintf(c.buf, "%f TL\n", f)
}

// Sets the current font size and font (c.TextState.Font and c.TextState.FontSize) to size and font.
func (c *ContentStream) SetFont(size float64, font *Font) {
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

// Sets the current text rendering mode (c.TextState.RenderMode) to r.
func (c *ContentStream) SetRenderMode(r RenderMode) {
	c.RenderMode = r
	fmt.Fprintf(c.buf, "%d Tr\n", r)
}

// Sets the current text rise (c.TextState.RenderMode) to f.
func (c *ContentStream) SetRise(f float64) {
	c.Rise = f
	fmt.Fprintf(c.buf, "%f Ts\n", f)
}
