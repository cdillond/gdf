package gdf

import (
	"fmt"
	"io"
)

// The acroform type represents a document-level parent form. There can be only one acroform per document, but it can have multiple fields.
// Inclusion of an acroform in a PDF complicates the document structure, since Widgets must be children of both an AcroField and a Page.
type acroform struct {
	acrofields []obj
	resources  resourceDict
	refnum     int
}

func (a *acroform) mark(i int) { a.refnum = i }
func (a *acroform) id() int    { return a.refnum }
func (a *acroform) children() []obj {
	objs := make([]obj, len(a.acrofields)+len(a.resources.Fonts)+len(a.resources.Images)+len(a.resources.XForms))
	n := copy(objs, a.acrofields)
	for i := range a.resources.Fonts {
		objs[n] = a.resources.Fonts[i]
		n++
	}
	for i := range a.resources.Images {
		objs[n] = a.resources.Images[i]
		n++
	}
	for i := range a.resources.XForms {
		objs[n] = a.resources.XForms[i]
		n++
	}
	return objs
}
func (a *acroform) encode(w io.Writer) (int, error) {
	fields := []field{{"/Fields", a.acrofields}}
	if a.resources.Fonts != nil {
		fields = append(fields, field{"/DR", a.resources.bytes()})
	}
	return w.Write(dict(512, fields))
}

func (p *PDF) newAcroform() *acroform {
	p.catalog.acroform = new(acroform)
	return p.catalog.acroform
}

// An acroType represents one of the various kinds of interactive AcroForm fields.
type acroType uint

const (
	AcroButton acroType = iota // Form type that includes checkboxes.
	AcroText
	acroChoice    // TODO
	acroSignature // TODO
	badAcroType
)

var acroTypes = [...]string{"/Btn", "/Tx", "/Ch", "/Sig"}

// compile time check to make sure badAcroType == len(acroTypes)
var _ = int8(int(badAcroType)-len(acroTypes)) << 8

func (a acroType) isValid() bool { return a < badAcroType }
func (a acroType) String() string {
	if a.isValid() {
		return acroTypes[a]
	}
	return ""
}

// All AcroFields should be created using the PDF.NewAcroField() method. Once an AcroField is created, it must be paired
// with a Widget and added to a Page that must, in turn, be appended to p.
func (p *PDF) NewAcroField() *AcroField {
	if p.catalog.acroform == nil {
		p.newAcroform()
	}
	aform := p.catalog.acroform
	out := &AcroField{
		parent: aform,
	}
	aform.acrofields = append(aform.acrofields, out)
	return out
}

// An AcroField is an interactive element within a PDF. All AcroField elements are children of a document-level
// acroform node. Each AcroField must be instantiated on a Page by a Widget, and can be associated with only 1 Widget.
type AcroField struct {
	Name  string     // the partial field name
	Flags fieldFlags // field flags

	fieldType acroType // field type
	child     *Widget  // each acrofield must be instatiated by a Widget
	da        []byte   // "default appearance" directive for variable text-type fields
	refnum    int
	parent    *acroform
}

func (a *AcroField) mark(i int) { a.refnum = i }
func (a *AcroField) id() int    { return a.refnum }
func (a *AcroField) children() []obj {
	return nil
}
func (a *AcroField) encode(w io.Writer) (int, error) {
	out := []field{
		{"/FT", a.fieldType.String()},
		{"/Parent", iref(a.parent)},
		{"/Kids", []obj{a.child}},
		{"/T", acrofieldname(a.Name + "_" + itoa(a.id()))}, // should be unique and not empty
		{"/Ff", uint32(a.Flags)},
	}
	if a.fieldType == acroChoice {
		out = append(out, field{"/V", pdfstring(a.child.Opts[0])}) // really pushing it lol
	}
	if a.fieldType == AcroText {

		out = append(out, field{"/DA", htxt(a.da)})
		if cfg, ok := a.child.cfg.(AcroTextCfg); ok && cfg.MaxLen > 0 {
			out = append(out, field{"/MaxLen", cfg.MaxLen})
		}
	}
	return w.Write(dict(256, out))
}

type fieldFlags uint32

const (
	DefaultFieldFlags fieldFlags = 0

	// fieldFlags common to all AcroField types (Table 227)
	FfReadOnly fieldFlags = 1
	FfRequired fieldFlags = 1 << 1
	FfNoExport fieldFlags = 1 << 2

	// fieldFlags specific to button-type AcroFields (Table 229)
	FfNoToggleToOff  fieldFlags = 1 << 14
	FfRadio          fieldFlags = 1 << 15
	FfPushbutton     fieldFlags = 1 << 16
	FfRadiosInUnison fieldFlags = 1 << 25

	// fieldFlags specific to text-type AcroFields (Table 231)
	FfMultiline       fieldFlags = 1 << 12
	FfPassword        fieldFlags = 1 << 13
	FfFileSelect      fieldFlags = 1 << 20
	FfDoNotSpellCheck fieldFlags = 1 << 22 // can also be used with choice-type AcroFields
	FfDoNotScroll     fieldFlags = 1 << 23
	FfComb            fieldFlags = 1 << 24
	FfRichText        fieldFlags = 1 << 25

	// fieldFlags specific to choice-type AcroFields (Table 233)
	FfCombo             fieldFlags = 1 << 17
	FfEdit              fieldFlags = 1 << 18
	FfSort              fieldFlags = 1 << 19
	FfMultiSelect       fieldFlags = 1 << 21
	FfCommitOnSelChange fieldFlags = 1 << 26
)

var (
	ErrChildren = fmt.Errorf("acrofields supported by gdf may have at most 1 child")
)

// AddAcroField adds an AcroField, whose visible component is w, to p. dst specifies the area of p's user space onto which w is drawn.
// If strokeBorder is true, dst is added to the page's path, and the border is stroked with the current stroke width and stroke color.
// Once f has been added to p, p must be appended to the PDF from which f was derived prior to the invocation of PDF.WriteTo().
func (p *Page) AddAcroField(w *Widget, f *AcroField, dst Rect, strokeBorder bool) error {
	if f.child != nil {
		return ErrChildren
	}
	f.child = w
	f.fieldType = w.AcroType

	// Calculate default appearance and add the font to the field's resources
	if w.AcroType == AcroText {
		if cfg, ok := w.cfg.(AcroTextCfg); ok {
			// Quick way to ensure all /WinAnsiEncoding glyphs are included
			for i := range w1252 {
				cfg.Font.GlyphAdvance(w1252[i])
			}
			fonts := f.parent.resources.Fonts
			var i int
			for i < len(fonts) && fonts[i] != cfg.Font {
				i++
			}
			if i == len(fonts) {
				f.parent.resources.Fonts = append(f.parent.resources.Fonts, cfg.Font)
			}
			var buf []byte
			buf = append(buf, []byte("/F"+itoa(i)+"\x20")...)
			buf = cmdf(buf, op_Tf, cfg.FontSize)
			f.da = buf
			if cfg.IsMultiLine {
				f.Flags |= FfMultiline
			}
		}
	}

	w.rect = dst
	w.page = p
	w.acrofield = f
	p.C.resources.Widgets = append(p.C.resources.Widgets, w)
	if strokeBorder {
		p.C.Re2(dst)
		p.C.Stroke()
	}
	return nil
}
