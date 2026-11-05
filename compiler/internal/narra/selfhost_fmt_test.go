package narra

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// runFmtCommand runs `narrac fmt` in dir with the given stdin.
func runFmtCommand(t *testing.T, exe, dir, stdin string, args ...string) (string, string, int) {
	t.Helper()
	return runFmtWithPWD(t, exe, dir, dir, stdin, args...)
}

// runFmtWithPWD runs `narrac fmt` in dir with the PWD variable set to pwd; an
// empty pwd leaves it empty, as a shell without the variable would.
func runFmtWithPWD(t *testing.T, exe, dir, pwd, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(exe, append([]string{"fmt"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0", "PWD="+pwd)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return stdout.String(), stderr.String(), code
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

var widthSuffix = regexp.MustCompile(`_w(\d+)(_|\.)`)

// optionsFromName reads the options of a golden from its file name: `_w<N>`
// sets printWidth and `_none` sets trailingComma to "none". It returns the
// narra.toml text, or "" without options.
func optionsFromName(name string) string {
	var lines []string
	if match := widthSuffix.FindStringSubmatch(name); match != nil {
		lines = append(lines, "printWidth = "+match[1])
	}
	if strings.Contains(strings.TrimSuffix(name, ".nra"), "_none") {
		lines = append(lines, `trailingComma = "none"`)
	}
	if len(lines) == 0 {
		return ""
	}
	return "[fmt]\n" + strings.Join(lines, "\n") + "\n"
}

// fmtStdin formats text through --stdin and fails the test on an error.
func fmtStdin(t *testing.T, exe, dir, text string) string {
	t.Helper()
	out, errText, code := runFmtCommand(t, exe, dir, text, "--stdin")
	if code != 0 {
		t.Fatalf("fmt --stdin exit code %d\nstderr:\n%s", code, errText)
	}
	return out
}

// TestSelfhostFmt checks `narrac fmt`:
//   - testdata/fmt/{decl,stmt,comment}_*.nra format to the .fmt golden through
//     --stdin, and through --check (the input is reported, the golden is not);
//     a `_w<N>` or `_none` in the name sets printWidth or trailingComma;
//   - testdata/fmt/config_*/ hold a narra.toml, input.nra and expected.fmt (or
//     expected.err for a bad config), run through --stdin and on a copy of the
//     directory in place;
//   - every .nra file under testdata/ and selfhost/src formats to a text that
//     formats to itself, keeps the token sequence and the comment texts, and
//     parses to the same tree once positions are removed.
//
// NARRA_UPDATE_GOLDEN=1 rewrites the .fmt goldens.
func TestSelfhostFmt(t *testing.T) {
	exe := buildSelfhost(t)
	fmtDir := filepath.Join(testdata, "fmt")

	var goldens []string
	for _, prefix := range []string{"decl_", "stmt_", "comment_", "expr_", "type_", "pattern_"} {
		found, _ := filepath.Glob(filepath.Join(fmtDir, prefix+"*.nra"))
		goldens = append(goldens, found...)
	}
	if len(goldens) == 0 {
		t.Fatal("no golden inputs in testdata/fmt")
	}
	for _, path := range goldens {
		t.Run("golden/"+filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			golden := strings.TrimSuffix(path, ".nra") + ".fmt"
			input := readFile(t, path)
			scratch := t.TempDir()
			if toml := optionsFromName(filepath.Base(path)); toml != "" {
				writeFile(t, filepath.Join(scratch, "narra.toml"), toml)
			}
			if os.Getenv("NARRA_UPDATE_GOLDEN") != "" {
				writeFile(t, golden, fmtStdin(t, exe, scratch, input))
			}
			want := readFile(t, golden)
			if got := fmtStdin(t, exe, scratch, input); got != want {
				t.Fatalf("stdout differs: %s", firstDiff(want, got))
			}
			// The golden is already formatted.
			if got := fmtStdin(t, exe, scratch, want); got != want {
				t.Fatalf("golden is not stable: %s", firstDiff(want, got))
			}
			// --check lists the input and not the golden.
			inputPath := filepath.Join(scratch, "input.nra")
			goldenPath := filepath.Join(scratch, "golden.nra")
			writeFile(t, inputPath, input)
			writeFile(t, goldenPath, want)
			out, errText, code := runFmtCommand(t, exe, scratch, "", "--check", "input.nra", "golden.nra")
			wantCode, wantOut := 0, ""
			if input != want {
				wantCode, wantOut = 1, "input.nra\n"
			}
			if code != wantCode || out != wantOut || errText != "" {
				t.Fatalf("--check: code %d stdout %q stderr %q, want code %d stdout %q", code, out, errText, wantCode, wantOut)
			}
			// In place, the input becomes the golden.
			if _, errText, code := runFmtCommand(t, exe, scratch, "", "input.nra"); code != 0 || errText != "" {
				t.Fatalf("in place: code %d stderr %q", code, errText)
			}
			if got := readFile(t, inputPath); got != want {
				t.Fatalf("in place differs: %s", firstDiff(want, got))
			}
		})
	}

	configs, _ := filepath.Glob(filepath.Join(fmtDir, "config_*"))
	for _, dir := range configs {
		t.Run("config/"+filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			input := readFile(t, filepath.Join(dir, "input.nra"))
			// Run in a copy under a temporary directory, so the nearest narra.toml
			// is the case's own and the input can sit in a subdirectory.
			scratch := t.TempDir()
			writeFile(t, filepath.Join(scratch, "narra.toml"), readFile(t, filepath.Join(dir, "narra.toml")))
			sub := filepath.Join(scratch, "sub")
			if err := os.Mkdir(sub, 0o755); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(sub, "input.nra"), input)
			if _, err := os.Stat(filepath.Join(dir, "expected.err")); err == nil {
				wantErr := readFile(t, filepath.Join(dir, "expected.err"))
				out, errText, code := runFmtCommand(t, exe, scratch, input, "--stdin")
				if code != 1 || out != "" || errText != wantErr {
					t.Fatalf("stdin: code %d stdout %q stderr %q, want code 1 stderr %q", code, out, errText, wantErr)
				}
				out, errText, code = runFmtCommand(t, exe, scratch, "", "sub/input.nra")
				if code != 1 || out != "" || errText != wantErr {
					t.Fatalf("path: code %d stdout %q stderr %q, want code 1 stderr %q", code, out, errText, wantErr)
				}
				if got := readFile(t, filepath.Join(sub, "input.nra")); got != input {
					t.Fatal("the input changed after a config error")
				}
				return
			}
			golden := filepath.Join(dir, "expected.fmt")
			if os.Getenv("NARRA_UPDATE_GOLDEN") != "" {
				writeFile(t, golden, fmtStdin(t, exe, scratch, input))
			}
			want := readFile(t, golden)
			if got := fmtStdin(t, exe, scratch, input); got != want {
				t.Fatalf("stdin differs: %s", firstDiff(want, got))
			}
			if _, errText, code := runFmtCommand(t, exe, scratch, "", "sub/input.nra"); code != 0 || errText != "" {
				t.Fatalf("in place: code %d stderr %q", code, errText)
			}
			if got := readFile(t, filepath.Join(sub, "input.nra")); got != want {
				t.Fatalf("in place differs: %s", firstDiff(want, got))
			}
		})
	}

	// A file that does not parse is reported and left unchanged.
	t.Run("syntax_error", func(t *testing.T) {
		scratch := t.TempDir()
		writeFile(t, filepath.Join(scratch, "bad.nra"), "function f( {\n")
		writeFile(t, filepath.Join(scratch, "good.nra"), "function g() {\n}\n")
		out, errText, code := runFmtCommand(t, exe, scratch, "", "bad.nra", "good.nra")
		if code != 1 || out != "" || !strings.HasPrefix(errText, "error bad.nra:1:") {
			t.Fatalf("code %d stdout %q stderr %q", code, out, errText)
		}
		if got := readFile(t, filepath.Join(scratch, "bad.nra")); got != "function f( {\n" {
			t.Fatalf("bad.nra changed: %q", got)
		}
		if got := readFile(t, filepath.Join(scratch, "good.nra")); got != "function g() {}\n" {
			t.Fatalf("good.nra not formatted: %q", got)
		}
		_, errText, code = runFmtCommand(t, exe, scratch, "function f( {\n", "--stdin")
		if code != 1 || !strings.HasPrefix(errText, "error <stdin>:1:") {
			t.Fatalf("stdin: code %d stderr %q", code, errText)
		}
	})

	// A comment the printer cannot place is an error, not a silent loss.
	t.Run("lost_comment", func(t *testing.T) {
		scratch := t.TempDir()
		source := "enum E {\n  A(/* inside */)\n}\n"
		writeFile(t, filepath.Join(scratch, "lost.nra"), source)
		want := "error lost.nra:2:5: comment not printed: /* inside */\n"
		out, errText, code := runFmtCommand(t, exe, scratch, "", "lost.nra")
		if code != 1 || out != "" || errText != want {
			t.Fatalf("code %d stdout %q stderr %q, want code 1 stderr %q", code, out, errText, want)
		}
		if got := readFile(t, filepath.Join(scratch, "lost.nra")); got != source {
			t.Fatalf("lost.nra changed: %q", got)
		}
		out, errText, code = runFmtCommand(t, exe, scratch, source, "--stdin")
		if code != 1 || out != "" || errText != "error <stdin>:2:5: comment not printed: /* inside */\n" {
			t.Fatalf("stdin: code %d stdout %q stderr %q", code, out, errText)
		}
	})

	// Without the PWD variable the config is still found, above a relative path.
	t.Run("no_pwd", func(t *testing.T) {
		scratch := t.TempDir()
		writeFile(t, filepath.Join(scratch, "narra.toml"), "[fmt]\nprintWidth = 40\n")
		if err := os.MkdirAll(filepath.Join(scratch, "sub", "deeper"), 0o755); err != nil {
			t.Fatal(err)
		}
		// Fits in 100 columns, not in 40.
		source := "function f(firstParameter: int, secondParameter: int): int {\n  return 1\n}\n"
		writeFile(t, filepath.Join(scratch, "sub", "b.nra"), source)
		writeFile(t, filepath.Join(scratch, "sub", "deeper", "c.nra"), source)
		for _, path := range []string{"sub/b.nra", "./sub/deeper/../b.nra", "sub/deeper/c.nra"} {
			for _, pwd := range []string{scratch, ""} {
				out, errText, code := runFmtWithPWD(t, exe, scratch, pwd, "", "--check", path)
				if code != 1 || out != path+"\n" || errText != "" {
					t.Fatalf("%s with PWD %q: code %d stdout %q stderr %q", path, pwd, code, out, errText)
				}
			}
		}
		// From a subdirectory the file is one level below the config.
		out, _, code := runFmtWithPWD(t, exe, filepath.Join(scratch, "sub"), "", "", "--check", "b.nra")
		if code != 1 || out != "b.nra\n" {
			t.Fatalf("from sub: code %d stdout %q", code, out)
		}
	})

	// A directory means every .nra file under it; no path means the current directory.
	t.Run("directory", func(t *testing.T) {
		scratch := t.TempDir()
		if err := os.MkdirAll(filepath.Join(scratch, "a", "b"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(scratch, "a", "b", "x.nra"), "function  x() {\n}\n")
		writeFile(t, filepath.Join(scratch, "a", "y.nra"), "function y() {}\n")
		writeFile(t, filepath.Join(scratch, "a", "note.txt"), "function  z() {\n}\n")
		out, _, code := runFmtCommand(t, exe, scratch, "", "--check")
		if code != 1 || out != "a/b/x.nra\n" {
			t.Fatalf("--check: code %d stdout %q", code, out)
		}
		if _, errText, code := runFmtCommand(t, exe, scratch, "", "a"); code != 0 || errText != "" {
			t.Fatalf("code %d stderr %q", code, errText)
		}
		if got := readFile(t, filepath.Join(scratch, "a", "b", "x.nra")); got != "function x() {}\n" {
			t.Fatalf("x.nra: %q", got)
		}
		if got := readFile(t, filepath.Join(scratch, "a", "note.txt")); got != "function  z() {\n}\n" {
			t.Fatalf("note.txt changed: %q", got)
		}
	})

	files, names := selfhostInputs(t, map[string]string{
		"only_comments.nra": "// a\n\n\n/* b */\n",
		"crlf.nra":          "// a\r\nfunction f() {\r\n  // b\r\n}\r\n",
		"empty.nra":         "",
	})
	for _, path := range files {
		t.Run("stable/"+names[path], func(t *testing.T) {
			t.Parallel()
			checkFormatted(t, exe, path)
		})
	}
}

var (
	positionPattern = regexp.MustCompile(`^( *\S+) \d+:\d+`)
	endPattern      = regexp.MustCompile(` end=\d+:\d+`)
	hiddenPattern   = regexp.MustCompile(`__for_\d+_\d+`)
	tokenPattern    = regexp.MustCompile(`^\d+:\d+ `)
	commentPattern  = regexp.MustCompile(`^comment \d+ \w+ \S+ \w+ "(.*)"$`)
)

// dumpOf runs a narrac dump command and returns its stdout and exit code.
func dumpOf(t *testing.T, exe string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		return stdout.String(), exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return stdout.String(), 0
}

// stripPositions removes the positions of an AST dump.
func stripPositions(dump string) string {
	lines := strings.Split(dump, "\n")
	for i, line := range lines {
		line = positionPattern.ReplaceAllString(line, "$1")
		line = endPattern.ReplaceAllString(line, "")
		lines[i] = hiddenPattern.ReplaceAllString(line, "__for_")
	}
	return strings.Join(lines, "\n")
}

// tokensOf is the token dump without positions, line ends, commas and
// grouping parentheses. The formatter sets line ends and commas, and it drops
// a redundant pair `((a))`, keeps a written pair, and adds one around a
// conditional that spans lines. A pair is a grouping pair when the token
// before the `(` cannot end an expression (an operator, `,`, `(`, `?`, `:`,
// `return` or nothing); call, parameter and condition parentheses stay in the
// sequence. The tree comparison covers what the grouping pairs mean.
func tokensOf(dump string) string {
	var tokens []string
	for _, line := range strings.Split(dump, "\n") {
		line = tokenPattern.ReplaceAllString(line, "")
		if strings.HasPrefix(line, "newline ") {
			continue
		}
		tokens = append(tokens, line)
	}
	drop := make([]bool, len(tokens))
	var open []int
	for index, token := range tokens {
		switch token {
		case "punct (":
			grouping := index == 0
			if index > 0 {
				before := tokens[index-1]
				grouping = before == "keyword return" ||
					(strings.HasPrefix(before, "punct ") && before != "punct )" && before != "punct ]" && before != "punct >" && before != "punct }")
			}
			drop[index] = grouping
			open = append(open, index)
		case "punct )":
			if len(open) > 0 {
				last := open[len(open)-1]
				open = open[:len(open)-1]
				drop[index] = drop[last]
			}
		}
	}
	var kept []string
	for index, token := range tokens {
		if drop[index] || token == "punct ," {
			continue
		}
		kept = append(kept, token)
	}
	return strings.Join(kept, "\n")
}

// commentsOf lists the comments of a trivia dump: the text without trailing
// blanks, the kind letter of the node it is attached to and its placement.
func commentsOf(dump string) string {
	var texts []string
	attached := map[string]string{}
	for _, line := range strings.Split(dump, "\n") {
		if match := commentPattern.FindStringSubmatch(line); match != nil {
			text := match[1]
			for {
				trimmed := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(strings.TrimRight(text, " "), `\t`), `\r`), `\n`)
				if trimmed == text {
					break
				}
				text = trimmed
			}
			texts = append(texts, text)
		} else if fields := strings.Fields(line); len(fields) == 4 && fields[0] == "attached" {
			kind := fields[1][:1]
			if strings.HasPrefix(fields[1], "file") {
				kind = "file"
			}
			attached[fields[3]] = kind + " " + fields[2]
		}
	}
	for i, text := range texts {
		texts[i] = text + "  [" + attached[strconv.Itoa(i)] + "]"
	}
	return strings.Join(texts, "\n")
}

// checkFormatted formats one file and checks that the result is stable, keeps
// the tokens and comments, and has the same tree. A file that does not parse
// must be rejected with exit code 1.
func checkFormatted(t *testing.T, exe, path string) {
	t.Helper()
	source := readFile(t, path)
	scratch := t.TempDir()
	if _, code := dumpOf(t, exe, "parse", path); code != 0 {
		_, errText, fmtCode := runFmtCommand(t, exe, scratch, source, "--stdin")
		if fmtCode != 1 || !strings.HasPrefix(errText, "error <stdin>:") {
			t.Fatalf("a file that does not parse: code %d stderr %q", fmtCode, errText)
		}
		return
	}
	once := fmtStdin(t, exe, scratch, source)
	if twice := fmtStdin(t, exe, scratch, once); twice != once {
		t.Fatalf("formatting twice differs: %s", firstDiff(once, twice))
	}
	formatted := filepath.Join(scratch, "formatted.nra")
	writeFile(t, formatted, once)

	beforeTokens, _ := dumpOf(t, exe, path)
	afterTokens, _ := dumpOf(t, exe, formatted)
	if want, got := tokensOf(beforeTokens), tokensOf(afterTokens); want != got {
		t.Fatalf("tokens differ: %s", firstDiff(want, got))
	}
	beforeComments, _ := dumpOf(t, exe, "trivia", path)
	afterComments, _ := dumpOf(t, exe, "trivia", formatted)
	if want, got := commentsOf(beforeComments), commentsOf(afterComments); want != got {
		t.Fatalf("comments differ: %s", firstDiff(want, got))
	}
	beforeTree, _ := dumpOf(t, exe, "parse", path)
	afterTree, _ := dumpOf(t, exe, "parse", formatted)
	if want, got := stripPositions(beforeTree), stripPositions(afterTree); want != got {
		t.Fatalf("tree differs: %s", firstDiff(want, got))
	}
}
