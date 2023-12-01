package gdf

import (
	"fmt"
)

type TextObj struct {
	Matrix
	LineMatrix Matrix
}

// Returns the x and y coordinates of the point (0,0,1) transformed only by the current text matrix.
func (c *ContentStream) RawTextCursor() Point {
	return Transform(Point{0, 0}, c.TextObj.Matrix)
}

// Returns the x and y coordinates of the point (0,0,1) transformed by the current text matrix and the current transformation matrix.
func (c *ContentStream) TextCursor() Point {
	p := Transform(Point{0, 0}, c.TextObj.Matrix)
	return Transform(p, c.GS.Matrix)
}

// Sets the current text object's text matrix and line matrix to m; Tm.
func (c *ContentStream) SetTextMatrix(m Matrix) {
	c.TextObj.Matrix = m
	c.TextObj.LineMatrix = m
	fmt.Fprintf(c.buf, "%f %f %f %f %f %f Tm\n", m.A, m.B, m.C, m.D, m.E, m.F)
}

// Offsets the current text object's text matrix by x and y, and sets the text object's line matrix equal to its text matrix.
func (c *ContentStream) TextOffset(x, y float64) {
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, x, y})
	c.LineMatrix = c.TextObj.Matrix
	fmt.Fprintf(c.buf, "%f %f Td\n", x, y)
}

// Sets the content stream's current leading to y and then calls c.TextOffset(x, y).
func (c *ContentStream) TextOffsetLeading(x, y float64) {
	c.SetLeading(-y)
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, x, y})
	c.LineMatrix = c.TextObj.Matrix
	fmt.Fprintf(c.buf, "%f %f Td\n", x, y)
}

// Begins a new text line by setting the current text matrix and line matrix equal to the line matrix offset by (0, -c.Leading); T*.
func (c *ContentStream) NextLine() {
	c.TextObj.Matrix = Mul(c.LineMatrix, Matrix{1, 0, 0, 1, 0, -c.Leading})
	c.LineMatrix = c.TextObj.Matrix
	c.buf.Write([]byte("T*\n"))
}

// Writes s (without kerning) and advances the text matrix by the extent of s; Tj.
func (c *ContentStream) ShowString(s string) {
	ext := c.RawExtentPts([]rune(s))
	b, _ := c.Font.enc.Bytes([]byte(s))
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, ext, 0})
	fmt.Fprintf(c.buf, "<%X> Tj\n", b)
}

// Writes t (with kerning) and advances the text matrix by the extent of t; TJ.
func (c *ContentStream) ShowText(t []rune, kerns []int) error {
	if len(t) != len(kerns) {
		return fmt.Errorf("equal number of runes and kerns required. rune count: %d, kern count: %d", len(t), len(kerns))
	}
	c.buf.WriteByte('[')
	tmp := []byte{}
	for i, r := range t {
		b, _ := c.Font.enc.Bytes([]byte(string(r)))
		tmp = append(tmp, b...)
		if kerns[i] != 0 {
			fmt.Fprintf(c.buf, "<%X>%d", tmp, -(kerns[i]))
			tmp = tmp[:0]
		}
	}
	if len(tmp) != 0 {
		fmt.Fprintf(c.buf, "<%X>", tmp)
	}
	c.buf.Write([]byte("] TJ\n"))
	ext, err := c.ExtentKernsPts(t, kerns)
	if err != nil {
		return err
	}
	c.TextObj.Matrix = Mul(c.TextObj.Matrix, Matrix{1, 0, 0, 1, ext, 0})
	return nil
}
