package gdf

import (
	"fmt"
	"strings"
)

type ShapedLine struct {
	BreakPoint int
	Dif        float64
	Text       []rune
	Kerns      []float64
	Advs       []int
	SquishDif  float64
	StretchDif float64
	StretchBP  int
	SquishBP   int
	Hyphenated bool
}

func BreakShaped(text []rune, ts TextState, maxW float64) []ShapedLine {
	// maxW is in font units
	lines := []ShapedLine{}
	sl := ShapedLine{}
	var newline bool
	tmpText := []rune{}
	tmpAdvs := []int{}
	tmpKerns := []float64{}

	for i := 0; i < len(text)-1; i++ {
		if text[i] == '\n' || text[i] == '\r' {
			newline = true
		}

		tmpText = append(tmpText, text[i])
		adv, kern := ShapedGlyphAdv(text[i], text[i+1], ts.Font)
		tmpKerns = append(tmpKerns, float64(kern))
		tmpAdvs = append(tmpAdvs, adv)

		if text[i] == ' ' && ShapedTextExtent(tmpText[:len(tmpText)-1], ts.Font) < maxW {
			sl.BreakPoint = len(tmpText) - 1
			sl.StretchBP = sl.BreakPoint
			numspaces := strings.Count(string(tmpText[:sl.BreakPoint]), " ")
			if numspaces != 0 {
				dif := (maxW - ShapedTextExtent(tmpText[:sl.BreakPoint], ts.Font)) / float64(numspaces)
				sl.Dif = dif
				sl.StretchDif = dif
				dstText := make([]rune, len(tmpText)-1)
				dstKerns := make([]float64, len(tmpKerns)-1)
				dstAdvs := make([]int, len(tmpAdvs)-1)
				copy(dstText, tmpText)
				copy(dstKerns, tmpKerns)
				copy(dstAdvs, tmpAdvs)
				sl.Advs = dstAdvs
				sl.Kerns = dstKerns
				sl.Text = dstText
			}
		}
		if ShapedTextExtent(tmpText, ts.Font) >= maxW && (text[i] == ' ' || newline) {
			extraText := []rune{}
			extraAdvs := []int{}
			extraKerns := []float64{}

			sl.SquishBP = len(tmpText) - 1
			// calculate hyphenated text
			hyphIndex, existing := IntraWordBP([]byte(string(tmpText[sl.StretchBP:sl.SquishBP])))
			if hyphIndex > 0 {
				var hyphchar rune
				sl.SquishBP = sl.StretchBP + hyphIndex + 1

				if !existing {
					dst1 := make([]rune, len(tmpText[sl.SquishBP:])-1)
					dst2 := make([]int, len(tmpAdvs[sl.SquishBP:])-1)
					dst3 := make([]float64, len(tmpKerns[sl.SquishBP:])-1)
					copy(dst1, tmpText[sl.SquishBP:])
					copy(dst2, tmpAdvs[sl.SquishBP:])
					copy(dst3, tmpKerns[sl.SquishBP:])
					extraText = append(dst1, ' ')
					extraAdvs = append(dst2, GlyphAdvance(' ', ts.Font))
					extraKerns = append(dst3, 0)

					hyphchar = '\u002D' //'\u00AD'
					hyphKerns := append(tmpKerns[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], 0)
					hyphAdvs := append(tmpAdvs[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], GlyphAdvance(hyphchar, ts.Font))
					hyphText := append(tmpText[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], hyphchar)

					tmpKerns = append(tmpKerns[:sl.StretchBP+1], hyphKerns...)
					tmpAdvs = append(tmpAdvs[:sl.StretchBP+1], hyphAdvs...)
					tmpText = append(tmpText[:sl.StretchBP+1], hyphText...)

				} else {
					// TO DO: FIX THIS
					dst1 := make([]rune, len(tmpText[sl.SquishBP:])-1)
					dst2 := make([]int, len(tmpAdvs[sl.SquishBP:])-1)
					dst3 := make([]float64, len(tmpKerns[sl.SquishBP:])-1)
					copy(dst1, tmpText[sl.SquishBP:])
					copy(dst2, tmpAdvs[sl.SquishBP:])
					copy(dst3, tmpKerns[sl.SquishBP:])
					extraText = append(dst1, ' ')
					extraAdvs = append(dst2, GlyphAdvance(' ', ts.Font))
					extraKerns = append(dst3, 0)

					hyphchar = '\u002D'
					hyphKerns := append(tmpKerns[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], 0)
					hyphAdvs := append(tmpAdvs[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], GlyphAdvance(hyphchar, ts.Font))
					hyphText := append(tmpText[sl.StretchBP+1:sl.StretchBP+hyphIndex+1], hyphchar)

					tmpKerns = append(tmpKerns[:sl.StretchBP], hyphKerns...)
					tmpAdvs = append(tmpAdvs[:sl.StretchBP], hyphAdvs...)
					tmpText = append(tmpText[:sl.StretchBP], hyphText...)
				}

			}

			// calculate squish dif
			numspaces := strings.Count(string(tmpText[:len(tmpText)-1]), " ")
			if numspaces != 0 && hyphIndex < 0 {
				dif := (maxW - ShapedTextExtent(tmpText[:len(tmpText)-1], ts.Font)) / float64(numspaces)
				sl.SquishDif = dif
			}
			if numspaces != 0 && hyphIndex >= 0 {
				dif := (maxW - ShapedTextExtent(tmpText, ts.Font)) / float64(numspaces)
				sl.SquishDif = dif
			}

			// determine which version to use
			if -2.5*sl.SquishDif <= sl.StretchDif {
				sl.Dif = sl.SquishDif
				sl.BreakPoint = sl.SquishBP
				if hyphIndex < 0 {
					dstText := make([]rune, len(tmpText)-1)
					dstKerns := make([]float64, len(tmpKerns)-1)
					dstAdvs := make([]int, len(tmpAdvs)-1)
					copy(dstText, tmpText)
					copy(dstKerns, tmpKerns)
					copy(dstAdvs, tmpAdvs)
					sl.Advs = dstAdvs
					sl.Kerns = dstKerns
					sl.Text = dstText

				} else {
					dstText := make([]rune, len(tmpText))
					dstKerns := make([]float64, len(tmpKerns))
					dstAdvs := make([]int, len(tmpAdvs))
					copy(dstText, tmpText)
					copy(dstKerns, tmpKerns)
					copy(dstAdvs, tmpAdvs)
					sl.Advs = dstAdvs
					sl.Kerns = dstKerns
					sl.Text = dstText
				}

			}
			if len(tmpText)-sl.BreakPoint > 1 {
				extraText = tmpText[sl.BreakPoint+1:]
				extraAdvs = tmpAdvs[sl.BreakPoint+1:]
				extraKerns = tmpKerns[sl.BreakPoint+1:]
			}

			lines = append(lines, sl)
			tmpText = extraText
			tmpAdvs = extraAdvs
			tmpKerns = extraKerns
			sl = ShapedLine{Text: extraText, Advs: extraAdvs, Kerns: extraKerns}
			fmt.Println(sl)
			newline = false
		} else if newline {
			sl.Dif = 0
			sl.Text = tmpText[:len(tmpText)-1]
			sl.Kerns = tmpKerns[:len(tmpKerns)-1]
			sl.Advs = tmpAdvs[:len(tmpAdvs)-1]
			lines = append(lines, sl)
			sl = ShapedLine{}
			tmpText = []rune{}
			tmpAdvs = []int{}
			tmpKerns = []float64{}
			newline = false
		}

	}
	sl.Dif = 0
	if text[len(text)-1] != ' ' && text[len(text)-1] != '\n' && text[len(text)-1] != '\r' && text[len(text)-1] != '\t' {
		sl.Text = append(sl.Text, text[len(text)-1])
		sl.Kerns = append(sl.Kerns, 0)
		sl.Advs = append(sl.Advs, GlyphAdvance(text[len(text)-1], ts.Font))
		sl.Dif = 0
	}

	lines = append(lines, sl)
	return lines
}

func PotentialOrphan(text []rune, ts TextState, maxW float64) []ShapedLine {
	if len(text) == 0 {
		return []ShapedLine{}
	}

	return []ShapedLine{}
}
