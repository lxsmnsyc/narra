package narra

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testdata = "../../testdata"

type buildMode struct {
	name string
	opts BuildOptions
}

// Debug builds run under AddressSanitizer and the runtime leak check, so
// reference counting mistakes fail the tests.
var modes = []buildMode{
	{"debug", BuildOptions{CFlags: []string{"-fsanitize=address,undefined", "-fno-omit-frame-pointer"}}},
	{"release", BuildOptions{Release: true}},
}

func runProgram(t *testing.T, src string, opts BuildOptions) (string, string, int) {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "prog")
	if err := Build(src, exe, opts); err != nil {
		t.Fatalf("build %s: %v", src, err)
	}
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s: %v", src, err)
	}
	return stdout.String(), stderr.String(), code
}

func TestPrograms(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(testdata, "*.nra"))
	for _, src := range files {
		want, err := os.ReadFile(strings.TrimSuffix(src, ".nra") + ".out")
		if err != nil {
			t.Fatalf("missing expected output for %s", src)
		}
		for _, m := range modes {
			t.Run(filepath.Base(src)+"/"+m.name, func(t *testing.T) {
				opts := m.opts
				// tests.nra and std_test*.nra run their test blocks.
				base := filepath.Base(src)
				opts.Tests = base == "tests.nra" || strings.HasPrefix(base, "std_test")
				out, stderr, code := runProgram(t, src, opts)
				if code != 0 {
					t.Fatalf("exit code %d\nstderr:\n%s", code, stderr)
				}
				if out != string(want) {
					t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, out)
				}
			})
		}
	}
}

func TestPanics(t *testing.T) {
	cases := map[string]string{
		"bounds.nra":   "index 5 out of range for length 2",
		"overflow.nra": "integer overflow",
		"divzero.nra":  "division by zero",
	}
	for file, msg := range cases {
		t.Run(file, func(t *testing.T) {
			_, stderr, code := runProgram(t, filepath.Join(testdata, "panics", file), modes[0].opts)
			if code != 101 || !strings.Contains(stderr, msg) {
				t.Fatalf("want panic %q with exit 101, got exit %d\nstderr:\n%s", msg, code, stderr)
			}
		})
	}
}

func TestCompileErrors(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(testdata, "errors", "*.nra"))
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			text, _ := os.ReadFile(src)
			first := strings.SplitN(string(text), "\n", 2)[0]
			want := strings.TrimSpace(strings.TrimPrefix(first, "// error:"))
			_, err := CompileToC(src, false)
			if err == nil {
				t.Fatalf("expected an error containing %q", want)
			}
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("want error containing %q, got:\n%v", want, err)
			}
		})
	}
}
