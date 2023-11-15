package gdf

import (
	"fmt"
	"io"
)

type catalog struct {
	Pages   *pages
	streams []obj
	refnum  int
}

func (c *catalog) refNum() int { return c.refnum }
func (c *catalog) children() []obj {
	return append([]obj{c.Pages}, c.streams...)
}
func (c *catalog) setRef(i int) { c.refnum = i }
func (c *catalog) encode(w io.Writer) (int, error) {
	if len(c.streams) != 0 {
		return fmt.Fprintf(w, "<<\n/Type /Catalog\n/Pages %d 0 R\n/Metadata %d 0 R\n>>\n", c.Pages.refNum(), c.streams[0].refNum())
	}
	return fmt.Fprintf(w, "<<\n/Type /Catalog\n/Pages %d 0 R\n>>\n", c.Pages.refNum())
}
