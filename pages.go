package gdf

import (
	"io"
)

type pages struct {
	P      []*Page
	refnum int
}

func (p *pages) id() int { return p.refnum }
func (p *pages) children() []obj {
	out := make([]obj, 0, len(p.P))
	for _, page := range p.P {
		out = append(out, page)
	}
	return out
}
func (p *pages) mark(i int) { p.refnum = i }
func (p *pages) encode(w io.Writer) (int, error) {
	l := len(p.P)
	kids := make([]string, l)
	for i := range p.P {
		kids[i] = iref(p.P[i])
	}
	return w.Write(dict(1024, []field{
		{"/Type", "/Pages"},
		{"/Kids", kids},
		{"/Count", l},
	}))
}
