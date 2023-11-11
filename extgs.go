package gdf

import (
	"fmt"
	"io"
)

type ExtGState struct {
	Dict   map[GStateKey]any
	refnum int
}

func NewExtGState() ExtGState {
	return ExtGState{
		Dict: make(map[GStateKey]any),
	}
}

func (e *ExtGState) SetRef(i int) { e.refnum = i }
func (e *ExtGState) RefNum() int  { return e.refnum }
func (e *ExtGState) Children() []Obj {
	out := []Obj{}
	for _, val := range e.Dict {
		if v, ok := val.(Obj); ok {
			out = append(out, v)
		}
	}
	return out
}
func (e *ExtGState) Encode(w io.Writer) (int, error) {
	var n int
	t, err := w.Write([]byte("<<\n/Type /ExtGState\n"))
	if err != nil {
		return t, err
	}
	n += t
	for key, val := range e.Dict {
		switch v := val.(type) {
		case string, Name:
			t, err = fmt.Fprintf(w, "/%s /%v\n", key, v)
			if err != nil {
				return n + t, err
			}
		case bool, LineCap, LineJoin, []float64:
			t, err = fmt.Fprintf(w, "/%s %v\n", key, v)
			if err != nil {
				return n + t, err
			}
		case DashPattern:
			t, err = fmt.Fprintf(w, "/%s %v %d\n", key, v.Array, v.Phase)
			if err != nil {
				return n + t, err
			}
		case Obj:
			t, err = fmt.Fprintf(w, "/%s %d 0 R\n", key, v)
			if err != nil {
				return n + t, err
			}
		}
		n += t
	}
	t, err = w.Write([]byte(">>\n"))
	if err != nil {
		return n + t, err
	}
	return n, nil
}

type GStateKey string

const (
	// GSType         GStateKey = "Type"           // ExtGState ADDED BY DEFAULT
	LW             GStateKey = "LW"             // float64
	LC             GStateKey = "LC"             // LineCap
	LJ             GStateKey = "LJ"             // LineJoin
	ML             GStateKey = "ML"             // float64
	D              GStateKey = "D"              // DashPattern
	RI             GStateKey = "RI"             // name
	OP             GStateKey = "OP"             // bool
	Op             GStateKey = "op"             // bool
	OPM            GStateKey = "OPM"            // int
	GSFont         GStateKey = "Font"           // Tf params
	BG2            GStateKey = "BG2"            // name
	UCR2           GStateKey = "UCR2"           // name
	HT             GStateKey = "HT"             // name, dictionary, stream
	FL             GStateKey = "FL"             // float64
	SM             GStateKey = "SM"             // float64
	SA             GStateKey = "SA"             // bool
	BM             GStateKey = "BM"             // name
	SMask          GStateKey = "SMask"          // dictionary, name
	CA             GStateKey = "CA"             // float64
	Ca             GStateKey = "ca"             // float64
	AIS            GStateKey = "AIS"            // bool
	TK             GStateKey = "TK"             // bool
	UseBlackPtComp GStateKey = "UseBlackPtComp" // name
	HTO            GStateKey = "HTO"            // [2]float64

	// BG UNSUPPORTED
	// UCR UNSUPPORTED
)
