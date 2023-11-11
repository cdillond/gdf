package gdf

import (
	"bytes"
)

/*
	type LineWriter struct {
		Lines       []string
		SpaceWidths []float64
		n           int
	}

	type LW struct {
		Lines []ShapedLine
		n     int
	}

	func (l *LW) Draw(cs *ContentStream, startX, startY float64) int {
		to := NewTextObject(cs)
		to.BT()
		to.Tm(1, 0, 0, 1, startX, startY)
		var i int
		for ; to.f*to.d >= cs.Parent.Bottom; i++ {
			if i+l.n >= len(l.Lines) {
				l.n += i
				return i
			}
			err := to.TJSpace(l.Lines[l.n+i].Text, l.Lines[l.n+i].Kerns, l.Lines[l.n+i].Dif)
			if err != nil {
				fmt.Println(err.Error())
			}
			to.TStar()
			fmt.Println(to.f * to.d)
		}
		to.ET()
		return i
	}

	func (l *LineWriter) Draw(cs *ContentStream, startX, startY float64) int {
		spaceAdv := float64(GlyphAdvance(' ', cs.Font))
		to := NewTextObject(cs)
		to.BT()
		cs.Tf(12, cs.Font)
		cs.TL(14)
		to.Td(startX, startY)
		var i int
		for yPos := to.f; yPos >= cs.Parent.Bottom+cs.Parent.CropBox.LLY && l.n+i < len(l.Lines); {
			if len(l.Lines[l.n+i]) == 0 {
				to.TStar()
			} else {
				difs := []float64{}
				runs := []string{}
				tr := []rune(l.Lines[l.n+i])
				for j, run := range tr {
					if j == len(tr)-1 {
						break
					}
					_, kern := ShapedGlyphAdv(run, tr[j+1], cs.Font)
					difs = append(difs, float64(kern))
					runs = append(runs, string(run))
				}
				runs = append(runs, string(tr[len(tr)-1]))
				difs = append(difs, 0)

				// indent first lines
				if l.n+i == 0 || (l.n+i-1 > 0 && len(l.Lines[l.n+i-1]) == 0) {
					to.Td(5*FUToPt(float64(spaceAdv), cs.FontSize), 0)
					//to.TJShaped(runs, difs)
					to.TJSpace(runs, difs, spaceAdv)
					to.TStar()
					to.Td(-5*FUToPt(spaceAdv, cs.FontSize), 0)
				} else {
					to.TJSpace(runs, difs, spaceAdv)
					//to.TJShaped(runs, difs)
					//to.Tj(l.Lines[l.n+i])
					to.TStar()
				}

			} // 2 egg rolls; general tso's chicken; egg drop soup

			tmp := Mul(to.LineMatrix, Matrix{1, 0, 0, 1, 0, -to.Leading})
			yPos = tmp.f
			i++
		}
		l.n += i
		to.ET()
		return i
	}

// MaxWidth is specified in points

	func BreakLines(text string, ts TextState, maxWidth float64) LineWriter {
		broken := strings.Split(text, "\n")
		maxW := PtToFU(maxWidth, ts.FontSize)
		spaceAdv := float64(GlyphAdvance(' ', ts.Font))
		lines := []string{}
		difs := []float64{}
		for _, run := range broken {
			if len(run) == 0 {
				lines = append(lines, "")
				difs = append(difs, 0)
				continue
			}

			var start int
			var nextSpace, lastSpace int // indices of the next (squish) and last (stretch) breakpoint
			var lastExt, nextExt float64
			nextSpace = strings.Index(run, " ")
			if nextSpace == -1 {
				lines = append(lines, "")
				difs = append(difs, 0)
				continue
			}
			firstLine := true
			for {
				nextExt = ts.ShapedTextExtent(run[start:nextSpace])
				if (firstLine && nextExt > maxW-5*spaceAdv) || nextExt > maxW {
					// we're in business
					lastDif := maxW - lastExt
					nextDif := maxW - nextExt
					if firstLine {
						lastDif -= 5 * spaceAdv
						nextDif -= 5 * spaceAdv
					}
					if strings.Count(run[start:nextSpace], " ") < 1 {
						// we need to append last
						lines = append(lines, run[start:lastSpace])
						difs = append(difs, 0)
						start = lastSpace + 1
					} else {
						lastDifPerSpace := lastDif / float64(strings.Count(run[start:lastSpace], " "))
						nextDifPerSpace := nextDif / float64(strings.Count(run[start:nextSpace], " "))
						if lastDifPerSpace <= -1.5*nextDifPerSpace || -nextDifPerSpace >= 100 || -nextDifPerSpace >= spaceAdv {
							// stretch Saudi
							difs = append(difs, lastDifPerSpace)
							lines = append(lines, run[start:lastSpace])
							start = lastSpace + 1 // do not include the space
						} else {
							// check if the word can be broken
							ibp, existingHyphen := IntraWordBP([]byte(run[lastSpace+1 : nextSpace]))
							if ibp != -1 && !existingHyphen {
								hyphline0 := make([]byte, len(run[start:lastSpace+1+ibp+1])+1)
								copy(hyphline0, run[start:lastSpace+1+ibp+1])
								hyphline0[len(hyphline0)-1] = '\u00AD'
								hyphline := string(hyphline0)
								hext := ts.ShapedTextExtent(hyphline)
								hdfi := maxW - hext
								if firstLine {
									hdfi -= 5 * spaceAdv
								}
								difs = append(difs, hdfi/float64(strings.Count(hyphline, " ")))
								lines = append(lines, hyphline)
								start = lastSpace + 1 + ibp + 1
							} else if ibp != -1 && existingHyphen {
								hyphline0 := make([]byte, len(run[start:lastSpace+1+ibp+1]))
								copy(hyphline0, run[start:lastSpace+1+ibp+1])
								hyphline := string(hyphline0)
								hext := ts.ShapedTextExtent(hyphline)
								hdfi := maxW - hext
								if firstLine {
									hdfi -= 5 * spaceAdv
								}
								difs = append(difs, hdfi/float64(strings.Count(hyphline, " ")))
								lines = append(lines, hyphline)
								start = lastSpace + 1 + ibp + 1
							} else {
								// squish
								difs = append(difs, nextDifPerSpace)
								lines = append(lines, run[start:nextSpace])
								start = nextSpace + 1
							}

						}
					}
					firstLine = false
				}
				lastSpace = nextSpace
				lastExt = nextExt
				nextIndex := strings.IndexByte(run[nextSpace+1:], ' ')
				if nextIndex == -1 {
					// check if it overruns
					lineExt := ts.ShapedTextExtent(run[start:])
					adjMaxXFU := maxW
					if firstLine {
						adjMaxXFU = maxW - 5*spaceAdv
					}
					if lineExt > adjMaxXFU {
						// prefer squishing to orphaning; don't hyphenate
						squishDif := (adjMaxXFU - lineExt) / float64(strings.Count(run[start:], " "))
						if -squishDif >= spaceAdv-100 {
							// squishing would look bad, so go for the orphan
							// or... create a new line with 2 words and stretch the old line
							lastSpace := start + strings.LastIndexByte(run[start:nextSpace], ' ')
							//penultimateSpace := start + bytes.LastIndexByte(run[start:lastSpace], ' ')
							stretchLineExt := ts.ShapedTextExtent(run[start:lastSpace])
							stretchDif := (adjMaxXFU - stretchLineExt) / float64(strings.Count(run[start:lastSpace], " "))
							difs = append(difs, stretchDif)
							lines = append(lines, run[start:lastSpace])
							difs = append(difs, 0)
							lines = append(lines, run[lastSpace+1:])

						} else {
							difs = append(difs, squishDif)
							lines = append(lines, run[start:])
						}

						break
					}
					// it might be tricky to deal with orphans...
					if sp := strings.Count(run[start:], " "); sp == 0 {
						// we now need to recalculate the previous line
						lastLine := lines[len(lines)-1]
						lastLineNewBP := strings.LastIndexByte(lastLine, ' ')

						lines[len(lines)-1] = lastLine[:lastLineNewBP]
						adjMaxXFU := maxW
						// recalculate dif for previous line
						// 1. check if it has an indent and adjust accordingly
						if len(lines) == 1 || len(lines[len(lines)-2]) == 0 {
							adjMaxXFU = maxW - 5*spaceAdv
						}
						newLastExt := ts.ShapedTextExtent(lastLine[:lastLineNewBP])
						newLastDif := (adjMaxXFU - newLastExt) / float64(strings.Count(lastLine[:lastLineNewBP], " "))
						difs[len(difs)-1] = newLastDif

						// calculate dif for current last line
						// if the last line was squished, check if it has a hyphen
						finalLine := ""                        //[]byte{}
						lastWord := lastLine[lastLineNewBP+1:] //append([]byte{}, lastLine[lastLineNewBP+1:]...)
						if lastWord[len(lastWord)-1] == '\u00AD' {
							// remove soft hyphen and don't add a space
							finalLine += lastWord[:len(lastWord)-1] //= append(finalLine, lastWord[:len(lastWord)-1]...)
						} else if lastWord[len(lastWord)-1] == '\u002D' {
							// dont add a space for regular hyphens
							finalLine += lastWord //append(finalLine, lastWord...)
						} else {
							// add a space for all other characters
							finalLine += lastWord // append(finalLine, lastWord...)
							finalLine += " "      //append(finalLine, ' ')
						}
						finalLine += run[start:] //append(finalLine, run[start:]...)
						// this line now should not be indented
						// check if it overflows, and squish if so
						finalLineExt := ts.ShapedTextExtent(finalLine)
						if finalLineExt > maxW {
							finalDif := maxW - finalLineExt
							difs = append(difs, finalDif/float64(strings.Count(finalLine, " ")))
							lines = append(lines, finalLine)
							break
						}
						lines = append(lines, finalLine)
						difs = append(difs, 0)
						break
					}

					difs = append(difs, 0)
					lines = append(lines, run[start:])
					break
				}
				nextSpace = nextSpace + 1 + nextIndex
			}
		}
		return LineWriter{Lines: lines, SpaceWidths: difs}
	}
*/
func IsConsonant(b byte) bool {
	if ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z') {
		switch b {
		case 'A', 'a', 'E', 'e', 'I', 'i', 'O', 'o', 'U', 'u':
			return false
		default:
			return true
		}
	}
	return false
}

func IsVowel(b byte) bool {
	switch b {
	case 'A', 'a', 'E', 'e', 'I', 'i', 'O', 'o', 'U', 'u':
		return true
	default:
		return false
	}
}

func IsDigraph(b []byte) bool {
	if bytes.EqualFold(b, []byte("th")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ch")) {
		return true
	}
	if bytes.EqualFold(b, []byte("sh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ph")) {
		return true
	}
	if bytes.EqualFold(b, []byte("dg")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wn")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wd")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wl")) {
		return true
	}
	if bytes.EqualFold(b, []byte("gh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ng")) {
		return true
	}
	if bytes.EqualFold(b, []byte("sc")) {
		return true
	}
	if bytes.EqualFold(b, []byte("nx")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ck")) {
		return true
	}
	if bytes.EqualFold(b, []byte("kn")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("nd")) {
		return true
	}
	if bytes.EqualFold(b, []byte("tr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("dr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("cr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ll")) {
		return true
	}
	return false
}

func WordStart(b []byte) bool {
	if bytes.EqualFold(b, []byte("nc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("bc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("bz")) {
		return false
	}
	if bytes.EqualFold(b, []byte("dc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("dd")) {
		return false
	}

	return true
}

func IsAlpha(b byte) bool {
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

// Returns the index after which a hyphen should be inserted
func IntraWordBP(word []byte) (int, bool) {
	// don't break words shorter than 5 characters
	if len(word) < 5 {
		return -1, false
	}
	// if a word itself has a hyphen, break at the hyphen
	// does not distinguish between \u002D and \u00AD
	if i := bytes.IndexAny(word, "\u002D\u00AD"); i != -1 {
		return i, true
	}
	// don't break proper nouns
	if word[0] >= 'A' && word[0] <= 'Z' {
		return -1, false
	}
	// only break between consonants that do not form a digraph
	// and do not leave an unacceptable beginning consonant pair
	for i := 2; i < len(word)-4; i++ {
		if IsConsonant(word[i]) {
			if !IsConsonant(word[i+1]) {
				continue
			}
			if IsDigraph(word[i : i+2]) {
				continue
			}
			if !WordStart(word[i+1 : i+3]) {
				continue
			}
			if !IsAlpha(word[i+3]) {
				continue
			}
			return i, false
		}
	}
	return -1, false
}
