package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// effectsDir holds the goldens of the effects node: <case>.nra with
// <case>.dump. A case named err_ must fail.
var effectsDir = filepath.Join(testdata, "typesdump", "effects")

// TestTypesDumpEffects compares the goldens in testdata/typesdump/effects
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDumpEffects(t *testing.T) {
	cases := declDumpCases(t, effectsDir)
	if len(cases) == 0 {
		t.Fatal("no effects cases")
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

// The goldens print the effects lines, inferred and declared sets, effect
// variables and each kind of effect error.
func TestTypesDumpEffectsCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, effectsDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, want := range []string{
		"\neffects module=", " func noisy io\n", " func quiet -\n", " func map E,alloc\n", " func shout alloc,io\n",
		" impl Named basics.Cat.name io\n", " func ping -\n",
		"unknown effect `network`", "must appear in the function type of a parameter",
		"(through `noisy`) but its signature does not declare it", "(through `map literal`)",
		"the argument for `f` of `strict` uses `io`",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no golden prints %q", want)
		}
	}
}

// effectsInputs are the inputs of TestSelfhostBodies, the cases of
// testdata/typesdump/effects and the selfhost package itself.
func effectsInputs(t *testing.T) []string {
	inputs := bodiesInputs(t)
	inputs = append(inputs, declDumpCases(t, effectsDir)...)
	return append(inputs, "../../../selfhost")
}

// TestSelfhostEffects checks that `narrac types <path>` prints the same dump
// (now with the effects lines) as `narra types`, or the same errors on
// stderr with exit 1, on every input of effectsInputs. Only an input that
// does not load is skipped. It logs the time of both on the selfhost package.
func TestSelfhostEffects(t *testing.T) {
	exe := buildSelfhost(t)
	t.Run("inputs", func(t *testing.T) {
		for _, path := range effectsInputs(t) {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				if _, err := Load(path); err != nil {
					t.Skip("does not load")
				}
				goStart := time.Now()
				want, ok := TypesDump(path)
				goTime := time.Since(goStart)
				narraStart := time.Now()
				out, errOut, code := runNarracRaw(t, exe, "types", path)
				narraTime := time.Since(narraStart)
				if strings.HasSuffix(path, "/selfhost") {
					t.Logf("types selfhost/: Go %v, narrac %v", goTime, narraTime)
				}
				if construct := unsupportedConstruct(errOut); construct != "" {
					t.Fatalf("unsupported: %s", construct)
				}
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
}

// TestSelfhostPackageIsClean checks that the effect pass accepts the selfhost
// package, as `narra check` does.
func TestSelfhostPackageIsClean(t *testing.T) {
	got, ok := TypesDump("../../../selfhost")
	if !ok {
		t.Fatalf("the selfhost package has errors:\n%s", got)
	}
}
