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
// returns a gdf.XImage and an error.
func Decode(b []byte) (gdf.XImage, error) {
	cfg, err := std.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return *new(gdf.XImage), err
	}
	x := gdf.XImage{
		Height:           cfg.Height,
		Width:            cfg.Width,
		BitsPerComponent: 8,
		AppliedFilter:    gdf.DCTDecode,
	}

	switch cfg.ColorModel {
	case color.YCbCrModel, color.RGBAModel:
		x.ColorSpace = gdf.DeviceRGB
	default:
		// CMYK and Gray color models should be re-encoded to avoid errors.
		return slowPath(b, x)
	}
	x.Data = b
	return x, nil
}

// DecodeFile reads the contents of the file at the specified path and then calls Decode.
func DecodeFile(path string) (gdf.XImage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return *new(gdf.XImage), err
	}
	return Decode(b)
}

func imageToXImg(img image.Image, x gdf.XImage) (gdf.XImage, error) {
	switch img.ColorModel() {
	case color.CMYKModel:
		x.ColorSpace = gdf.DeviceCMYK
	case color.GrayModel, color.Gray16Model:
		x.ColorSpace = gdf.DeviceGray
	default:
		x.ColorSpace = gdf.DeviceRGB
	}
	w := new(bytes.Buffer)

	if err := std.Encode(w, img, &std.Options{Quality: 100}); err != nil {
		return *new(gdf.XImage), err
	}
	x.Data = w.Bytes()
	return x, nil
}

func slowPath(b []byte, x gdf.XImage) (gdf.XImage, error) {
	img, err := std.Decode(bytes.NewReader(b))
	if err != nil {
		return *new(gdf.XImage), err
	}
	return imageToXImg(img, x)
}
