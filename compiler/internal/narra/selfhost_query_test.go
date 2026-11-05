package narra

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// queryCase is one `> command` of a queries.txt file and the output it expects.
type queryCase struct {
	command string
	want    string
}

// parseQueries reads a queries.txt file: lines before the first `>` are
// comments, then each `> <arguments>` line is followed by the expected
// output up to the next `>` line: a first line `# exit N` when the exit
// code is not 0, the stdout, then a line `# stderr` and the stderr when
// stderr is not empty.
func parseQueries(text string) []queryCase {
	var cases []queryCase
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.HasPrefix(line, "> ") {
			cases = append(cases, queryCase{command: strings.TrimSpace(line[2:])})
			continue
		}
		if len(cases) > 0 {
			cases[len(cases)-1].want += line
		}
	}
	return cases
}

// runQuery runs `narrac query --package dir <arguments>` and returns the
// output in the queries.txt form.
func runQuery(t *testing.T, exe, dir, command string) string {
	t.Helper()
	args := append([]string{"query", "--package", dir}, strings.Fields(command)...)
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if code != 0 {
		out = "# exit " + strconv.Itoa(code) + "\n" + out
	}
	if stderr.Len() > 0 {
		if out != "" && !strings.HasSuffix(out, "\n") {
			t.Fatalf("stdout of %q does not end with a newline", command)
		}
		out += "# stderr\n" + stderr.String()
	}
	return out
}

// TestSelfhostQuery runs every command of compiler/testdata/query/<case>/queries.txt
// against the package in that folder and compares the output byte for byte
// (contracts/query-output.md). NARRA_UPDATE=1 rewrites the expected output.
func TestSelfhostQuery(t *testing.T) {
	exe := buildSelfhost(t)
	dirs, err := filepath.Glob("../../testdata/query/*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) == 0 {
		t.Fatal("no query cases")
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(dir, "queries.txt")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			var header strings.Builder
			for _, line := range strings.SplitAfter(text, "\n") {
				if strings.HasPrefix(line, "> ") {
					break
				}
				header.WriteString(line)
			}
			cases := parseQueries(text)
			if len(cases) == 0 {
				t.Fatalf("%s has no queries", path)
			}
			var updated strings.Builder
			updated.WriteString(header.String())
			for _, c := range cases {
				got := runQuery(t, exe, dir, c.command)
				updated.WriteString("> " + c.command + "\n" + got)
				if os.Getenv("NARRA_UPDATE") == "" && got != c.want {
					t.Errorf("> %s\n%s", c.command, firstDiff(c.want, got))
				}
			}
			if os.Getenv("NARRA_UPDATE") != "" {
				if err := os.WriteFile(path, []byte(updated.String()), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
