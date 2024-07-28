package gdf

import (
	"io"
	"os"

	"github.com/cdillond/gdf/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// A FontSubsetter is intended to give the user control over how a Font is subset when it is embedded in the output PDF.
// By default, gdf uses the DefaultSubsetter, which is equivalent to the BasicSubsetter type from github.com/cdillond/gdf/font,
// but there are good reasons to choose a different implementation (the gdf/font package provides several).
// For example, Harfbuzz's hb-subset tool and Microsoft's Win32 CreateFontPackage are robust alternatives written in C++
// that can be wrapped by a user-defined FontSubsetter.
// If you do not want the embedded font to be subset at all, you can set the font's FontSubsetter to nil.
// A FontSubsetter should not alter the glyph ID of any rune in the cutset. It must also be sure to include a .notdef glyph.
type FontSubsetter interface {
	Init(SFNT *sfnt.Font, src []byte, path string)
	Subset(cutset map[rune]struct{}) ([]byte, error)
}

type DefaultSubsetter struct {
	font.BasicSubsetter
}

const ppem = 1000

// A Font represents a TrueType/OpenType font. Any given Font struct should be used on at most 1 PDF. To use the same underlying
// font on multiple PDF files, derive a new Font struct from the source font file or bytes for each PDF.
type Font struct {
	SFNT      *sfnt.Font // The source TrueType or OpenType font.
	Subsetter FontSubsetter

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
	srcb      []byte
	srcPath   string
}

// LoadTrueType returns a *Font object, which can be used for drawing text to a ContentStream or XObject, and an error.
func LoadTrueType(b []byte, flag FontFlag) (*Font, error) {
	b2 := b
	fnt, err := sfnt.Parse(b)
	if err != nil {
		return nil, err
	}
	out := &Font{
		Subsetter: new(DefaultSubsetter),
		subtype:   "/TrueType",
		encName:   "/WinAnsiEncoding",
		charset:   make(map[rune]int),
		enc:       charmap.Windows1252.NewEncoder(),
		source: &stream{
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
	out.SFNT = fnt
	return out, nil
}

// LoadTrueTypeFile returns a *Font object, which can be used for drawing text to a ContentStream or XObject, and an error.
func LoadTrueTypeFile(path string, flag FontFlag) (*Font, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := LoadTrueType(b, flag)
	if err != nil {
		f.srcPath = path
	}
	return f, err
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

// newSimpleFD returns a font descriptor suitable for use with simple (i.e. non Type3, Type0, or MMType1) fonts.
func newSimpleFD(fnt *sfnt.Font, flag FontFlag, buf *sfnt.Buffer) *simpleFD {
	fd := new(simpleFD)
	if flag&Nonsymbolic == 0 && flag&Symbolic == 0 {
		flag |= Nonsymbolic
	}
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
	Symbolic    FontFlag = 1 << 2
	Script      FontFlag = 1 << 3
	Nonsymbolic FontFlag = 1 << 5
	Italic      FontFlag = 1 << 6
	AllCap      FontFlag = 1 << 16
	SmallCap    FontFlag = 1 << 17
	ForceBold   FontFlag = 1 << 18
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
		{"/FontDescriptor", iref(f.simpleFD)},
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
	var ref string
	if fd.FontFile2 == nil {
		vers = 3
		ref = iref(fd.FontFile3)
	} else {
		vers = 2
		ref = iref(fd.FontFile2)
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
		{"/FontFile" + itoa(vers), ref},
	}))
}
