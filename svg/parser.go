package svg

import (
	"strconv"
	"strings"

	"github.com/cdillond/gdf"
)

type buf struct {
	b []byte
	n int
}

func (b *buf) next() (c byte, ok bool) {
	if ok = (b.n < len(b.b)) && b.n > -1; ok {
		c = b.b[b.n]
	}
	b.n++
	return c, ok
}

func (b *buf) skip() {
	b.n++
}
func (b *buf) skipWSPComma() {
	for c, ok := b.peek(); ok && (c == ',' || c < '!'); c, ok = b.peek() {
		b.skip()
	}
}
func (b *buf) consumeDigit() {
	for c, _ := b.next(); c >= '0' && c <= '9'; c, _ = b.next() {
	}
	b.undo()
}

func (b *buf) peek() (c byte, ok bool) {
	if ok = (b.n < len(b.b)) && b.n > -1; ok {
		c = b.b[b.n]
	}
	return c, ok
}

func (b *buf) undo() {
	b.n--
}

func (b *buf) upTo(c byte) []byte {
	i := b.n
	for i < len(b.b) && b.b[i] != c {
		i++
	}
	res := b.b[b.n:i]
	b.n = i
	return res
}

func (b *buf) upTo2(c1, c2 byte) []byte {
	i := b.n
	for i < len(b.b) && b.b[i] != c1 && b.b[i] != c2 {
		i++
	}
	res := b.b[b.n:i]
	b.n = i
	return res
}

// https://www.w3.org/TR/css-syntax-3/#typedef-number-token
func (b *buf) ConsumeNumber() string {
	// skip over leading whitespace
	b.skipWSPComma()
	start := b.n
	// optionally include a sign
	c, ok := b.peek()
	if !ok {
		return ""
	}
	if c == '-' || c == '+' {
		b.skip()
		c, ok = b.peek()
		if !ok {
			return ""
		}
	}

	// digit dot digit
	// digit
	// dot digit
	b.consumeDigit()
	c, ok = b.peek()
	if !ok {
		return string(b.b[start:b.n])
	}
	if c == '.' {
		b.skip()
		b.consumeDigit()
		c, _ = b.peek()
	}
	if c != 'e' && c != 'E' {
		return string(b.b[start:b.n])
	}
	b.skip()
	c, _ = b.peek()
	if c == '-' || c == '+' {
		b.skip()
		_, ok = b.peek()
		if !ok {
			return ""
		}
	}
	b.consumeDigit()
	return string(b.b[start:b.n])
}

func parseTransform(s string) []gdf.Matrix {
	var out []gdf.Matrix
	var buf buf
	buf.b = []byte(s)
	for buf.n < len(buf.b) {
		s := string(buf.upTo('('))
		switch s {
		case "matrix":
			// TODO validate that this is accurate
			buf.skip()
			s = string(buf.upTo(')'))
			s = strings.ReplaceAll(s, ",", " ")
			nums := strings.Fields(s)

			var snums = [6]float64{}
			for i := 0; i < len(nums) && i < len(snums); i++ {
				snums[i] = pf(nums[i])
			}
			m := gdf.NewMatrix()
			m.A = snums[0] // x scale
			m.B = snums[1] // x shear
			m.C = snums[2] // y shear
			m.D = snums[3] // y scale
			m.E = snums[4] // x offset
			m.F = snums[5] // y offset
			out = append(out, m)
		case "translate":
			buf.skip()
			buf.skipWSPComma()
			s = string(buf.upTo2(',', '\x20'))
			dx, _ := strconv.ParseFloat(s, 64)
			buf.skip()
			buf.skipWSPComma()
			dy, _ := strconv.ParseFloat(string(buf.upTo(')')), 64)
			out = append(out, gdf.Translate(dx, dy))
		case "skewX":
			s = string(buf.upTo(')'))
			sx, _ := strconv.ParseFloat(s, 64)
			out = append(out, gdf.Skew(sx*gdf.Deg, 0))
		case "skewY":
			s = string(buf.upTo(')'))
			sy, _ := strconv.ParseFloat(s, 64)
			out = append(out, gdf.Skew(0, sy*gdf.Deg))
		case "scale":
			buf.skip()
			buf.skipWSPComma()
			s = string(buf.upTo2(',', '\x20'))
			scaleX, _ := strconv.ParseFloat(s, 64)
			buf.skip()
			buf.skipWSPComma()
			scaleY, _ := strconv.ParseFloat(string(buf.upTo(')')), 64)
			out = append(out, gdf.ScaleBy(scaleX, scaleY))
		case "rotate":
			// TODO
			// by default the rotation is about the origin, but
			// this may yield odd results given the difference between pdf and svg coordinate spaces.
		}
		buf.skip()
		buf.skipWSPComma()
	}
	return out
}
