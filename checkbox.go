package gdf

func checkOn(r Rect) *XObject {
	box := Point{0, 0}.ToRect(r.Width(), r.Height())
	on := NewXObj(XForm, box)
	box = box.Bounds(Margins{1, 1, 1, 1})
	on.Re2(box)
	on.Stroke()
	on.DrawLine(box.LLX+3, box.LLY+3, box.URX-3, box.URY-3)
	on.DrawLine(box.LLX+3, box.URY-3, box.URX-3, box.LLY+3)
	return on
}

func checkOnNoBorder(r Rect) *XObject {
	box := Point{0, 0}.ToRect(r.Width(), r.Height())
	on := NewXObj(XForm, box)
	on.DrawLine(box.LLX+3, box.LLY+3, box.URX-3, box.URY-3)
	on.DrawLine(box.LLX+3, box.URY-3, box.URX-3, box.LLY+3)
	return on
}

func checkOff(r Rect) *XObject {
	box := Point{0, 0}.ToRect(r.Width(), r.Height())
	off := NewXObj(XForm, box)
	box = box.Bounds(Margins{1, 1, 1, 1})
	off.Re2(box)
	off.Stroke()
	return off
}

func checkOffNoBorder(r Rect) *XObject {
	box := Point{0, 0}.ToRect(r.Width(), r.Height())
	off := NewXObj(XForm, box)
	return off
}

/*
The PDF spec requires a PDF document to describe the appearance of any acroform fields it contains. This means
that the appearance of each form field is customizable, but it also means that the 'On' and 'Off' states (as well as the default state)
of a checkbox need to be specified in advance. DefaultCheckbox is a convenience function that returns a Checkbox design
suitable for use with most sizes of Checkbox. Checkboxes with dimensions smaller than 4 pts will look bad. The returned
Checkbox design can be reused for any Checkbox-type (Button) acroform fields of the same size. NOTE: The appearance of a Checkbox may be dependent
on the PDF viewer used to render it. This can be especially noticeable when withBorder is true. It is often preferable to draw the Checkbox
border separately.
*/
func DefaultCheckbox(size Rect, withBorder bool) CheckboxCfg {
	if withBorder {
		return CheckboxCfg{
			On:  checkOn(size),
			Off: checkOff(size),
		}
	}
	return CheckboxCfg{
		On:  checkOnNoBorder(size),
		Off: checkOffNoBorder(size),
	}
}
