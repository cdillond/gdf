package gdf

/*
import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type encoding uint

const (
	WIN_ANSI_ENCODING Encoding = iota
	SYMBOLIC_ENCODING
	MAC_ROMAN_ENCODING
)

var encs = []string{"WinAnsiEncoding", "Symbolic", "MacRomanEncoding"}

func (e Encoding) String() string {
	return encs[e]
}

func toEncoder(e Encoding) *encoding.Encoder {
	switch e {
	case SYMBOLIC_ENCODING:
		//u := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
		return nil //u.NewEncoder()
	case WIN_ANSI_ENCODING:
		return charmap.Windows1252.NewEncoder()
	case MAC_ROMAN_ENCODING:
		return charmap.Macintosh.NewEncoder()
	default:
		return charmap.Windows1252.NewEncoder()
	}
}
*/
