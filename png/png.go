package png

import (
	"bytes"
	"encoding/binary"
	"image"
	std "image/png"
	"os"

	"github.com/cdillond/gdf"
)

// DecodeFile reads the contents of the file at the provided path and then calls Decode.
func DecodeFile(path string) (gdf.XImage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return *new(gdf.XImage), err
	}
	return Decode(b)
}

// Decode interprets b as data representing a PNG image. It decodes the image and returns
// a *gdf.XImage and an error. This function may not be ideal for all varieties of PNG.
// In particular, grayscale images with alpha channels are converted to their NRGBA equivalents,
// which may have the effect of significantly increasing the image's encoding size. Applications
// sensitive to performance may benefit from processing the image data separately and then
// generating the image's XImage representation by way of the gdf.LoadXImage function.
func Decode(b []byte) (gdf.XImage, error) {
	// See http://www.libpng.org/pub/png/spec/1.2/png-1.2-pdg.html for details.
	cfg, err := std.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return *new(gdf.XImage), err
	}

	x := gdf.XImage{
		Height:     cfg.Height,
		Width:      cfg.Width,
		ColorSpace: gdf.DeviceGray,
	}
	img, err := std.Decode(bytes.NewReader(b))
	if err != nil {
		return *new(gdf.XImage), err
	}

	var colors, alpha []byte
	box := img.Bounds()

	csize := 8
	var p, q int
	var hasAlpha bool
	switch v := img.(type) {
	case *image.NRGBA:
		colors = make([]byte, 3*(cfg.Height)*(cfg.Width))
		alpha = make([]byte, (cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.NRGBAAt(x, y)
				if px.A != 0 {
					hasAlpha = true
				}
				colors[p] = px.R
				colors[p+1] = px.G
				colors[p+2] = px.B
				p += 3
				alpha[q] = px.A
				q++
			}
		}
	case *image.NRGBA64:
		csize = 16
		colors = make([]byte, 6*(cfg.Height)*(cfg.Width))
		alpha = make([]byte, 2*(cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.NRGBA64At(x, y)
				if px.A != 0 {
					hasAlpha = true
				}
				binary.BigEndian.PutUint16(colors[p:p+2], px.R)
				binary.BigEndian.PutUint16(colors[p+2:p+4], px.G)
				binary.BigEndian.PutUint16(colors[p+4:p+6], px.B)
				p += 6
				binary.BigEndian.PutUint16(alpha[q:q+2], px.A)

			}
		}
	case *image.RGBA:
		// PNG does not use premultiplied alpha; we therefore assume that if
		// an image is of type image.RGBA or image.RGBA64, the alpha channel contains no data.
		colors = make([]byte, 3*(cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.RGBAAt(x, y)
				colors[p] = px.R
				colors[p+1] = px.G
				colors[p+2] = px.B
				p += 3
			}
		}
	case *image.RGBA64:
		csize = 16
		colors = make([]byte, 6*(cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.RGBA64At(x, y)
				binary.BigEndian.PutUint16(colors[p:p+2], px.R)
				binary.BigEndian.PutUint16(colors[p+2:p+4], px.G)
				binary.BigEndian.PutUint16(colors[p+4:p+6], px.B)
				p += 6
			}
		}
	case *image.Gray:
		// Technically, grayscale PNG images can also include alpha channels and/or
		// have "bit depths" of less than 8. While image/png can read these files,
		// go's standard library image package doesn't export these image types.
		colors = make([]byte, (cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.GrayAt(x, y)
				colors[p] = px.Y
				p++
			}
		}
	case *image.Gray16:
		csize = 16
		colors = make([]byte, 2*(cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.Gray16At(x, y)
				binary.BigEndian.PutUint16(colors[p:p+2], px.Y)
				p += 2
			}
		}

	}
	x.BitsPerComponent = csize
	if hasAlpha {
		mask := x
		mask.Data = alpha
		x.Alpha = &mask
	}

	x.ColorSpace = gdf.DeviceRGB
	x.Data = colors
	return x, nil
}
