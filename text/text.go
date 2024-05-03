package text

import (
	"github.com/cdillond/gdf"
)

// CenterH returns the x offset, in points, from the start of rect, needed to center t horizontally, if drawn with c's current
// font at c's current font size.
func CenterH(c *gdf.ContentStream, t []rune, rect gdf.Rect) float64 {
	ext := gdf.FUToPt(c.Extent(t), c.FontSize)
	dif := rect.URX - rect.LLX - ext
	return dif / 2
}

// CenterV returns the y offset, in points, from the start of rect, needed to center a line of text vertically based on the text's height.
// This is a naive approach.
func CenterV(height float64, rect gdf.Rect) float64 {
	return -(rect.URY - rect.LLY - height) / 2
}
