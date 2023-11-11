package gdf

import (
	"bytes"
	"fmt"
	"sort"

	cfnt "github.com/tdewolff/canvas/font"
)

// Writes a TrueType font representing the subset of f defined by f's internal charset to f's internal source buffer.
func (f *Font) subset() error {
	if f.noSubset {
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

	//b := cfnt.FromGoSFNT(f.Font)
	//c, err := cfnt.ParseSFNT(b, 0)
	//if err != nil {
	//	return err
	//}
	f.m.Lock()
	subfnt, _ := f.c.Subset(glyphs, cfnt.WritePDFTables)
	f.m.Unlock()
	sbuf := new(bytes.Buffer)
	sbuf.Read(subfnt)
	f.source.buf = sbuf
	return nil
}
