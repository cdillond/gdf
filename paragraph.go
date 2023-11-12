package gdf

import (
	"fmt"
	"strings"
)

type Paragraph struct {
	Lines   [][]rune
	Difs    []float64 // must be the same len as Lines
	Kerns   []int
	Advs    []int
	Hyphens map[int]bool // indices of lines ending with hyphens
	Indent  bool
	MaxW    float64
}

func WriteParagraph(p Paragraph, startIndex int, to *ContentStream) int {
	if startIndex >= len(p.Lines) {
		return -1
	}
	i := startIndex
	var j int // get Kerns index
	for ; j < i; j++ {
		j += len(p.Lines[j])
	}

	if i == 0 && p.Indent {
		tabW := float64(5 * GlyphAdvance(' ', to.Font))
		to.Td(FUToPt(tabW, to.FontSize), 0)
		var line []rune
		var kerns []int
		if exists, ok := p.Hyphens[i]; ok {
			line = make([]rune, len(p.Lines[i]))
			copy(line, p.Lines[i])
			kerns = make([]int, len(p.Lines[i]))
			copy(kerns, p.Kerns[j+1:j+len(line)+1])
			if !exists {
				line = append(line, '\u00AD')
				kerns = append(kerns, 0)
			}
		} else {
			line = p.Lines[i]
			kerns = p.Kerns[j+1 : j+len(line)+1]
		}
		to.TJSpace(line, kerns, p.Difs[i])
		j += len(p.Lines[i])
		to.TStar()
		to.Td(FUToPt(-1*tabW, to.FontSize), 0)
		i++
	}
	for ; i < len(p.Lines); i++ {
		var line []rune
		if exists, ok := p.Hyphens[i]; ok {
			line = make([]rune, len(p.Lines[i]))
			copy(line, p.Lines[i])
			if !exists {
				line = append(line, '\u00AD')
			}
		} else {
			line = p.Lines[i]
		}
		err := to.TJSpace(line, p.Kerns[j+1:j+len(line)+1], p.Difs[i])
		if err != nil {
			fmt.Println(err)
		}
		j += len(p.Lines[i])
		to.TStar()
	}
	return i - startIndex
}

func ParseParagraph(src []rune, indent bool, maxW float64, ts TextState) (Paragraph, error) {
	// trim leading whitespace
	var text []rune
	for i := 0; i < len(src); i++ {
		switch src[i] {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			text = src[i:]
		}
		break
	}
	if len(text) == 0 {
		return Paragraph{}, fmt.Errorf("text cannot be empty or contain only whitespace")
	}

	// trim trailing whitespace
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			text = text[:i+1]
		}
		break
	}

	// replace any non-space whitespace characters with \u0020
	var spaceCount int
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\t', '\r', '\n':
			text[i] = ' '
			spaceCount++
		case ' ':
			spaceCount++
		}
	}
	// parse with a line parser instead if there are fewer than 2 spaces
	if spaceCount < 2 {
		return Paragraph{}, fmt.Errorf("paragraph must contain at least two spaces")
	}

	// the last line must contain at least two words
	var lastLineIndex int
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == ' ' {
			if lastLineIndex != 0 {
				lastLineIndex = i
				break
			}
			lastLineIndex = i
		}
	}
	if lastLineIndex == 0 {
		return Paragraph{}, fmt.Errorf("this should not happen")
	}

	P := Paragraph{
		Hyphens: make(map[int]bool),
		Indent:  indent,
		MaxW:    maxW,
	}
	var start int
	var lineExt int
	var stretchBP, squishBP int
	var stretchExt, squishExt int
	var i int
	defaultSpace := GlyphAdvance(' ', ts.Font)
	// treat the first line differently if indented
	if indent {
		maxIndent := maxW - float64(defaultSpace)*5
		for ; i < lastLineIndex; i++ {
			r := text[i]
			adv, kern := ShapedGlyphAdv(r, text[i+1], ts.Font)
			P.Kerns = append(P.Kerns, kern)
			P.Advs = append(P.Advs, adv)
			if float64(lineExt) < maxIndent && r == ' ' && (start < lastLineIndex && i <= lastLineIndex) {
				stretchBP = i
				stretchExt = lineExt
			}
			if float64(lineExt) >= maxIndent && (r == ' ' || i >= lastLineIndex) {
				squishBP = i
				squishExt = lineExt

				// if exisiting, text[stretchBP+1+n] is already a hyphen
				n, existing := intraWordBP([]byte(string(text[stretchBP+1 : squishBP])))
				if n > 0 {
					squishBP = stretchBP + 1 + n + 1
					squishExt = stretchExt
					if !existing {
						// adjust kern of letter before hyphen
						_, kern = ShapedGlyphAdv(text[stretchBP+1+n], '\u002D', ts.Font)
						squishExt += kern - P.Kerns[stretchBP+1+n]
						//P.Kerns[stretchBP+1+n] = kern
						// add advance of hyphen
						squishExt += GlyphAdvance('\u002D', ts.Font)
					}
					for j := stretchBP; j <= stretchBP+n+1; j++ {
						squishExt += P.Kerns[j]
						squishExt += P.Advs[j]
					}
				}

				stretchSpaces := strings.Count(string(text[start:stretchBP]), " ")
				squishSpaces := strings.Count(string(text[start:squishBP]), " ")
				var stretchDif, squishDif float64
				if stretchSpaces != 0 {
					stretchDif = (maxIndent - float64(stretchExt)) / float64(stretchSpaces)
				}
				if squishSpaces != 0 {
					squishDif = (maxIndent - float64(squishExt)) / float64(squishSpaces)
				}
				if -4*squishDif < stretchDif && -squishDif <= float64(defaultSpace)-100 { // preference stretching over squishing
					P.Lines = append(P.Lines, text[start:squishBP])
					P.Difs = append(P.Difs, squishDif)
					start = squishBP + 1
					lineExt = 0
					if n > 0 {
						start = squishBP
						P.Hyphens[len(P.Lines)-1] = existing
						for j := start; j < i; j++ {
							lineExt += P.Kerns[j]
							lineExt += P.Advs[j]
						}
						lineExt += adv
					}

				} else {
					P.Lines = append(P.Lines, text[start:stretchBP])
					P.Difs = append(P.Difs, stretchDif)
					start = stretchBP + 1
					lineExt = 0
					for j := start; j <= i; j++ {
						lineExt += P.Kerns[j]
						lineExt += P.Advs[j]
					}
					i++
				}
				break
			}
			lineExt += adv + kern
		}
	}

	// ultimately, we need to avoid a last line that starts and ends after lastLineIndex
	// this means that the second to last line cannot end after lastLineIndex IF it does not end at len(text)-1
	for ; i < len(text)-1; i++ {
		r := text[i]
		adv, kern := ShapedGlyphAdv(r, text[i+1], ts.Font)
		P.Kerns = append(P.Kerns, kern)
		P.Advs = append(P.Advs, adv)
		if float64(lineExt) < maxW && r == ' ' && (start < lastLineIndex && i <= lastLineIndex) {
			stretchBP = i
			stretchExt = lineExt
		}
		if float64(lineExt) >= maxW && (r == ' ' || i >= lastLineIndex) {
			squishBP = i
			squishExt = lineExt

			var n int
			var existing bool
			// if exisiting, text[stretchBP+1+n] is already a hyphen
			n, existing = intraWordBP([]byte(string(text[stretchBP+1 : squishBP])))

			if n > 0 {
				squishBP = stretchBP + 1 + n + 1
				squishExt = stretchExt
				if !existing {
					// adjust kern of letter before hyphen
					_, kern = ShapedGlyphAdv(text[stretchBP+1+n], '\u002D', ts.Font)
					squishExt += kern - P.Kerns[stretchBP+1+n]
					//P.Kerns[stretchBP+1+n] = kern
					// add advance of hyphen
					squishExt += GlyphAdvance('\u002D', ts.Font)
				}
				for j := stretchBP; j <= stretchBP+n+1; j++ {
					squishExt += P.Kerns[j]
					squishExt += P.Advs[j]
				}
			}

			stretchSpaces := strings.Count(string(text[start:stretchBP]), " ")
			squishSpaces := strings.Count(string(text[start:squishBP]), " ")
			var stretchDif, squishDif float64
			if stretchSpaces != 0 {
				stretchDif = (maxW - float64(stretchExt)) / float64(stretchSpaces)
			}
			if squishSpaces != 0 {
				squishDif = (maxW - float64(squishExt)) / float64(squishSpaces)
			}

			if -4*squishDif < stretchDif && -squishDif <= float64(defaultSpace)-100 && (start < lastLineIndex && i < lastLineIndex) { // preference stretching over squishing
				// if squish is used, set the lineExt to 0
				// and only recalculate it if the line is hyphenated
				P.Lines = append(P.Lines, text[start:squishBP])
				P.Difs = append(P.Difs, squishDif)
				start = squishBP + 1
				lineExt = 0
				if n > 0 {
					start = squishBP
					P.Hyphens[len(P.Lines)-1] = existing
					for j := start; j < i; j++ {
						lineExt += P.Kerns[j]
						lineExt += P.Advs[j]
					}
					lineExt += adv
				}
			} else {
				// if stretch is used, recalculate the lineExt between
				// the stretch BP and the current position
				P.Lines = append(P.Lines, text[start:stretchBP])
				P.Difs = append(P.Difs, stretchDif)
				start = stretchBP + 1
				lineExt = 0

				for j := start; j <= i; j++ {
					lineExt += P.Kerns[j]
					lineExt += P.Advs[j]
				}
			}
			if start >= lastLineIndex {
				break
			}
			continue
		}
		lineExt += adv + kern
	}
	// append the final line, with no justification
	if start < len(text)-1 {
		P.Lines = append(P.Lines, text[start:])
		P.Difs = append(P.Difs, 0)
	}

	return P, nil
}
