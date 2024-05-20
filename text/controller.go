package text

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/cdillond/gdf"
)

var (
	// errors returned by NewController
	ErrTolerance = fmt.Errorf("unable to break lines using current tolerances")
	ErrWordSize  = fmt.Errorf("source text contains an unbreakable word that is longer than the maximum line length")

	// errors returned by DrawText
	ErrWidth   = fmt.Errorf("target area must be at least as wide as the maximum line width")
	ErrEmpty   = fmt.Errorf("source text buffer is empty")
	ErrLeading = fmt.Errorf("font leading must be greater than 0")
	ErrHeight  = fmt.Errorf("target area must be at least as tall as the font leading")
)

// A Controller is a struct that aids in writing text to a ContentStream. The Controller can break text into lines and paragraphs,
// determine the appropriate kerning for glyphs in a string, and draw text according to the format specified by the ControllerCfg struct.
type Controller struct {
	src            []rune        // source text
	family         FontFamily    // the set of fonts to be used. On each call to DrawText, the supplied ContentStream's font will be set to one of the fonts from f, according to the TextConroller's font weight state. The fonts need not be of the same actual family - chosen families are valid, too!
	isBold, isItal bool          // font weight state
	curFont        *gdf.Font     // Used for calculating line widths, but not for actually writing the text.
	fontSize       float64       // font size. On each call to DrawText, the supplied ContentStream's FontSize will be set to this value
	leading        float64       // text leading. On each call to DrawText, the supplied ContentStream's Leading will be set to this value.
	lineWidth      float64       // ideal line width in Font Units
	alignment      Alignment     // paragraph alignment
	just           Justification // paragraph justification style
	firstIndent    float64       // indent of the first line of each paragraph
	tokens         []token       // source text tokens
	breakpoints    []int         // indices of forced breaks in k.tokens
	lineWidths     []float64     // actual line widths, unjustified
	adjs           []float64     // adjustments (in font units) to the spaces ('\x20') in each line
	tightness      float64       // the ratio of the minimum allowable space advance and the normal space advance in justified text
	looseness      float64       // the ratio of the maximum allowable space advance and the normal space advance in justified text
	scolor, ncolor gdf.Color
	renderMode     gdf.RenderMode
	n              int // token index
	ln             int // line index
}

// A ControllerCfg specifies options for the formatting of text drawn by a Controller.
type ControllerCfg struct {
	Alignment
	Justification
	RenderMode     gdf.RenderMode
	FontSize       float64
	Leading        float64
	IsIndented     bool
	NColor, SColor gdf.Color // default nonstroking and stroking colors
	Looseness      float64   // the ratio of the maximum allowable space advance and the normal space advance in justified text
	Tightness      float64   // the ratio of the minimum allowable space advance and the normal space advance in justified text
	IsBold, IsItal bool
}

func NewControllerCfg(fontSize, leading float64) ControllerCfg {
	return ControllerCfg{
		Alignment:     Left,
		Justification: Ragged,
		FontSize:      fontSize,
		Leading:       leading,
		IsIndented:    false,
		SColor:        nil,
		NColor:        nil,
		Looseness:     2,
		Tightness:     .25,
	}
}

type FontFamily struct {
	Regular, Bold, Ital, BoldItal *gdf.Font
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

const (
	eot_tok  rune = -1
	col_tok  rune = -2
	bold_tok rune = -3
	ital_tok rune = -4
)

/*
FormatText represents a slice of runes that can specify the formatting and content of the source text. The TextController applies the following rules
to the formatting directives contained in FormatText:
 1. rune(-1) is interpreted as end of text. Any runes that appear after this character will not be parsed.
 2. rune(-2) is interpreted as a color indicator. This character must be followed by three comma-separated 3-digit integers in [0,255]
    that specify the red, green, and blue components of an RGBColor. (This is equivalent to setting the nonstroking color of the document to
    RGBColor{R:float64(red)/255, G:float64(green)/255, B:float64(blue)/255}).
 3. rune(-3) toggles bold text on and off.
 4. rune(-4) toggles italic text on and off.
    The bold and italic indicators can be used to switch among fonts in a given font family.
*/
type FormatText []rune

// NewController returns a Controller that is ready to write src to ContentStreams. It returns an invalid Controller
// and an error if it encounters a problem while parsing and shaping src. lineWidth should be the maximum desired
// width, in points, of each line of the output text when drawn to a gdf.ContentStream.
func NewController(src FormatText, lineWidth float64, f FontFamily, cfg ControllerCfg) (Controller, error) {
	tc := Controller{
		src:        src,
		family:     f,
		lineWidth:  gdf.PtToFU(lineWidth, cfg.FontSize),
		fontSize:   cfg.FontSize,
		alignment:  cfg.Alignment,
		just:       cfg.Justification,
		tightness:  cfg.Tightness,
		looseness:  cfg.Looseness,
		leading:    cfg.Leading,
		isBold:     cfg.IsBold,
		isItal:     cfg.IsItal,
		scolor:     cfg.SColor,
		ncolor:     cfg.NColor,
		renderMode: cfg.RenderMode,
	}
	if tc.just == Ragged {
		tc.tightness = 0
		tc.looseness = 1
	}
	if tc.just == Justified {
		// stretch tolerance cannot be 0
		if tc.looseness == 0 {
			tc.looseness = .5
		}
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

	spaceAdv := f.Regular.GlyphAdvance('\u0020') // use the regular font as the baseline regardless
	if cfg.IsIndented {
		// indent is 4 spaces; subject to change.
		tc.firstIndent = 4.0 * float64(spaceAdv)
	}

	tc.tokens = tc.tokenize(tc.src)
	breakpoints, lineWidths, adjs, err := tc.breakLines(tc.tightness, tc.looseness)
	// keep trying until it's clear there's no acceptable solution
	if err != nil {
		for squish, stretch := tc.tightness*1.25, tc.looseness*2; squish < 1 && errors.Is(err, ErrTolerance); {
			breakpoints, lineWidths, adjs, err = tc.breakLines(squish, stretch)
			squish *= 1.25
			stretch *= 2
		}

		if err != nil {
			return *new(Controller), err
		}
	}
	tc.breakpoints = breakpoints
	tc.lineWidths = lineWidths
	tc.adjs = adjs
	return tc, nil
}

// DrawText draws the text from the Controller to the area of c specified by area. The return gdf.Point is
// the position of c's TextCursor after the text has been drawn. (This value would not be otherwise accessible because
// each call to DrawText encompasses a c.BeginText/EndText pair.) The returned bool indicates whether the Controller's
// buffer still contains additional source text. If this value is true, then future calls to DrawText can be used to
// draw the remaining source text - usually to different areas or ContentStreams.
func (tc *Controller) DrawText(c *gdf.ContentStream, area gdf.Rect) (gdf.Point, bool, error) {
	if gdf.PtToFU(area.Width(), tc.fontSize) < tc.lineWidth {
		return *new(gdf.Point), false, ErrWidth
	}
	if tc.n >= len(tc.tokens) {
		return *new(gdf.Point), false, ErrEmpty
	}
	if tc.leading <= 0 {
		return *new(gdf.Point), false, ErrLeading
	}
	maxLines := int(area.Height()/tc.leading) + tc.ln
	if maxLines < 1 {
		return *new(gdf.Point), false, ErrHeight
	}
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
	if c.RenderMode != tc.renderMode {
		c.SetRenderMode(tc.renderMode)
	}
	c.SetTextOffset(area.LLX, area.URY-tc.leading)
	if err != nil {
		return *new(gdf.Point), false, err
	}
	tc.writeLines(c, maxLines)
	endPt := c.RawTextCursor()
	err = et()
	if err != nil {
		return *new(gdf.Point), false, err
	}
	return endPt, tc.n == len(tc.tokens)-1, nil
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

// nonstroking color change
type ncChange struct {
	r, g, b uint
}

func (n ncChange) Width() float64 { return 0 }

type hyphen float64

func (h hyphen) Width() float64 { return float64(h) }

// parses the source text and returns a slice of raw tokens
func (tc *Controller) tokenize(src FormatText) []token {
	// make local copies of these values so future calls to DrawText are not affected by operations here that nonetheless depend on future text states
	curFont := tc.curFont
	isBold := tc.isBold
	isItal := tc.isItal

	src = append(src, []rune{'\n', eot_tok}...) // this simplifies some of the logic
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
				out = append(out, flIndent(gdf.FUToPt(tc.firstIndent, tc.fontSize)))
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
			adv := curFont.GlyphAdvance('\u0020')
			out = append(out, skip(adv))
		case col_tok:
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
		case bold_tok: // bold
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
				curFont = tc.family.Ital
				out = append(out, ital)
			} else if isBold {
				curFont = tc.family.Regular
				out = append(out, regular)
			} else if isItal {
				curFont = tc.family.BoldItal
				out = append(out, boldItal)
			} else {
				curFont = tc.family.Bold
				out = append(out, bold)
			}
			isBold = !isBold
		case ital_tok: // italic
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
				curFont = tc.family.Bold
				out = append(out, bold)
			} else if isItal {
				curFont = tc.family.Regular
				out = append(out, regular)
			} else if isBold {
				curFont = tc.family.BoldItal
				out = append(out, boldItal)
			} else {
				curFont = tc.family.Ital
				out = append(out, ital)
			}
			isItal = !isItal
		case eot_tok: // eot
			if len(run) != 0 {
				out = append(out, box{chars: run, kerns: kerns, advs: advs, width: width(advs, kerns)})
			}
			return out
		default:
			run = append(run, src[i])
			adv, kern := curFont.ShapedGlyphAdv(src[i], src[i+1])
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

// The algorithm used here - a modified version of the Knuth-Plass line-breaking algorithm - has O(n²) time complexity, but the value of n is
// effectively limited by the squishTolerance and stretchTolerance. There are pathological cases that can break the algorithm. In such cases,
// the tolerances can be expanded - but this is doubly bad because it can result in worse-looking paragraphs that take much longer to process.
// Alternative algorithms either cannot be adopted to text that includes optional hyphenated breaks and/or negative glyph advances, or find
// potentially suboptimal line fits.
// TODO: gracefully handle pathological cases.
// breakpoints, lineWidths, adjs, err
func (tc *Controller) breakLines(squishTolerance, stretchTolerance float64) (breakpoints []int, lineWidths []float64, adjs []float64, err error) {
	curFont := tc.curFont

	breakIndices := []int{}
	lineLengths := []float64{}
	//adjs := []float64{}
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
		case flIndent:
			// TODO
		case box:
			curWidth += v.Width()
			runWidth += v.Width()
			if runWidth > tc.lineWidth {
				return nil, nil, nil, fmt.Errorf("%w: %s", ErrWordSize, string(v.chars))
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
			spAdv := float64(curFont.GlyphAdvance('\u0020'))
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
				return nil, nil, nil, fmt.Errorf("%w, squish: %f stretch: %f", ErrTolerance, squishTolerance, stretchTolerance)
			}
		case fWeight:
			switch v {
			case regular:
				curFont = tc.family.Regular
			case bold:
				curFont = tc.family.Bold
			case ital:
				curFont = tc.family.Ital
			case boldItal:
				curFont = tc.family.BoldItal
			}
		case ncChange:
		case hyphen:
		case newline:
			runWidth = 0
			// this should already have been caught, but it might help to do a bounds check just to be safe
			if len(activeNodes) == 0 {
				return nil, nil, nil, fmt.Errorf("%w, squish: %f stretch: %f", ErrTolerance, squishTolerance, stretchTolerance)
			}
			// the remaining active node should be the one with the optimal endpoint
			endNode := activeNodes[len(activeNodes)-1]

			nodes := []node{endNode}
			for next := nodes[0].bestStart; next > 0; {
				nextNode := activeNodes[next]
				nodes = append(nodes, nextNode)
				next = nextNode.bestStart
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
	if tc.just == Ragged {
		adjs = make([]float64, len(adjs))
	}
	return breakIndices, lineLengths, adjs, nil
}

// the run needs to be considered in absence of the formatting directives, but then it needs to be reconstituted with those
// directives in mind
func (tc *Controller) writeLines(c *gdf.ContentStream, numLines int) {
	lineCount := tc.ln
	breaks := map[int]struct{}{}
	for _, ind := range tc.breakpoints {
		breaks[ind] = struct{}{}
	}

	run := []rune{}
	kerns := []int{}
	indented := tc.firstIndent != 0 && tc.n == 0
	if indented {
		c.Concat(gdf.Translate(gdf.FUToPt(tc.firstIndent, tc.fontSize), 0))
	}
	i := tc.n
	for ; i < len(tc.tokens) && lineCount < len(tc.breakpoints); i++ {
		if _, ok := breaks[i]; ok {
			var dif float64
			var alignAdj float64
			if tc.adjs[lineCount] == 0 && (tc.alignment == Center || tc.alignment == Right) && len(run) != 0 {
				dif = tc.lineWidth - tc.lineWidths[lineCount]
				switch tc.alignment {
				case Center:
					alignAdj = dif / 2
				case Right:
					alignAdj = dif
				}
				c.Concat(gdf.Translate(gdf.FUToPt(alignAdj, tc.fontSize), 0))
			}
			if len(run) != 0 {
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(gdf.FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}
			}
			if dif != 0 {
				c.Concat(gdf.Translate(-gdf.FUToPt(alignAdj, tc.fontSize), 0))
			}
			c.NextLine()
			if indented {
				c.Concat(gdf.Translate(-gdf.FUToPt(tc.firstIndent, tc.fontSize), 0))
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
				if tc.adjs[lineCount] == 0 && (tc.alignment == Center || tc.alignment == Right) && len(run) != 0 {
					dif = tc.lineWidth - tc.lineWidths[lineCount]
					switch tc.alignment {
					case Center:
						alignAdj = dif / 2
					case Right:
						alignAdj = dif
					}
					c.Concat(gdf.Translate(gdf.FUToPt(alignAdj, tc.fontSize), 0))
				}
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(gdf.FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}

				if dif != 0 {
					c.Concat(gdf.Translate(-gdf.FUToPt(alignAdj, tc.fontSize), 0))
				}
			}
			run = run[:0]
			kerns = kerns[:0]
			switch v {
			case regular:
				c.SetFont(tc.fontSize, tc.family.Regular)
			case bold:
				c.SetFont(tc.fontSize, tc.family.Bold)
			case ital:
				c.SetFont(tc.fontSize, tc.family.Ital)
			case boldItal:
				c.SetFont(tc.fontSize, tc.family.BoldItal)
			}
		case ncChange:
			if len(run) != 0 {
				var dif float64
				var alignAdj float64
				if tc.adjs[lineCount] == 0 && (tc.alignment == Center || tc.alignment == Right) && len(run) != 0 {
					dif = tc.lineWidth - tc.lineWidths[lineCount]
					switch tc.alignment {
					case Center:
						alignAdj = dif / 2
					case Right:
						alignAdj = dif
					}
					c.Concat(gdf.Translate(gdf.FUToPt(alignAdj, tc.fontSize), 0))
				}
				if tc.adjs[lineCount] != 0 {
					c.SetWordSpace(gdf.FUToPt(tc.adjs[lineCount], c.FontSize))
					c.ShowText(run, kerns)
					c.SetWordSpace(0)
				} else {
					c.ShowText(run, kerns)
				}

				if dif != 0 {
					c.Concat(gdf.Translate(-gdf.FUToPt(alignAdj, tc.fontSize), 0))
				}
			}
			run = run[:0]
			kerns = kerns[:0]
			c.SetColor(gdf.RGBColor{R: float64(v.r) / 255, G: float64(v.g) / 255, B: float64(v.b) / 255})
		case skip:
			if v.Width() != 0 {
				run = append(run, ' ')
				kerns = append(kerns, 0)
			}

		case flIndent:
			indented = true
			c.Concat(gdf.Translate(float64(v), 0))
		case hyphen:
		}
	}
	tc.n = i
	tc.ln = lineCount
}
