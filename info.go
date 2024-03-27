package gdf

import (
	"io"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

func (p *PDF) SetInfo(id InfoDict) {
	p.info = &id
}

var utf16 *encoding.Encoder

// An InfoDict represents a PDF's document information dictionary, which describes document level metadata. The PDF 2.0 Specification deprecates the use of all InfoDict fields except
// CreationDate and ModDate in favor of XMP Metadata. Adding an InfoDict to a PDF therefore has the effect of setting the output PDF version to 1.7. Although this is feature deprecated
// for PDF 2.0, info dictionary metadata has broader support among PDF readers than XMP, and is the most common method for representing certain document properties.
type InfoDict struct {
	Title        string
	Author       string
	Subject      string
	Keywords     string
	Creator      string
	Producer     string
	CreationDate time.Time
	ModDate      time.Time
	//Trapped      string
	refnum int
}

func utf16BEstring(s string) string {
	// lazy load the encoder
	if utf16 == nil {
		utf16 = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
	}
	t, err := utf16.String(s)
	if err != nil {
		return "()"
	}
	return "(\xFE\xFF" + t + ")"
}

func (I *InfoDict) mark(i int)      { I.refnum = i }
func (I *InfoDict) id() int         { return I.refnum }
func (I *InfoDict) children() []obj { return nil }
func (I *InfoDict) encode(w io.Writer) (int, error) {
	fields := make([]field, 0, 8)

	if I.Title != "" {
		fields = append(fields, field{"/Title", utf16BEstring(I.Title)})
	}
	if I.Author != "" {
		fields = append(fields, field{"/Author", utf16BEstring(I.Author)})
	}
	if I.Subject != "" {
		fields = append(fields, field{"/Subject", utf16BEstring(I.Subject)})
	}
	if I.Keywords != "" {
		fields = append(fields, field{"/Keywords", utf16BEstring(I.Keywords)})
	}
	if I.Creator != "" {
		fields = append(fields, field{"/Creator", utf16BEstring(I.Creator)})
	}
	if I.Producer != "" {
		fields = append(fields, field{"/Producer", utf16BEstring(I.Producer)})
	}

	if !I.CreationDate.IsZero() {
		fields = append(fields, field{"/CreationDate", date(I.CreationDate)})
	}
	if !I.ModDate.IsZero() {
		fields = append(fields, field{"/ModDate", date(I.CreationDate)})
	}

	return w.Write(dict(256, fields))
}
