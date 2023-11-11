package content

import (
	"fmt"

	"github.com/cdillond/gdf"
)

func WriteParagraph(cs *gdf.ContentStream, p gdf.Paragraph, startIndex int, bounds gdf.Rect) (int, error) {
	width := gdf.FUToPt(bounds.Width(), cs.FontSize)
	if width > p.MaxW {
		return 0, fmt.Errorf("paragraph is too wide for target area %v  %v", width, p.MaxW)
	}
	if startIndex >= len(p.Lines) {
		return 0, fmt.Errorf("EOP")
	}
	i := startIndex
	var j int // get Kerns index
	for ; j < i; j++ {
		j += len(p.Lines[j])
	}
	cs.TStar()

	if pt := cs.TextCursor(); pt.Y > bounds.LLY && i == 0 && p.Indent {
		tabW := float64(5 * gdf.GlyphAdvance(' ', cs.Font))
		cs.Td(gdf.FUToPt(tabW, cs.FontSize), 0)
		var line []rune
		var kerns []int
		if exists, ok := p.Hyphens[i]; ok {
			line = make([]rune, len(p.Lines[i]))
			copy(line, p.Lines[i])
			kerns = make([]int, len(p.Lines[i]))
			copy(kerns, p.Kerns[j+1:j+len(line)+1])
			if !exists {
				line = append(line, '\u00AD')
				kerns = append(kerns, 0)
			}
		} else {
			line = p.Lines[i]
			kerns = p.Kerns[j+1 : j+len(line)+1]
		}
		cs.TJSpace(line, kerns, p.Difs[i])
		j += len(p.Lines[i])
		cs.TStar()
		cs.Td(gdf.FUToPt(-1*tabW, cs.FontSize), 0)
		i++
	}

	for pt := cs.TextCursor(); pt.Y > bounds.LLY && i < len(p.Lines); pt = cs.TextCursor() {
		var line []rune
		if exists, ok := p.Hyphens[i]; ok {
			line = make([]rune, len(p.Lines[i]))
			copy(line, p.Lines[i])
			if !exists {
				line = append(line, '\u00AD')
			}
		} else {
			line = p.Lines[i]
		}
		err := cs.TJSpace(line, p.Kerns[j+1:j+len(line)+1], p.Difs[i])
		if err != nil {
			fmt.Println(err)
		}
		j += len(p.Lines[i])
		cs.TStar()
		i++
	}
	return i - startIndex, nil
}
