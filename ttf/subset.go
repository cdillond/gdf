package ttf

import (
	"bytes"
	"encoding/binary"
	"sort"

	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
	"golang.org/x/image/font/sfnt"
)

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

// Subset is something of a poor man's subsetting function. It works - for TrueType fonts with 'glyf' tables only - by zeroing out
// the outlines of all glyphs not corresponding to or directly referenced by f's glyphs for the runes in cutset,
// truncating f's glyf and loca tables, and then writing only the required tables to the returned byte slice. The final subset font
// contains cmap, glyf, head, hhea, hmtx, loca, and maxp tables. The glyph indices are not affected. src should be a copy of the source
// bytes for f, since the underlying bytes used by f should not be accessed while f is in use.
func Subset(f *sfnt.Font, src []byte, cutset map[rune]struct{}) ([]byte, error) {
	sbuf := new(sfnt.Buffer)
	glyphs := make([]uint32, 0, 256)
	glyphs = append(glyphs, 0) // must include .notdef
	glyphset := make(map[uint32]struct{}, len(glyphs))
	glyphset[0] = struct{}{}
	for key := range cutset {
		gid, _ := f.GlyphIndex(sbuf, key)
		if gid == 0 {
			continue
		}
		glyphs = append(glyphs, uint32(gid))
		glyphset[uint32(gid)] = struct{}{}
	}
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })

	srcR := bytes.NewReader(src)
	ld, err := loader.NewLoader(srcR)
	if err != nil {
		return *new([]byte), err
	}

	head, err := ld.RawTable(1751474532)
	if err != nil {
		return *new([]byte), err
	}
	headP, _, err := tables.ParseHead(head)
	if err != nil {
		return *new([]byte), err
	}
	isLong := headP.IndexToLocFormat == 1

	loca, err := ld.RawTable(1819239265)
	if err != nil {
		return *new([]byte), err
	}
	locaP, err := tables.ParseLoca(loca, f.NumGlyphs(), isLong)
	if err != nil {
		return *new([]byte), err
	}

	glyf, err := ld.RawTable(1735162214)
	if err != nil {
		return *new([]byte), err
	}
	// include glyphs that are components of composite glyphs
	var composits = []uint32{}
	for i := range glyphset { //i := 0; i < len(locaP); i++ {
		var offset, next uint32
		if i < uint32(len(locaP)-1) {
			offset = locaP[i]
			next = locaP[i+1]
		} else if i == uint32(len(locaP)-1) {
			offset = locaP[i]
			next = offset
		} else {
			continue // should not happen
		}
		// per the spec, loca[n] must always be less than or equal to loca[n+1]
		if next < offset {
			continue
		}
		// this should not happen
		if next > uint32(len(glyf)) {
			break
		}
		g, _, err := tables.ParseGlyph(glyf[offset:next])
		if err != nil {
			continue
		}
		switch v := g.Data.(type) {
		case tables.CompositeGlyph:
			for _, gp := range v.Glyphs {
				if _, seen := glyphset[uint32(gp.GlyphIndex)]; !seen {
					composits = append(composits, uint32(gp.GlyphIndex))
				}
			}
		default:
			continue
		}
	}
	// do these here instead of modifying the object that's being ranged over in the previous loop
	for _, comp := range composits {
		glyphset[comp] = struct{}{}
	}
	glyphs = append(glyphs, composits...)
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })

	var finalOffset uint32
	var final uint32
	for i := 0; i < len(locaP); i++ {
		// assuming this works the same for short and long formats once they've been parsed
		var offset, next uint32
		if i < len(locaP)-1 {
			offset = locaP[i]
			next = locaP[i+1]
		} else {
			offset = locaP[i]
			next = offset
		}

		if next < offset {
			continue
		}
		if next > uint32(len(glyf)) {
			break
		}

		if _, used := glyphset[uint32(i)]; used {
			continue
		}

		// zero out old glyph outlines
		for j := offset; j < next; j++ {
			glyf[j] = 0
		}
	}
	// the loca table needs to be no more than final GID long
	finalOffset = locaP[glyphs[len(glyphs)-1]+1]
	final = glyphs[len(glyphs)-1]

	// update the number of glyphs in the maxp table
	// https://learn.microsoft.com/en-us/typography/opentype/spec/maxp
	maxp, err := ld.RawTable(1835104368)
	if err != nil {
		return *new([]byte), err
	}
	if len(maxp) >= 6 {
		// we can proceed
		binary.BigEndian.PutUint16(maxp[4:], uint16(final+1))
	}

	// truncate the loca table
	// https://learn.microsoft.com/en-us/typography/opentype/spec/loca
	if isLong {
		// each offset is 4 bytes
		loca = loca[:4*(final+2)+1]
	} else {
		// each offset is 2 bytes
		loca = loca[:2*(final+2)+1]
	}

	// truncate the glyf table
	glyf = glyf[:finalOffset]

	tbl := make([]loader.Table, len(pdfTables))
	for i, tag := range pdfTables {
		switch tag {
		case 1735162214:
			tbl[i] = loader.Table{Content: glyf, Tag: loader.Tag(tag)}
		case 1751474532:
			tbl[i] = loader.Table{Content: head, Tag: loader.Tag(tag)}
		case 1819239265:
			tbl[i] = loader.Table{Content: loca, Tag: loader.Tag(tag)}
		case 1835104368:
			tbl[i] = loader.Table{Content: maxp, Tag: loader.Tag(tag)}
		default:
			cnt, err := ld.RawTable(loader.Tag(tag))
			if err != nil {
				return *new([]byte), err
			}
			tbl[i] = loader.Table{Content: cnt, Tag: loader.Tag(tag)}
		}
	}
	return loader.WriteTTF(tbl), nil
}
