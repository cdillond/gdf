package gdf

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
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
	Charset   map[rune]int // maps a font's runes to their default glyph advances
	enc       *encoding.Encoder
	source    *resourceStream
	buf       *sfnt.Buffer
	noSubset  bool // whether the font represents a subset
	srcb      []byte
}

// Returns a *Font object, which can be used for drawing text to a ContentStream, and an error.
func LoadTrueType(b []byte, flag FontFlag) (*Font, error) {
	b2 := make([]byte, len(b))
	copy(b2, b)
	fnt, err := sfnt.Parse(b)
	if err != nil {
		return nil, err
	}
	out := &Font{
		Type:     Name("Font"),
		Subtype:  Name("TrueType"),
		Encoding: Name("WinAnsiEncoding"),
		Charset:  make(map[rune]int),
		enc:      charmap.Windows1252.NewEncoder(),
		source: &resourceStream{
			buf:    new(bytes.Buffer),
			Filter: FILTER_FLATE,
		},
		buf:  new(sfnt.Buffer),
		srcb: b2,
	}
	bf, err := fnt.Name(out.buf, sfnt.NameIDPostScript)
	if err != nil {
		return nil, err
	}
	out.BaseFont = Name(bf)
	fd := NewSimpleFD(fnt, flag, out.buf)
	fd.FontFile2 = out.source
	fd.FontName = Name(bf)
	out.SimpleFD = fd
	out.Font = fnt
	return out, nil
}

// Returns a *Font object, which can be used for drawing text to a ContentStream, and an error.
func LoadTrueTypeFile(path string, flag FontFlag) (*Font, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadTrueType(b, flag)
}

type SimpleFD struct {
	Type        Name
	FontName    Name
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
	FontFile     *resourceStream // Type1 fonts
	FontFile2    *resourceStream // TrueType fonts
	FontFile3    *resourceStream // Program defined by the Subtype entry in the stream dictionray

	refnum int
}

// Returns a font descriptor suitable for use with simple (i.e. non Type3, Type0, or MMType1) fonts.
func NewSimpleFD(fnt *sfnt.Font, flag FontFlag, buf *sfnt.Buffer) *SimpleFD {
	fd := new(SimpleFD)
	fd.Type = Name("FontDescriptor")
	fd.Flags = flag
	res, _ := fontBBox(fnt, buf)
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
	SYMBOLIC    FontFlag = 1 << 2 // Must be set when NONSYMBOLIC is not set and vice versa.
	SCRIPT      FontFlag = 1 << 3
	NONSYMBOLIC FontFlag = 1 << 5 // Must be set when SYMBOLIC is not set and vice versa.
	ITALIC      FontFlag = 1 << 6
	ALL_CAP     FontFlag = 1 << 16
	SMALL_CAP   FontFlag = 1 << 17
	FORCE_BOLD  FontFlag = 1 << 18

	NO_SUBSET FontFlag = 1 << 19 // Prevent gdf from subsetting a font. Not in the PDF spec.
)

func (f *Font) setRef(i int) { f.refnum = i }
func (f *Font) refNum() int  { return f.refnum }
func (f *Font) children() []obj {
	return []obj{f.SimpleFD, f.source}
}
func (f *Font) encode(w io.Writer) (int, error) {
	var encstr string
	if f.Encoding != Name("Symbolic") {
		encstr = fmt.Sprintf("/Encoding %s\n", f.Encoding)
	}
	return fmt.Fprintf(w, "<<\n/Type %s\n/Subtype %s\n/BaseFont %s\n/FirstChar %d\n/LastChar %d\n/Widths %v\n%s/FontDescriptor %d 0 R\n>>\n",
		f.Type, f.Subtype, f.BaseFont, f.FirstChar, f.LastChar, f.Widths, encstr, f.SimpleFD.refNum())
}

func (fd *SimpleFD) setRef(i int)    { fd.refnum = i }
func (fd *SimpleFD) refNum() int     { return fd.refnum }
func (fd *SimpleFD) children() []obj { return []obj{} } // no need to include FontFile2
func (fd *SimpleFD) encode(w io.Writer) (int, error) {
	var vers int
	var ref int
	if fd.FontFile2 == nil {
		vers = 3
		ref = fd.FontFile3.refNum()
	} else {
		vers = 2
		ref = fd.FontFile2.refNum()
	}
	return fmt.Fprintf(w, "<<\n/Type %s\n/FontName %s\n/Flags %d\n/FontBBox %v\n/ItalicAngle %d\n/Ascent %d\n/Descent %d\n/CapHeight %d\n/StemV %d\n/XHeight %d\n/FontFile%d %d 0 R\n>>\n",
		fd.Type, fd.FontName, fd.Flags, fd.FontBBox, fd.ItalicAngle, fd.Ascent, fd.Descent, fd.CapHeight, fd.StemV, fd.XHeight, vers, ref)
}
