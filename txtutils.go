package gdf

import "fmt"

// Extent returns the extent (width) in font units of text, if it were to be set as a single line
// in c's current font. The returned value accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) Extent(text []rune) float64 {
	var ext float64
	var spaceCount float64
	var i int
	for ; i < len(text)-1; i++ {
		adv, kern := c.Font.ShapedGlyphAdv(text[i], text[i+1])
		if text[i] == ' ' {
			spaceCount++
		}
		ext += float64(adv + kern)
	}
	ext += float64(c.Font.GlyphAdvance(text[i]))
	if text[i] == ' ' {
		spaceCount++
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// RawExtent returns the extent in font units of text, when set in c's current font as a single line. The returned value
// accounts for word spacing and horizontal scaling, but does not account for kerning.
func (c *ContentStream) RawExtent(text []rune) float64 {
	var ext float64
	var spaceCount float64
	for i := 0; i < len(text); i++ {
		ext += float64(c.Font.GlyphAdvance(text[i]))
		if text[i] == ' ' {
			spaceCount++
		}
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// rawExtentPts returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for word spacing and horizontal scaling, but does not account for kerning.
func (c *ContentStream) rawExtentPts(t []rune) float64 {
	return FUToPt(c.RawExtent(t), c.FontSize)
}

// ExtentKerns returns the extent in font units of t, when set in c's current font as a single line
// using the supplied kerning values. The returned value accounts for word spacing, and horizontal scaling.
func (c *ContentStream) ExtentKerns(t []rune, kerns []int) (float64, error) {
	if len(t) != len(kerns) {
		return 0, fmt.Errorf("equal number of runes and kerns required. rune count: %d, kern count: %d", len(t), len(kerns))
	}
	var ext float64
	var spaceCount float64
	for i := 0; i < len(t); i++ {
		ext += float64(c.Font.GlyphAdvance(t[i])) + float64(kerns[i])
		if t[i] == ' ' {
			spaceCount++
		}
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100, nil
}

// extentKernsPts returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for kerning, word spacing, and horizontal scaling.
func (c *ContentStream) extentKernsPts(t []rune, kerns []int) (float64, error) {
	fu, err := c.ExtentKerns(t, kerns)
	if err != nil {
		return *new(float64), err
	}
	return FUToPt(fu, c.FontSize), nil
}
