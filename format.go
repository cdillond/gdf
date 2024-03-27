package gdf

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// This file contains utility functions related to the formatting of PDF objects.

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

// Returns the PDF Date string representation of t.
func date(t time.Time) []byte {
	dst := make([]byte, 0, len("(D:YYYYMMDDHHmmSSOHH'mm)"))
	dst = append(dst, "(D:"...)
	dst = itobuf(t.Year(), dst)

	// to be used for 0-padded digits
	two := [2]byte{'0', '0'}

	tmp := itobuf(int64(t.Month()), two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)

	tmp = itobuf(int64(t.Day()), two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)

	tmp = itobuf(int64(t.Hour()), two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)

	tmp = itobuf(int64(t.Minute()), two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)

	tmp = itobuf(int64(t.Second()), two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)

	_, offset := t.Zone()

	if offset < 0 {
		dst = append(dst, '-')

	} else if offset == 0 {
		return append(dst, 'Z')

	} else {
		dst = append(dst, '+')
	}
	offset = abs(offset)
	hours := offset / (60 * 60)
	offset -= hours * 60 * 60
	minutes := offset / 60

	tmp = itobuf(hours, two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)
	dst = append(dst, '\'')
	tmp = itobuf(minutes, two[:])
	dst = append(dst, tmp[len(tmp)-2:]...)
	return append(dst, ')')

}

// Returns the hex encoding of b (including <> chars) suitable for use as a PDF byte string.
func htxt(b []byte) []byte {
	dst := make([]byte, 0, 2*len(b)+2)
	dst = append(dst, '<')
	dst = hex.AppendEncode(dst, b)
	return append(dst, '>')
}

func acrofieldname(s string) []byte {
	dst := make([]byte, 0, len(s)+4)
	dst = append(dst, '(')

	for _, r := range s {
		c := byte(r)
		if rune(c) != r || c < '!' || c > '~' || c == '\\' || c == '(' || c == ')' || c == '.' {
			c = '_'
		}
		dst = append(dst, c)
	}

	return append(dst, ')')
}

func pdfstring(s string) string {
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\\", "\\\\)")
	return "(" + s + ")"
}

// Returns the string representation of i as an indirect reference.
func iref(i int) string {
	return itoa(i) + "\x200\x20R"
}

// Returns the s as a PDF Name object literal.
func name(s string) string {
	return "/" + s
}

// Returns true iff c is an \n or \r byte.
func isEOL(c byte) bool { return c == '\n' || c == '\r' }

// Appends a byte slice representing a command that takes one or more float64 args to buf and returns the extended slice.
func cmdf(buf []byte, op string, args ...float64) []byte {
	for i := 0; i < len(args); i++ {
		buf = strconv.AppendFloat(buf, args[i], 'f', -1, 64)
		buf = append(buf, '\x20')
	}
	return append(buf, op...)
}

// Appends a byte slice representing a command that takes one or more int64 args to buf and returns the extended slice.
func cmdi(buf []byte, op string, args ...int64) []byte {
	for i := 0; i < len(args); i++ {
		buf = strconv.AppendInt(buf, args[i], 10)
		buf = append(buf, '\x20')
	}
	return append(buf, op...)
}

// Returns fields formatted as a byte slice. Size is a hint for how large the buffer should be.
func dict(size int, fields []field) []byte {
	if len(fields) == 0 {
		return nil
	}
	out := make([]byte, 0, size)
	out = append(out, "<<\n"...)
	for i := range fields {
		if fields[i].val == nil {
			continue
		}
		out = append(out, fields[i].key...)
		out = append(out, '\x20')
		switch v := fields[i].val.(type) {
		case string:
			out = append(out, v...)
		case float64:
			out = strconv.AppendFloat(out, v, 'f', -1, 64)
		case int:
			out = strconv.AppendInt(out, int64(v), 10)
		case uint:
			out = strconv.AppendInt(out, int64(v), 10)
		case uint32:
			out = strconv.AppendInt(out, int64(v), 10)
		case []byte:
			out = append(out, v...)
		case bool:
			if v {
				out = append(out, "true"...)
			} else {
				out = append(out, "false"...)
			}
		default:
			out = sbuf(out, v)
		}
		out = append(out, '\n')
	}
	return append(out, ">>\n"...)
}

// Wrapper for dict() that removes the trailing '\n'; to be used when the dict is embedded in another dict (i.e. is a 'subdictionary').
func subdict(size int, fields []field) []byte {
	b := dict(size, fields)
	if len(b) < 1 {
		return nil
	}
	return b[:len(b)-1]
}

// Appends the string representation of a Rect, []string, []float64, or []int to dst and returns the extended slice. Returns dst unaltered if s is not one of these types.
func sbuf(dst []byte, s any) []byte {
	switch v := s.(type) {
	case Rect:
		dst = append(dst, '[')
		dst = strconv.AppendFloat(dst, v.LLX, 'f', -1, 64)
		dst = append(dst, '\x20')
		dst = strconv.AppendFloat(dst, v.LLY, 'f', -1, 64)
		dst = append(dst, '\x20')
		dst = strconv.AppendFloat(dst, v.URX, 'f', -1, 64)
		dst = append(dst, '\x20')
		dst = strconv.AppendFloat(dst, v.URY, 'f', -1, 64)
		dst = append(dst, ']')
	case []string:
		if len(v) == 0 {
			return append(dst, "[]"...)
		}
		dst = append(dst, '[')
		for i := range v {
			dst = append(dst, v[i]...)
			dst = append(dst, '\x20')
		}
		dst[len(dst)-1] = ']'
	case []float64:
		if len(v) == 0 {
			return append(dst, "[]"...)
		}
		dst = append(dst, '[')
		for i := range v {
			dst = strconv.AppendFloat(dst, v[i], 'f', -1, 64)
			dst = append(dst, '\x20')
		}
		dst[len(dst)-1] = ']'
	case []int:
		if len(v) == 0 {
			return append(dst, "[]"...)
		}
		dst = append(dst, '[')
		for i := range v {
			dst = strconv.AppendInt(dst, int64(v[i]), 10)
			dst = append(dst, '\x20')
		}
		dst[len(dst)-1] = ']'
	case []obj:
		if len(v) == 0 {
			return append(dst, "[]"...)
		}
		dst = append(dst, '[')
		for i := range v {
			dst = append(dst, iref(v[i].id())...)
			dst = append(dst, '\x20')
		}
		dst[len(dst)-1] = ']'
	default:
	}
	return dst
}

// To be used for writing 0-padded offsets to the xref table.
func pad10(n int, b []byte) bool {
	if len(b) < 10 {
		return false
	}
	i := 9
	for i > -1 && n > 0 {
		b[i] = byte(n%10) + '0'
		n /= 10
		i--
	}
	for i > -1 {
		b[i] = '0'
		i--
	}
	return true
}

type integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | int | ~int8 | ~int16 | ~int32 | ~int64
}

// Wrapper for strconv.AppendInt.
func itobuf[T integer](i T, dst []byte) []byte {
	return strconv.AppendInt(dst, int64(i), 10)
}

func itoa[T integer](i T) string {
	return strconv.Itoa(int(i))
}
func itob[T integer](i T) []byte {
	return strconv.AppendInt(nil, int64(i), 10)
}

// Wrapper for strconv.FormatFloat.
func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func oneif(b bool) int {
	var x int
	if b {
		x = 1
	}
	return x
}
