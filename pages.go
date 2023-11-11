package gdf

import (
	"fmt"
	"io"
)

type Pages struct {
	P      []*Page
	refnum int
}

func (p *Pages) RefNum() int { return p.refnum }
func (p *Pages) Children() []Obj {
	out := make([]Obj, 0, len(p.P))
	for _, page := range p.P {
		out = append(out, page)
	}
	return out
}
func (p *Pages) SetRef(i int) { p.refnum = i }
func (p *Pages) Encode(w io.Writer) (int, error) {
	kids := make([]string, len(p.P))
	for i := range p.P {
		kids[i] = fmt.Sprintf("%d 0 R", p.P[i].RefNum())
	}
	return fmt.Fprintf(w, "<<\n/Type /Pages\n/Kids %v\n/Count %d\n>>\n",
		kids, len(kids))
}
