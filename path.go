package gdf

import "fmt"

// Begin a new path starting at (x, y); m.
func (c *ContentStream) MoveTo(x, y float64) {
	c.PathState = PATH_BUILDING
	c.CurPt = Point{x, y}
	fmt.Fprintf(c.buf, "%f %f m\n", x, y)
}

// Append a straight line from the current point to (x, y); l.
func (c *ContentStream) LineTo(x, y float64) {
	fmt.Fprintf(c.buf, "%f %f l\n", x, y)
}

// Append a cubic Bézier curve to the current path; c.
func (c *ContentStream) CSmall(x1, y1, x2, y2, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f %f %f c\n", x1, y1, x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; v.
func (c *ContentStream) VSmall(x2, y2, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f v\n", x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; y.
func (c *ContentStream) YSmall(x1, y1, x3, y3 float64) {
	switch c.PathState {
	case PATH_NONE, PATH_CLIPPING:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f c", x1, y1, x3, y3)
	}
}

// Close the current path; h.
func (c *ContentStream) ClosePath() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("h\n"))
		c.PathState = PATH_BUILDING
	}
}

// Stroke path; S.
func (c *ContentStream) Stroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("S\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Close and stroke path; s.
func (c *ContentStream) CloseStroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("s\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}

}

// Fill path using the non-zero winding rule; f.
func (c *ContentStream) Fill() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("f\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Fill path using the even-odd rule; f*.
func (c *ContentStream) FillEO() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("f*\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Fill path using the non-zero winding rule and then stroke; B.
func (c *ContentStream) FillStroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("B\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Fill path using the even-odd winding rule and then stroke; B*.
func (c *ContentStream) FillEOStroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("B*\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Close path, fill path using the non-zero winding rule, then stroke path; b.
func (c *ContentStream) CloseFillStroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("b\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Close path, fill path using the even-odd winding rule, then stroke path; b*.
func (c *ContentStream) CloseFillEOStroke() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("b*\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// End the current path. Used primarily to apply changes to the current clipping path; n.
func (c *ContentStream) EndPath() {
	switch c.PathState {
	case PATH_BUILDING, PATH_CLIPPING:
		c.buf.Write([]byte("n\n"))
		c.PathState = PATH_NONE
		c.CurPt = *new(Point)
	}
}

// Append a rectangle, of width w and height h, starting at the point (X, Y), to the current path; re.
func (c *ContentStream) Re(x, y, w, h float64) {
	c.PathState = PATH_BUILDING
	c.CurPt = Point{x, y}
	fmt.Fprintf(c.buf, "%f %f %f %f re\n", x, y, w, h)
}
