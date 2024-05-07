package gdf

import "fmt"

// Extent returns the extent in font units of t, when set as a single line in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling of the text.
func (c *ContentStream) Extent(t []rune) float64 {
	var ext float64
	var spaceCount float64
	var i int
	for ; i < len(t)-1; i++ {
		adv, kern := c.Font.ShapedGlyphAdv(t[i], t[i+1])
		if t[i] == ' ' {
			spaceCount++
		}
		ext += float64(adv + kern)
	}
	ext += float64(c.Font.GlyphAdvance(t[i]))
	if t[i] == ' ' {
		spaceCount++
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// RawExtent returns the extent in font units of t, when set as a single line in c's current font. The returned value
// accounts for word spacing and horizontal scaling of the text, but does not account for kerning.
func (c *ContentStream) RawExtent(t []rune) float64 {
	var ext float64
	var spaceCount float64
	for i := 0; i < len(t); i++ {
		ext += float64(c.Font.GlyphAdvance(t[i]))
		if t[i] == ' ' {
			spaceCount++
		}
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100
}

// ExtentKerns returns the extent in font units of t, when set as a single line and using the supplied kerning in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling of the text.
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
