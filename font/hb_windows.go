//go:build windows

package font

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"golang.org/x/image/font/sfnt"
)

// HBSubsetPath returns a func that can be used as a gdf.FontSubsetFunc. The returned func
// is similar to HBSubset and has the same limitations. The path parameter represents the
// path of the source font.
func HBSubsetPath(path string) func(*sfnt.Font, []byte, map[rune]struct{}) ([]byte, error) {
	return func(_ *sfnt.Font, _ []byte, cutset map[rune]struct{}) ([]byte, error) {
		var out *os.File
		var err error
		outPath := "tmp-font-out"

		// make sure we're not overwriting anything
		var i int
		for _, err = os.Stat(outPath); err == nil; i++ {
			outPath += strconv.Itoa(i)
		}
		// set up output file
		if out, err = os.Create(outPath); err != nil {
			return nil, err
		}
		defer os.Remove(outPath)
		if err = out.Close(); err != nil {
			return nil, err
		}

		u := make([]byte, 0, 512)
		for key := range cutset {
			u = strconv.AppendInt(u, int64(key), 16)
			u = append(u, ',')
		}
		if len(u) < 1 {
			return nil, fmt.Errorf("cutset is too small")
		}
		cmd := exec.Command("hb-subset",
			"--font-file="+path, // must be passed explicitly as an arg
			"-u", string(u[:len(u)-1]),
			"--retain-gids",
			"-o", outPath, // ditto for output
		)
		if err = cmd.Run(); err != nil {
			return nil, err
		}
		return os.ReadFile(outPath)
	}
}

// HBSubset can be used as a gdf.FontSubsetFunc.
// For this function to work, the HarfBuzz hb-subset tool must be installed.
// The HBSubset func may handle edge cases that the TTFSubset func does not. hb-subset
// has a mature, well-tested API and is capable of handling more font formats than TTFSubset.
// However, this approach requires os/exec, so it might not be suitable for all environments.
func HBSubset(_ *sfnt.Font, src []byte, cutset map[rune]struct{}) ([]byte, error) {
	// Instead of using /dev/stdin and /dev/stdout, on Windows, this function creates temp files.
	// It's unclear if this is necessary, but the hb-subset tool is finnicky.
	var in, out *os.File
	var err error

	var (
		inPath  = "tmp-font-in"
		outPath = "tmp-font-out"
	)
	// make sure we're not overwriting anything
	var i int
	for _, err = os.Stat(inPath); err == nil; i++ {
		inPath += strconv.Itoa(i)
	}
	i = 0
	for _, err = os.Stat(outPath); err == nil; i++ {
		outPath += strconv.Itoa(i)
	}

	// set up input file; this would be faster if it were possible to pass the font file path instead.
	if in, err = os.Create(inPath); err != nil {
		return nil, err
	}
	defer os.Remove(inPath)
	if _, err = in.Write(src); err != nil {
		in.Close()
		return nil, err
	}
	if err = in.Close(); err != nil {
		return nil, err
	}
	// set up output file
	if out, err = os.Create(outPath); err != nil {
		return nil, err
	}
	defer os.Remove(outPath)
	if err = out.Close(); err != nil {
		return nil, err
	}

	u := make([]byte, 0, 512)
	for key := range cutset {
		u = strconv.AppendInt(u, int64(key), 16)
		u = append(u, ',')
	}
	if len(u) < 1 {
		return nil, fmt.Errorf("cutset is too small")
	}

	cmd := exec.Command("hb-subset",
		"--font-file="+inPath, // must be passed explicitly as an arg
		"-u", string(u[:len(u)-1]),
		"--retain-gids",
		"-o", outPath, // ditto for output
	)
	if err = cmd.Run(); err != nil {
		return nil, err
	}
	return os.ReadFile(outPath)
}
