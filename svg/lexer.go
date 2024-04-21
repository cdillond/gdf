package svg

type lexer struct {
	src []byte
	n   int
	out chan token
}

// See https://www.youtube.com/watch?v=HxaD_trXwRE&t for a discussion of this design pattern.
type stateFn func(*lexer) stateFn

type svgCmd struct {
	op   svgPathOp
	args []float64
}

type parseFn func(chan token) ([]svgCmd, error)

func run(src []byte, start stateFn, parse parseFn) ([]svgCmd, error) {
	l := &lexer{
		src: src,
		out: make(chan token),
	}
	go func() {
		for fn := start; fn != nil; fn = fn(l) {
		}
		close(l.out)
	}()
	return parse(l.out)
}

func (l *lexer) next() (byte, bool) {
	ok := l.n < len(l.src)
	var c byte
	if ok {
		c = l.src[l.n]
		l.n++
	}
	return c, ok
}

func (l *lexer) last() (byte, bool) {
	i := l.n - 1
	ok := (i >= 0) && (i < len(l.src))
	var c byte
	if ok {
		c = l.src[i]
	}
	return c, ok
}

func (l *lexer) peek() (byte, bool) {
	ok := l.n < len(l.src)
	var c byte
	if ok {
		c = l.src[l.n]
	}
	return c, ok
}

func (l *lexer) undo() {
	if l.n > 0 {
		l.n--
	}
}

func (l *lexer) skip() {
	l.n++
}

func (l *lexer) skipWSP() {
	for l.n < len(l.src) && l.src[l.n] < '!' {
		l.n++
	}
}

func (l *lexer) skipWSPComma() {
	for l.n < len(l.src) && (l.src[l.n] == ',' || l.src[l.n] < '!') {
		l.n++
	}
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

type typ int

const (
	bad typ = iota - 1
	num
	op
)

type token struct {
	typ  typ
	text []byte
}

type svgPathOp byte

const (
	repeat svgPathOp = 0

	// moveto
	mAbs svgPathOp = 'M'
	mRel svgPathOp = 'm'

	// closepath
	zAbs svgPathOp = 'Z'
	zRel svgPathOp = 'z'

	// lineto
	lAbs svgPathOp = 'L'
	lRel svgPathOp = 'l'

	// horizontal lineto
	hAbs svgPathOp = 'H'
	hRel svgPathOp = 'h'

	// vertical lineto
	vAbs svgPathOp = 'V'
	vRel svgPathOp = 'v'

	// curveto
	cAbs svgPathOp = 'C'
	cRel svgPathOp = 'c'

	// shorthand/smooth curveto
	sAbs svgPathOp = 'S'
	sRel svgPathOp = 's'

	// quadratic Bézier curveto
	qAbs svgPathOp = 'Q'
	qRel svgPathOp = 'q'

	// Shorthand/smooth quadratic Bézier curveto
	tAbs svgPathOp = 'T'
	tRel svgPathOp = 't'

	badPathToken = 'A'
)

func (p svgPathOp) isValid() bool {
	return p == mAbs || p == mRel || p == lAbs || p == lRel
}

func lexSVGPathOp(l *lexer) stateFn {
	l.skipWSPComma()
	c, ok := l.next()
	if !ok {
		return nil
	}
	l.out <- token{typ: op, text: l.src[l.n-1 : l.n]}
	if c == 'z' || c == 'Z' {
		return lexSVGPathOp
	}
	return lexPathNum
}

func lexPathNum(l *lexer) stateFn {
	l.skipWSPComma()
	i := l.n

	var seenDot bool
	// lex significand, which can be a decimal number
	for c, ok := l.next(); ok; c, ok = l.next() {
		if isDigit(c) {
			continue
		}
		switch c {
		case 'E', 'e':
			goto exponent
		case '+', '-':
			if l.n != i+1 {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPathNum
			}
		case '.':
			if seenDot {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPathNum
			}
			seenDot = true
		case ',', '\x20', '\t', '\n', '\r':
			l.out <- token{typ: num, text: l.src[i : l.n-1]}
			l.skipWSPComma()
			v, ok := l.peek()
			if !ok {
				return nil
			}
			if isDigit(v) || v == '+' || v == '-' || v == '.' {
				return lexPathNum
			}
			return lexSVGPathOp
		default:
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return lexSVGPathOp

		}
	}
	l.out <- token{typ: num, text: l.src[i:l.n]}
	return nil

	// lex exponent, which must be an integer
exponent:
	for c, ok := l.next(); ok; c, ok = l.next() {
		if isDigit(c) {
			continue
		}
		switch c {
		case '+', '-':
			prev, _ := l.last()
			if prev != 'E' && prev != 'e' {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPathNum
			}
		case '.':
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return lexPathNum
		case ',', '\x20', '\t', '\n', '\r':
			l.out <- token{typ: num, text: l.src[i : l.n-1]}
			l.skipWSPComma()
			v, ok := l.peek()
			if !ok {
				return nil
			}
			if isDigit(v) || v == '+' || v == '-' || v == '.' {
				return lexPathNum
			}
			return lexSVGPathOp
		default:
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return lexSVGPathOp
		}
	}
	l.out <- token{typ: num, text: l.src[i:l.n]}
	return nil
}

// Very similar to LexPathNum.
func lexPolyNum(l *lexer) stateFn {
	l.skipWSPComma()
	i := l.n

	var seenDot bool
	// lex significand, which can be a decimal number
	for c, ok := l.next(); ok; c, ok = l.next() {
		if isDigit(c) {
			continue
		}
		switch c {
		case 'E', 'e':
			goto exponent
		case '+', '-':
			if l.n != i+1 {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPolyNum
			}
		case '.':
			if seenDot {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPolyNum
			}
			seenDot = true
		case ',', '\x20', '\t', '\n', '\r':
			l.out <- token{typ: num, text: l.src[i : l.n-1]}
			l.skipWSPComma()
			v, ok := l.peek()
			if !ok {
				return nil
			}
			if isDigit(v) || v == '+' || v == '-' || v == '.' {
				return lexPolyNum
			}
			return nil
		default:
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return nil

		}
	}
	l.out <- token{typ: num, text: l.src[i:l.n]}
	return nil

	// lex exponent, which must be an integer
exponent:
	for c, ok := l.next(); ok; c, ok = l.next() {
		if isDigit(c) {
			continue
		}
		switch c {
		case '+', '-':
			prev, _ := l.last()
			if prev != 'E' && prev != 'e' {
				l.undo()
				l.out <- token{typ: num, text: l.src[i:l.n]}
				return lexPolyNum
			}
		case '.':
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return lexPolyNum
		case ',', '\x20', '\t', '\n', '\r':
			l.out <- token{typ: num, text: l.src[i : l.n-1]}
			l.skipWSPComma()
			v, ok := l.peek()
			if !ok {
				return nil
			}
			if isDigit(v) || v == '+' || v == '-' || v == '.' {
				return lexPolyNum
			}
			return nil
		default:
			l.undo()
			l.out <- token{typ: num, text: l.src[i:l.n]}
			return nil
		}
	}
	l.out <- token{typ: num, text: l.src[i:l.n]}
	return nil
}

func parseCmd(b []byte) svgPathOp {
	if len(b) != 1 {
		return 0
	}
	valid := []byte("CcHhLlMmQqSsTtVvZz")
	for i := range valid {
		if b[0] == valid[i] {
			return svgPathOp(b[0])
		}
	}
	return badPathToken
}
