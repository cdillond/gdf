package subset

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HBSubset returns the source bytes of a font containing only the characters included in the cutset.
// The path parameter represents the path of the source font. This function has the same limitations as HBSubset.
func HBSubsetPath(path string, cutset map[rune]struct{}) ([]byte, error) {
	if len(cutset) == 0 {
		return nil, fmt.Errorf("cutset is too small")
	}
	bldr := new(strings.Builder)
	bldr.Grow(512)
	for r := range cutset {
		bldr.WriteRune(r)
	}
	cmd := exec.Command("hb-subset",
		"--font-file="+path, // must be passed explicitly as an arg
		"-t", bldr.String(),
		"--retain-gids",
		"-o", os.Stdout.Name(), // ditto for stdout
	)
	return cmd.Output()
}

// HBSubset returns the source bytes of a font containing only the characters included in the cutset.
// For this function to work, the HarfBuzz hb-subsettool must be installed.
// The HBSubset func may handle edge cases that the TTFSubset func does not. hb-subset
// has a mature, well-tested API and is capable of handling more font formats than TTFSubset.
// However, this approach requires os/exec, so it might not be suitable for all environments.
func HBSubset(src []byte, cutset map[rune]struct{}) ([]byte, error) {
	if len(cutset) == 0 {
		return nil, fmt.Errorf("cutset is too small")
	}
	bldr := new(strings.Builder)
	bldr.Grow(512)
	for r := range cutset {
		bldr.WriteRune(r)
	}

	cmd := exec.Command("hb-subset",
		"--font-file="+os.Stdin.Name(), // must be passed explicitly as an arg
		"-t", bldr.String(),
		"--retain-gids",
		"-o", os.Stdout.Name(), // ditto for stdout
	)
	cmd.Stdin = bytes.NewReader(src)
	return cmd.Output()
}
