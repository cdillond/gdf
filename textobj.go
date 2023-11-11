package gdf

import (
	"fmt"
	"strings"
)

type TextObj struct {
	Matrix
	LineMatrix Matrix
}

// returns the x an y coordinates of the cursor relative to the text object's coordinate space
func (c *ContentStream) RawTextCursor() Point {
	return Transform(Point{0, 0}, c.TextObj.Matrix)
}

// returns the x and y coordinates of the text cursor relative to the content stream's coordinate space
func (c *ContentStream) TextCursor() Point {
	p := Transform(Point{0, 0}, c.TextObj.Matrix)
	return Transform(p, c.GS.Matrix)
}

func (c *ContentStream) TextExtentPts(s string) float64 {
	extFU := TextExtent(s, c.Font)
	extFU += float64(len(s)) * c.CharSpace
	extFU += float64(strings.Count(s, "\u0020")) * c.WordSpace
	return FUToPt(extFU*c.Scale/100, c.FontSize)
}
func (c *ContentStream) ShapedTextExtentPts(s string) float64 {
	extFU := ShapedTextExtent([]rune(s), c.Font)
	extFU += float64(len(s)) * c.CharSpace
	extFU += float64(strings.Count(s, "\u0020")) * c.WordSpace
	return FUToPt(extFU*c.Scale/100, c.FontSize)
}
func (c *ContentStream) UnscaledTextExtentPts(s string) float64 {
	extFU := TextExtent(s, c.Font)
	extFU += float64(len(s)) * c.CharSpace
	extFU += float64(strings.Count(s, "\u0020")) * c.WordSpace
	return FUToPt(extFU, c.FontSize)
}

// sets the current text object's text matrix and line matrix to m
func (c *ContentStream) Tm(m Matrix) {
	c.TextObj.Matrix = m
	c.TextObj.LineMatrix = m
	fmt.Fprintf(c.buf, "%f %f %f %f %f %f Tm\n", m.a, m.b, m.c, m.d, m.e, m.f)
}

// offsets the current text object's text matrix by x and y, and sets the text object's line matrix equal to its text matrix
func (c *ContentStream) Td(x, y float64) {
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, x, y})
	c.LineMatrix = c.TextObj.Matrix
	fmt.Fprintf(c.buf, "%f %f Td\n", x, y)
}

// sets the content stream's current leading to y and then calls c.Td(x, y)
func (c *ContentStream) TD(x, y float64) {
	c.TL(-y)
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, x, y})
	c.LineMatrix = c.TextObj.Matrix
	fmt.Fprintf(c.buf, "%f %f Td\n", x, y)
}

// sets the current text matrix and line matrix equal to the line matrix offset by (0, -c.Leading)
func (c *ContentStream) TStar() {
	c.TextObj.Matrix = Mul(c.LineMatrix, Matrix{1, 0, 0, 1, 0, -c.Leading})
	c.LineMatrix = c.TextObj.Matrix
	c.buf.Write([]byte("T*\n"))
}

// writes s and advances the text matrix by the extent of s
func (c *ContentStream) Tj(s string) {
	ext := c.TextExtentPts(s)
	b, _ := c.Font.enc.Bytes([]byte(s))
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, ext, 0})
	fmt.Fprintf(c.buf, "<%X> Tj\n", b)
}

func (c *ContentStream) TJ(runs []string, adjs []float64) error {
	if len(runs) != len(adjs) {
		return fmt.Errorf("equal number of runs and adjs required")
	}
	var ext1, ext2 float64
	c.buf.WriteByte('[')
	tmp := []byte{}
	for i := range runs {
		ext1 += adjs[i]
		ext2 += c.UnscaledTextExtentPts(string(runs[i]))
		b, _ := c.Font.enc.Bytes([]byte(runs[i]))
		tmp = append(tmp, b...)
		if adjs[i] != 0 {
			fmt.Fprintf(c.buf, "<%X>%.3f", tmp, -adjs[i])
			tmp = tmp[:0]
		}
	}
	if len(tmp) != 0 {
		fmt.Fprintf(c.buf, "<%X>", tmp)
	}
	c.buf.Write([]byte("] TJ\n"))
	ext := ext2 - FUToPt(ext1, c.FontSize)
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, ext * c.Scale / 100, 0})
	return nil
}

func (c *ContentStream) TJSpace(run []rune, kerns []int, spaceAdj float64) error {
	if len(run) != len(kerns) {
		return fmt.Errorf("equal number of runes and kerns required %d %d", len(run), len(kerns))
	}
	var ext float64
	if spaceAdj != 0 {
		c.Tw(FUToPt(spaceAdj, c.FontSize))
	}
	c.buf.WriteByte('[')
	tmp := []byte{}
	var kerntotal int
	for i, r := range run {
		b, _ := c.Font.enc.Bytes([]byte(string(r)))
		tmp = append(tmp, b...)
		if kerns[i] != 0 {
			fmt.Fprintf(c.buf, "<%X>%d", tmp, -(kerns[i]))
			tmp = tmp[:0]
			kerntotal += kerns[i]
		}
	}
	if len(tmp) != 0 {
		fmt.Fprintf(c.buf, "<%X>", tmp)
	}
	c.buf.Write([]byte("] TJ\n"))
	ext = c.UnscaledTextExtentPts(string(run))
	ext += FUToPt(float64(-kerntotal), c.FontSize)
	ext *= c.Scale / 100
	if spaceAdj != 0 {
		c.Tw(0)
	}
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, ext, 0})
	return nil
}
