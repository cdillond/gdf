//go:build !cgo || !hbsubsetc

package subset

import "fmt"

// HBSubsetC calls functions in libharfbuzz and libharfbuzz-subset via CGo and returns the source bytes of a font containing only the
// characters included in the cutset. In order for this function to work, CGo must be enabled, HarfBuzz v>=2.9.0 must be installed on
// your system, and `hbsubsetc` must be passed to the Go compiler as a build tag.
func HBSubsetC(src []byte, charset map[rune]struct{}) ([]byte, error) {
	return nil, fmt.Errorf("HBSubetC is not enabled; use the build tag hbsubsetc and consult github.com/cdillond/gdf/font for further information")
}
