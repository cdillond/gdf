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
	ExtGState
	GStack []*GS        // Graphics state stack
	Stack  []StackState // Records the type of the most recent object pushed to a stack
	Parent *Page
	buf    *bytes.Buffer
	refnum int
}

type StackState uint8

const (
	G_STATE StackState = iota
	T_STATE
)

func (c *ContentStream) Close() {
	for i := len(c.Stack) - 1; i >= 0; i-- {
		switch c.Stack[i] {
		case G_STATE:
			c.Q()
		case T_STATE:
			c.ET()
		}
	}
}

func (c *ContentStream) BT() error {
	if c.TextObj != nil {
		return fmt.Errorf("text objects cannot be statically nested")
	}
	c.TextObj = &TextObj{
		Matrix:     NewMatrix(),
		LineMatrix: NewMatrix(),
	}
	c.Stack = append(c.Stack, T_STATE)
	_, err := c.buf.Write([]byte("BT\n"))
	if err != nil {
		return err
	}
	return nil
}

func (c *ContentStream) ET() error {
	if c.TextObj == nil {
		return fmt.Errorf("text object is already closed")
	}
	c.Stack = c.Stack[:len(c.Stack)-1]
	_, err := c.buf.Write([]byte("ET\n"))
	if err != nil {
		return err
	}
	return nil
}

func (c *ContentStream) setRef(i int) { c.refnum = i }
func (c *ContentStream) refNum() int  { return c.refnum }
func (c *ContentStream) children() []Obj {
	if c.ExtGState.Dict != nil {
		return []Obj{&c.ExtGState}
	}
	return []Obj{}
}
func (c *ContentStream) encode(w io.Writer) (int, error) {
	if c.buf.Len() > 1024 {
		c.Filter = FILTER_FLATE
	}
	var n int
	switch c.Filter {
	case FILTER_FLATE:
		encbuf := new(bytes.Buffer)
		l1 := c.buf.Len()
		_, err := FlateCompress(encbuf, c.buf)
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
	t, err := w.Write([]byte(">>\nendstream\n"))
	if err != nil {
		return n + t, err
	}
	return n + t, nil
}

func (c *ContentStream) SetRGB(r, g, b float64) {
	fmt.Fprintf(c.buf, "%f %f %f rg\n", r, g, b)
}
func (c *ContentStream) SetRGBStroking(r, g, b float64) {
	fmt.Fprintf(c.buf, "%f %f %f RG\n", r, g, b)
}
func (c *ContentStream) SetG(g float64) {
	fmt.Fprintf(c.buf, "%f g\n", g)
}
func (c *ContentStream) SetGStroking(g float64) {
	fmt.Fprintf(c.buf, "%f G\n", g)
}

// Clip path (non-zero winding).
// A clipping path operator (W or W*, shown in "Table 60 — Clipping path operators") may appear after
// the last path construction operator and before the path-painting operator that terminates a path object.
func (c *ContentStream) W() {
	switch c.PathState {
	case PATH_BUILDING:
		c.PathState = PATH_CLIPPING
		c.buf.Write([]byte("W\n"))
	}
}

// Clip path (even-odd).
func (c *ContentStream) WStar() {
	switch c.PathState {
	case PATH_BUILDING:
		c.PathState = PATH_CLIPPING
		c.buf.Write([]byte("W*\n"))
	}
}
