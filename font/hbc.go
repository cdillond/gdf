//go:build !cgo || !hbsubsetc

package font

import (
	"golang.org/x/image/font/sfnt"
)

// HBSubsetC can be used as a gdf.FontSubsetFunc. It calls functions in libharfbuzz and libharfbuzz-subset via CGo. In order
// for this function to work, CGo must be enabled, HarfBuzz v>=2.9.0 must be installed on your system, and `hbsubsetc` must be passed
// as a build tag to the Go compiler. If these conditions are not met, the function instead calls HBSubset (which may fail if hb-subset)
// has not been installed.
func HBSubsetC(_ *sfnt.Font, src []byte, charset map[rune]struct{}) ([]byte, error) {
	return HBSubset(nil, src, charset)
}
