package gdf

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type Encoding uint

const (
	SYMBOLIC_ENCODING Encoding = iota
	WIN_ANSI_ENCODING
	MAC_ROMAN_ENCODING
)

var encs = []string{"Symbolic", "WinAnsiEncoding", "MacRomanEncoding"}

func toNameString(e Encoding) string {
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
