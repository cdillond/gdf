package gdf

import (
	"io"
)

type catalog struct {
	Pages    *pages
	streams  []obj
	prefs    ViewPrefs
	Acroform *acroform

	images []*Image
	xforms []*XContent
	//xobjs []*XObject
	pageLayout
	pageMode
	lang   string
	refnum int
}

func (c *catalog) id() int { return c.refnum }
func (c *catalog) children() []obj {
	var i int
	out := make([]obj, 1+oneif(len(c.Acroform.acrofields) > 0)+len(c.images)+len(c.xforms)+len(c.streams))
	out[i] = c.Pages
	i++
	if len(c.Acroform.acrofields) > 0 {
		out[i] = c.Acroform
		i++
	}
	for j := range c.images {
		out[i] = c.images[j]
		i++
	}
	for j := range c.xforms {
		out[i] = c.xforms[j]
		i++
	}
	for j := range c.streams {
		out[i] = c.streams[j]
		i++
	}
	return out
}
func (c *catalog) mark(i int) { c.refnum = i }
func (c *catalog) encode(w io.Writer) (int, error) {
	fields := []field{
		{"/Type", "/Catalog"},
		{"/Pages", iref(c.Pages)},
	}
	if len(c.streams) > 0 {
		fields = append(fields, field{
			"/Metadata", iref(c.streams[0]),
		})
	}
	if len(c.Acroform.acrofields) > 0 {
		fields = append(fields, field{
			"/AcroForm", iref(c.Acroform),
		})
	}
	if b := c.prefs.bytes(); b != nil {
		fields = append(fields, field{"/ViewerPreferences", b})
	}
	if s := c.pageLayout.String(); s != "" {
		fields = append(fields, field{"/PageLayout", s})
	}
	if s := c.pageMode.String(); s != "" {
		fields = append(fields, field{"/PageMode", s})
	}
	if c.lang != "" {
		fields = append(fields, field{"/Lang", htxt([]byte(c.lang))})
	}
	return w.Write(dict(64, fields))
}

// SetLanguage sets the default natural language of all text in the PDF to s, which must be a string representation of a valid BCP 47 language tag. (See golang.org/x/text/language).
// NOTE: gdf currently only supports Windows-1252 ("WinAnsiEncoding") for most textual elements in a PDF document. Text that appears in annotations may represent a wider range
// of characters, depending on the reader used to view the PDF.
func (p *PDF) SetLanguage(s string) {
	p.catalog.lang = s
}

func (p *PDF) SetPageLayout(pl pageLayout) {
	p.catalog.pageLayout = pl
}

func (p *PDF) SetPageMode(pm pageMode) {
	p.catalog.pageMode = pm
}

type pageLayout uint

const (
	SinglePage pageLayout = 1 + iota
	OneColumn
	TwoColumnLeft
	TwoColumnRight
	TwoPageLeft
	TwoPageRight
	badPageLayout
)

func (p pageLayout) String() string {
	if p < badPageLayout {
		return pageLayouts[p]
	}
	return ""
}

var pageLayouts = [...]string{"", "/SinglePage", "/OneColumn", "/TwoColumnLeft", "TwoColumnRight", "/TwoPageLeft", "/TwoPageRight"}

var _ = int8(int(badPageLayout)-len(pageLayouts)) << 8

type pageMode uint

const (
	DefaultMode pageMode = 1 + iota
	OutlinesMode
	ThumbsMode
	FullScreenMode
	OCMode
	AttachmentsMode
	badMode
)

var pageModes = [...]string{"", "/UseNone", "/UseOutlines", "/UseThumbs", "/FullScreen", "/UseOC", "/UseAttachments"}

var _ = int8(int(badMode)-len(pageLayouts)) << 8

func (p pageMode) String() string {
	if p < badMode {
		return pageModes[p]
	}
	return ""
}
