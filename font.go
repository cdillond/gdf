package gdf

import (
	"io"
	"os"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

const ppem = 1000

type Font struct {
	*sfnt.Font
	*simpleFD

	refnum    int
	subtype   string
	baseFont  string
	firstChar int
	lastChar  int
	widths    []int
	encName   string
	charset   map[rune]int // maps a font's runes to their default glyph advances
	enc       *encoding.Encoder
	source    *stream
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
		subtype: "/TrueType",
		encName: "/WinAnsiEncoding",
		charset: make(map[rune]int),
		enc:     charmap.Windows1252.NewEncoder(),
		source: &stream{
			//buf:    new(bytes.Buffer),
			Filter: Flate,
		},
		buf:  new(sfnt.Buffer),
		srcb: b2,
	}
	bf, err := fnt.Name(out.buf, sfnt.NameIDPostScript)
	if err != nil {
		return nil, err
	}
	out.baseFont = name(bf)
	fd := newSimpleFD(fnt, flag, out.buf)
	fd.FontFile2 = out.source
	fd.FontName = name(bf)
	out.simpleFD = fd
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

type simpleFD struct {
	FontName    string
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
	FontFile     *stream // Type1 fonts
	FontFile2    *stream // TrueType fonts
	FontFile3    *stream // Program defined by the Subtype entry in the stream dictionray

	refnum int
}

// Returns a font descriptor suitable for use with simple (i.e. non Type3, Type0, or MMType1) fonts.
func newSimpleFD(fnt *sfnt.Font, flag FontFlag, buf *sfnt.Buffer) *simpleFD {
	fd := new(simpleFD)
	fd.Flags = flag
	res, _ := fontBBox(fnt, buf)
	fd.FontBBox = []int{int(res.Min.X), int(res.Min.Y), int(res.Max.X), int(res.Max.Y)}

	pt := fnt.PostTable()
	fd.ItalicAngle = int(pt.ItalicAngle)

	met, _ := fnt.Metrics(buf, ppem, 0)
	fd.Ascent = int(met.Ascent)
	fd.Descent = int(met.Descent)
	fd.CapHeight = int(met.CapHeight)
	fd.StemV = 0
	fd.XHeight = int(met.XHeight)
	return fd
}

type FontFlag uint32

const (
	FixedPitch  FontFlag = 1 << 0
	Serif       FontFlag = 1 << 1
	Symbolic    FontFlag = 1 << 2 // Must be set when Nonsymbolic is not set and vice versa.
	Script      FontFlag = 1 << 3
	Nonsymbolic FontFlag = 1 << 5 // Must be set when Symbolic is not set and vice versa.
	Italic      FontFlag = 1 << 6
	AllCap      FontFlag = 1 << 16
	SmallCap    FontFlag = 1 << 17
	ForceBold   FontFlag = 1 << 18

	NoSubset FontFlag = 1 << 19 // Prevent gdf from subsetting a font. Not in the PDF spec.
)

func (f *Font) mark(i int) { f.refnum = i }
func (f *Font) id() int    { return f.refnum }
func (f *Font) children() []obj {
	return []obj{f.simpleFD, f.source}
}
func (f *Font) encode(w io.Writer) (int, error) {
	return w.Write(dict(1024, []field{
		{"/Type", "/Font"},
		{"/Subtype", f.subtype},
		{"/BaseFont", f.baseFont},
		{"/FirstChar", f.firstChar},
		{"/LastChar", f.lastChar},
		{"/Widths", f.widths},
		{"/Encoding", f.encName},
		{"/FontDescriptor", iref(f.simpleFD.id())},
	}))
}

type field struct {
	key string
	val any
}

func (fd *simpleFD) mark(i int)      { fd.refnum = i }
func (fd *simpleFD) id() int         { return fd.refnum }
func (fd *simpleFD) children() []obj { return nil } // no need to include FontFile2
func (fd *simpleFD) encode(w io.Writer) (int, error) {
	var vers int
	var ref int
	if fd.FontFile2 == nil {
		vers = 3
		ref = fd.FontFile3.id()
	} else {
		vers = 2
		ref = fd.FontFile2.id()
	}

	return w.Write(dict(1024, []field{
		{"/Type", "/FontDescriptor"},
		{"/FontName", fd.FontName},
		{"/Flags", int(fd.Flags)},
		{"/FontBBox", fd.FontBBox},
		{"/ItalicAngle", fd.ItalicAngle},
		{"/Ascent", fd.Ascent},
		{"/Descent", fd.Descent},
		{"/CapHeight", fd.CapHeight},
		{"/StemV", fd.StemV},
		{"/XHeight", fd.XHeight},
		{"/FontFile" + itoa(vers), iref(ref)},
	}))
}
