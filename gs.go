package gdf

import "fmt"

// Graphics State
type GS struct {
	Matrix // CTM
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

// save graphic state to the stack
func (c *ContentStream) QSmall() {
	g := c.GS
	c.GStack = append(c.GStack, g)
	c.Stack = append(c.Stack, G_STATE)
	c.buf.Write([]byte("q\n"))
}

// restore graphic state from the stack
func (c *ContentStream) Q() error {
	if c.Stack[len(c.Stack)-1] != G_STATE {
		return fmt.Errorf("cannot interleave q/Q and BT/ET pairs")
	}
	c.Stack = c.Stack[:len(c.Stack)-1]
	c.GS = c.GStack[len(c.GStack)-1]
	c.GStack = c.GStack[:len(c.GStack)-1]
	c.buf.Write([]byte("Q\n"))
	return nil
}

// Concatenate t to the CTM
func (c *ContentStream) Cm(m Matrix) {
	c.GS.Matrix = Mul(c.GS.Matrix, m)
	fmt.Fprintf(c.buf, "%f %f %f %f %f %f cm\n", m.A, m.B, m.C, m.D, m.E, m.F)
}

// Set linewidth to f
func (c *ContentStream) WSmall(f float64) {
	c.LineWidth = f
	fmt.Fprintf(c.buf, "%f w\n", f)
}

// Set the line cap style
func (c *ContentStream) J(lc LineCap) {
	c.LineCap = lc
	fmt.Fprintf(c.buf, "%d J\n", lc)
}

// set the line join style
func (c *ContentStream) JSmall(lj LineJoin) {
	c.LineJoin = lj
	fmt.Fprintf(c.buf, "%d j\n", lj)
}

// set miter limit
func (c *ContentStream) M(ml float64) {
	c.MiterLimit = ml
	fmt.Fprintf(c.buf, "%f M\n", ml)
}

// set the dash pattern
func (c *ContentStream) DSmall(d DashPattern) {
	c.DashPattern = d
	fmt.Fprintf(c.buf, "%v %d d\n", d.Array, d.Phase)
}

// set the rednering intent
func (c *ContentStream) Ri(n Name) {
	c.RenderingIntent = n
	fmt.Fprintf(c.buf, "%s ri\n", ToString(n))
}

// set the flatness
func (c *ContentStream) ISmall(f float64) {
	c.Flatness = f
	fmt.Fprintf(c.buf, "%f i\n", f)
}

func (c *ContentStream) XGraphicState(e ExtGState) {
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
