package gdf

import (
	"bytes"
	"sort"

	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
	"golang.org/x/text/encoding/charmap"
)

var wdec [256]rune

func init() {
	out := [256]rune{}
	dec := charmap.Windows1252.NewDecoder()
	for i := 0; i < 256; i++ {
		db, err := dec.String(string([]byte{byte(i)}))
		if err != nil {
			continue
		}
		rdb := []rune(db)
		if len(db) > 0 {
			out[i] = rdb[0]
		}
	}
	wdec = out
}

var pdfTables = []uint{
	//1330851634, // OS/2
	1668112752, // cmap
	1735162214, // glyf
	1751474532, // head
	1751672161, // hhea
	1752003704, // hmtx
	1819239265, // loca
	1835104368, // maxp
	//1851878757, // name
	//1886352244, // post
}

// Zeroes out unused outlines contained in the font's glyf table and writes only the required tables from the source font to the dst font. Does not work for OpenType fonts with CFF outlines.
func (f *Font) subset() error {
	if f.noSubset {
		return nil
	}
	f.noSubset = true
	glyphs := make([]uint32, 256)
	glyphset := make(map[uint32]struct{}, len(glyphs))
	for i, r := range wdec {
		gid, _ := f.GlyphIndex(f.buf, r)
		if gid == 0 {
			continue
		}
		glyphs[i] = uint32(gid)
		glyphset[uint32(gid)] = struct{}{}
	}

	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })

	rsrc := bytes.NewReader(f.srcb)
	ld, err := loader.NewLoader(rsrc)
	if err != nil {
		return err
	}
	head, err := ld.RawTable(1751474532)
	if err != nil {
		return err
	}
	headP, _, err := tables.ParseHead(head)
	if err != nil {
		return err
	}
	isLong := headP.IndexToLocFormat == 1

	loca, err := ld.RawTable(1819239265)
	if err != nil {
		return err
	}
	locaP, err := tables.ParseLoca(loca, f.NumGlyphs(), isLong)
	if err != nil {
		panic(err)
	}
	glyfs, err := ld.RawTable(1735162214)
	if err != nil {
		panic(err)
	}
	var composits = []uint32{}
	for i := 0; i < len(glyphs)-1; i++ {
		if glyphs[i]+1 >= uint32(len(locaP)) {
			continue
		}
		offset := locaP[glyphs[i]]
		next := locaP[glyphs[i]+1]
		if offset > uint32(len(glyfs))-2 || next > uint32(len(glyfs))-1 || offset >= next {
			break
		}
		cg, _, err := tables.ParseCompositeGlyph(glyfs[offset:next])
		if err != nil {
			continue
		}
		for _, gp := range cg.Glyphs {
			composits = append(composits, uint32(gp.GlyphIndex))
			glyphset[uint32(gp.GlyphIndex)] = struct{}{}
		}
	}

	glyphs = append(glyphs, composits...)
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })

	for i := 0; i < len(locaP)-1; i++ {
		// assuming this works the same for short and long formats once they've been parsed
		offset := locaP[i]
		next := locaP[i+1]

		if next == 0 || next < offset {
			continue
		}
		if offset >= uint32(len(glyfs)) || next >= uint32(len(glyfs)) {
			break
		}

		if _, used := glyphset[uint32(i)]; used {
			continue
		}

		for j := offset; j < next; j++ {
			glyfs[j] = 0
		}

	}

	tbl := make([]loader.Table, len(pdfTables))
	for i, tag := range pdfTables {
		cnt, err := ld.RawTable(loader.Tag(tag))
		if err != nil {
			return err
		}

		if tag == 1735162214 {
			tbl[i] = loader.Table{Content: glyfs, Tag: loader.Tag(tag)}
			continue
		}
		tbl[i] = loader.Table{Content: cnt, Tag: loader.Tag(tag)}
	}
	f.source.buf.Write(loader.WriteTTF(tbl))
	return nil
}
