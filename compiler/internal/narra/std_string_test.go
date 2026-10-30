package narra

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStringPanics(t *testing.T) {
	cases := map[string]string{
		"strings2_inside.nra": "position 2 is inside a character",
		"strings2_range.nra":  "position 9 out of range for length 3",
	}
	for file, msg := range cases {
		t.Run(file, func(t *testing.T) {
			_, stderr, code := runProgram(t, filepath.Join(testdata, "panics", file), modes[0].opts)
			if code != 101 || !strings.Contains(stderr, msg) {
				t.Fatalf("want panic %q with exit 101, got exit %d\nstderr:\n%s", msg, code, stderr)
			}
		})
	}
}
