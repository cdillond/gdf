package gdf

import (
	"bytes"
)

func isConsonant(b byte) bool {
	if ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z') {
		switch b {
		case 'A', 'a', 'E', 'e', 'I', 'i', 'O', 'o', 'U', 'u':
			return false
		default:
			return true
		}
	}
	return false
}

func isDigraph(b []byte) bool {
	if bytes.EqualFold(b, []byte("th")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ch")) {
		return true
	}
	if bytes.EqualFold(b, []byte("sh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ph")) {
		return true
	}
	if bytes.EqualFold(b, []byte("dg")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wn")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wd")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wl")) {
		return true
	}
	if bytes.EqualFold(b, []byte("gh")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ng")) {
		return true
	}
	if bytes.EqualFold(b, []byte("sc")) {
		return true
	}
	if bytes.EqualFold(b, []byte("nx")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ck")) {
		return true
	}
	if bytes.EqualFold(b, []byte("kn")) {
		return true
	}
	if bytes.EqualFold(b, []byte("wr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("nd")) {
		return true
	}
	if bytes.EqualFold(b, []byte("tr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("dr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("cr")) {
		return true
	}
	if bytes.EqualFold(b, []byte("ll")) {
		return true
	}
	return false
}

func wordStart(b []byte) bool {
	if bytes.EqualFold(b, []byte("nc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("bc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("bz")) {
		return false
	}
	if bytes.EqualFold(b, []byte("dc")) {
		return false
	}
	if bytes.EqualFold(b, []byte("dd")) {
		return false
	}

	return true
}

func isAlpha(b byte) bool {
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

// Returns the index after which a hyphen should be inserted
func intraWordBP(word []byte) (int, bool) {
	// don't break words shorter than 5 characters
	if len(word) < 5 {
		return -1, false
	}
	// if a word itself has a hyphen, break at the hyphen
	// does not distinguish between \u002D and \u00AD
	if i := bytes.IndexAny(word, "\u002D\u00AD"); i != -1 {
		return i, true
	}
	// don't break proper nouns
	if word[0] >= 'A' && word[0] <= 'Z' {
		return -1, false
	}
	// only break between consonants that do not form a digraph
	// and do not leave an unacceptable beginning consonant pair
	for i := 2; i < len(word)-4; i++ {
		if isConsonant(word[i]) {
			if !isConsonant(word[i+1]) {
				continue
			}
			if isDigraph(word[i : i+2]) {
				continue
			}
			if !wordStart(word[i+1 : i+3]) {
				continue
			}
			if !isAlpha(word[i+3]) {
				continue
			}
			return i, false
		}
	}
	return -1, false
}
