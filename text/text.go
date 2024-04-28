package text

import (
	"fmt"

	"github.com/cdillond/gdf"
)

// ExtentFU returns the extent in font units of t, when set in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func ExtentFU(c *gdf.ContentStream, t []rune) float64 {
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

// ExtentPts returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for kerning, word spacing, and horizontal scaling.
func ExtentPts(c *gdf.ContentStream, t []rune) float64 {
	return gdf.FUToPt(ExtentFU(c, t), c.FontSize)
}

// RawExtentFU returns the extent in font units of t, when set in c's current font. The returned value
// accounts for word spacing and horizontal scaling, but does not account for kerning.
func RawExtentFU(c *gdf.ContentStream, t []rune) float64 {
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

// RawExtentPts returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for word spacing and horizontal scaling, but does not account for kerning.
func RawExtentPts(c *gdf.ContentStream, t []rune) float64 {
	return gdf.FUToPt(RawExtentFU(c, t), c.FontSize)
}

// ExtentKernsFU returns the extent in font units of t, when set in c's current font. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func ExtentKernsFU(c *gdf.ContentStream, t []rune, kerns []int) (float64, error) {
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

// ExtentKernsPts returns the extent in points of t, when set in c's current font at the current font size.
// The returned value accounts for kerning, word spacing, and horizontal scaling.
func ExtentKernsPts(c *gdf.ContentStream, t []rune, kerns []int) (float64, error) {
	fu, err := c.ExtentKernsFU(t, kerns)
	if err != nil {
		return *new(float64), err
	}
	return gdf.FUToPt(fu, c.FontSize), nil
}

// ExtentFontFU returns the extent in font units of t, when set in font f. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func ExtentFontFU(c *gdf.ContentStream, t []rune, f *gdf.Font) float64 {
	var ext float64
	var spaceCount float64
	var i int
	for ; i < len(t)-1; i++ {
		adv, kern := f.ShapedGlyphAdv(t[i], t[i+1])
		if t[i] == ' ' {
			spaceCount++
		}
		ext += float64(adv + kern)
	}
	ext += float64(f.GlyphAdvance(t[i]))
	if t[i] == ' ' {
		spaceCount++
	}
	ext += c.WordSpace * spaceCount
	return ext * c.HScale / 100

}

// ExtentFontPts returns the extent in points of text, when set in font f at the given size. The returned value
// accounts for kerning, word spacing, and horizontal scaling.
func ExtentFontPts(c *gdf.ContentStream, t []rune, f *gdf.Font, size float64) float64 {
	return gdf.FUToPt(ExtentFontFU(c, t, f), size)
}

// CenterH returns the x offset, in points, from the start of rect, needed to center t horizontally, if drawn with c's current
// font at c's current font size.
func CenterH(c *gdf.ContentStream, t []rune, rect gdf.Rect) float64 {
	ext := c.ExtentPts(t)
	dif := rect.URX - rect.LLX - ext
	return dif / 2
}

// CenterV returns the y offset, in points, from the start of rect, needed to center a line of text vertically based on the text's height.
// This is a naive approach.
func CenterV(height float64, rect gdf.Rect) float64 {
	return -(rect.URY - rect.LLY - height) / 2
}
