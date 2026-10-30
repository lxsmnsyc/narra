package narra

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestLiteralTypeArgsOneError checks that each mistake in
// testdata/errors/literal_type_args_*.nra gives exactly one diagnostic, not a
// follow-up such as "found void".
func TestLiteralTypeArgsOneError(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(testdata, "errors", "literal_type_args_*.nra"))
	if len(files) == 0 {
		t.Fatal("no literal_type_args error tests")
	}
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			_, err := CompileToC(src, false)
			if err == nil {
				t.Fatal("expected an error")
			}
			lines := strings.Split(strings.TrimSpace(err.Error()), "\n")
			if len(lines) != 1 {
				t.Fatalf("want one diagnostic, got %d:\n%v", len(lines), err)
			}
		})
	}
}
