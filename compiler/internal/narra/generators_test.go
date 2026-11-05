package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generatorsDir holds the types dump goldens of the generator checks
// (node generators-check): <case>.nra with <case>.dump. TestSelfhostBodies
// compares the Narra checker with them, as it does every testdata file.
var generatorsDir = filepath.Join(testdata, "typesdump", "generators")

// TestTypesDumpGenerators compares the goldens in testdata/typesdump/generators
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpGenerators(t *testing.T) {
	cases := declDumpCases(t, generatorsDir)
	if len(cases) == 0 {
		t.Fatal("no generator cases")
	}
	for _, path := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got, ok := TypesDump(path)
			if strings.HasPrefix(filepath.Base(path), "err_") == ok {
				t.Fatalf("a case named err_ must fail and no other may (ok=%v):\n%s", ok, got)
			}
			golden := strings.TrimSuffix(path, ".nra") + ".dump"
			if os.Getenv("NARRA_UPDATE") != "" {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("dump differs: %s", firstDiff(string(want), got))
			}
		})
	}
}
