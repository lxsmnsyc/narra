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

// emitNodes are the roadmap nodes whose stubs report `unsupported:`. Each has
// a skip file compiler/testdata/emit/skips/<node>.txt that lists, one per line
// relative to testdata/, the inputs on which narrac reports that node. A node
// removes its inputs from its file as its emitter grows, and deletes the file
// content when it is done.
var emitNodes = []string{
	"selfhost-emit-program",
	"selfhost-emit-decls",
	"selfhost-emit-bodies",
	"selfhost-emit-match",
	"selfhost-emit-closures",
	"selfhost-emit-traits",
	"selfhost-emit-maps",
	"selfhost-emit-std",
	"selfhost-emit-options",
}

// emitSections are the headers that split the C file, in order. The text
// before the first header is the preamble.
var emitSections = []string{
	"\n// Types\n",
	"\n// Reference counting\n",
	"\n// Constants\n",
	"\n// Functions\n",
	"\n// String literals\n",
}

// emitInput is one program the emit tests run.
type emitInput struct {
	path  string // as passed to the compiler
	name  string // relative to testdata/, the name in the skip files
	tests bool
}

// emitInputs lists testdata/*.nra (test mode for tests.nra and std_test*),
// testdata/panics/*.nra, the packages except errors/ (test mode for runner),
// and testdata/runner/mixed.nra (test mode).
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
	var kept []emitInput
	for _, in := range inputs {
		if in.name == "packages/errors" {
			continue
		}
		kept = append(kept, in)
	}
	return kept
}

// readEmitSkips returns, for each input name, the nodes whose skip file lists it.
func readEmitSkips(t *testing.T) map[string]map[string]bool {
	t.Helper()
	skips := map[string]map[string]bool{}
	for _, node := range emitNodes {
		data, err := os.ReadFile(filepath.Join(testdata, "emit", "skips", node+".txt"))
		if err != nil {
			t.Fatalf("skip file of %s: %v", node, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if skips[line] == nil {
				skips[line] = map[string]bool{}
			}
			skips[line][node] = true
		}
	}
	return skips
}

// emitUnsupportedNodes returns the nodes named by the `unsupported: <construct> (<node>)`
// lines of narrac's stderr.
func emitUnsupportedNodes(stderr string) map[string]bool {
	nodes := map[string]bool{}
	for _, line := range strings.Split(stderr, "\n") {
		rest, ok := strings.CutPrefix(line, "unsupported: ")
		if !ok {
			continue
		}
		open := strings.LastIndex(rest, " (")
		if open < 0 || !strings.HasSuffix(rest, ")") {
			continue
		}
		nodes[rest[open+2:len(rest)-1]] = true
	}
	return nodes
}

// splitEmitSections cuts the C text at the section headers. A header that is
// missing leaves the following parts empty.
func splitEmitSections(text string) []string {
	parts := make([]string, 0, len(emitSections)+1)
	rest := text
	for _, header := range emitSections {
		at := strings.Index(rest, header)
		if at < 0 {
			parts = append(parts, rest)
			rest = ""
			continue
		}
		parts = append(parts, rest[:at])
		rest = rest[at:]
	}
	return append(parts, rest)
}

// beforeMarker is the text of a part before its first marker line, and whether there is one.
func beforeMarker(part string) (string, bool) {
	at := strings.Index(part, "/* unsupported:")
	if at < 0 {
		return part, false
	}
	// Cut at the start of the marker's line.
	if nl := strings.LastIndex(part[:at], "\n"); nl >= 0 {
		return part[:nl+1], true
	}
	return "", true
}

// compareEmitted checks the first five parts of narrac's C against Go's. The
// last part is not compared: a stub shifts the literal and temp counters. A
// part with a marker must agree up to the marker line; one without must be equal.
func compareEmitted(want, got string) string {
	wantParts, gotParts := splitEmitSections(want), splitEmitSections(got)
	for i := 0; i < len(emitSections); i++ {
		prefix, marked := beforeMarker(gotParts[i])
		if marked {
			if !strings.HasPrefix(wantParts[i], prefix) {
				return "part " + strconv.Itoa(i) + " differs before its marker: " + firstDiff(wantParts[i], prefix)
			}
		} else if wantParts[i] != gotParts[i] {
			return "part " + strconv.Itoa(i) + " differs: " + firstDiff(wantParts[i], gotParts[i])
		}
	}
	return ""
}

// TestSelfhostEmit runs `narrac emit` on every input and checks it against
// Go's CompileBinary. The nodes narrac reports (U) must equal the nodes whose
// skip file lists the input (S). With U empty the C must be byte-equal and the
// exit code 0; otherwise the exit code is 3 and the parts of the file agree
// up to the first marker. The counts per node are logged.
func TestSelfhostEmit(t *testing.T) {
	exe := buildSelfhost(t)
	skips := readEmitSkips(t)
	var mu sync.Mutex
	compared, skipped := map[string]int{}, map[string]int{}
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
				got := emitUnsupportedNodes(errOut)
				listed := skips[in.name]
				for node := range got {
					if !listed[node] {
						t.Errorf("narrac reports %s but its skip file does not list %s", node, in.name)
					}
				}
				for node := range listed {
					if !got[node] {
						t.Errorf("the skip file of %s lists %s but narrac does not report it", node, in.name)
					}
				}
				if t.Failed() {
					return
				}
				mu.Lock()
				for node := range got {
					skipped[node]++
				}
				mu.Unlock()
				if len(got) == 0 {
					if code != 0 {
						t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
					}
					if out != want {
						t.Fatalf("C differs: %s", firstDiff(want, out))
					}
					mu.Lock()
					identical++
					mu.Unlock()
					return
				}
				if code != 3 {
					t.Fatalf("exit code %d, want 3; stderr:\n%s", code, errOut)
				}
				if diff := compareEmitted(want, out); diff != "" {
					t.Fatal(diff)
				}
				mu.Lock()
				for _, node := range emitNodes {
					if !got[node] {
						compared[node]++
					}
				}
				mu.Unlock()
			})
		}
	})
	t.Logf("byte-identical %d, did not compile %d", identical, failedToCompile)
	for _, node := range emitNodes {
		t.Logf("  %-24s skipped %3d, compared without it %3d", node, skipped[node], compared[node])
	}
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
