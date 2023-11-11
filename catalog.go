package gdf

import (
	"fmt"
	"io"
)

type Catalog struct {
	Pages  *Pages
	refnum int
}

func (c *Catalog) refNum() int { return c.refnum }
func (c *Catalog) children() []Obj {
	return []Obj{c.Pages}
}
func (c *Catalog) setRef(i int) { c.refnum = i }
func (c *Catalog) encode(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "<<\n/Type /Catalog\n/Pages %d 0 R\n>>\n", c.Pages.refNum())
}
