package gdf

import "fmt"

// Begin a new path starting at (x, y); m.
func (c *ContentStream) MoveTo(x, y float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	fmt.Fprintf(c.buf, "%f %f m\n", x, y)
}

// Append a straight line from the current point to (x, y); l.
func (c *ContentStream) LineTo(x, y float64) {
	fmt.Fprintf(c.buf, "%f %f l\n", x, y)
}

// Append a cubic Bézier curve to the current path; c. The curve extends
// from the current point to (x3, y3) using using (x1, y1) and (x2, y2) as the Bézier control points.
func (c *ContentStream) CubicBezier1(x1, y1, x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f %f %f c\n", x1, y1, x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; v. The curve extends from
// the current point to (x3, y3), using the current point and (x2, y2) as the Bézier control points.
func (c *ContentStream) CubicBezier2(x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f v\n", x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; y. The curve extends
// from the current point to (x3, y3), using (x1, y1) and (x3, y3) as the Bézier control points
func (c *ContentStream) CubicBezier3(x1, y1, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		fmt.Fprintf(c.buf, "%f %f %f %f c", x1, y1, x3, y3)
	}
}

// Close the current path; h.
func (c *ContentStream) ClosePath() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("h\n"))
		c.PathState = Building
	}
}

// Stroke path; S.
func (c *ContentStream) Stroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("S\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Close and stroke path; s.
func (c *ContentStream) ClosePathStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("s\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}

}

// Fill path using the non-zero winding rule; f.
func (c *ContentStream) Fill() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("f\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Fill path using the even-odd rule; f*.
func (c *ContentStream) FillEO() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("f*\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Fill path using the non-zero winding rule and then stroke; B.
func (c *ContentStream) FillStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("B\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Fill path using the even-odd winding rule and then stroke; B*.
func (c *ContentStream) FillEOStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("B*\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Close path, fill path using the non-zero winding rule, then stroke path; b.
func (c *ContentStream) ClosePathFillStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("b\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Close path, fill path using the even-odd winding rule, then stroke path; b*.
func (c *ContentStream) ClosePathFillEOStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("b*\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// End the current path. Used primarily to apply changes to the current clipping path; n.
func (c *ContentStream) EndPath() {
	switch c.PathState {
	case Building, Clipping:
		c.buf.Write([]byte("n\n"))
		c.PathState = NoPath
		c.CurPt = *new(Point)
	}
}

// Append a rectangle, of width w and height h, starting at the point (X, Y), to the current path; re.
func (c *ContentStream) Re(x, y, w, h float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	fmt.Fprintf(c.buf, "%f %f %f %f re\n", x, y, w, h)
}

// Append r to the current path; a possibly more convenient version of Re.
func (c *ContentStream) Re2(r Rect) {
	c.PathState = Building
	c.CurPt = Point{r.LLX, r.LLY}
	fmt.Fprintf(c.buf, "%f %f %f %f re\n", r.LLX, r.LLY, r.URX-r.LLX, r.URY-r.LLY)
}

// Clip path (non-zero winding). A clipping path operator. May appear after the last path
// construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) Clip() {
	switch c.PathState {
	case Building:
		c.PathState = Clipping
		c.buf.Write([]byte("W\n"))
	}
}

// Clip path (even-odd). A clipping path operator. May appear after the last path
// construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) ClipEO() {
	switch c.PathState {
	case Building:
		c.PathState = Clipping
		c.buf.Write([]byte("W*\n"))
	}
}
