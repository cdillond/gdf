package text

import (
	"github.com/cdillond/gdf"
)

// NewFontFamily provides a convenience wrapper around gdf.LoadSFNTFile. The arguments are paths identifying the locations
// of the font files for the members of the FontFamily. If an error is generated when loading any of the fonts in the
// family, all fonts in the FontFamily returned by NewFontFamily will be nil.
func NewFontFamily(regular, bold, italic, boldItal string) (FontFamily, error) {
	var r, b, i, bi *gdf.Font
	var err error
	if r, err = gdf.LoadSFNTFile(regular, gdf.Nonsymbolic); err != nil {
		return FontFamily{}, err
	}
	if b, err = gdf.LoadSFNTFile(regular, gdf.Nonsymbolic); err != nil {
		return FontFamily{}, err
	}
	if i, err = gdf.LoadSFNTFile(regular, gdf.Nonsymbolic|gdf.Italic); err != nil {
		return FontFamily{}, err
	}
	if bi, err = gdf.LoadSFNTFile(regular, gdf.Nonsymbolic|gdf.Italic); err != nil {
		return FontFamily{}, err
	}
	return FontFamily{
		Regular:  r,
		Bold:     b,
		Ital:     i,
		BoldItal: bi,
	}, nil
}
