package narra

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// typesDumpCallsDir holds the call cases of the types dump (constructors,
// generic and trait calls, distinct and module calls): <case>.nra or a
// package directory <case>/, with <case>.dump.
var typesDumpCallsDir = filepath.Join(testdata, "typesdump", "calls")

// TestTypesDumpCalls compares the goldens in testdata/typesdump/calls with
// the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpCalls(t *testing.T) {
	cases := declDumpCases(t, typesDumpCallsDir)
	if len(cases) == 0 {
		t.Fatal("no typesdump/calls cases")
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

// The call goldens print every attribute and header this node owns.
func TestTypesDumpCallsCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, typesDumpCallsDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, want := range []string{
		" impl ", " targs=", " ctor=", " trait=", " distinct=", " func=",
		" ident module ", "\nerror ",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no call golden prints %q", want)
		}
	}
}

// The instance bodies come in discovery order: ping and pong call each
// other's instances, so pong<int> follows ping<int>, and ping<string>
// follows pong<string>, after the instances run discovers first.
func TestTypesDumpInstanceOrder(t *testing.T) {
	got, ok := TypesDump(filepath.Join(typesDumpCallsDir, "generics.nra"))
	if !ok {
		t.Fatalf("dump failed:\n%s", got)
	}
	var headers []string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "body ") && strings.Contains(line, " targs=") && (strings.Contains(line, " ping ") || strings.Contains(line, " pong ")) {
			headers = append(headers, strings.TrimPrefix(line, `body module="generics" func `))
		}
	}
	want := []string{"ping targs=int", "pong targs=string", "pong targs=int", "ping targs=string"}
	if strings.Join(headers, "|") != strings.Join(want, "|") {
		t.Fatalf("instance order %q, want %q", headers, want)
	}
}

// callsInputs are the inputs of TestSelfhostBodies plus the cases of
// testdata/typesdump/calls, once each. The bodies inputs already hold the
// single-file cases and the files of the package cases.
func callsInputs(t *testing.T) []string {
	inputs := bodiesInputs(t)
	seen := map[string]bool{}
	for _, p := range inputs {
		seen[p] = true
	}
	for _, p := range declDumpCases(t, typesDumpCallsDir) {
		if !seen[p] {
			inputs = append(inputs, p)
		}
	}
	return inputs
}

// TestSelfhostCalls checks that `narrac types <path>` prints the same dump
// as `narra types`, or the same errors with exit 1, on callsInputs. As in
// TestSelfhostBodies, an input is skipped only when it does not load or
// when narrac reports a construct a later node ports (`unsupported:`), and
// the counts per reason are printed at the end. The call cases must all be
// compared.
func TestSelfhostCalls(t *testing.T) {
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
		for _, path := range callsInputs(t) {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				own := strings.HasPrefix(path, typesDumpCallsDir)
				if _, err := Load(path); err != nil {
					skip(t, "does not load")
				}
				want, ok := TypesDump(path)
				out, errOut, code := runNarracRaw(t, exe, "types", path)
				if construct := unsupportedConstruct(errOut); construct != "" {
					if own {
						t.Fatalf("a call case must be compared, found unsupported: %s", construct)
					}
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
