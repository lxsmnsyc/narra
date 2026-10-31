package narra

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// closuresDir holds the goldens of the closures node: <case>.nra or a
// package directory <case>/, with <case>.dump.
var closuresDir = filepath.Join(testdata, "typesdump", "closures")

// TestTypesDumpClosures compares the goldens in testdata/typesdump/closures
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpClosures(t *testing.T) {
	cases := declDumpCases(t, closuresDir)
	if len(cases) == 0 {
		t.Fatal("no closures cases")
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

// The goldens print the stack and captures attributes, value calls and
// function values.
func TestTypesDumpClosuresCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, closuresDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, want := range []string{
		" arrow stack ", " arrow captures=rate ", " arrow stack captures=rate ", " arrow captures=a,rate ",
		" call value ", " func=arrows.double ", " func=p/ops.same ", " func=p/ops.double ", " const=p/ops.triple ",
		" local x ", "\nerror ",
		"parameter \"f\" cannot escape", "a closure cannot capture", "cannot tell the type of parameter",
		"cannot also hold the function being called",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no golden prints %q", want)
		}
	}
}

// closuresInputs are the cases of testdata/typesdump/closures and the
// closure samples and error cases, which TestSelfhostBodies also reaches.
// The cases of the closures folder may never be skipped.
func closuresInputs(t *testing.T) []string {
	inputs := declDumpCases(t, closuresDir)
	for _, pattern := range []string{"closures*.nra", filepath.Join("errors", "closures*.nra")} {
		matches, err := filepath.Glob(filepath.Join(testdata, pattern))
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, matches...)
	}
	return inputs
}

// TestSelfhostClosures checks that `narrac types <path>` prints the same
// dump as `narra types`, or the same errors on stderr with exit 1, on every
// input of closuresInputs. As in TestSelfhostBodies, an input is skipped
// only when it does not load or the Narra checker reports a construct a
// later node ports (`unsupported:`).
func TestSelfhostClosures(t *testing.T) {
	exe := buildSelfhost(t)
	var mu sync.Mutex
	compared := 0
	skipped := map[string]int{}
	skip := func(t *testing.T, reason string) {
		if strings.Contains(t.Name(), "typesdump/closures/") {
			t.Fatalf("a closures case may not be skipped: %s", reason)
		}
		mu.Lock()
		skipped[reason]++
		mu.Unlock()
		t.Skip(reason)
	}
	t.Run("inputs", func(t *testing.T) {
		for _, path := range closuresInputs(t) {
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
