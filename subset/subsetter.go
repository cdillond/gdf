package subset

import (
	"fmt"

	"golang.org/x/image/font/sfnt"
)

type BasicSubsetter struct {
	sFNT *sfnt.Font
	src  []byte
}

func (b *BasicSubsetter) Subset(cutset map[rune]struct{}) ([]byte, error) {
	return TTFSubset(b.sFNT, b.src, cutset)
}
func (b *BasicSubsetter) Init(SFNT *sfnt.Font, src []byte, _ string) {
	b.sFNT = SFNT
	b.src = src
}

type HarfBuzzSubsetter struct {
	path string // Path to the font file; may be empty if Src is non-nil
	src  []byte // Font source bytes; may be nil if Path is not empty
}

func (h *HarfBuzzSubsetter) Subset(cutset map[rune]struct{}) ([]byte, error) {
	if h.path != "" {
		return HBSubsetPath(h.path, cutset)
	}
	if h.src != nil {
		return HBSubset(h.src, cutset)
	}
	return nil, fmt.Errorf("no font source data provided")
}

func (h *HarfBuzzSubsetter) Init(_ *sfnt.Font, src []byte, path string) {
	h.path = path
	h.src = src
}

type HarfBuzzCGoSubsetter struct {
	src []byte // Font source bytes
}

func (h *HarfBuzzCGoSubsetter) Subset(cutset map[rune]struct{}) ([]byte, error) {
	return HBSubsetC(h.src, cutset)
}

func (h *HarfBuzzCGoSubsetter) Init(_ *sfnt.Font, src []byte, _ string) {
	h.src = src
}
