package narra

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// patternsDir holds the goldens of the patterns node: <case>.nra or a
// package directory <case>/, with <case>.dump.
var patternsDir = filepath.Join(testdata, "typesdump", "patterns")

// TestTypesDumpPatterns compares the goldens in testdata/typesdump/patterns
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpPatterns(t *testing.T) {
	cases := declDumpCases(t, patternsDir)
	if len(cases) == 0 {
		t.Fatal("no patterns cases")
	}
	for _, path := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got, ok := TypesDump(path)
			if !ok && !strings.HasPrefix(got, "error ") {
				t.Fatalf("a failed dump must start with `error `: %q", got)
			}
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

// The goldens print the pattern lines, bindings and narrowing.
func TestTypesDumpPatternsCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, patternsDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, want := range []string{
		" pat-wildcard ", " pat-name ", " pat-ctor ", " pat-struct ", " pat-tuple ", " pat-lit ",
		" ctor=Shape.Circle ", " narrow=Circle ", " narrow=Square ", " is ", " match ", " map ",
		" local radius ", " local a mutable ", "\nerror ",
		"match does not cover",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no golden prints %q", want)
		}
	}
}

// patternsInputs are the cases of testdata/typesdump/patterns. The earlier
// inputs are compared by TestSelfhostBodies.
func patternsInputs(t *testing.T) []string {
	return declDumpCases(t, patternsDir)
}

// TestSelfhostPatterns checks that `narrac types <path>` prints the same
// dump as `narra types`, or the same errors on stderr with exit 1, on every
// input of patternsInputs. As in TestSelfhostBodies, an input is skipped
// only when it does not load or the Narra checker reports a construct a
// later node ports (`unsupported:`).
func TestSelfhostPatterns(t *testing.T) {
	exe := buildSelfhost(t)
	var mu sync.Mutex
	compared := 0
	skipped := map[string]int{}
	skip := func(t *testing.T, reason string) {
		mu.Lock()
		skipped[reason]++
		mu.Unlock()
		t.Skip(reason)
	}
	t.Run("inputs", func(t *testing.T) {
		for _, path := range patternsInputs(t) {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				if _, err := Load(path); err != nil {
					skip(t, "does not load")
				}
				want, ok := TypesDump(path)
				out, errOut, code := runNarracRaw(t, exe, "types", path)
				if construct := unsupportedConstruct(errOut); construct != "" {
					skip(t, "unsupported: "+construct)
				}
				mu.Lock()
				compared++
				mu.Unlock()
				if !ok {
					if code != 1 || out != "" {
						t.Fatalf("exit code %d, want 1; stdout:\n%s\nstderr:\n%s", code, out, errOut)
					}
					if errOut != want {
						t.Fatalf("errors differ: %s", firstDiff(want, errOut))
					}
					return
				}
				if code != 0 {
					t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
				}
				if want != out {
					t.Fatalf("dump differs: %s", firstDiff(want, out))
				}
			})
		}
	})
	total := 0
	var reasons []string
	for r, n := range skipped {
		total += n
		reasons = append(reasons, r)
	}
	sort.Slice(reasons, func(i, j int) bool {
		if skipped[reasons[i]] != skipped[reasons[j]] {
			return skipped[reasons[i]] > skipped[reasons[j]]
		}
		return reasons[i] < reasons[j]
	})
	t.Logf("compared %d, skipped %d", compared, total)
	for _, r := range reasons {
		t.Logf("  %4d %s", skipped[r], r)
	}
}
