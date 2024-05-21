package raster

/*
Package raster provides functions for decoding raster images to gdf.XImage objects.
To include JPEG images in a PDF, use the github.com/cdillond/jpeg package; it is
likely to be significantly more efficient.

Example:
```go
import (
	"image/png"

	"github.com/cdillond/gdf/raster"
)

func main() {
	pngDec := raster.NewDecoder(png.Decode, png.DecodeConfig)
	ximg, err := pngDec.DecodeFile("/path/to/image.png")
	...
	etc
}
```
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"github.com/cdillond/gdf"
)

// A DecodeFunc reads an image fom the source io.Reader and returns an image.Image and an error.
// Each of the standard library's image package subdirectories contain functions that can
// be used here, although alternative implementations are fine too.
type DecodeFunc func(io.Reader) (image.Image, error)

// A DecodeConfigFunc reads from a source io.Reader and returns an image.Config and an error.
type DecodeConfigFunc func(io.Reader) (image.Config, error)

// A Decoder is a struct that can read gdf.XImages from source bytes. Most image
// codecs designed to be used with the standard library's image.RegisterFormat() function
// export functions that are useable as DecodeFunc and DecodeConfigFuncs. A decoder can be
// reused on multiple images of the same format, although it is not concurrency safe.
// NOTE: For JPEG images, the gdf/jpeg package is more efficient.
type Decoder struct {
	DecodeFunc
	DecodeConfigFunc
}

func NewDecoder(df DecodeFunc, dcf DecodeConfigFunc) Decoder {
	return Decoder{DecodeFunc: df, DecodeConfigFunc: dcf}
}

var (
	ErrDec    = fmt.Errorf("improperly initialized Decoder")
	ErrFormat = fmt.Errorf("unsupported image color model")
)

// DecodeFile reads the contents of the file at the provided path and then calls Decode.
func (d Decoder) DecodeFile(path string) (gdf.XImage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return *new(gdf.XImage), err
	}
	return d.Decode(b)
}

type rgba64 interface {
	RGBA64At(int, int) color.RGBA64
}

// The standard library's image package treats RGBA64 as a sort of lingua franca for images.
// It's not ideal, especially for images with 8-bit depth, but it does the trick for images with formats
// that don't cleanly map to PDF image headers.
func processImage(img rgba64, box image.Rectangle, cfg image.Config) (colors, alpha []byte, hasAlpha bool) {
	var p, q int
	colors = make([]byte, 6*(cfg.Height)*(cfg.Width))
	alpha = make([]byte, 2*(cfg.Height)*(cfg.Width))
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			px := img.RGBA64At(x, y)
			binary.BigEndian.PutUint16(colors[p:p+2], px.R)
			binary.BigEndian.PutUint16(colors[p+2:p+4], px.G)
			binary.BigEndian.PutUint16(colors[p+4:p+6], px.B)
			p += 6
			if px.A != 0 {
				hasAlpha = true
				binary.BigEndian.PutUint16(alpha[q:q+2], px.A)
			}
			q += 2
		}
	}
	return colors, alpha, hasAlpha
}

// Decode interprets b as data representing a PNG image. It decodes the image and returns
// a gdf.XImage and an error. This function may not be ideal for all varieties of PNG.
// In particular, grayscale images with alpha channels are converted to their NRGBA equivalents,
// which may have the effect of significantly increasing the image's encoding size. Applications
// sensitive to performance may benefit from processing the image data separately and then
// generating the image's XImage representation by way of the gdf.LoadXImage function.
func (d Decoder) Decode(b []byte) (gdf.XImage, error) {
	if d.DecodeFunc == nil || d.DecodeConfigFunc == nil {
		return *new(gdf.XImage), ErrDec
	}

	cfg, err := d.DecodeConfigFunc(bytes.NewReader(b))
	if err != nil {
		return *new(gdf.XImage), err
	}

	x := gdf.XImage{
		Height: cfg.Height,
		Width:  cfg.Width,
	}
	img, err := d.DecodeFunc(bytes.NewReader(b))
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
		colors = make([]byte, 3*(cfg.Height)*(cfg.Width))
		alpha = make([]byte, (cfg.Height)*(cfg.Width))
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				px := v.RGBAAt(x, y)
				colors[p] = px.R
				colors[p+1] = px.G
				colors[p+2] = px.B
				p += 3
				if px.A != 0 {
					hasAlpha = true
					alpha[q] = px.A
				}
				q++
			}
		}

	case *image.Gray:
		// Technically, grayscale PNG images can also include alpha channels and/or
		// have "bit depths" of less than 8. While image/png can read these files,
		// Go's standard library image package doesn't export these image types.
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
	case *image.RGBA64:
		csize = 16
		colors, alpha, hasAlpha = processImage(v, box, cfg)
	case *image.CMYK:
		csize = 16
		colors, alpha, hasAlpha = processImage(v, box, cfg)
	case *image.YCbCr:
		csize = 16
		colors, alpha, hasAlpha = processImage(v, box, cfg)
	case *image.NYCbCrA:
		csize = 16
		colors, alpha, hasAlpha = processImage(v, box, cfg)
	default:
		return x, ErrFormat
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
