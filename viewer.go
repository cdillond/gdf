package gdf

type ViewPrefs struct {
	HideToolbar           bool
	HideMenu              bool
	HideWindowUI          bool
	FitWindow             bool
	CenterWindow          bool
	DisplayDocTitle       bool
	NonFullScreenPageMode bool
	Direction             preference
	PrintScaling          preference
	Duplex                preference
	PrintPageRange        []PageRange
	NumCopies             uint
}

// SetViewPrefs indicates the desired display settings for the PDF viewer to use when displaying p.
func (p *PDF) SetViewPrefs(v ViewPrefs) {
	p.catalog.prefs = v
}

func (v ViewPrefs) bytes() []byte {
	fields := make([]field, 0, 11)
	if v.HideToolbar {
		fields = append(fields, field{"/HideToolBar", v.HideToolbar})
	}
	if v.HideMenu {
		fields = append(fields, field{"/HideMenu", v.HideMenu})
	}
	if v.HideWindowUI {
		fields = append(fields, field{"/HideWindowUI", v.HideWindowUI})
	}
	if v.FitWindow {
		fields = append(fields, field{"/FitWindow", v.FitWindow})
	}

	if v.DisplayDocTitle {
		fields = append(fields, field{"/DisplayDocTitle", v.DisplayDocTitle})
	}
	if v.NonFullScreenPageMode {
		fields = append(fields, field{"/NonFullScreenPageMode", v.NonFullScreenPageMode})
	}

	if v.Direction != 0 {
		if t := v.Direction.String(); t != "" {
			fields = append(fields, field{"/Direction", t})
		}
	}
	if v.PrintScaling != 0 {
		if t := v.PrintScaling.String(); t != "" {
			fields = append(fields, field{"/PrintScaling", t})
		}

	}
	if v.Duplex != 0 {
		if t := v.Duplex.String(); t != "" {
			fields = append(fields, field{"/Duplex", t})
		}
	}
	if v.PrintPageRange != nil {
		if t := ptob(v.PrintPageRange); t != nil {
			fields = append(fields, field{"/PageRange", t})
		}
	}
	if v.NumCopies > 0 {
		fields = append(fields, field{"/NumCopies", v.NumCopies})
	}

	if len(fields) == 0 {
		return nil
	}
	d := dict(len(fields)*32, fields)
	return d[:len(d)-1]
}

type PageRange struct {
	Start, End uint
}

func (p PageRange) isValid() bool { return p.Start > 0 && p.Start <= p.End }

func ptob(p []PageRange) []byte {
	if len(p) == 0 {
		return nil
	}
	out := make([]byte, len(p)*8)
	out = append(out, '[')
	for i := range p {
		if !p[i].isValid() {
			continue
		}
		out = itobuf(p[i].Start, out)
		out = append(out, '\x20')
		out = itobuf(p[i].End, out)
		out = append(out, '\x20')
	}
	out[len(out)-1] = ']'
	return out
}

type preference uint

func (p preference) String() string {
	if p < invalidPref {
		return prefs[p]
	}
	return ""
}

const (
	noPref preference = iota
	UseNone
	UseOutlines
	UseThumbs
	UseOC
	L2R
	R2L
	AppDefaultScaling
	NoScaling
	Simplex
	DuplexFlipShortEdge
	DuplexFlipLongEdge
	invalidPref
)

var prefs = [...]string{"", "/UseNone", "/UseOutlines", "/UseThumbs", "/UseOC", "/L2R", "/R2L", "/AppDefault", "/None", "/Simplex", "/DuplexFlipShortEdge", "/DuplexFlipShortEdge"}
var _ = int8(len(prefs)-int(invalidPref)) << 8
