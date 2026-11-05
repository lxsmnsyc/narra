package narra

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSelfhostFmtExprsStable prints every expression of every .nra file under
// testdata/ and selfhost/src at a narrow and a wide width, with and without
// trailing commas, reads the text back and prints it again. The driver's
// --check mode fails when the text does not parse or changes.
func TestSelfhostFmtExprsStable(t *testing.T) {
	exe := buildFmtExprs(t)
	var inputs []string
	for _, root := range []string{"../../testdata", "../../../selfhost/src"} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && strings.HasSuffix(path, ".nra") {
				inputs = append(inputs, path)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(inputs) == 0 {
		t.Fatal("no inputs")
	}
	for _, input := range inputs {
		name := strings.TrimPrefix(strings.ReplaceAll(input, "../", ""), "/")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, args := range [][]string{{"30", "all"}, {"100", "none"}} {
				out, err := exec.Command(exe, "--check", args[0], args[1], input).CombinedOutput()
				if err != nil {
					t.Fatalf("width %s trailing %s: %v\n%s", args[0], args[1], err, out)
				}
			}
		})
	}
}
