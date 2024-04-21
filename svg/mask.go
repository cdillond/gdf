package svg

func parseMaskURL(s string) string {
	buf := newBuffer([]byte(s))
	buf.upTo('(')
	buf.skip()
	buf.skip()
	return string(buf.upTo(')'))
}
