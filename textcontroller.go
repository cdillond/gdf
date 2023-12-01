package gdf

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

var (
	errTolerance = fmt.Errorf("unable to break lines using current tolerances")
	errWordSize  = fmt.Errorf("source text contains an unbreakable word that is longer than the maximum line length")
)

type TextController struct {
	src              []rune        // source text
	f                FontFamily    // the set of fonts to be used. On each call to DrawText, the supplied ContentStream's font will be set to one of the fonts from f, according to the TextConroller's font weight state. The fonts need not be of the same actual family - chosen families are valid, too!
	isBold, isItal   bool          // font weight state
	curFont          *Font         // Used for calculating line widths, but not for actually writing the text.
	fontSize         float64       // font size. On each call to DrawText, the supplied ContentStream's FontSize will be set to this value
	leading          float64       // text leading. On each call to DrawText, the supplied ContentStream's Leading will be set to this value.
	lineWidth        float64       // ideal line width in Font Units
	a                Alignment     // paragraph alignment
	j                Justification // paragraph justification style
	firstIndent      float64       // indent of the first line of each paragraph
	tokens           []token       // source text tokens
	breakpoints      []int         // indices of forced breaks in k.tokens
	lineWidths       []float64     // actual line widths, unjustified
	adjs             []float64     // adjustments (in font units) to the spaces ('\x20') in each line
	squishTolerance  float64       // maximum allowable squish factor for the spaces in each line. lower = tighter spacing. default = 4. values under 1.0 can result in overlapping text.
	stretchTolerance float64       // maximum allowable stretch tolerance for the spaces in each line. higher = wider spacing. default = 10.
	scolor, ncolor   Color
	r                RenderMode
	n                int // token index
	ln               int // line index
}

type ControllerCfg struct {
	Alignment
	Justification
	RenderMode
	FontSize         float64
	Leading          float64
	IsIndented       bool
	NColor, SColor   Color // default non-stroking and stroking colors
	StretchTolerance float64
	SquishTolerance  float64
	IsBold, IsItal   bool
}

func NewControllerCfg(fontSize, leading float64) ControllerCfg {
	return ControllerCfg{
		Alignment:        Right,
		Justification:    Justified,
		FontSize:         fontSize,
		Leading:          leading,
		IsIndented:       true,
		SColor:           nil,
		NColor:           nil,
		StretchTolerance: 5,
		SquishTolerance:  .5,
	}
}

type FontFamily struct {
	Regular, Bold, Ital, BoldItal *Font
}

type Alignment uint

const (
	Left Alignment = iota
	Right
	Center
)

type Justification uint

const (
	Ragged Justification = iota
	Justified
)

/*
FormatText represents a slice of runes that can specify the formatting and content of the source text. The TextController applies the following rules
to the formatting directives contained in FormatText:
 1. \x03 (U+0003) is interpreted as end of text. Any runes that appear after this character will not be parsed.
 2. \x07 (U+0007) is interpreted as a color indicator. This character must be followed by three comma-separated 3-digit integers in [0,255]
    that specify the red, green, and blue components of an RGBColor. (This is equivalent to setting the non-stroking color of the document to
    RGBColor{R:float64(red)/255, G:float64(green)/255, B:float64(blue)/255}). Example: Hello, \x07127,000,090world\x07000,000,000!
 3. \x0E (U+000E) toggles bold text on and off.
 4. \x0F (U+000F) toggles italic text on and off.
    The bold and italic indicators can be used to switch among fonts in a given font family. Example:
    This is regular text. \x0EThis is bold text. \x0FThis is bold-italic text. \x0EThis is italic text. \x0FThis is regular text.
*/
type FormatText []rune

func NewTextController(src FormatText, lineWidth float64, f FontFamily, cfg ControllerCfg) (TextController, error) {
	tc := TextController{
		src:              src,
		f:                f,
		lineWidth:        lineWidth,
		fontSize:         cfg.FontSize,
		a:                cfg.Alignment,
		j:                cfg.Justification,
		squishTolerance:  cfg.SquishTolerance,
		stretchTolerance: cfg.StretchTolerance,
		leading:          cfg.Leading,
		isBold:           cfg.IsBold,
		isItal:           cfg.IsItal,
		scolor:           cfg.SColor,
		ncolor:           cfg.NColor,
		r:                cfg.RenderMode,
	}
	if tc.j == Ragged {
		tc.squishTolerance = 0
	}
	if tc.isBold && tc.isItal {
		tc.curFont = f.BoldItal
	} else if tc.isBold {
		tc.curFont = f.Bold
	} else if tc.isItal {
		tc.curFont = f.Ital
	} else {
		tc.curFont = f.Regular
	}

	spaceAdv := GlyphAdvance('\u0020', f.Regular) // use the regular font as the baseline regardless
	if cfg.IsIndented {
		tc.firstIndent = 5.0 * float64(spaceAdv)
	}

	tc.tokens = tc.tokenize(tc.src)
	breakpoints, lineWidths, adjs, err := tc.breakLines(tc.squishTolerance, tc.stretchTolerance)
	// keep trying until it's clear there's no acceptable solution
	if err != nil {
		if errors.Is(err, errTolerance) {
			squish, stretch := tc.squishTolerance*1.25, tc.stretchTolerance*2
			tc.squishTolerance = 1 / tc.squishTolerance
			for squish < 1 && err != nil {
				breakpoints, lineWidths, adjs, err = tc.breakLines(squish, stretch)
				squish *= 2
				stretch *= 2
			}
		}
		if err != nil {
			return *new(TextController), err
		}
	}
	tc.breakpoints = breakpoints
	tc.lineWidths = lineWidths
	tc.adjs = adjs
	return tc, nil
}

func (tc *TextController) DrawText(c *ContentStream, area Rect) (error, bool) {
	if PtToFU(area.Width(), tc.fontSize) < tc.lineWidth {
		return fmt.Errorf("target area must be at least as wide as the max line width"), false
	}
	if tc.n >= len(tc.tokens) {
		return fmt.Errorf("src text buffer is empty"), false
	}
	if tc.leading <= 0 {
		return fmt.Errorf("font leading must be greater than 0"), false
	}
	maxLines := area.Height() / tc.leading
	c.QSave()
	if c.Leading != tc.leading {
		c.SetLeading(tc.leading)
	}
	if c.Font != tc.curFont || c.FontSize != tc.fontSize {
		c.SetFont(tc.fontSize, tc.curFont)
	}
	if c.NColor != tc.ncolor && tc.ncolor != nil {
		c.SetColor(tc.ncolor)
	}
	if c.SColor != tc.scolor && tc.scolor != nil {
		c.SetColorStroke(tc.scolor)
	}
	et, err := c.BeginText()
	if c.RenderMode != tc.r {
		c.SetRenderMode(tc.r)
	}
	c.TextOffset(area.LLX, area.URY-tc.leading)
	if err != nil {
		c.QRestore()
		return err, false
	}
	tc.writeLines(c, min(int(maxLines), len(tc.breakpoints)))
	err = et()
	if err != nil {
		c.QRestore()
		return err, false
	}
	err = c.QRestore()
	if err != nil {
		return err, false
	}
	return nil, tc.n == len(tc.tokens)
}

type token interface {
	Width() float64
}

type box struct {
	chars []rune
	advs  []int
	kerns []int
	width float64
}

func (b box) Width() float64 { return b.width }

func width(advs, kerns []int) float64 {
	var out int
	_ = kerns[len(advs)-1]
	for i := 0; i < len(advs); i++ {
		out += advs[i] + kerns[i]
	}
	return float64(out)
}

type skip float64

func (s skip) Width() float64 { return float64(s) }

type newline struct{}

func (n newline) Width() float64 { return 0 }

type fWeight uint

func (f fWeight) Width() float64 { return 0 }

const (
	regular fWeight = iota
	bold
	ital
	boldItal
)

type flIndent float64

func (f flIndent) Width() float64 { return 0 }

// non-stroking color change
type ncChange struct {
	r, g, b uint
}

func (n ncChange) Width() float64 { return 0 }

type hyphen float64

func (h hyphen) Width() float64 { return float64(h) }

// parses the source text and returns a slice of raw tokens
func (tc *TextController) tokenize(src FormatText) []token {
	// make local copies of these values so future calls to DrawText are not affected by operations here that nonetheless depend on future text states
	curFont := tc.curFont
	isBold := tc.isBold
	isItal := tc.isItal

	src = append(src, []rune{'\n', '\u0003'}...) // this simplifies some of the logic
	out := make([]token, 0, len(src))
	run := []rune{}
	kerns := []int{}
	advs := []int{}
	i := 0
	for ; ; i++ {
		switch src[i] {
		case '\n', '\r':
			if len(run) != 0 {
				chars := make([]rune, len(run))
				copy(chars, run)
				kerns2 := make([]int, len(run))
				copy(kerns2, kerns)
				advs2 := make([]int, len(run))
				copy(advs2, advs)
				out = append(out, box{chars: chars, kerns: kerns2, advs: advs2, width: width(advs2, kerns2)})
			}
			run = run[:0]
			advs = advs[:0]
			kerns = kerns[:0]

			// finishing glue
			out = append(out, skip(0))
			out = append(out, newline{})
			if tc.firstIndent != 0 {
				out = append(out, flIndent(FUToPt(tc.firstIndent, tc.fontSize)))
			}
		case '\u0020':
			if len(run) != 0 {
				chars := make([]rune, len(run))
				copy(chars, run)
				kerns2 := make([]int, len(run))
				copy(kerns2, kerns)
				advs2 := make([]int, len(run))
				copy(advs2, advs)
				out = append(out, box{chars: chars, kerns: kerns2, advs: advs2, width: width(advs2, kerns2)})
			}
			run = run[:0]
			advs = advs[:0]
			kerns = kerns[:0]
			adv := GlyphAdvance('\u0020', curFont)
			out = append(out, skip(adv))
		case '\u0007':
			if len(run) != 0 {
				chars := make([]rune, len(run))
				copy(chars, run)
				kerns2 := make([]int, len(run))
				copy(kerns2, kerns)
				advs2 := make([]int, len(run))
				copy(advs2, advs)
				out = append(out, box{chars: chars, kerns: kerns2, advs: advs2, width: width(advs2, kerns2)})
			}
			run = run[:0]
			advs = advs[:0]
			kerns = kerns[:0]

			if i > len(src)-10 {
				continue
			}
			var r, g, b uint
			_, err := fmt.Sscanf(string(src[i+1:i+1+9+2]), "%03d,%03d,%03d", &r, &g, &b)
			if err != nil {
				continue
			}
			out = append(out, ncChange{r: r, g: g, b: b})
			i += 11
		case '\u000E': // bold
			if len(run) != 0 {
				chars := make([]rune, len(run))
				copy(chars, run)
				kerns2 := make([]int, len(run))
				copy(kerns2, kerns)
				advs2 := make([]int, len(run))
				copy(advs2, advs)
				out = append(out, box{chars: chars, kerns: kerns2, advs: advs2, width: width(advs2, kerns2)})
			}
			run = run[:0]
			advs = advs[:0]
			kerns = kerns[:0]
			if isBold && isItal {
				curFont = tc.f.Ital
				out = append(out, ital)
			} else if isBold {
				curFont = tc.f.Regular
				out = append(out, regular)
			} else if isItal {
				curFont = tc.f.BoldItal
				out = append(out, boldItal)
			} else {
				curFont = tc.f.Bold
				out = append(out, bold)
			}
			isBold = !isBold
		case '\u000F': // italic
			if len(run) != 0 {
				chars := make([]rune, len(run))
				copy(chars, run)
				kerns2 := make([]int, len(run))
				copy(kerns2, kerns)
				advs2 := make([]int, len(run))
				copy(advs2, advs)
				out = append(out, box{chars: chars, kerns: kerns2, advs: advs2, width: width(advs2, kerns2)})
			}
			run = run[:0]
			advs = advs[:0]
			kerns = kerns[:0]
			if isItal && isBold {
				curFont = tc.f.Bold
				out = append(out, bold)
			} else if isItal {
				curFont = tc.f.Regular
				out = append(out, regular)
			} else if isBold {
				curFont = tc.f.BoldItal
				out = append(out, boldItal)
			} else {
				curFont = tc.f.Ital
				out = append(out, ital)
			}
			isItal = !isItal
		case '\u0003': // eot
			if len(run) != 0 {
				out = append(out, box{chars: run, kerns: kerns, advs: advs, width: width(advs, kerns)})
			}
			return out
		default:
			run = append(run, src[i])
			adv, kern := ShapedGlyphAdv(src[i], src[i+1], curFont)
			kerns = append(kerns, kern)
			advs = append(advs, adv)
		}
	}
}

type node struct {
	nIndex    int     // node index in the slice of active nodes
	tIndex    int     // token index in k.tokens
	pWidth    float64 // total width of the current paragraph text up to and including the current node
	pSpaces   float64 // number of spaces in the current paragraph, including the current node
	bestStart int     // nIndex value of the optimal line start node
	bestLW    float64 // width of a line starting at at activeNodes[bestStart] and terminating at the current node
	bestR     float64 // adjustment ratio of a line starting at activeNodes[bestStart] and terminating at the current node
	dSum      float64 // sum of the demerits for all nodes in the optimal path leading to and including the current node
}

// The algorithm used here - a modified version of the Knuth-Plass linebreaking algorithm - has O(n²) time complexity, but the value of n is
// effectively limited by the squishTolerance and stretchTolerance. There are pathological cases that can break the algorithm. In such cases,
// the tolerances can be expanded - but this is doublely bad because it can result in worse-looking paragraphs that take much longer to process.
// Alternative algorithms either cannot be adopted to text that includes optional hyphenated breaks and/or negative glyph advances, or find
// potentially suboptimal line fits.
// TODO: gracefully handle pathological cases.
func (tc *TextController) breakLines(squishTolerance, stretchTolerance float64) ([]int, []float64, []float64, error) {
	curFont := tc.curFont

	breakIndices := []int{}
	lineLengths := []float64{}
	adjs := []float64{}
	activeNodes := []node{{
		tIndex:    0,
		pWidth:    0,
		pSpaces:   0,
		bestStart: -1,
	}}

	var numSpaces float64
	var lineStart int
	curWidth := tc.firstIndent
	var runWidth float64
	for i := 0; i < len(tc.tokens); i++ {
		switch v := tc.tokens[i].(type) {
		case box:
			curWidth += v.Width()
			runWidth += v.Width()
			if runWidth > tc.lineWidth {
				return *new([]int), *new([]float64), *new([]float64), fmt.Errorf("%w: %s", errWordSize, string(v.chars))
			}
		case skip:
			runWidth = 0
			curWidth += v.Width()
			numSpaces++

			newNode := node{tIndex: i, pWidth: curWidth, pSpaces: numSpaces}
			var bestStart int
			var bestLW, bestR float64
			bestDemerits := math.Inf(0)
			bestSumDemerits := math.Inf(0)
			// this has to be done on a token-by token basis because the font can change
			spAdv := float64(GlyphAdvance(' ', curFont))
			// check if we have a feasible breakpoint
			for j := lineStart; j < len(activeNodes); j++ {

				r := (tc.lineWidth + activeNodes[j].pWidth - curWidth + v.Width()) / (numSpaces - activeNodes[j].pSpaces - 1)
				if v.Width() == 0 { // we've hit finishing glue
					if r > 0 || math.IsNaN(r) {
						r = 0
						lineStart = j + 1 // disable nodes that don't terminate here
					}
				}
				if math.IsNaN(r) {
					continue
				}

				// remove node j from future consideration as an active node
				if r < -spAdv*squishTolerance {
					lineStart = j + 1
				}

				if r >= -spAdv*squishTolerance && r <= spAdv*stretchTolerance {
					demerits := r * r
					if demerits+activeNodes[j].dSum < bestSumDemerits {
						bestSumDemerits = demerits + activeNodes[j].dSum
						bestDemerits = demerits
						bestStart = activeNodes[j].nIndex
						bestLW = curWidth - activeNodes[j].pWidth - v.Width()
						bestR = r
					}
				}
			}

			if !math.IsInf(bestDemerits, 0) {
				newNode.dSum = bestSumDemerits
				newNode.bestStart = bestStart
				newNode.bestLW = bestLW
				newNode.bestR = bestR
				newNode.nIndex = len(activeNodes)
				activeNodes = append(activeNodes, newNode)
			}
			// unable to proceed
			if lineStart == len(activeNodes) {
				return []int{}, []float64{}, []float64{}, fmt.Errorf("%w, squish: %f stretch: %f", errTolerance, squishTolerance, stretchTolerance)
			}
		case fWeight:
			switch v {
			case regular:
				curFont = tc.f.Regular
			case bold:
				curFont = tc.f.Bold
			case ital:
				curFont = tc.f.Ital
			case boldItal:
				curFont = tc.f.BoldItal
			}
		case ncChange:
		case hyphen:
		case newline:
			runWidth = 0
			// this should already have been caught, but it might help to do a bounds check just to be safe
			if len(activeNodes) == 0 {
				return []int{}, []float64{}, []float64{}, fmt.Errorf("%w, squish: %f stretch: %f", errTolerance, squishTolerance, stretchTolerance)
			}
			// the remaining active node should be the one with the optimal endpoint
			endNode := activeNodes[len(activeNodes)-1]

			nodes := []node{endNode}
			for next := nodes[0].bestStart; next > 0; {
				nnode := activeNodes[next]
				nodes = append(nodes, nnode)
				next = nnode.bestStart
			}
			slices.Reverse(nodes)

			for _, n := range nodes {
				breakIndices = append(breakIndices, n.tIndex)
				lineLengths = append(lineLengths, n.bestLW)
				adjs = append(adjs, n.bestR)
			}
			activeNodes = []node{{
				tIndex:    i,
				pWidth:    0,
				pSpaces:   0,
				bestStart: -1,
			}}

			curWidth = tc.firstIndent
			numSpaces = 0
			lineStart = 0
		}
	}
	if tc.j == Ragged {
		adjs = make([]float64, len(adjs))
	}
	return breakIndices, lineLengths, adjs, nil
}

// the run needs to be considered in absence of the formatting directives, but then it needs to be reconstituted with those
// directives in mind
func (tc *TextController) writeLines(c *ContentStream, numLines int) {
	lineCount := tc.ln
	breaks := map[int]struct{}{}
	for _, ind := range tc.breakpoints {
		breaks[ind] = struct{}{}
	}

	run := []rune{}
	kerns := []int{}
	indented := tc.firstIndent != 0 && tc.n == 0
	if indented {
		c.Concat(Translate(FUToPt(tc.firstIndent, tc.fontSize), 0))
	}
	i := tc.n
	for ; i < len(tc.tokens) && lineCount < len(tc.breakpoints); i++ {
		if _, ok := breaks[i]; ok {
			var dif float64
			var alignAdj float64
			if tc.adjs[lineCount] == 0 && (tc.a == Center || tc.a == Left) && len(run) != 0 {
				dif = tc.lineWidth - tc.lineWidths[lineCount]
				switch tc.a {
				case Center:
					alignAdj = dif / 2
				case Left:
					alignAdj = dif
				}
				c.Concat(Translate(FUToPt(alignAdj, tc.fontSize), 0))
			}
			if len(run) != 0 {
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}
			}
			if dif != 0 {
				c.Concat(Translate(-FUToPt(alignAdj, tc.fontSize), 0))
			}
			c.NextLine()
			if indented {
				c.Concat(Translate(-FUToPt(tc.firstIndent, tc.fontSize), 0))
				indented = false
			}
			lineCount++
			run = run[:0]
			kerns = kerns[:0]
			if lineCount == numLines {
				tc.n = i + 1
				tc.ln = lineCount
				return
			}
			continue
		}
		switch v := tc.tokens[i].(type) {
		case box:
			run = append(run, v.chars...)
			kerns = append(kerns, v.kerns...)
		case fWeight:
			if len(run) != 0 {
				var dif float64
				var alignAdj float64
				if tc.adjs[lineCount] == 0 && (tc.a == Center || tc.a == Left) && len(run) != 0 {
					dif = tc.lineWidth - tc.lineWidths[lineCount]
					switch tc.a {
					case Center:
						alignAdj = dif / 2
					case Left:
						alignAdj = dif
					}
					c.Concat(Translate(FUToPt(alignAdj, tc.fontSize), 0))
				}
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}

				if dif != 0 {
					c.Concat(Translate(-FUToPt(alignAdj, tc.fontSize), 0))
				}
			}
			run = run[:0]
			kerns = kerns[:0]
			switch v {
			case regular:
				c.SetFont(tc.fontSize, tc.f.Regular)
			case bold:
				c.SetFont(tc.fontSize, tc.f.Bold)
			case ital:
				c.SetFont(tc.fontSize, tc.f.Ital)
			case boldItal:
				c.SetFont(tc.fontSize, tc.f.BoldItal)
			}
		case ncChange:
			if len(run) != 0 {
				var dif float64
				var alignAdj float64
				if tc.adjs[lineCount] == 0 && (tc.a == Center || tc.a == Left) && len(run) != 0 {
					dif = tc.lineWidth - tc.lineWidths[lineCount]
					switch tc.a {
					case Center:
						alignAdj = dif / 2
					case Left:
						alignAdj = dif
					}
					c.Concat(Translate(FUToPt(alignAdj, tc.fontSize), 0))
				}
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}

				if dif != 0 {
					c.Concat(Translate(-FUToPt(alignAdj, tc.fontSize), 0))
				}
			}
			run = run[:0]
			kerns = kerns[:0]
			c.SetColor(RGBColor{R: float64(v.r) / 255, G: float64(v.g) / 255, B: float64(v.b) / 255})
		case skip:
			if v.Width() != 0 {
				run = append(run, ' ')
				kerns = append(kerns, 0)
			}

		case flIndent:
			indented = true
			c.Concat(Translate(float64(v), 0))
		case hyphen:
		}
	}
	tc.n = i
	tc.ln = lineCount
}
