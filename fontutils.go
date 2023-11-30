package gdf

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func calculateWidths(f *Font) {
	bs := make([]int, 256)
	for char, adv := range f.Charset {
		b, err := f.enc.Bytes([]byte(string(char)))
		if err == nil && len(b) == 1 {
			bs[b[0]] = adv
		}
	}
	var fc, lc int
	var first bool
	for i, b := range bs {
		if b != 0 {
			lc = i
			if !first {
				fc = i
				first = true
			}
		}
	}
	f.FirstChar = fc
	f.LastChar = lc
	f.Widths = bs[f.FirstChar : f.LastChar+1]
}

// Returns the advance of r in font units.
func GlyphAdvance(r rune, f *Font) int {
	adv, ok := f.Charset[r]
	if ok {
		return adv
	}
	gid, err := f.GlyphIndex(nil, r)
	if err != nil || gid == 0 {
		// try an encoded version instead
		f.Charset[r] = 0
		return 0
	}

	adv26_6, err := f.GlyphAdvance(f.buf, gid, 1000, font.HintingNone)
	if err != nil {
		f.Charset[r] = 0
		return 0
	}
	f.Charset[r] = int(adv26_6)
	return int(adv26_6)
}

// Returns the advance and kerning of r1 when set before r2.
func ShapedGlyphAdv(r1, r2 rune, f *Font) (int, int) {
	adv := GlyphAdvance(r1, f)
	gid1, err := f.GlyphIndex(f.buf, r1)
	if err != nil {
		return adv, 0
	}
	gid2, err := f.GlyphIndex(f.buf, r2)
	if err != nil {
		return adv, 0
	}
	kern, err := f.Kern(f.buf, gid1, gid2, 1000, 0)
	if err != nil {
		return adv, 0
	}
	return adv, int(kern)
}

func fontBBox(font *sfnt.Font, buf *sfnt.Buffer) (fixed.Rectangle26_6, error) {
	bbox := *new(fixed.Rectangle26_6)
	for i := sfnt.GlyphIndex(0); i < sfnt.GlyphIndex(font.NumGlyphs()); i++ {
		bounds, _, err := font.GlyphBounds(buf, i, 1000, 0)
		if err != nil {
			return *new(fixed.Rectangle26_6), err
		}
		bbox = bbox.Union(bounds)
	}
	return bbox, nil
}

// Returns the ascent and descent of the glyph corresponding to the given rune.
func AscDesc(r rune, f *Font) (float64, float64) {
	gid, err := f.GlyphIndex(f.buf, r)
	if err != nil {
		return 0, 0
	}
	bounds, _, _ := f.GlyphBounds(nil, gid, 1000, 0)
	// In Go's standard graphics library, y increases downward, so the values are
	// the reverse of what it seems they should be.
	return float64(-bounds.Min.Y), float64(bounds.Max.Y)
}

// Returns the maximum ascent and descent of the glyphs in the given text.
func TextAscDesc(text []rune, f *Font) (float64, float64) {
	var maxA, maxD float64
	for i := range text {
		asc, desc := AscDesc(text[i], f)
		if asc > maxA {
			maxA = asc
		}
		if desc > maxD {
			maxD = desc
		}
	}
	return maxA, maxD
}
