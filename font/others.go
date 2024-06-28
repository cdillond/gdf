//go:build !windows

package font

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"

	"golang.org/x/image/font/sfnt"
)

// HBSubset returns a func that can be used as a gdf.FontSubsetFunc on systems with /dev/stdin
// and /dev/stdout device files. For this function to work, the HarfBuzz hb-subset tool must
// be installed. The HBSubset func may handle edge cases that the TTFSubset func does not. hb-subset
// has a mature, well-tested API and is capable of handling more font formats than TTFSubset.
// However, this approach requires os/exec, so it might not be suitable for all environments.
func HBSubset(_ *sfnt.Font, src []byte, cutset map[rune]struct{}) ([]byte, error) {
	u := make([]byte, 0, 512)
	for key := range cutset {
		u = strconv.AppendInt(u, int64(key), 16)
		u = append(u, ',')
	}
	if len(u) < 1 {
		return nil, fmt.Errorf("cutset is too small")
	}
	cmd := exec.Command("hb-subset",
		"--font-file=/dev/stdin", // must be passed explicitly as an arg
		"-u", string(u[:len(u)-1]),
		"--retain-gids",
		"-o", "/dev/stdout", // ditto for stdout
	)
	cmd.Stdin = bytes.NewReader(src)
	return cmd.Output()
}
