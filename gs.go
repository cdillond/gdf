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
	RenderingIntent string
	StrokeAdj       bool
	BlendMode       string
	SoftMask        string
	AlphaConstant   float64
	AlphaSource     bool
	BPComp          string
	Overprint       bool
	OverprintMode   uint
	BlackGen        string
	UndercolorRem   string
	Transfer        string
	Halftone        string
	Flatness        float64
	Smoothness      float64
}

type LineCap uint

const (
	ButtCap LineCap = iota
	RoundCap
	SquareCap
)

type LineJoin uint

const (
	MiterJoin LineJoin = iota
	RoundJoin
	BevelJoin
)

type DashPattern struct {
	Array []int
	Phase int
}

type PathState uint

const (
	NoPath PathState = iota
	Building
	Clipping
)

func NewGS() GS {
	out := new(GS)
	out.HScale = 100
	out.Matrix = NewMatrix()
	return *out
}

// QSave pushes the current GS (graphics sate) to c's GSStack (graphics state stack).
func (c *ContentStream) QSave() {
	g := c.GS
	c.gSStack = append(c.gSStack, g)
	c.stack = append(c.stack, gState)
	c.buf = append(c.buf, op_q...)
	//c.buf.Write([]byte("q\n"))
}

// QRestore pops the most recent addition to c's GSStack off the stack and
// sets c's current GS equal to that value. It returns an error if c's GSStack
// is empty or if c.BeginText() has been called after the last call to c.QSave(),
// and the returned EndText function has not yet been called.
func (c *ContentStream) QRestore() error {
	if len(c.stack) == 0 {
		return fmt.Errorf("current GSStack is empty")
	}
	if c.stack[len(c.stack)-1] != gState {
		return fmt.Errorf("cannot interleave q/Q and BT/ET pairs")
	}
	c.stack = c.stack[:len(c.stack)-1]
	c.GS = c.gSStack[len(c.gSStack)-1]
	c.gSStack = c.gSStack[:len(c.gSStack)-1]
	c.buf = append(c.buf, op_Q...)
	//c.buf.Write([]byte("Q\n"))
	return nil
}

// Sets c's Current Transformation Matrix (c.GS.Matrix) to the matrix product of m and c.GS.Matrix.
func (c *ContentStream) Concat(m Matrix) {
	c.GS.Matrix = Mul(m, c.GS.Matrix) // NOT COMMUTATIVE, THIS ORDER MUST REMAIN THE SAME
	c.buf = cmdf(c.buf, op_cm, m.A, m.B, m.C, m.D, m.E, m.F)
}

// Sets the linewidth (c.GS.LineWidth) to f.
func (c *ContentStream) SetLineWidth(f float64) {
	c.LineWidth = f
	c.buf = cmdf(c.buf, op_w, f)
}

// Sets the line cap style (c.GS.LineCap) to lc.
func (c *ContentStream) SetLineCap(lc LineCap) {
	c.LineCap = lc
	c.buf = cmdi(c.buf, op_J, int64(lc))
}

// Sets the line join style (c.GS.LineJoin) to lj.
func (c *ContentStream) SetLineJoin(lj LineJoin) {
	c.LineJoin = lj
	c.buf = cmdi(c.buf, op_j, int64(lj))
}

// Sets miter limit (c.GS.MiterLimit) to ml.
func (c *ContentStream) SetMiterLimit(ml float64) {
	c.MiterLimit = ml
	c.buf = cmdf(c.buf, op_M, ml)
}

// Sets the dash pattern (c.GS.DashPattern) to d.
func (c *ContentStream) SetDashPattern(d DashPattern) {
	c.DashPattern = d
	c.buf = append(c.buf, fmt.Sprintf("%v %d d\n", d.Array, d.Phase)...)
}

// Sets the rendering intent (c.GS.RenderingIntent) to n.
func (c *ContentStream) SetRenderIntent(n string) {
	c.RenderingIntent = n
	c.buf = append(c.buf, n...)
	c.buf = append(c.buf, '\x20')
	c.buf = append(c.buf, op_ri...)
}

// Set the flatness (c.GS.Flatness) to f.
func (c *ContentStream) SetFlatness(f float64) {
	c.Flatness = f
	c.buf = cmdf(c.buf, op_i, f)
}
