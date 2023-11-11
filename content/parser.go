package content

// control characters

const (
	LF rune = '\n'
	CR rune = '\r'
	HT rune = '\t'

	SO rune = '\x0E' // mark start of bold text
	SI rune = '\x0F' // mark start of italic text
)

type TextStateFlag uint

const (
	FLAG_INDENTED TextStateFlag = 1 << iota
	FLAG_BOLD
	FLAG_ITALIC
)
const (
	BOLD_ITALIC          = FLAG_BOLD | FLAG_ITALIC
	BOLD_ITALIC_INDENTED = BOLD_ITALIC | FLAG_INDENTED
	BOLD_INDENTED        = FLAG_BOLD | FLAG_INDENTED
	ITALIC_INDENTED      = FLAG_ITALIC | FLAG_INDENTED
)
