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
	// We can get away with not copying b, since bytes.Reader leaves b unaltered, but the io.Reader
	// interface explicitly states that b may be used as a scratch space by certain implementations.
	r := bytes.NewReader(b)
	cfg, err := std.DecodeConfig(r)
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
		r.Reset(b)
		// CMYK and Gray color models should be re-encoded to avoid errors.
		return slowPath(r, x)
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

func slowPath(r *bytes.Reader, x gdf.XImage) (gdf.XImage, error) {
	img, err := std.Decode(r)
	if err != nil {
		return *new(gdf.XImage), err
	}
	return imageToXImg(img, x)
}
