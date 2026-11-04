package narra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// bootstrapRoot is the selfhost package, as emitInputs passes it.
const bootstrapRoot = "../../../selfhost"

// stage2Build holds the stage-2 compiler: stage 1 emits its own package and
// the C compiler builds that C with the flags of the stage-1 build.
var stage2Build struct {
	once sync.Once
	dir  string
	exe  string
	c    string // the C that stage 1 emitted
	err  error
}

func buildStage2Executable() {
	dir, err := os.MkdirTemp("", "narrac2")
	if err != nil {
		stage2Build.err = err
		return
	}
	stage2Build.dir = dir
	cmd := exec.Command(selfhostBuild.exe, "emit", bootstrapRoot)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		stage2Build.err = fmt.Errorf("stage 1 emit: %w\n%s", err, stderr.String())
		return
	}
	stage2Build.c = string(out)
	cfile := filepath.Join(dir, "stage2.c")
	if err := os.WriteFile(cfile, out, 0o644); err != nil {
		stage2Build.err = err
		return
	}
	stage2Build.exe = filepath.Join(dir, "narrac2")
	// The same command and flags as Build runs for modes[0] (driver.go, Build).
	args := append([]string{"-O0", "-g", "-std=c11", "-w"}, modes[0].opts.CFlags...)
	args = append(args, "-o", stage2Build.exe, cfile, "-lm")
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	build := exec.Command(cc, args...)
	build.Stdout, build.Stderr = os.Stderr, os.Stderr
	if err := build.Run(); err != nil {
		stage2Build.err = fmt.Errorf("C compiler failed: %w", err)
	}
}

// buildStage2 builds the stage-2 compiler once per test run.
func buildStage2(t *testing.T) string {
	t.Helper()
	buildSelfhost(t)
	stage2Build.once.Do(buildStage2Executable)
	if stage2Build.err != nil {
		t.Fatalf("build stage 2: %v", stage2Build.err)
	}
	return stage2Build.exe
}

// TestSelfhostBootstrap closes the loop. Stage 1 (built by Go) emits the
// selfhost package; the C compiler builds that C into stage 2. Stage 2 must
// emit the same C (the fixed point), and must match Go on every emit input.
func TestSelfhostBootstrap(t *testing.T) {
	stage2 := buildStage2(t)
	goC, err := CompileBinary(bootstrapRoot, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if stage2Build.c != goC {
		t.Fatalf("stage 1 C differs from Go: %s", firstDiff(goC, stage2Build.c))
	}
	// Fixed point: stage 2 emitting its own package gives stage 1's output.
	out, errOut, code := runNarracRaw(t, stage2, "emit", bootstrapRoot)
	if code != 0 {
		t.Fatalf("stage 2 emit selfhost: exit code %d\nstderr:\n%s", code, errOut)
	}
	if out != stage2Build.c {
		t.Fatalf("not a fixed point: %s", firstDiff(stage2Build.c, out))
	}
	// Stage 2 against Go on every emit input.
	t.Run("inputs", func(t *testing.T) {
		for _, in := range emitInputs(t) {
			t.Run(in.name, func(t *testing.T) {
				t.Parallel()
				want, goErr := CompileBinary(in.path, "", in.tests)
				args := []string{"emit"}
				if in.tests {
					args = append(args, "--tests")
				}
				got, errOut, code := runNarracRaw(t, stage2, append(args, in.path)...)
				if goErr != nil {
					if code != 1 || got != "" {
						t.Fatalf("exit code %d, want 1; stderr:\n%s", code, errOut)
					}
					return
				}
				if code != 0 {
					t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
				}
				if got != want {
					t.Fatalf("C differs: %s", firstDiff(want, got))
				}
			})
		}
	})
}
