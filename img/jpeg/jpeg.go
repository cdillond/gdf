package jpeg

import (
	"bytes"
	"image"
	"image/color"
	std "image/jpeg"
	"os"

	"github.com/cdillond/gdf"
)

// Decode interprets b as data representing a JPEG image, as specified by ISO/IEC 10918. It
// returns a *gdf.XObject and an error. If the error is nil, the XObject can be drawn to a
// gdf.ContentStream.
func Decode(b []byte) (*gdf.XObject, error) {
	cfg, err := std.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	d := gdf.ImageDict{
		Height:           cfg.Height,
		Width:            cfg.Width,
		BitsPerComponent: 8,
		Format:           gdf.JPEG,
	}

	switch cfg.ColorModel {
	case color.YCbCrModel, color.RGBAModel:
		d.ColorSpace = gdf.DeviceRGB
	default:
		// CMYK and Gray color models should be re-encoded to avoid errors.
		return slowPath(b, d)
	}
	return gdf.NewImageXObj(b, d), nil
}

// DecodeFile reads the contents of the file at the specified path and then calls Decode.
func DecodeFile(path string) (*gdf.XObject, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(b)
}

func imageToXObj(img image.Image, d gdf.ImageDict) (*gdf.XObject, error) {
	switch img.ColorModel() {
	case color.CMYKModel:
		d.ColorSpace = gdf.DeviceCMYK
	case color.GrayModel, color.Gray16Model:
		d.ColorSpace = gdf.DeviceGray
	default:
		d.ColorSpace = gdf.DeviceRGB
	}
	w := new(bytes.Buffer)

	if err := std.Encode(w, img, &std.Options{Quality: 100}); err != nil {
		return nil, err
	}
	x := gdf.NewImageXObj(w.Bytes(), d)
	return x, nil
}

func slowPath(b []byte, d gdf.ImageDict) (*gdf.XObject, error) {
	img, err := std.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return imageToXObj(img, d)
}
