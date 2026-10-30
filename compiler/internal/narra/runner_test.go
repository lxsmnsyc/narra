package narra

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// runTests builds src in test mode and runs it with args and extra env.
func runTests(t *testing.T, src string, m buildMode, args []string, env ...string) (string, string, int) {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "prog")
	opts := m.opts
	opts.Tests = true
	if err := Build(src, exe, opts); err != nil {
		t.Fatalf("build %s: %v", src, err)
	}
	cmd := exec.Command(exe, append([]string{"--golden-dir", goldenDir(src)}, args...)...)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	cmd.Env = append(cmd.Env, env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

// TestRunRunnerMixed: every test runs, failures are listed, the exit code is 1.
func TestRunRunnerMixed(t *testing.T) {
	src := filepath.Join(testdata, "runner", "mixed.nra")
	want, err := os.ReadFile(filepath.Join(testdata, "runner", "mixed.out"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out, stderr, code := runTests(t, src, m, nil)
			if code != 1 || out != string(want) {
				t.Fatalf("want exit 1 and\n%s\ngot exit %d\n%s\nstderr:\n%s", want, code, out, stderr)
			}
		})
	}
}

// TestRunRunnerFilterAndTime: --filter selects by substring, --time adds
// timing only when asked.
func TestRunRunnerFilterAndTime(t *testing.T) {
	src := filepath.Join(testdata, "runner", "mixed.nra")
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out, _, code := runTests(t, src, m, []string{"--filter", "passes"})
			if code != 0 || out != "ok - passes first\nok - passes after failures\n2 passed, 0 failed\n" {
				t.Fatalf("filter: exit %d\n%s", code, out)
			}
			out, _, code = runTests(t, src, m, []string{"--filter=nothing matches"})
			if code != 1 || out != "no tests match \"nothing matches\"\n" {
				t.Fatalf("empty filter: exit %d\n%s", code, out)
			}
			out, _, _ = runTests(t, src, m, []string{"-filter", "passes first", "-time"})
			re := regexp.MustCompile(`^ok - passes first \(\d+ ms\)\n1 passed, 0 failed\n$`)
			if !re.MatchString(out) {
				t.Fatalf("time:\n%s", out)
			}
		})
	}
}

// TestRunRunnerGolden covers a match, a mismatch, a missing file and the
// update mode, in a temporary directory.
func TestRunRunnerGolden(t *testing.T) {
	text := `import * as Test from "std/test"

test "golden" {
  Test.golden("out", "one\ntwo\nthree\n")
}
`
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "g.nra")
			if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			golden := filepath.Join(dir, "out.golden")

			out, _, code := runTests(t, src, m, nil)
			if code != 1 || !strings.Contains(out, "golden file out.golden does not exist") {
				t.Fatalf("missing: exit %d\n%s", code, out)
			}
			out, _, code = runTests(t, src, m, nil, "NARRA_UPDATE_GOLDEN=1")
			if code != 0 {
				t.Fatalf("update: exit %d\n%s", code, out)
			}
			got, _ := os.ReadFile(golden)
			if string(got) != "one\ntwo\nthree\n" {
				t.Fatalf("golden file holds %q", got)
			}
			if out, _, code = runTests(t, src, m, nil); code != 0 {
				t.Fatalf("match: exit %d\n%s", code, out)
			}
			if err := os.WriteFile(golden, []byte("one\n2\nthree\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			out, _, code = runTests(t, src, m, nil)
			if code != 1 || !strings.Contains(out, `golden out differs at line 2: expected "2", got "two"`) {
				t.Fatalf("mismatch: exit %d\n%s", code, out)
			}
			// The actual text is longer than the file: the line after the last.
			if err := os.WriteFile(golden, []byte("one\ntwo\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			out, _, code = runTests(t, src, m, nil)
			if code != 1 || !strings.Contains(out, `differs at line 3: expected "", got "three"`) {
				t.Fatalf("longer: exit %d\n%s", code, out)
			}
		})
	}
}

// TestRunRunnerGoldenPackage: a package reads golden files from its testdata/.
func TestRunRunnerGoldenPackage(t *testing.T) {
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out, _, code := runTests(t, filepath.Join(testdata, "packages", "runner"), m, nil)
			if code != 0 || out != "ok - golden in a package\n1 passed, 0 failed\n" {
				t.Fatalf("exit %d\n%s", code, out)
			}
		})
	}
}

func writeProg(t *testing.T, text string) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "p.nra")
	if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return src
}

// TestRunRunnerTimeout: a test that never returns fails after the limit and
// the run goes on.
func TestRunRunnerTimeout(t *testing.T) {
	src := writeProg(t, "test \"spins\" {\n  while (true) {}\n}\n\ntest \"after\" {\n  assert(true)\n}\n")
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out, _, code := runTests(t, src, m, nil, "NARRA_TEST_TIMEOUT=1")
			want := "FAIL - spins: timed out after 1 s\nok - after\n1 passed, 1 failed\n"
			if code != 1 || out != want {
				t.Fatalf("exit %d\n%s", code, out)
			}
		})
	}
}

// TestRunRunnerLeak: a leak in one test fails that test and names the leak.
// NARRA_TEST_FORCE_LEAK is a hook of the runner for this test.
func TestRunRunnerLeak(t *testing.T) {
	src := writeProg(t, "test \"a\" {\n  assert(true)\n}\n")
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out, _, code := runTests(t, src, m, nil, "NARRA_TEST_FORCE_LEAK=1")
			if code != 1 || out != "FAIL - a: leaked 1 allocations (leak check)\n0 passed, 1 failed\n" {
				t.Fatalf("exit %d\n%s", code, out)
			}
		})
	}
}

// TestRunRunnerArguments: unknown options and a missing value are errors, a
// file without tests passes, and golden names are checked.
func TestRunRunnerArguments(t *testing.T) {
	m := modes[1]
	src := writeProg(t, "test \"a\" {\n  assert(true)\n}\n")
	if _, stderr, code := runTests(t, src, m, []string{"--bogus"}); code != 2 || !strings.Contains(stderr, "unknown argument --bogus") {
		t.Fatalf("unknown: exit %d\n%s", code, stderr)
	}
	if _, stderr, code := runTests(t, src, m, []string{"--filter"}); code != 2 || !strings.Contains(stderr, "needs a value") {
		t.Fatalf("no value: exit %d\n%s", code, stderr)
	}
	none := writeProg(t, "function f() {}\n")
	if out, _, code := runTests(t, none, m, nil); code != 0 || out != "0 passed, 0 failed\n" {
		t.Fatalf("no tests: exit %d\n%s", code, out)
	}
	names := writeProg(t, `import * as Test from "std/test"

test "dots" {
  Test.golden("a..b", "x")
}

test "up" {
  Test.golden("a/../b", "x")
}

test "empty" {
  Test.golden("", "x")
}
`)
	out, _, code := runTests(t, names, m, nil, "NARRA_UPDATE_GOLDEN=1")
	for _, want := range []string{"ok - dots", "FAIL - up: Test.golden: the name must be a relative path without a `..` segment", "FAIL - empty: Test.golden: the name is empty"} {
		if !strings.Contains(out, want) {
			t.Fatalf("exit %d, missing %q\n%s", code, want, out)
		}
	}
}

// TestRunRunnerGoldenOutsideTest:Test.golden is test-only like the others.
func TestRunRunnerGoldenOutsideTest(t *testing.T) {
	src := filepath.Join(t.TempDir(), "g.nra")
	text := "import * as Test from \"std/test\"\n\nfunction f() {\n  Test.golden(\"x\", \"y\")\n}\n"
	if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CompileToC(src, true); err == nil || !strings.Contains(err.Error(), "can only be called inside a `test` block") {
		t.Fatalf("got %v", err)
	}
}
