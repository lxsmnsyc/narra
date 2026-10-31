package narra

import (
	"sort"
	"strings"
	"sync"
	"testing"
)

// bodiesInputs are the inputs of TestSelfhostImports plus the package
// cases of testdata/typesdump/bodies (its single files are already in
// declInputs).
func bodiesInputs(t *testing.T) []string {
	inputs := importInputs(t)
	for _, p := range declDumpCases(t, typesDumpDir) {
		if !strings.HasSuffix(p, ".nra") {
			inputs = append(inputs, p)
		}
	}
	return inputs
}

// unsupportedConstruct returns the construct of the first `unsupported:`
// line of a narrac output, or "".
func unsupportedConstruct(out string) string {
	_, rest, ok := strings.Cut(out, "unsupported: ")
	if !ok {
		return ""
	}
	line, _, _ := strings.Cut(rest, "\n")
	return line
}

// TestSelfhostBodies checks that `narrac types <path>` prints the same dump
// as `narra types` (TypesDump), or the same errors on stderr with exit 1,
// on every input of bodiesInputs. An input is skipped only when it does not
// load or when the Narra checker reports a construct a later node ports
// (`unsupported:`); each skip is logged with its construct, and the counts
// per reason are printed at the end.
func TestSelfhostBodies(t *testing.T) {
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
		for _, path := range bodiesInputs(t) {
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

func TestUnsupportedConstruct(t *testing.T) {
	out := "error a.nra:1:2: unknown name \"x\"\nerror a.nra:3:4: unsupported: `is` (selfhost-check-patterns)\n"
	if got := unsupportedConstruct(out); got != "`is` (selfhost-check-patterns)" {
		t.Fatalf("got %q", got)
	}
	if got := unsupportedConstruct("error a.nra:1:2: unknown name \"x\"\n"); got != "" {
		t.Fatalf("got %q", got)
	}
}
