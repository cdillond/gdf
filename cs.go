package gdf

import (
	"fmt"
	"io"
)

var (
	ErrNested = fmt.Errorf("text objects cannot be statically nested")
	ErrClosed = fmt.Errorf("text object is already closed")
)

type ContentStream struct {
	stream
	GS
	*TextObj
	//TODO: ExtGState
	gSStack   []GS         // Graphics state stack
	stack     []stackState // used for recording the order of calls to QSave/QRestore and BeginText/EndText
	resources resourceDict
	refnum    int
}

type stackState uint8

const (
	gState stackState = iota
	tState
)

// An EndText function return by c.BeginText() must be invoked to close a section of text written to c.
type EndText func() error

// BeginText declares a new text object within the ContentStream. It must be called before drawing
// any text to c. It returns an EndText function, which must be called to close the text object, and
// an error. All successive calls to BeginText before EndText is called will result in an error.
// Pairs of BeginText/EndText calls should not be interleaved with pairs of QSave/Restore calls,
// although each pair can fully contain instances of the other pair.
// BeginText automatically sets the current Text Matrix and the Line Matrix equal to the identity matrix.
// If you do not wish for all glyphs to appear at the origin, you must also adjust the current Text Matrix.
func (c *ContentStream) BeginText() (EndText, error) {
	if c.TextObj != nil {
		return nil, ErrNested
	}
	c.TextObj = &TextObj{
		Matrix:     NewMatrix(),
		LineMatrix: NewMatrix(),
	}
	c.stack = append(c.stack, tState)
	c.buf = append(c.buf, op_BT...)

	return func() error {
		if c.TextObj == nil {
			return ErrClosed
		}
		c.TextObj = nil
		c.stack = c.stack[:len(c.stack)-1]
		c.buf = append(c.buf, op_ET...)
		return nil
	}, nil
}

func (c *ContentStream) mark(i int) { c.refnum = i }
func (c *ContentStream) id() int    { return c.refnum }
func (c *ContentStream) children() []obj {
	return c.stream.children()
}
func (c *ContentStream) encode(w io.Writer) (int, error) {
	return c.stream.encode(w)
}
