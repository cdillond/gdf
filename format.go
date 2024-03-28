package gdf

/* This file contains utility functions related to the formatting of PDF objects. */

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

// Returns the PDF Date string representation of t.
func date(t time.Time) []byte {
	dst := make([]byte, len("(D:YYYYMMDDHHmmSSOHH'mm)"))
	n := copy(dst, []byte("(D:"))
	dst = itobuf(t.Year(), dst[n:])
	n += 4

	// to be used for 0-padded digits
	two := [2]byte{'0', '0'}

	buf := itobuf(int64(t.Month()), two[:])
	n += copy(dst[n:], buf[len(buf)-2:])

	buf = itobuf(int64(t.Day()), two[:])
	n += copy(dst[n:], buf[len(buf)-2:])

	buf = itobuf(int64(t.Hour()), two[:])
	n += copy(dst[n:], buf[len(buf)-2:])

	buf = itobuf(int64(t.Minute()), two[:])
	n += copy(dst[n:], buf[len(buf)-2:])

	buf = itobuf(int64(t.Second()), two[:])
	n += copy(dst[n:], buf[len(buf)-2:])

	_, offset := t.Zone()

	if offset < 0 {
		dst[n] = '-'

	} else if offset == 0 {
		dst[n] = 'Z'

	} else {
		dst[n] = '+'
	}
	n++

	offset = abs(offset)
	hours := offset / (60 * 60)
	offset -= hours * 60 * 60
	minutes := offset / 60

	buf = itobuf(hours, two[:])
	n += copy(dst[n:], buf[len(buf)-2:])
	dst[n] = '\''
	n++
	buf = itobuf(minutes, two[:])
	n += copy(dst[n:], buf[len(buf)-2:])
	dst[n] = ')'
	return dst

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
func iref(o obj) string {
	return itoa(o.id()) + "\x200\x20R"
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
			dst = append(dst, iref(v[i])...)
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

// Returns 1 if b == true and 0 otherwise.
func oneif(b bool) int {
	var x int
	if b {
		x = 1
	}
	return x
}

// returns the win1252 encoding of r.
func rtoc(r rune) byte {
	if r < 0 {
		return 0
	}
	if r < 128 {
		return byte(r)
	}
	if r < 256 && w1252[r] == r {
		return byte(r)
	}
	for i := 0x80; i < len(w1252); i++ {
		if r == w1252[i] {
			return byte(i)
		}
	}
	return 0
}

// maps indices to unicode runes.
var w1252 = [256]rune{
	0x0000, 0x0001, 0x0002, 0x0003, 0x0004, 0x0005, 0x0006, 0x0007,
	0x0008, 0x0009, 0x000A, 0x000B, 0x000C, 0x000D, 0x000E, 0x000F,
	0x0010, 0x0011, 0x0012, 0x0013, 0x0014, 0x0015, 0x0016, 0x0017,
	0x0018, 0x0019, 0x001A, 0x001B, 0x001C, 0x001D, 0x001E, 0x001F,
	0x0020, 0x0021, 0x0022, 0x0023, 0x0024, 0x0025, 0x0026, 0x0027,
	0x0028, 0x0029, 0x002A, 0x002B, 0x002C, 0x002D, 0x002E, 0x002F,
	0x0030, 0x0031, 0x0032, 0x0033, 0x0034, 0x0035, 0x0036, 0x0037,
	0x0038, 0x0039, 0x003A, 0x003B, 0x003C, 0x003D, 0x003E, 0x003F,
	0x0040, 0x0041, 0x0042, 0x0043, 0x0044, 0x0045, 0x0046, 0x0047,
	0x0048, 0x0049, 0x004A, 0x004B, 0x004C, 0x004D, 0x004E, 0x004F,
	0x0050, 0x0051, 0x0052, 0x0053, 0x0054, 0x0055, 0x0056, 0x0057,
	0x0058, 0x0059, 0x005A, 0x005B, 0x005C, 0x005D, 0x005E, 0x005F,
	0x0060, 0x0061, 0x0062, 0x0063, 0x0064, 0x0065, 0x0066, 0x0067,
	0x0068, 0x0069, 0x006A, 0x006B, 0x006C, 0x006D, 0x006E, 0x006F,
	0x0070, 0x0071, 0x0072, 0x0073, 0x0074, 0x0075, 0x0076, 0x0077,
	0x0078, 0x0079, 0x007A, 0x007B, 0x007C, 0x007D, 0x007E, 0x007F,
	0x20AC, 0x0000, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x0000, 0x017D, 0x0000,
	0x0000, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x0000, 0x017E, 0x0178,
	0x00A0, 0x00A1, 0x00A2, 0x00A3, 0x00A4, 0x00A5, 0x00A6, 0x00A7,
	0x00A8, 0x00A9, 0x00AA, 0x00AB, 0x00AC, 0x00AD, 0x00AE, 0x00AF,
	0x00B0, 0x00B1, 0x00B2, 0x00B3, 0x00B4, 0x00B5, 0x00B6, 0x00B7,
	0x00B8, 0x00B9, 0x00BA, 0x00BB, 0x00BC, 0x00BD, 0x00BE, 0x00BF,
	0x00C0, 0x00C1, 0x00C2, 0x00C3, 0x00C4, 0x00C5, 0x00C6, 0x00C7,
	0x00C8, 0x00C9, 0x00CA, 0x00CB, 0x00CC, 0x00CD, 0x00CE, 0x00CF,
	0x00D0, 0x00D1, 0x00D2, 0x00D3, 0x00D4, 0x00D5, 0x00D6, 0x00D7,
	0x00D8, 0x00D9, 0x00DA, 0x00DB, 0x00DC, 0x00DD, 0x00DE, 0x00DF,
	0x00E0, 0x00E1, 0x00E2, 0x00E3, 0x00E4, 0x00E5, 0x00E6, 0x00E7,
	0x00E8, 0x00E9, 0x00EA, 0x00EB, 0x00EC, 0x00ED, 0x00EE, 0x00EF,
	0x00F0, 0x00F1, 0x00F2, 0x00F3, 0x00F4, 0x00F5, 0x00F6, 0x00F7,
	0x00F8, 0x00F9, 0x00FA, 0x00FB, 0x00FC, 0x00FD, 0x00FE, 0x00FF,
}
