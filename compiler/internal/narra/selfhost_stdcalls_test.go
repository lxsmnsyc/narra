package narra

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// stdcallsDir holds the goldens of the builtins node: <case>.nra with
// <case>.dump.
var stdcallsDir = filepath.Join(testdata, "typesdump", "stdcalls")

// TestTypesDumpStdCalls compares the goldens in testdata/typesdump/stdcalls
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpStdCalls(t *testing.T) {
	cases := declDumpCases(t, stdcallsDir)
	if len(cases) == 0 {
		t.Fatal("no stdcalls cases")
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

// Every check name of the registry has a call in the goldens.
func TestTypesDumpStdCallsCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, stdcallsDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, b := range builtinsByName {
		generic := false
		for _, prefix := range []string{"Array.", "Map.", "Box.", "Range.", "Option.", "Result.", "Test."} {
			generic = generic || strings.HasPrefix(b.Name, prefix)
		}
		if generic && !strings.Contains(text, "builtin="+b.Name+" ") {
			t.Errorf("no golden calls %s", b.Name)
		}
	}
	for _, want := range []string{
		"can only be called inside a `test` block", "needs a function that returns a value",
		"needs a function that returns Option", "needs an Option<T>", "needs a Result<T, E>",
		"changes the array", "changes the map", "is passed as mutable in this call",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no golden prints %q", want)
		}
	}
}

// stdcallsInputs are the cases of testdata/typesdump/stdcalls and the
// closure samples, which use the Array builtins (moved here from the
// closures node), plus every input of TestSelfhostBodies.
func stdcallsInputs(t *testing.T) []string {
	inputs := declDumpCases(t, stdcallsDir)
	inputs = append(inputs, closuresInputs(t)...)
	return append(inputs, bodiesInputs(t)...)
}

// TestSelfhostStdCalls checks that `narrac types <path>` prints the same dump
// as `narra types`, or the same errors on stderr with exit 1, on every input
// of stdcallsInputs. A std call may never be skipped; an input is skipped
// only when it does not load or reports a construct a later node ports.
func TestSelfhostStdCalls(t *testing.T) {
	exe := buildSelfhost(t)
	var mu sync.Mutex
	compared := 0
	skipped := map[string]int{}
	skip := func(t *testing.T, reason string) {
		if strings.Contains(t.Name(), "typesdump/stdcalls/") {
			t.Fatalf("a std call case may not be skipped: %s", reason)
		}
		if strings.Contains(reason, "builtin") {
			t.Fatalf("a std call may not be skipped: %s", reason)
		}
		mu.Lock()
		skipped[reason]++
		mu.Unlock()
		t.Skip(reason)
	}
	t.Run("inputs", func(t *testing.T) {
		for _, path := range stdcallsInputs(t) {
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
