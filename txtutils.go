package gdf

type ShapedRun struct {
	Txt   []rune
	Kerns []int
	Advs  []int
	FUExt float64
}

// returns the extent of the text in points
func (s ShapedRun) Extent(fs float64) float64 {
	return FUToPt(s.FUExt, fs)
}

func Shape(s string, f *Font) ShapedRun {
	txt := []rune(s)
	st := ShapedRun{
		Txt:   txt,
		Kerns: make([]int, len(txt)),
		Advs:  make([]int, len(txt)),
	}
	i := 0
	var ext int
	for ; i < len(txt)-1; i++ {
		adv, kern := ShapedGlyphAdv(txt[i], txt[i+1], f)
		st.Kerns[i] = kern
		st.Advs[i] = adv
		ext += adv + kern
	}
	adv := GlyphAdvance(txt[i], f)
	st.Advs[i] = adv
	ext += adv
	st.FUExt = float64(ext)
	return st
}

// returns the x offset, in points, from the start of rect, needed to center s
func CenterH(s ShapedRun, rect Rect, fs float64) float64 {
	ext := s.Extent(fs)
	dif := rect.URX - rect.LLX - ext
	return dif / 2
}

// returns the y offset, in points, from the start of rect, needed to center s vertically (based on the text's leading)
func CenterV(leading float64, rect Rect) float64 {
	return -(rect.URY - rect.LLY - leading) / 2
}
