package narra

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSelfhostFmtDoc runs `narrac fmt-doc` on every testdata/fmt/doc_*.txt
// and compares stdout with the .out file next to it. The cases cover the
// rules of contracts/fmt-doc.md; the options are `@` lines in the .txt file.
func TestSelfhostFmtDoc(t *testing.T) {
	exe := buildSelfhost(t)
	inputs, err := filepath.Glob("../../testdata/fmt/doc_*.txt")
	if err != nil || len(inputs) == 0 {
		t.Fatalf("no golden cases: %v", err)
	}
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(input, ".txt") + ".out")
			if err != nil {
				t.Fatal(err)
			}
			runNarrac(t, exe, []string{"fmt-doc", input}, string(want), "", 0)
		})
	}
}

// TestSelfhostFmtDocErrors checks the reader of the text form.
func TestSelfhostFmtDocErrors(t *testing.T) {
	exe := buildSelfhost(t)
	cases := map[string]string{
		"newline.txt":   "\"a\\nb\"\n",
		"unclosed.txt":  "(concat \"a\"\n",
		"unknown.txt":   "(frob \"a\")\n",
		"two.txt":       "\"a\" \"b\"\n",
		"option.txt":    "@bogus\n\"a\"\n",
		"ifbreak.txt":   "(ifbreak \"a\")\n",
		"extra.txt":     "\"a\")\n",
		"unterm.txt":    "\"abc\n",
		"groupid.txt":   "(group-id \"a\")\n",
		"badnumber.txt": "(group-id x \"a\")\n",
	}
	dir := t.TempDir()
	for name, source := range cases {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) {
			// Only the exit code is pinned; the message text may change.
			runNarracFails(t, exe, []string{"fmt-doc", path})
		})
	}
}

// TestSelfhostFmtDocErrorText pins the error line: one position prefix.
func TestSelfhostFmtDocErrorText(t *testing.T) {
	exe := buildSelfhost(t)
	path := filepath.Join(t.TempDir(), "word.txt")
	if err := os.WriteFile(path, []byte("bogus\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runNarrac(t, exe, []string{"fmt-doc", path}, "", "error 1:1: unknown word bogus\n", 1)
}

// TestSelfhostFmtDocDeep prints documents nested 50000 levels deep; the
// printer, the reader and the release of the document must not use the C stack
// per level.
func TestSelfhostFmtDocDeep(t *testing.T) {
	exe := buildSelfhost(t)
	const depth = 50000
	dir := t.TempDir()
	cases := []struct {
		name, source, want string
	}{
		{
			"indent",
			strings.Repeat("(indent ", depth) + "(concat hardline \"x\")" + strings.Repeat(")", depth) + "\n",
			"\n" + strings.Repeat("  ", depth) + "x",
		},
		{
			"concat",
			strings.Repeat("(concat \"a\" ", depth) + "\"b\"" + strings.Repeat(")", depth) + "\n",
			strings.Repeat("a", depth) + "b",
		},
		{
			"group",
			strings.Repeat("(group (concat \"a\" ", depth) + "\"b\"" + strings.Repeat("))", depth) + "\n",
			strings.Repeat("a", depth) + "b",
		},
		{
			"group_line",
			"@width 10\n" + strings.Repeat("(group (concat \"a\" line ", 5) + "\"b\"" + strings.Repeat("))", 5) + "\n",
			"a\na a a a b",
		},
	}
	for _, c := range cases {
		path := filepath.Join(dir, c.name+".txt")
		if err := os.WriteFile(path, []byte(c.source), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Run(c.name, func(t *testing.T) {
			runNarrac(t, exe, []string{"fmt-doc", path}, c.want, "", 0)
		})
	}
}

// runNarracFails checks that narrac exits with code 1 and prints an error line.
func runNarracFails(t *testing.T, exe string, args []string) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("want failure, got output %q", out)
	}
	if !strings.Contains(string(out), "error ") && !strings.Contains(string(out), "needs") && !strings.Contains(string(out), "unknown option") {
		t.Fatalf("no error message: %q", out)
	}
}
