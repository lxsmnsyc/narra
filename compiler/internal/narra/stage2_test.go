package narra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Samples under testdata/stage2/ import std modules that only the self-hosted
// compiler has (see compiler/README.md, "Standard modules written in Narra").
// The Go tests skip the folder: TestPrograms, TestCompileErrors and the emit
// comparisons glob testdata/*.nra and a few named subfolders. These tests
// reuse the stage 2 binary of TestSelfhostBootstrap.

// TestStage2Programs emits each testdata/stage2/*.nra with stage 2, compiles
// the C with the sanitizer flags, runs it with the leak check and compares
// stdout with the .out file.
func TestStage2Programs(t *testing.T) {
	stage2 := buildStage2(t)
	files, _ := filepath.Glob(filepath.Join(testdata, "stage2", "*.nra"))
	if len(files) == 0 {
		t.Fatal("no stage 2 samples")
	}
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			want, err := os.ReadFile(strings.TrimSuffix(src, ".nra") + ".out")
			if err != nil {
				t.Fatalf("missing expected output for %s", src)
			}
			c, errOut, code := runNarracRaw(t, stage2, "emit", src)
			if code != 0 {
				t.Fatalf("stage 2 emit: exit code %d\nstderr:\n%s", code, errOut)
			}
			dir := t.TempDir()
			cfile := filepath.Join(dir, "prog.c")
			if err := os.WriteFile(cfile, []byte(c), 0o644); err != nil {
				t.Fatal(err)
			}
			exe := filepath.Join(dir, "prog")
			args := append([]string{"-O0", "-g", "-std=c11", "-w"}, modes[0].opts.CFlags...)
			args = append(args, "-o", exe, cfile, "-lm")
			cc := os.Getenv("CC")
			if cc == "" {
				cc = "cc"
			}
			if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
				t.Fatalf("C compiler failed: %v\n%s", err, out)
			}
			run := exec.Command(exe)
			run.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
			var stdout, stderr strings.Builder
			run.Stdout, run.Stderr = &stdout, &stderr
			err = run.Run()
			if err != nil {
				t.Fatalf("run: %v\nstderr:\n%s", err, stderr.String())
			}
			if err := stage2OutputError(string(want), stdout.String()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestStage2Errors checks that stage 2 rejects each testdata/stage2/errors
// sample with the error on the first line ("// error: text").
func TestStage2Errors(t *testing.T) {
	stage2 := buildStage2(t)
	files, _ := filepath.Glob(filepath.Join(testdata, "stage2", "errors", "*.nra"))
	if len(files) == 0 {
		t.Fatal("no stage 2 error samples")
	}
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			text, _ := os.ReadFile(src)
			first := strings.SplitN(string(text), "\n", 2)[0]
			want := strings.TrimSpace(strings.TrimPrefix(first, "// error:"))
			out, errOut, code := runNarracRaw(t, stage2, "emit", src)
			if code != 1 || out != "" {
				t.Fatalf("exit code %d, want 1 with no output; stdout:\n%s", code, out)
			}
			if !strings.Contains(errOut, want) {
				t.Fatalf("want error containing %q, got:\n%s", want, errOut)
			}
		})
	}
}

// stage2OutputError is nil when the program printed the expected output.
func stage2OutputError(want, got string) error {
	if want != got {
		return fmt.Errorf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
	return nil
}

// The comparison must fail when an expected output differs.
func TestStage2OutputComparison(t *testing.T) {
	if stage2OutputError("42\n", "42\n") != nil {
		t.Fatal("equal output was rejected")
	}
	if stage2OutputError("42\n", "43\n") == nil {
		t.Fatal("different output was accepted")
	}
}

// A std source module (path std/...) may import std/internal/..., and the
// call to the primitive checks and emits.
func TestStdInternalFromStdModule(t *testing.T) {
	src := "import * as Probe from \"std/internal/probe\"\n\nexport function code(c: char): int {\n  return Probe.code(c)\n}\n"
	f, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	f.Path = "std/src/probeuser.nra"
	pkg := &Package{Name: "std", Modules: []*SourceModule{newSourceModule("std/probeuser", f)}, single: true}
	if _, err := CheckPackage(pkg); err != nil {
		t.Fatalf("a std module must import std/internal: %v", err)
	}
}
