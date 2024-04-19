package gdf

func checkOn() *XObject {
	box := Rect{0, 0, 72, 72}
	on := NewFormXObj(make([]byte, 0, 512), box)
	on.SetLineWidth(2)
	on.DrawLine(box.LLX+12, box.LLY+12, box.URX-12, box.URY-12)
	on.DrawLine(box.LLX+12, box.URY-12, box.URX-12, box.LLY+12)
	return on
}

func checkOff() *XObject {
	box := Rect{0, 0, 72, 72}
	off := NewFormXObj(make([]byte, 0, 512), box)
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
func DefaultCheckbox() CheckboxCfg {
	return CheckboxCfg{
		On:    checkOn(),
		Off:   checkOff(),
		Flags: PrintAnnot,
	}
}
