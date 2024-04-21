package gdf

type TextState struct {
	*Font
	FontSize  float64 // points
	CharSpace float64 // points
	WordSpace float64 // points
	HScale    float64 // horizontal scale, as a percentage of the normal width
	Leading   float64 // points
	Rise      float64 // points
	RenderMode
}

type RenderMode uint

const (
	FillText RenderMode = iota
	StrokeText
	FillStrokeText
	Invisible
	FillTextAddPath
	StrokeTextAddPath
	FillStrokeTextAddPath
	AddTextPath
	badRenderMode
)

func (r RenderMode) isValid() bool { return r < badRenderMode }

// SetCharSpace sets the content stream's character spacing (c.TextState.CharSpace) f.
func (c *ContentStream) SetCharSpace(f float64) {
	c.CharSpace = f
	c.buf = cmdf(c.buf, op_Tc, f)
}

// SetWordSpace sets the content stream's word spacing (c.TextState.WordSpace) to f.
func (c *ContentStream) SetWordSpace(f float64) {
	c.WordSpace = f
	c.buf = cmdf(c.buf, op_Tw, f)
}

// SetHScale sets the content stream's horizontal text scaling percentage (c.TextState.Scale) to f. The default value is 100.0 percent.
func (c *ContentStream) SetHScale(f float64) {
	c.HScale = f
	c.buf = cmdf(c.buf, op_Tz, f)
}

// SetLeading sets the content stream's text leading/line height (c.TextState.Leading) to f.
func (c *ContentStream) SetLeading(f float64) {
	c.Leading = f
	c.buf = cmdf(c.buf, op_TL, f)
}

// SetFont sets the current font size and font (c.TextState.Font and c.TextState.FontSize) to size and font.
func (c *ContentStream) SetFont(size float64, font *Font) {
	var i int
	for ; i < len(c.resources.Fonts); i++ {
		if c.resources.Fonts[i] == font {
			break
		}
	}
	if i == len(c.resources.Fonts) {
		c.resources.Fonts = append(c.resources.Fonts, font)
	}
	c.FontSize = size
	c.Font = font
	c.buf = append(c.buf, []byte("/F"+itoa(i)+"\x20")...)
	c.buf = cmdf(c.buf, op_Tf, size)
}

// SetRenderMode sets the current text rendering mode (c.TextState.RenderMode) to r.
func (c *ContentStream) SetRenderMode(r RenderMode) {
	if !r.isValid() {
		return
	}
	c.RenderMode = r
	c.buf = cmdi(c.buf, op_Tr, int64(r))
}

// SetRise sets the current text rise (c.TextState.RenderMode) to f.
func (c *ContentStream) SetRise(f float64) {
	c.Rise = f
	c.buf = cmdf(c.buf, op_Ts, f)
}

// DrawXContent draws x to c at c's current point.
func (c *ContentStream) DrawXContent(x *XContent) {
	var i int
	for ; i < len(c.resources.XForms); i++ {
		if c.resources.XForms[i] == x {
			break
		}
	}
	if i == len(c.resources.XForms) {
		c.resources.XForms = append(c.resources.XForms, x)
	}

	c.buf = append(c.buf, "/P"...)
	c.buf = itobuf(i, c.buf)
	c.buf = append(c.buf, '\x20')
	c.buf = append(c.buf, op_Do...)
}

func (c *ContentStream) DrawImage(img *Image) {
	var i int
	for ; i < len(c.resources.Images); i++ {
		if c.resources.Images[i] == img {
			break
		}
	}
	if i == len(c.resources.Images) {
		c.resources.Images = append(c.resources.Images, img)
	}
	c.buf = append(c.buf, "/Im"...)
	c.buf = itobuf(i, c.buf)
	c.buf = append(c.buf, '\x20')
	c.buf = append(c.buf, op_Do...)
}

func (c *ContentStream) DrawImageTo(dst Rect, x *Image) {
	xScale := dst.Width()
	yScale := dst.Height()
	c.QSave()
	c.Concat(Translate(dst.LLX, dst.LLY))
	c.Concat(ScaleBy(xScale, yScale))
	var i int
	for ; i < len(c.resources.Images); i++ {
		if c.resources.Images[i] == x {
			break
		}
	}
	if i == len(c.resources.Images) {
		c.resources.Images = append(c.resources.Images, x)
	}

	c.buf = append(c.buf, "/Im"...)
	c.buf = itobuf(i, c.buf)
	c.buf = append(c.buf, '\x20')
	c.buf = append(c.buf, op_Do...)
	c.QRestore()
}

// DrawXContent adjusts the CTM such that the contents of x are drawn to dst.
func (c *ContentStream) DrawXContentTo(dst Rect, x *XContent) {
	var i int
	for ; i < len(c.resources.XForms); i++ {
		if c.resources.XForms[i] == x {
			break
		}
	}
	if i == len(c.resources.XForms) {
		c.resources.XForms = append(c.resources.XForms, x)
	}
	xScale := dst.Width() / x.BBox.Width()
	yScale := dst.Height() / x.BBox.Height()
	c.QSave()
	c.Concat(Translate(dst.LLX-x.BBox.LLX*xScale, dst.LLY-x.BBox.LLY*yScale))
	c.Concat(ScaleBy(xScale, yScale))
	c.buf = append(c.buf, "/P"...)
	c.buf = itobuf(i, c.buf)
	c.buf = append(c.buf, '\x20')
	c.buf = append(c.buf, op_Do...)
	c.QRestore()
}
