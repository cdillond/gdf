package gdf

import "fmt"

// Graphics State
type GS struct {
	Matrix // Current Transformation Matrix
	TextState
	LineCap
	LineJoin
	DashPattern
	PathState
	CurPt           Point      // Current path point
	NColorSpace     ColorSpace // non-stroking
	SColorSpace     ColorSpace // stroking
	NColor          Color      // non-stroking
	SColor          Color      // stroking
	LineWidth       float64
	MiterLimit      float64
	RenderingIntent Name
	StrokeAdj       bool
	BlendMode       Name
	SoftMask        Name
	AlphaConstant   float64
	AlphaSource     bool
	BPComp          Name
	Overprint       bool
	OverprintMode   uint
	BlackGen        Name
	UndercolorRem   Name
	Transfer        Name
	Halftone        Name
	Flatness        float64
	Smoothness      float64
}

type LineCap uint

const (
	BUTT_CAP LineCap = iota
	ROUND_CAP
	SQUARE_CAP
)

type LineJoin uint

const (
	MITER_JOIN LineJoin = iota
	ROUND_JOIN
	BEVEL_JOIN
)

type DashPattern struct {
	Array []int
	Phase int
}

type PathState uint

const (
	PATH_NONE PathState = iota
	PATH_BUILDING
	PATH_CLIPPING
)

func NewGS() *GS {
	out := new(GS)
	out.Scale = 100
	out.Matrix = NewMatrix()
	return out
}

// QSave pushes the current GS (graphics sate) to c's GSStack (graphics state stack).
func (c *ContentStream) QSave() {
	g := c.GS
	c.gSStack = append(c.gSStack, g)
	c.stack = append(c.stack, g_STATE)
	c.buf.Write([]byte("q\n"))
}

// QRestore pops the most recent addition to c's GSStack off the stack and
// sets c's current GS equal to that value. It returns an error if c's GSStack
// is empty or if c.BeginText() has been called after the last call to c.QSave(),
// and the returned EndText function has not yet been called.
func (c *ContentStream) QRestore() error {
	if len(c.stack) == 0 {
		return fmt.Errorf("current GSStack is empty")
	}
	if c.stack[len(c.stack)-1] != g_STATE {
		return fmt.Errorf("cannot interleave q/Q and BT/ET pairs")
	}
	c.stack = c.stack[:len(c.stack)-1]
	c.GS = c.gSStack[len(c.gSStack)-1]
	c.gSStack = c.gSStack[:len(c.gSStack)-1]
	c.buf.Write([]byte("Q\n"))
	return nil
}

// Sets c's Current Transformation Matrix (c.GS.Matrix) to the matrix product of m and c.GS.Matrix.
func (c *ContentStream) Cm(m Matrix) {
	c.GS.Matrix = Mul(c.GS.Matrix, m)
	fmt.Fprintf(c.buf, "%f %f %f %f %f %f cm\n", m.A, m.B, m.C, m.D, m.E, m.F)
}

// Sets the linewidth (c.GS.LineWidth) to f.
func (c *ContentStream) SetLineWidth(f float64) {
	c.LineWidth = f
	fmt.Fprintf(c.buf, "%f w\n", f)
}

// Sets the line cap style (c.GS.LineCap) to lc.
func (c *ContentStream) SetLineCap(lc LineCap) {
	c.LineCap = lc
	fmt.Fprintf(c.buf, "%d J\n", lc)
}

// Sets the line join style (c.GS.LineJoin) to lj.
func (c *ContentStream) SetLineJoin(lj LineJoin) {
	c.LineJoin = lj
	fmt.Fprintf(c.buf, "%d j\n", lj)
}

// Sets miter limit (c.GS.MiterLimit) to ml.
func (c *ContentStream) SetMiterLimit(ml float64) {
	c.MiterLimit = ml
	fmt.Fprintf(c.buf, "%f M\n", ml)
}

// Sets the dash pattern (c.GS.DashPattern) to d.
func (c *ContentStream) SetDashPattern(d DashPattern) {
	c.DashPattern = d
	fmt.Fprintf(c.buf, "%v %d d\n", d.Array, d.Phase)
}

// Sets the rendering intent (c.GS.RenderingIntent) to n.
func (c *ContentStream) SetRenderIntent(n Name) {
	c.RenderingIntent = n
	fmt.Fprintf(c.buf, "%s ri\n", n)
}

// Set the flatness (c.GS.Flatness) to f.
func (c *ContentStream) SetFlattness(f float64) {
	c.Flatness = f
	fmt.Fprintf(c.buf, "%f i\n", f)
}

func (c *ContentStream) XGraphicsState(e ExtGState) {
	var i int
	for ; i < len(c.Parent.ExtGState); i++ {
		if c.Parent.ExtGState[i] == &e {
			break
		}
	}
	if i == len(c.Parent.ExtGState) {
		c.Parent.ExtGState = append(c.Parent.ExtGState, &e)
	}
	c.ExtGState = e
	fmt.Fprintf(c.buf, "/GS%d gs\n", i)
}
