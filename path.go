package gdf

// Begin a new path starting at (x, y); m.
func (c *ContentStream) MoveTo(x, y float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	c.buf = cmdf(c.buf, op_m, x, y)
}

// Append a straight line from the current point to (x, y), which becomes the new current point; l.
func (c *ContentStream) LineTo(x, y float64) {
	if c.PathState == Building {
		c.CurPt = Point{x, y}
		c.buf = cmdf(c.buf, op_l, x, y)
	}
}

// A convenience function that strokes a line from Point{x1,y1} to Point{x2,y2}.
func (c *ContentStream) DrawLine(x1, y1, x2, y2 float64) {
	c.MoveTo(x1, y1)
	c.LineTo(x2, y2)
	c.Stroke()
}

// Append a cubic Bézier curve to the current path; c. The curve extends
// from the current point to (x3, y3) using using (x1, y1) and (x2, y2) as the Bézier control points.
func (c *ContentStream) CubicBezier1(x1, y1, x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.buf = cmdf(c.buf, op_c, x1, y1, x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; v. The curve extends from
// the current point to (x3, y3), using the current point and (x2, y2) as the Bézier control points.
func (c *ContentStream) CubicBezier2(x2, y2, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.buf = cmdf(c.buf, op_v, x2, y2, x3, y3)
	}
}

// Append a cubic Bézier curve to the current path; y. The curve extends
// from the current point to (x3, y3), using (x1, y1) and (x3, y3) as the Bézier control points
func (c *ContentStream) CubicBezier3(x1, y1, x3, y3 float64) {
	switch c.PathState {
	case NoPath, Clipping:
		return
	default:
		c.buf = cmdf(c.buf, op_y, x1, y1, x3, y3)
	}
}

// Close the current path; h.
func (c *ContentStream) ClosePath() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = Building
		c.buf = append(c.buf, op_h...)
	}
}

// Stroke path; S.
func (c *ContentStream) Stroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_S...)
	}
}

// Close and stroke path; s.
func (c *ContentStream) ClosePathStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_s...)
	}

}

// Fill path using the non-zero winding rule; f.
func (c *ContentStream) Fill() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_f...)
	}
}

// Fill path using the even-odd rule; f*.
func (c *ContentStream) FillEO() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_f_X...)
	}
}

// Fill path using the non-zero winding rule and then stroke; B.
func (c *ContentStream) FillStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_B...)
	}
}

// Fill path using the even-odd winding rule and then stroke; B*.
func (c *ContentStream) FillEOStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_B_X...)
	}
}

// Close path, fill path using the non-zero winding rule, then stroke path; b.
func (c *ContentStream) ClosePathFillStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_b...)
	}
}

// Close path, fill path using the even-odd winding rule, then stroke path; b*.
func (c *ContentStream) ClosePathFillEOStroke() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_b_X...)
	}
}

// End the current path. Used primarily to apply changes to the current clipping path; n.
func (c *ContentStream) EndPath() {
	switch c.PathState {
	case Building, Clipping:
		c.PathState = NoPath
		c.CurPt = *new(Point)
		c.buf = append(c.buf, op_n...)
	}
}

// Append a rectangle, of width w and height h, starting at the point (X, Y), to the current path; re.
func (c *ContentStream) Re(x, y, w, h float64) {
	c.PathState = Building
	c.CurPt = Point{x, y}
	c.buf = cmdf(c.buf, op_re, x, y, w, h)
}

// Append r to the current path; a possibly more convenient version of Re.
func (c *ContentStream) Re2(r Rect) {
	c.Re(r.LLX, r.LLY, r.URX-r.LLX, r.URY-r.LLY)
}

// Clip path (non-zero winding). A clipping path operator. May appear after the last path
// construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) Clip() {
	switch c.PathState {
	case Building:
		c.PathState = Clipping
		c.buf = append(c.buf, op_W...)
	}
}

// Clip path (even-odd). A clipping path operator. May appear after the last path
// construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) ClipEO() {
	switch c.PathState {
	case Building:
		c.PathState = Clipping
		c.buf = append(c.buf, op_W_X...)
	}
}
