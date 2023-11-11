package gdf

import "fmt"

// moveto
func (c *ContentStream) MSmall(x, y float64) {
	c.PathState = PATH_BUILDING
	c.CurPt = Point{x, y}
	fmt.Fprintf(c.buf, "%f %f m\n", x, y)
}

// lineto
func (c *ContentStream) LSmall(x, y float64) {
	fmt.Fprintf(c.buf, "%f %f l\n", x, y)
}

// append cubic bezier curve
func (c *ContentStream) CSmall(x1, y1, x2, y2, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f %f %f c\n", x1, y1, x2, y2, x3, y3)
	}
}

// append cubic bezier curve
func (c *ContentStream) VSmall(x2, y2, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f v\n", x2, y2, x3, y3)
	}
}

func (c *ContentStream) YSmall(x1, y1, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f c", x1, y1, x3, y3)
	}
}

func (c *ContentStream) HSmall() {
	c.buf.Write([]byte("h\n"))
}

// Table 59
// stroke path
func (c *ContentStream) S() {
	c.buf.Write([]byte("S\n"))
}

// close and stroke path
func (c *ContentStream) SSmall() {
	c.buf.Write([]byte("s\n"))
}

// fill path non-zero winding
func (c *ContentStream) FSmall() {
	c.buf.Write([]byte("f\n"))
}

// DEPRECATED
//func (c *ContentStream) F() {
//	c.buf.Write([]byte("F\n"))
//}

// fill path even-odd
func (c *ContentStream) FSmallStar() {
	c.buf.Write([]byte("f*\n"))
}

// fill path (non-zero winding) then stroke
func (c *ContentStream) B() {
	c.buf.Write([]byte("B\n"))
}

// fill path (even-odd) then stroke
func (c *ContentStream) BStar() {
	c.buf.Write([]byte("B*\n"))
}

// close path, fill path (non-zero winding), then stroke path
func (c *ContentStream) BSmall() {
	c.buf.Write([]byte("b\n"))
}

// close path, fill path (even-odd), then stroke path
func (c *ContentStream) BSmallStar() {
	c.buf.Write([]byte("b*\n"))
}

// end path
func (c *ContentStream) NSmall() {
	c.buf.Write([]byte("n\n"))
}
