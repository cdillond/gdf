package gdf

import (
	"fmt"
	"io"
)

type catalog struct {
	Pages  *Pages
	refnum int
}

func (c *catalog) refNum() int { return c.refnum }
func (c *catalog) children() []obj {
	return []obj{c.Pages}
}
func (c *catalog) setRef(i int) { c.refnum = i }
func (c *catalog) encode(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "<<\n/Type /Catalog\n/Pages %d 0 R\n>>\n", c.Pages.refNum())
}
