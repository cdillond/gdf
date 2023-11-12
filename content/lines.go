package content

import "github.com/cdillond/gdf"

type Alignment uint

const (
	ALIGN_LEFT Alignment = iota
	ALIGN_CENTER
	ALIGN_RIGHT
)

func Line(cs *gdf.ContentStream, start, end gdf.Point) {
	cs.MoveTo(start.X, start.Y)
	cs.LineTo(end.X, end.Y)
	cs.Stroke()
}

// Draws text to a rectangle beginning at start, strokes the rectangle, and returns the height and width of the resulting untransformed
func TextBox(cs *gdf.ContentStream, text string, start gdf.Point, margins gdf.Margins) (float64, float64) {
	ext := cs.TextExtentPts(text) + margins.Left + margins.Right
	ascFU, descFU := gdf.TextAscDesc([]rune(text), cs.Font)
	asc := gdf.FUToPt(ascFU, cs.FontSize)
	desc := gdf.FUToPt(descFU, cs.FontSize)
	height := asc + desc + margins.Top + margins.Bottom
	et, _ := cs.BeginText()
	cs.Td(start.X+margins.Left, start.Y)
	cs.Tj(text)
	et()
	cs.Re(start.X, start.Y-desc-margins.Bottom, ext, height)
	cs.Stroke()
	return height, ext
}
