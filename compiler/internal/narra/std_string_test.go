package narra

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStringPanics(t *testing.T) {
	cases := map[string]string{
		"strings2_inside.nra":        "position 2 is inside a character",
		"strings2_range.nra":         "position 9 out of range for length 3",
		"strings2_charat_end.nra":    "Str.charAt: position 2 is at the end of the string",
		"strings2_nextchar_past.nra": "Str.nextChar: position 5 out of range for length 2",
		"strings2_byteat_range.nra":  "Str.byteAt: position 2 out of range for length 2",
		"strings2_split_empty.nra":   "Str.split with an empty separator",
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
