package narra

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// emitInput is one program the emit tests run.
type emitInput struct {
	path  string // as passed to the compiler
	name  string // relative to testdata/, the name in the skip files
	tests bool
}

// emitInputs lists testdata/*.nra (test mode for tests.nra and std_test*),
// testdata/panics/*.nra, the packages except errors/ (test mode for runner),
// testdata/runner/mixed.nra (test mode) and the selfhost package.
func emitInputs(t *testing.T) []emitInput {
	t.Helper()
	var inputs []emitInput
	add := func(pattern string, tests func(string) bool) {
		matches, err := filepath.Glob(filepath.Join(testdata, pattern))
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(matches)
		for _, path := range matches {
			name, _ := filepath.Rel(testdata, path)
			inputs = append(inputs, emitInput{path: path, name: filepath.ToSlash(name), tests: tests(filepath.Base(path))})
		}
	}
	add("*.nra", func(base string) bool { return base == "tests.nra" || strings.HasPrefix(base, "std_test") })
	add("panics/*.nra", func(string) bool { return false })
	add("packages/*", func(base string) bool { return base == "runner" })
	add("runner/mixed.nra", func(string) bool { return true })
	// The selfhost package itself, the input of the bootstrap.
	inputs = append(inputs, emitInput{path: "../../../selfhost", name: "selfhost"})
	var kept []emitInput
	for _, in := range inputs {
		if in.name == "packages/errors" {
			continue
		}
		kept = append(kept, in)
	}
	return kept
}

// TestSelfhostEmit runs `narrac emit` on every input and checks it against
// Go's CompileBinary: the C must be byte-equal and the exit code 0, and stderr
// must hold no `unsupported:` line. An input Go cannot compile must fail the
// same way, with exit code 1.
func TestSelfhostEmit(t *testing.T) {
	exe := buildSelfhost(t)
	var mu sync.Mutex
	identical, failedToCompile := 0, 0
	t.Run("inputs", func(t *testing.T) {
		for _, in := range emitInputs(t) {
			t.Run(in.name, func(t *testing.T) {
				t.Parallel()
				want, goErr := CompileBinary(in.path, "", in.tests)
				args := []string{"emit"}
				if in.tests {
					args = append(args, "--tests")
				}
				out, errOut, code := runNarracRaw(t, exe, append(args, in.path)...)
				if goErr != nil {
					// The emitter cannot start: the same errors, exit 1.
					if code != 1 || out != "" {
						t.Fatalf("exit code %d, want 1; stdout:\n%s\nstderr:\n%s", code, out, errOut)
					}
					mu.Lock()
					failedToCompile++
					mu.Unlock()
					return
				}
				if strings.Contains(errOut, "unsupported:") {
					t.Fatalf("narrac reports an unsupported construct (exit code %d); stderr:\n%s", code, errOut)
				}
				if code != 0 {
					t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
				}
				if out != want {
					t.Fatalf("C differs: %s", firstDiff(want, out))
				}
				mu.Lock()
				identical++
				mu.Unlock()
			})
		}
	})
	t.Logf("byte-identical %d, did not compile %d", identical, failedToCompile)
}

// TestSelfhostEmitErrors checks the messages Go's CompileBinary gives for a
// file without an entry function and for an unknown binary.
func TestSelfhostEmitErrors(t *testing.T) {
	exe := buildSelfhost(t)
	dir := t.TempDir()
	noEntry := filepath.Join(dir, "noentry.nra")
	if err := os.WriteFile(noEntry, []byte("function helper(): int {\n  return 1\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		path string
		bin  string
	}{
		{"no entry", noEntry, ""},
		{"unknown binary", filepath.Join(testdata, "packages", "app"), "nope"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, goErr := CompileBinary(c.path, c.bin, false)
			if goErr == nil {
				t.Fatal("Go compiled it")
			}
			args := []string{"emit"}
			if c.bin != "" {
				args = append(args, "--bin", c.bin)
			}
			out, errOut, code := runNarracRaw(t, exe, append(args, c.path)...)
			if code != 1 || out != "" {
				t.Fatalf("exit code %d, want 1; stdout:\n%s", code, out)
			}
			// Go names the file as the loader does; the message text must match.
			want := goErr.Error() + "\n"
			if errOut != want {
				t.Fatalf("message differs: %s", firstDiff(want, errOut))
			}
		})
	}
}

// TestSelfhostXXH32 checks narrac's xxHash32 against the vectors of TestXXH32.
func TestSelfhostXXH32(t *testing.T) {
	exe := buildSelfhost(t)
	alpha := "abcdefghijklmnopqrstuvwxyz"
	cases := []struct {
		in   string
		seed uint32
		want uint32
	}{
		{"", 0, 0x02cc5d05},
		{"abc", 0, 0x32d153ff},
		{"Nobody inspects the spammish repetition", 0, 0xe2293b2f},
		{alpha[:15], 0, 0xb918a375},
		{alpha[:16], 0, 0x9d2d8b62},
		{alpha[:17], 0, 0xb3b873e1},
		{"abc", 0x9e3779b1, 0xa1ae7709},
		{alpha[:16], 0x9e3779b1, 0x3fab7bd8},
	}
	for _, c := range cases {
		out, errOut, code := runNarracRaw(t, exe, "hash", strconv.FormatUint(uint64(c.seed), 10), c.in)
		if code != 0 {
			t.Fatalf("hash %q: exit code %d: %s", c.in, code, errOut)
		}
		if want := hex8(c.want) + "\n"; out != want {
			t.Errorf("hash(%q, %#x) = %q, want %q", c.in, c.seed, out, want)
		}
	}
}
