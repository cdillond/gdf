package gdf

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func calculateWidths(f *Font) {
	var charWidths [256]int
	for char, adv := range f.charset {
		charWidths[rtoc(char)] = adv
	}
	// find first and last used chars
	var fc int
	for fc < len(charWidths) && charWidths[fc] == 0 {
		fc++
	}
	var lc = len(charWidths) - 1
	for lc > -1 && charWidths[lc] == 0 {
		lc--
	}
	if fc > lc {
		fc, lc = 0, 0
	}
	f.firstChar = fc
	f.lastChar = lc
	f.widths = charWidths[f.firstChar : f.lastChar+1]
}

// GlyphAdvance returns the advance of r in font units.
func (f *Font) GlyphAdvance(r rune) int {
	adv, ok := f.charset[r]
	if ok {
		return adv
	}
	gid, err := f.SFNT.GlyphIndex(nil, r)
	if err != nil || gid == 0 {
		// try an encoded version instead
		f.charset[r] = 0
		return 0
	}

	adv26_6, err := f.SFNT.GlyphAdvance(f.buf, gid, 1000, font.HintingNone)
	if err != nil {
		f.charset[r] = 0
		return 0
	}
	f.charset[r] = int(adv26_6)
	return int(adv26_6)
}

// ShapedGlyphAdv returns the advance and kerning of r1 when set before r2.
func (f *Font) ShapedGlyphAdv(r1, r2 rune) (adv int, kern int) {
	adv = f.GlyphAdvance(r1)
	gid1, err := f.SFNT.GlyphIndex(f.buf, r1)
	if err != nil {
		return adv, 0
	}
	gid2, err := f.SFNT.GlyphIndex(f.buf, r2)
	if err != nil {
		return adv, 0
	}
	fpKern, err := f.SFNT.Kern(f.buf, gid1, gid2, 1000, 0)
	if err != nil {
		return adv, 0
	}
	return adv, int(fpKern)
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

// AscDesc returns the ascent and descent, in font units, of the glyph corresponding to the given rune.
func (f *Font) AscDesc(r rune) (asc float64, desc float64) {
	gid, err := f.SFNT.GlyphIndex(f.buf, r)
	if err != nil {
		return 0, 0
	}
	bounds, _, _ := f.SFNT.GlyphBounds(nil, gid, 1000, 0)
	// In Go's standard graphics library, y increases downward, so the values are
	// the reverse of what it seems they should be.
	return float64(-bounds.Min.Y), float64(bounds.Max.Y)
}

// TextAscDesc returns the maximum ascent and descent, in font units, of the glyphs in the given text.
func (f *Font) TextAscDesc(text []rune) (asc float64, desc float64) {
	for i := range text {
		a, d := f.AscDesc(text[i])
		if a > asc {
			asc = a
		}
		if d > desc {
			desc = d
		}
	}
	return asc, desc
}
