package gdf

import (
	"io"
	"time"
)

type annotFlag uint32

const (
	// Flags specifying the behavior of TextAnnots and Widgets.
	InvisibleAnnot annotFlag = 1 << iota
	HiddenAnnot
	PrintAnnot // IMPORTANT: must be set for an annotation to appear when printed.
	NoZoomAnnot
	NoRotateAnnot
	NoViewAnnot
	ReadOnlyAnnot
	LockedAnnot
	ToggleNoViewAnnot
	LockedContentsAnnot
)

type textAnnotStyle uint

const (
	// PDF-viewer defined text annotation icon styles.
	CommentIcon textAnnotStyle = iota
	KeyIcon
	NoteIcon
	HelpIcon
	NewParagraphIcon
	ParagraphIcon
	InsertIcon
	badIcon
)

var ta_styles = [...]string{"/Comment", "/Key", "/Note", "/Help", "/NewParagraph", "/Paragraph", "/Insert"}
var _ = int8(int(badIcon)-len(ta_styles)) << 8

func (t textAnnotStyle) String() string {
	if t < badIcon {
		return ta_styles[t]
	}
	return "/Comment"
}

/*
A TextAnnot is text that appears in a pop-up when the user hovers over the annotated area of a page. The Appearance *XObject, if provided, describes
the appearance of the TextAnnot's note icon, which is normally always visible on the viewed page. The Color field specifies the color of the pop-up annotation.
*/
type TextAnnot struct {
	Contents     string // Text to be displayed by the annotation.
	User         string // Name of the user creating the comment.
	ModDate      time.Time
	CreationDate time.Time
	Subject      string
	Open         bool
	Name         string // Unique annotation name.
	Flags        annotFlag
	IconStyle    textAnnotStyle
	Appearance   *XContent // Optional; if nil the IconStyle is used; overrides the IconStyle if non-nil.
	Color

	rect   Rect //The annotation rectangle, defining the location of the annotation on the page in default user space units.
	refnum int
}

func (t *TextAnnot) mark(i int) { t.refnum = i }
func (t *TextAnnot) id() int    { return t.refnum }
func (t *TextAnnot) children() []obj {
	if t.Appearance == nil {
		return nil
	}
	return []obj{t.Appearance}
}
func (t *TextAnnot) encode(w io.Writer) (int, error) {
	fields := append(make([]field, 0, 10), []field{
		{"/Type", "/Annot"},
		{"/Subtype", "/Text"},
		{"/Rect", t.rect},
		{"/Contents", utf16BEstring(t.Contents)},
		{"/F", uint32(t.Flags)},
	}...)

	if t.Appearance != nil {
		fields = append(fields, field{"/AP", subdict(64, []field{{"/N", iref(t.Appearance)}})})
	} else {
		fields = append(fields, field{"/Name", t.IconStyle.String()})
	}
	if t.User != "" {
		fields = append(fields, field{"/T", utf16BEstring(t.User)})
	}
	if !t.CreationDate.IsZero() {
		fields = append(fields, field{"/CreationDate", date(t.CreationDate)})
	}
	if t.Subject != "" {
		fields = append(fields, field{"/Subj", utf16BEstring(t.User)})
	}
	if t.Open {
		fields = append(fields, field{"/Open", t.Open})
	}
	if t.Name != "" {
		fields = append(fields, field{"/NM", utf16BEstring(t.User)})
	}
	if !t.ModDate.IsZero() {
		fields = append(fields, field{"/M", date(t.ModDate)})
	}
	if t.Color != nil {
		fields = append(fields, field{"/C", t.color()})
	}
	return w.Write(dict(512, fields))
}
