package gdf

import (
	"bytes"
	"fmt"
	"io"
)

type metadata struct {
	refnum int
	buf    *bytes.Buffer
}

// Includes the XMP data represented by b as document-level metadata for the encoded PDF. Currently, gdf does not perform any validation on b.
func (p *PDF) AddXMPMetaData(b []byte) {
	m := new(metadata)
	m.buf = new(bytes.Buffer)
	m.buf.Write(b)
	p.catalog.streams = append(p.catalog.streams, m)
}

func (m *metadata) setRef(i int)    { m.refnum = i }
func (m *metadata) refNum() int     { return m.refnum }
func (m *metadata) children() []obj { return []obj{} }
func (m *metadata) encode(w io.Writer) (int, error) {
	var n int
	t, err := fmt.Fprintf(w, "<<\n/Type /Metadata\n/Subtype /XML\n/Length %d\n>>\nstream\n", m.buf.Len())
	if err != nil {
		return t, err
	}
	n += t
	t2, err := m.buf.WriteTo(w)
	n += int(t2)
	if err != nil {
		return n, err
	}

	t, err = w.Write([]byte("\nendstream\n"))
	if err != nil {
		return n + t, err
	}
	return n + t, nil
}
