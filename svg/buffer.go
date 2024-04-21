package svg

// custom buffer struct for common operations
type buffer struct {
	b []byte
	n int
}

func newBuffer(b []byte) buffer {
	return buffer{b: b}
}

func (b *buffer) upTo(c byte) []byte {
	i := b.n
	for i < len(b.b) && b.b[i] != c {
		i++
	}
	res := b.b[b.n:i]
	b.n = i
	return res
}

func (b *buffer) upTo2(c1, c2 byte) []byte {
	i := b.n
	for i < len(b.b) && b.b[i] != c1 && b.b[i] != c2 {
		i++
	}
	res := b.b[b.n:i]
	b.n = i
	return res
}

func (b *buffer) peek() (byte, bool) {
	ok := b.n+1 < len(b.b)
	var c byte
	if ok {
		c = b.b[b.n+1]
	}
	return c, ok
}

func (b *buffer) skip() {
	b.n++
}

func (b *buffer) skipWSP() {
	for b.n < len(b.b) && b.b[b.n] <= '\x20' {
		b.n++
	}
}
