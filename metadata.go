package gdf

import (
	"io"
)

type metadata struct {
	refnum int
	stream
}

// Includes the XMP data represented by b as document-level metadata for the encoded PDF. Currently, gdf does not perform any validation on b.
func (p *PDF) AddXMPMetaData(b []byte) {
	m := new(metadata)
	m.buf = b
	p.catalog.streams = append(p.catalog.streams, m)
}

func (m *metadata) mark(i int)      { m.refnum = i }
func (m *metadata) id() int         { return m.refnum }
func (m *metadata) children() []obj { return nil } // Since this is not compressed, the embedded stream will always be childless.
func (m *metadata) encode(w io.Writer) (int, error) {
	m.extras = []field{
		{"/Type", "/Metadata"},
		{"/Subtype", "/XML"},
	}
	return m.stream.encode(w)
}
