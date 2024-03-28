package gdf

import (
	"io"
	"time"
)

// A Widget is an annotation that serves as the visibile representation of an interactive acroForm field.
type Widget struct {
	AcroType     acroType
	Flags        annotFlag // 12.5.3 annotation flags
	User         string
	ModDate      time.Time
	CreationDate time.Time
	Subject      string
	Open         bool
	Name         string // unique text string identifiying the widget
	Opts         []string
	//H            string highlighting mode
	rect      Rect // location of the widget on the parent page
	cfg       WidgetCfger
	acrofield *AcroField
	page      *Page
	refnum    int
}

func (a *Widget) mark(i int) { a.refnum = i }
func (a *Widget) id() int    { return a.refnum }
func (a *Widget) children() []obj {
	if cfg, ok := a.cfg.(CheckboxCfg); ok {
		return []obj{cfg.Off, cfg.On}
	}
	if cfg, ok := a.cfg.(AcroTextCfg); ok {
		return []obj{cfg.Appearance, cfg.Font}
	}
	return nil
}
func (a *Widget) encode(w io.Writer) (int, error) {
	items := append(make([]field, 0, 20), []field{
		{"/Type", "/Annot"},
		{"/Subtype", "/Widget"},
		{"/FT", a.AcroType.String()},
		{"/F", uint32(a.Flags)}, // Always print widget annotations.
		{"/Rect", a.rect},
		{"/AP", a.cfg.bytes()}, // Required for all widgets supported by gdf.
		{"/Parent", iref(a.acrofield)},
		{"/P", iref(a.page)},
	}...)

	// optional fields
	if a.User != "" {
		items = append(items, field{"/T", utf16BEstring(a.User)})
	}
	if !a.ModDate.IsZero() {
		items = append(items, field{"/M", date(a.ModDate)})
	}
	if !a.CreationDate.IsZero() {
		items = append(items, field{"/CreationDate", date(a.CreationDate)})
	}
	if a.Name != "" {
		items = append(items, field{"/NM", utf16BEstring(a.Name)})
	}
	if a.Subject != "" {
		items = append(items, field{"/Subj", utf16BEstring(a.Subject)})
	}
	if a.Open {
		items = append(items, field{"/Open", a.Open})
	}

	switch a.AcroType {
	case AcroButton:
		var defState string
		if cfg, ok := a.cfg.(CheckboxCfg); ok && cfg.IsOnDefault {
			defState = "/On"
		} else {
			defState = "/Off"
		}
		items = append(items, field{"/AS", defState})
	case AcroText:
	case acroChoice:
		k := make([]string, len(a.Opts))
		for i := range a.Opts {
			k[i] = pdfstring(a.Opts[i])
		}
		items = append(items, field{"/Opt", k})
	case acroSignature:

	}

	return w.Write(dict(256, items))
}

func NewWidget(cfg WidgetCfger) *Widget {
	w := new(Widget)
	cfg.configure(w)
	return w
}

// An interface used to configure Widgets and their associated AcroFields. Implemented by AcroTextCfg and CheckboxCfg.
type WidgetCfger interface {
	configure(*Widget)
	bytes() []byte
}

/*
Implements WidgetCfger. AcroTextCfgs can be used to instantiate text-type AcroFields. The PrintAnnot
flag should be set in almost all cases. The Appearance *XObject can be left nil.
*/
type AcroTextCfg struct {
	Flags       annotFlag
	Appearance  *XObject
	Font        *Font
	FontSize    float64
	IsMultiLine bool
	MaxLen      int
}

func (a AcroTextCfg) configure(w *Widget) {
	w.Flags = a.Flags
	w.AcroType = AcroText
	w.cfg = a
}

func (a AcroTextCfg) bytes() []byte {
	if a.Appearance == nil {
		return nil
	}
	return subdict(64, []field{{"/N", iref(a.Appearance)}})
}

// Implements WidgetCfger. The PrintAnnot flag should be set in almost all cases. The Off and On *XObjects are mandatory,
// but can be reused by multiple Widgets.
type CheckboxCfg struct {
	Flags       annotFlag
	Off, On     *XObject
	IsOnDefault bool
}

func (c CheckboxCfg) configure(w *Widget) {
	w.Flags = c.Flags
	w.AcroType = AcroButton
	w.cfg = c
}

func (c CheckboxCfg) bytes() []byte {
	if c.Off == nil || c.On == nil {
		return nil
	}
	b := subdict(64, []field{
		{"/Off", iref(c.Off)},
		{"/On", iref(c.On)},
	})
	return subdict(128, []field{{"/N", b}})
}
