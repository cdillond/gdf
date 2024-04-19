package gdf

import (
	"io"
)

type xType bool

const (
	XForm  xType = false
	XImage xType = true
)

func (x xType) String() string {
	if x {
		return "/Image"
	}
	return "/Form"
}

// An XObject is a content stream that can be reused. It can either represent an image or an arbitrary sequence of objects, i.e. a "form". (Form XObjects are unrelated to AcroForms and XFA Forms; the nomenclature is confusing.)
type XObject struct {
	xType
	ContentStream
	ImageDict
	BBox Rect
}

type ImageDict struct {
	Format           ImageFormat
	Width            int // The width of the image in pixels.
	Height           int // The height of the image in pixels.
	ColorSpace       ColorSpace
	BitsPerComponent int      // The bit depth of the image's encoding.
	Alpha            *XObject // An image's alpha channel, if present, must be encoded as a separate image. The Alpha image's ColorSpace should be DeviceGray.
}

type ImageFormat = Filter

// Raster image formats supported by gdf.
const (
	JPEG         ImageFormat = DCTDecode
	PNG          ImageFormat = Flate
	PNGAlphaMask ImageFormat = Flate
)

// NewImageXObj returns an *XObject representing an image, using b as the image's raw
// data and d as information about the image. The returned image will always have a BBox
// of Rect{0,0,1,1}, so it must be properly scaled to be visible.
func NewImageXObj(b []byte, d ImageDict) *XObject {
	return &XObject{
		BBox:  Rect{0, 0, 1, 1},
		xType: XImage,
		ContentStream: ContentStream{
			stream: stream{
				buf:    b,
				Filter: d.Format,
			},
		},
		ImageDict: d,
	}
}

// NewFormXObj returns a form-type XObject with a buffer of b and a bounding box of BBox.
func NewFormXObj(b []byte, BBox Rect) *XObject {
	x := &XObject{
		BBox:  BBox,
		xType: XForm,
	}
	x.buf = b
	x.GS = newGS()
	x.Filter = Flate
	return x
}

// Bytes exposes the XObject's underlying byte slice.
func (x XObject) Bytes() []byte {
	return x.buf
}

func (x *XObject) mark(i int) { x.refnum = i }
func (x *XObject) id() int    { return x.refnum }
func (x *XObject) children() []obj {
	out := make([]obj, 0, len(x.resources.Fonts)+len(x.resources.XObjs)+oneif(x.Alpha != nil))
	for i := range x.resources.Fonts {
		out = append(out, x.resources.Fonts[i])
	}
	for i := range x.resources.XObjs {
		out = append(out, x.resources.XObjs[i])
	}
	if x.Alpha != nil {
		out = append(out, x.Alpha)
	}
	return out
}

func (x *XObject) encode(w io.Writer) (int, error) {
	switch x.xType {
	case XForm:
		x.stream.extras = []field{
			{"/Type", "/XObject"},
			{"/Subtype", x.xType.String()},
			{"/BBox", x.BBox},
			{"/Resources", x.resources.bytes()},
		}
	case XImage:
		x.Filter = x.ImageDict.Format
		x.stream.extras = []field{
			{"/Type", "/XObject"},
			{"/Subtype", x.xType.String()},
			{"/Width", x.Width},
			{"/Height", x.Height},
			{"/ColorSpace", x.ColorSpace.String()},
			{"/BitsPerComponent", x.BitsPerComponent},
		}
		if x.Alpha != nil {
			x.stream.extras = append(x.stream.extras, field{"/SMask", iref(x.Alpha)})
		}
	}
	return x.stream.encode(w)
}
