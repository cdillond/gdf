package gdf

import (
	"fmt"
	"io"
)

type Catalog struct {
	Pages  *Pages
	refnum int
}

func (c *Catalog) RefNum() int { return c.refnum }
func (c *Catalog) Children() []Obj {
	return []Obj{c.Pages}
}
func (c *Catalog) ToBytes() []byte { return []byte{} }
func (c *Catalog) SetRef(i int)    { c.refnum = i }
func (c *Catalog) Encode(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "<<\n/Type /Catalog\n/Pages %d 0 R\n>>\n", c.Pages.RefNum())
}
