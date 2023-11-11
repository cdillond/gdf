package content

import (
	"github.com/cdillond/gdf"
)

// Draws the bitmap as a QR code of height h starting at start. The fg and bg parameters specify the colors
// to use for the dark (foreground) and light (background) pixels. If a nil Color is passed to either of these parameters, the respective pixels will
// not be drawn. This can be used to paint bitmap with no background. The foreground color should not be nil.
func QRCode(c *gdf.ContentStream, bitmap [][]bool, h float64, start gdf.Point, fg, bg gdf.Color) {
	bitsize := h / float64(len(bitmap[0]))
	c.QSmall()
	if bg != nil {
		c.SetColor(bg)
		c.Re(start.X, start.Y, h, h)
		c.Fill()
	}
	if fg == nil {
		c.Q()
		return
	}
	c.SetColor(fg)
	// the origin of the bitmap is the (ULX,ULY) of the rectangle, whereas PDF rectangles are drawn with
	// the origin at (LLX, LLY). Each row needs to be offset vertically to adjust for this
	for i, row := range bitmap {
		for j, p := range row {
			if p {
				c.Re(start.X+float64(j)*bitsize, start.Y+h-(float64(i+1)*bitsize), bitsize, bitsize)
				c.Fill()
			}
		}
	}
	c.Q()
}
