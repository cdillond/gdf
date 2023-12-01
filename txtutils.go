package gdf

import "fmt"

// Returns the extent in font units of t, when set in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentFU(t []rune) float64 {
	var ext float64
	var spaceCount float64
	var i int
	for ; i < len(t)-1; i++ {
		adv, kern := ShapedGlyphAdv(t[i], t[i+1], c.Font)
		if t[i] == ' ' {
			spaceCount++
		}
		ext += float64(adv + kern)
	}
	ext += float64(GlyphAdvance(t[i], c.Font))
	if t[i] == ' ' {
		spaceCount++
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// Returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentPts(t []rune) float64 {
	return FUToPt(c.ExtentFU(t), c.FontSize)
}

// Returns the extent in font units of t, when set in c's current font. The returned value
// accounts for word spacing and horizontal scaling, but does not account for kerning.
func (c *ContentStream) RawExtentFU(t []rune) float64 {
	var ext float64
	var spaceCount float64
	for i := 0; i < len(t); i++ {
		ext += float64(GlyphAdvance(t[i], c.Font))
		if t[i] == ' ' {
			spaceCount++
		}
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// Returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for word spacing and horizontal scaling, but does not account for kerning.
func (c *ContentStream) RawExtentPts(t []rune) float64 {
	return FUToPt(c.RawExtentFU(t), c.FontSize)
}

// Returns the extent in font units of t, when set in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentKernsFU(t []rune, kerns []int) (float64, error) {
	if len(t) != len(kerns) {
		return 0, fmt.Errorf("equal number of runes and kerns required. rune count: %d, kern count: %d", len(t), len(kerns))
	}
	var ext float64
	var spaceCount float64
	for i := 0; i < len(t); i++ {
		ext += float64(GlyphAdvance(t[i], c.Font)) + float64(kerns[i])
		if t[i] == ' ' {
			spaceCount++
		}
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100, nil
}

// Returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentKernsPts(t []rune, kerns []int) (float64, error) {
	fu, err := c.ExtentKernsFU(t, kerns)
	if err != nil {
		return *new(float64), err
	}
	return FUToPt(fu, c.FontSize), nil
}

// Returns the extent in font units of t, when set in font f. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentFontFU(t []rune, f *Font) float64 {
	var ext float64
	var spaceCount float64
	var i int
	for ; i < len(t)-1; i++ {
		adv, kern := ShapedGlyphAdv(t[i], t[i+1], f)
		if t[i] == ' ' {
			spaceCount++
		}
		ext += float64(adv + kern)
	}
	ext += float64(GlyphAdvance(t[i], f))
	if t[i] == ' ' {
		spaceCount++
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100

}

// Returns the extent in points of text, when set in font f at the given size. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) ExtentFontPts(t []rune, f *Font, size float64) float64 {
	return FUToPt(c.ExtentFontFU(t, f), size)
}

// Returns the x offset, in points, from the start of rect, needed to center t horizontally, if drawn with c's current
// font at c's current font size.
func (c *ContentStream) CenterH(t []rune, rect Rect) float64 {
	ext := c.ExtentPts(t)
	dif := rect.URX - rect.LLX - ext
	return dif / 2
}

// Returns the y offset, in points, from the start of rect, needed to center a line of text vertically based on the text's height.
func CenterV(height float64, rect Rect) float64 {
	return -(rect.URY - rect.LLY - height) / 2
}
