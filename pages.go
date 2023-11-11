package gdf

import (
	"fmt"
	"io"
)

type Pages struct {
	P      []*Page
	refnum int
}

func (p *Pages) refNum() int { return p.refnum }
func (p *Pages) children() []Obj {
	out := make([]Obj, 0, len(p.P))
	for _, page := range p.P {
		out = append(out, page)
	}
	return out
}
func (p *Pages) setRef(i int) { p.refnum = i }
func (p *Pages) encode(w io.Writer) (int, error) {
	kids := make([]string, len(p.P))
	for i := range p.P {
		kids[i] = fmt.Sprintf("%d 0 R", p.P[i].refNum())
	}
	return fmt.Fprintf(w, "<<\n/Type /Pages\n/Kids %v\n/Count %d\n>>\n",
		kids, len(kids))
}
