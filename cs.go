package gdf

import (
	"bytes"
	"fmt"
	"io"
)

type ContentStream struct {
	Filter
	*GS
	*TextObj
	//TODO: ExtGState
	gSStack []*GS        // Graphics state stack
	stack   []stackState // used for recording the order of calls to QSave/QRestore and BeginText/EndText
	Parent  *Page
	buf     *bytes.Buffer
	refnum  int
	//*Type0Font
}

type stackState uint8

const (
	gState stackState = iota
	tState
)

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
		return nil, fmt.Errorf("text objects cannot be statically nested")
	}
	c.TextObj = &TextObj{
		Matrix:     NewMatrix(),
		LineMatrix: NewMatrix(),
	}
	c.stack = append(c.stack, tState)
	_, err := c.buf.Write([]byte("BT\n"))
	if err != nil {
		return nil, err
	}
	return func() error {
		if c.TextObj == nil {
			return fmt.Errorf("text object is already closed")
		}
		c.TextObj = nil
		c.stack = c.stack[:len(c.stack)-1]
		_, err := c.buf.Write([]byte("ET\n"))
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func (c *ContentStream) setRef(i int) { c.refnum = i }
func (c *ContentStream) refNum() int  { return c.refnum }
func (c *ContentStream) children() []obj {
	//if c.ExtGState.Dict != nil {
	//	return []obj{&c.ExtGState}
	//}
	return []obj{}
}
func (c *ContentStream) encode(w io.Writer) (int, error) {
	var n int
	switch c.Filter {
	case Flate:
		encbuf := new(bytes.Buffer)
		l1 := c.buf.Len()
		_, err := flateCompress(encbuf, c.buf)
		if err != nil {
			return 0, err
		}
		t, err := fmt.Fprintf(w, "<<\n/Filter /FlateDecode\n/Length1 %d\n/Length %d\n>>\nstream\n", l1, encbuf.Len())
		if err != nil {
			return t, err
		}
		encbuf.Write([]byte("\nendstream\n"))
		t2, err := encbuf.WriteTo(w)
		if err != nil {
			return t + int(t2), err
		}
		return t + int(t2), err
	default:
		t, err := fmt.Fprintf(w, "<<\n/Length %d\n>>\nstream\n", c.buf.Len())
		if err != nil {
			return t, err
		}
		n += t
		t2, err := c.buf.WriteTo(w)
		n += int(t2)
		if err != nil {
			return n, err
		}
	}
	t, err := w.Write([]byte("endstream\n"))
	if err != nil {
		return n + t, err
	}
	return n + t, nil
}
