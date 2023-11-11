package gdf

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/encoding"
)

type Font struct {
	*sfnt.Font
	*SimpleFD
	refnum    int
	Type      Name
	Subtype   Name
	BaseFont  Name
	FirstChar int
	LastChar  int
	Widths    []int
	Encoding  Name
	charset   map[rune]int // maps a font's runes to their default glyph advances
	enc       *encoding.Encoder
	source    *ResourceStream
	buf       *sfnt.Buffer
	path      string // the path of the font source file
	subset    bool   // whether the font represents a subset
}

func (f *Font) SetFilter(filter Filter) {
	f.source.Filter = filter
}

func LoadTrueTypeBytes(b []byte, flag FontFlag, encoding Encoding) (Font, error) {
	fnt, err := sfnt.Parse(b)
	if err != nil {
		return *new(Font), err
	}

	out := Font{
		Type:     Name("Font"),
		Subtype:  Name("TrueType"),
		Encoding: Name(toNameString(encoding)),
		charset:  make(map[rune]int),
		enc:      toEncoder(encoding),
		source: &ResourceStream{
			buf:    new(bytes.Buffer),
			Filter: FILTER_FLATE,
		},
		buf:    new(sfnt.Buffer),
		subset: true,
	}
	if flag&NO_SUBSET != 0 {
		out.subset = false
		out.source.buf.Write(b)
		flag ^= NO_SUBSET
	}
	bf, err := fnt.Name(nil, sfnt.NameIDPostScript)
	if err != nil {
		return *new(Font), err
	}
	out.BaseFont = Name(bf)
	fd := NewSimpleFD(fnt, flag, out.buf)
	fd.FontFile2 = out.source
	fd.FontName = Name(bf)
	out.SimpleFD = fd
	out.Font = fnt
	return out, nil
}

func LoadTrueTypeFont(path string, flag FontFlag, encoding Encoding) (Font, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return *new(Font), err
	}
	fnt, err := sfnt.Parse(b)
	if err != nil {
		return *new(Font), err
	}

	out := Font{
		Type:     Name("Font"),
		Subtype:  Name("TrueType"),
		Encoding: Name(toNameString(encoding)),
		charset:  make(map[rune]int),
		enc:      toEncoder(encoding),
		source: &ResourceStream{
			buf:    new(bytes.Buffer),
			Filter: FILTER_FLATE,
		},
		buf:    new(sfnt.Buffer),
		path:   path,
		subset: true,
	}
	if flag&NO_SUBSET != 0 {
		out.subset = false
		out.source.buf.Write(b)
		flag ^= NO_SUBSET // so that it doesn't mess with pdf readers
	}
	bf, err := fnt.Name(nil, sfnt.NameIDPostScript)
	if err != nil {
		return *new(Font), err
	}
	out.BaseFont = Name(bf)
	fd := NewSimpleFD(fnt, flag, out.buf)
	fd.FontFile2 = out.source
	fd.FontName = Name(bf)
	out.SimpleFD = fd
	out.Font = fnt
	return out, nil
}

type SimpleFD struct {
	Type        Name
	FontName    Name // same as /BaseFont
	Flags       FontFlag
	FontBBox    []int
	ItalicAngle int
	Ascent      int
	Descent     int
	CapHeight   int
	StemV       int

	// Optional entries
	FontFamily   string
	FontStretch  string
	FontWeight   int
	Leading      int
	XHeight      int
	StemH        int
	AvgWidth     int
	MaxWidth     int
	MissingWidth int
	FontFile     *ResourceStream // Type1 fonts
	FontFile2    *ResourceStream // TrueType fonts
	FontFile3    *ResourceStream // Program defined by the Subtype entry in the stream dictionray
	CharSet      string

	refnum int
}

// Returns a font descriptor suitable for use with simple (i.e. non Type3, Type0, or MMType1) fonts.
func NewSimpleFD(fnt *sfnt.Font, flag FontFlag, buf *sfnt.Buffer) *SimpleFD {
	fd := new(SimpleFD)
	fd.Type = Name("FontDescriptor")
	fd.Flags = flag
	res, _ := FontBBox(fnt, buf)
	fd.FontBBox = []int{int(res.Min.X), int(res.Min.Y), int(res.Max.X), int(res.Max.Y)}

	pt := fnt.PostTable()
	fd.ItalicAngle = int(pt.ItalicAngle)

	met, _ := fnt.Metrics(buf, 1000, 0) // ppem is alway 1000
	fd.Ascent = int(met.Ascent)
	fd.Descent = int(met.Descent)
	fd.CapHeight = int(met.CapHeight)
	fd.StemV = 0
	fd.XHeight = int(met.XHeight)
	return fd
}

type FontFlag uint32

const (
	FIXED_PITCH FontFlag = 1 << 0
	SERIF       FontFlag = 1 << 1
	SYMBOLIC    FontFlag = 1 << 2 // must be set when NONSYMBOLIC is not set and vice versa
	SCRIPT      FontFlag = 1 << 3
	NONSYMBOLIC FontFlag = 1 << 5
	ITALIC      FontFlag = 1 << 6
	ALL_CAP     FontFlag = 1 << 16
	SMALL_CAP   FontFlag = 1 << 17
	FORCE_BOLD  FontFlag = 1 << 18

	NO_SUBSET FontFlag = 1 << 19 // prevent gpdf from subsetting a font
)

func (f *Font) SetRef(i int) { f.refnum = i }
func (f *Font) RefNum() int  { return f.refnum }
func (f *Font) Children() []Obj {
	return []Obj{f.SimpleFD, f.source}
}
func (f *Font) Encode(w io.Writer) (int, error) {
	var encstr string
	if f.Encoding != Name("Symbolic") {
		encstr = fmt.Sprintf("/Encoding %s\n", ToString(f.Encoding))
	}
	return fmt.Fprintf(w, "<<\n/Type %s\n/Subtype %s\n/BaseFont %s\n/FirstChar %d\n/LastChar %d\n/Widths %v\n%s/FontDescriptor %d 0 R\n>>\n",
		ToString(f.Type), ToString(f.Subtype), ToString(f.BaseFont), f.FirstChar, f.LastChar, f.Widths, encstr, f.SimpleFD.RefNum())
}

func (fd *SimpleFD) SetRef(i int)    { fd.refnum = i }
func (fd *SimpleFD) RefNum() int     { return fd.refnum }
func (fd *SimpleFD) Children() []Obj { return []Obj{} } // no need to include FontFile2
func (fd *SimpleFD) Encode(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "<<\n/Type %s\n/FontName %s\n/Flags %d\n/FontBBox %v\n/ItalicAngle %d\n/Ascent %d\n/Descent %d\n/CapHeight %d\n/StemV %d\n/XHeight %d\n/FontFile2 %d 0 R\n>>\n",
		ToString(fd.Type), ToString(fd.FontName), fd.Flags, fd.FontBBox, fd.ItalicAngle, fd.Ascent, fd.Descent, fd.CapHeight, fd.StemV, fd.XHeight, fd.FontFile2.RefNum())
}
