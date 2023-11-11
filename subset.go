package gdf

import (
	"bytes"
	"fmt"
	"sort"

	cfnt "github.com/tdewolff/canvas/font"
)

func (f *Font) Subset() error {
	if !f.subset {
		return nil
	}
	glyphs := make([]uint16, 0, len(f.charset))
	for r := range f.charset {
		gid, _ := f.GlyphIndex(f.buf, r)
		if gid == 0 {
			continue
		}
		glyphs = append(glyphs, uint16(gid))
	}
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })
	if len(glyphs) < 1 {
		return fmt.Errorf("too few characters")
	}

	b := cfnt.FromGoSFNT(f.Font)
	c, err := cfnt.ParseSFNT(b, 0)
	if err != nil {
		return err
	}
	subfnt, _ := c.Subset(glyphs, cfnt.WritePDFTables)
	sbuf := new(bytes.Buffer)
	sbuf.Read(subfnt)
	f.source.buf = sbuf
	return nil
}
