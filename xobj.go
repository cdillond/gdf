package gdf

import (
	"io"
)

// An XContent struct represents a PDF form-type XObject. It is essentially a
// ContentStream that can be displayed by multiple Pages of a PDF.
type XContent struct {
	ContentStream
	BBox Rect
}

// An XImage represents an image that is external to any given PDF.
type XImage struct {
	Data             []byte
	Width            int // The width of the image in pixels.
	Height           int // The height of the image in pixels.
	ColorSpace       ColorSpace
	BitsPerComponent int    // The bit depth of the image's encoding.
	AppliedFilter    Filter // The filter used to pre-compress the image Data.
	RawDataLen       int    // The length (in bytes) of the uncompressed image Data; only needed if AppliedFilter is nonzero.
	Alpha            *XImage
}

// NewImage returns an Image that can be drawn to ContentStreams derived from p.
func (p PDF) NewImage(x XImage) *Image {
	img := &Image{
		Data:             x.Data,
		Width:            x.Width,
		Height:           x.Height,
		ColorSpace:       x.ColorSpace,
		BitsPerComponent: x.BitsPerComponent,
		AppliedFilter:    x.AppliedFilter,
		RawDataLen:       x.RawDataLen,
	}
	if x.Alpha != nil {
		img.Alpha = &Image{
			Data:             x.Alpha.Data,
			Width:            x.Alpha.Width,
			Height:           x.Alpha.Height,
			ColorSpace:       x.Alpha.ColorSpace,
			BitsPerComponent: x.Alpha.BitsPerComponent,
			AppliedFilter:    x.Alpha.AppliedFilter,
			RawDataLen:       x.Alpha.RawDataLen,
		}
	}
	return img
}

// An Image represents a raster image.
type Image struct {
	Data             []byte
	Width            int // The width of the image in pixels.
	Height           int // The height of the image in pixels.
	ColorSpace       ColorSpace
	BitsPerComponent int    // The bit depth of the image's encoding.
	Alpha            *Image // An image's alpha channel, if present, must be encoded as a separate image. The Alpha image's ColorSpace should be DeviceGray.
	AppliedFilter    Filter // The filter used to pre-compress the image Data.
	RawDataLen       int    // The length (in bytes) of the uncompressed image Data; only needed if AppliedFilter is nonzero.
	refnum           int
}

func (x *Image) Bytes() []byte {
	return x.Data
}

func NewXContent(b []byte, BBox Rect) *XContent {
	x := &XContent{
		BBox: BBox,
	}
	x.buf = b
	x.GS = newGS()
	x.Filter = Flate
	return x
}

func (x *XContent) Bytes() []byte {
	return x.stream.buf
}

func (x *XContent) mark(i int) { x.refnum = i }
func (x *XContent) id() int    { return x.refnum }
func (x *XContent) children() []obj {
	out := make([]obj, 0, len(x.resources.Fonts)+len(x.resources.XForms)+len(x.resources.Images))
	for i := range x.resources.Fonts {
		out = append(out, x.resources.Fonts[i])
	}
	for i := range x.resources.XForms {
		out = append(out, x.resources.XForms[i])
	}
	for i := range x.resources.Images {
		out = append(out, x.resources.Images[i])
	}
	return out
}
func (x *XContent) encode(w io.Writer) (int, error) {
	x.stream.extras = []field{
		{"/Type", "/XObject"},
		{"/Subtype", "/Form"},
		{"/BBox", x.BBox},
		{"/Resources", x.resources.bytes()},
	}
	return x.stream.encode(w)
}

func (x *Image) mark(i int) { x.refnum = i }
func (x *Image) id() int    { return x.refnum }
func (x *Image) children() []obj {
	if x.Alpha != nil {
		return []obj{x.Alpha}
	}
	return nil
}
func (x *Image) encode(w io.Writer) (int, error) {
	s := stream{
		buf:    x.Data,
		Filter: x.AppliedFilter,
	}
	s.extras = []field{
		{"/Type", "/XObject"},
		{"/Subtype", "/Image"},
		{"/Width", x.Width},
		{"/Height", x.Height},
		{"/ColorSpace", x.ColorSpace.String()},
		{"/BitsPerComponent", x.BitsPerComponent},
	}
	if x.Alpha != nil {
		s.extras = append(s.extras, field{"/SMask", iref(x.Alpha)})
	}
	if x.AppliedFilter != DefaultFilter && x.RawDataLen > 0 {
		s.extras = append(s.extras, []field{
			{"/Length1", x.RawDataLen},
			{"/Filter", x.AppliedFilter.String()},
		}...)
		s.Filter = NoFilter
	}
	return s.encode(w)
}
