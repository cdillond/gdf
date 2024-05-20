package svg

import "strconv"

// svg_path::= wsp* moveto? (moveto drawto_command*)?

func (l *lexer) rpath() {
	l.skipWSP()
	// look for moveto

	// look for drawto_commands
}

type mtcmd struct {
	op   byte
	args [][2]float64
}

// moveto::= ( "M" | "m" ) wsp* coordinate_pair_sequence
func (l *lexer) readMoveTo() (mtcmd, bool) {
	var out mtcmd
	var c byte
	var ok bool
	start := l.n
	if c, ok = l.next(); ok && (c == 'M' || c == 'm') {
		out.op = c
	} else {
		l.undo()
		return out, ok
	}
	l.skipWSP()
	out.args, ok = l.readCoordinatePairSequence()
	if !ok {
		l.backup(start)
	}
	return out, ok
}

// coordinate::= sign? number
func (l *lexer) readCoordinate() (float64, bool) {
	sign, _ := l.readSign()
	b := l.readUint()
	if b == nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return 0, false
	}
	if sign == '-' {
		f = -f
	}
	return f, true
}

// coordinate_pair::= coordinate comma_wsp? coordinate
func (l *lexer) readCoordinatePair() (out [2]float64, ok bool) {
	start := l.n
	out[0], ok = l.readCoordinate()
	if !ok {
		l.backup(start)
		return out, ok
	}
	l.skipCWSP()
	out[1], ok = l.readCoordinate()
	if !ok {
		l.backup(start)
	}
	return out, ok
}

// coordinate_pair_sequence::= coordinate_pair | (coordinate_pair comma_wsp? coordinate_pair_sequence)
func (l *lexer) readCoordinatePairSequence() ([][2]float64, bool) {
	var out [][2]float64
	for pair, ok := l.readCoordinatePair(); ok; pair, ok = l.readCoordinatePair() {
		out = append(out, pair)
		l.skipCWSP()
	}
	return out, out != nil
}

// sign::= "+"|"-"
func isSign(c byte) bool { return c == '+' || c == '-' }

// number ::= ([0-9])+
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// flag::=("0"|"1")
func isFlag(c byte) bool { return c == '0' || c == '1' }

// wsp ::= (#x9 | #x20 | #xA | #xC | #xD)
func isWSP(c byte) bool { return c < '!' }

// comma_wsp::=(wsp+ ","? wsp*) | ("," wsp*)
func isCWSP(c byte) bool { return c == ',' || c < '!' }

func (l *lexer) readUint() []byte {
	var out []byte
	start := l.n
	for c, ok := l.peek(); ok && isDigit(c); c, ok = l.peek() {
		l.skip()
	}
	if l.n > start {
		out = l.src[start:l.n]
	}
	return out
}
func (l *lexer) readSign() (c byte, ok bool) {
	if c, ok = l.peek(); ok && isSign(c) {
		l.skip()
		return c, ok
	}
	return 0, false
}
func (l *lexer) readFlag() []byte {
	var out []byte
	if c, ok := l.peek(); ok && isFlag(c) {
		l.skip()
		out = []byte{c}
	}
	return out
}
func (l *lexer) skipWSP() {
	for l.n < len(l.src) && isWSP(l.src[l.n]) {
		l.n++
	}
}
func (l *lexer) skipCWSP() {
	for l.n < len(l.src) && isCWSP(l.src[l.n]) {
		l.n++
	}
}

/*
elliptical_arc::= ( "A" | "a" ) wsp* elliptical_arc_argument_sequence

elliptical_arc_argument_sequence::=  elliptical_arc_argument  | (elliptical_arc_argument comma_wsp? elliptical_arc_argument_sequence)

elliptical_arc_argument::=number comma_wsp? number comma_wsp? number comma_wsp flag comma_wsp? flag comma_wsp? coordinate_pair

*/

func (l *lexer) readArcSequence() {}

func (l *lexer) readEArc() (out [7]float64, ok bool) {
	var err error
	var i int
	start := l.n
	for ; i < 3; i++ {
		b := l.readUint()
		if b == nil {
			l.backup(start)
			return out, false
		}
		out[i], err = strconv.ParseFloat(string(b), 64)
		if err != nil {
			l.backup(start)
			return out, false
		}
		l.skipCWSP()
	}
	// now search for flag
	for ; i < 5; i++ {
		b := l.readFlag()
		if b == nil {
			l.backup(start)
			return out, false
		}
		out[i], err = strconv.ParseFloat(string(b), 64)
		if err != nil {
			return
		}
		l.skipCWSP()
	}

	tuple, ok := l.readCoordinatePair()
	if !ok {
		l.backup(start)
		return out, false
	}
	out[5], out[6] = tuple[0], tuple[1]
	return out, true
}
