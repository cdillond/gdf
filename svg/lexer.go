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
	}
	l.n++
	return c, ok
}

func (l *lexer) again() (byte, bool) {
	i := l.n - 1
	ok := (i >= 0) && (i < len(l.src))
	var c byte
	if ok {
		c = l.src[i]
	}
	return c, ok
}

func (l *lexer) last() (byte, bool) {
	i := l.n - 2
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

// reverse to position i
func (l *lexer) backup(i int) {
	l.n = i
}

/*
func (l *lexer) skipWSP() {
	for l.n < len(l.src) && l.src[l.n] < '!' {
		l.n++
	}
}

func (l *lexer) skipWSPComma() {
	for l.n < len(l.src) && (l.src[l.n] == ',' || l.src[l.n] < '!') {
		l.n++
	}
}*/

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

	aAbs svgPathOp = 'A'
	aRel svgPathOp = 'a'

	badPathToken svgPathOp = '?'
)

func lexSVGPathOp(l *lexer) stateFn {
	l.skipCWSP()
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
	l.skipCWSP()
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
			l.skipCWSP()
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
			l.skipCWSP()
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

// annoying! The problem is that the flag args can be, e.g. 00 or 01 instead of 0 0 or 0 1.
// elliptical_arc::=( "A" | "a" ) wsp* elliptical_arc_argument_sequence
// elliptical_arc_argument_sequence::=elliptical_arc_argument | (elliptical_arc_argument comma_wsp? elliptical_arc_argument_sequence)
// elliptical_arc_argument::=number comma_wsp? number comma_wsp? number comma_wsp flag comma_wsp? flag comma_wsp? coordinate_pair

// Very similar to LexPathNum.
func lexPolyNum(l *lexer) stateFn {
	l.skipCWSP()
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
			l.skipCWSP()
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
			l.skipCWSP()
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
	valid := []byte("AaCcHhLlMmQqSsTtVvZz")
	for i := range valid {
		if b[0] == valid[i] {
			return svgPathOp(b[0])
		}
	}
	return badPathToken
}
