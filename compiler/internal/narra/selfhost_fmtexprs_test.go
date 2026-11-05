package narra

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// fmtExprsBuild is a copy of selfhost/ whose main.nra is the test driver
// testdata/fmt/expr_driver.txt. It formats the top-level consts of a file
// with the real expression, pattern and type printers, so the goldens run
// without the CLI of fmt-cli.
var fmtExprsBuild struct {
	once sync.Once
	dir  string
	exe  string
	err  error
}

func buildFmtExprs(t *testing.T) string {
	t.Helper()
	fmtExprsBuild.once.Do(func() {
		dir, err := os.MkdirTemp("", "narrac-fmtexprs")
		if err != nil {
			fmtExprsBuild.err = err
			return
		}
		fmtExprsBuild.dir = dir
		root := filepath.Join(dir, "selfhost")
		err = filepath.WalkDir("../../../selfhost", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel("../../../selfhost", path)
			target := filepath.Join(root, rel)
			if entry.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if rel == filepath.Join("src", "main.nra") {
				data, err = os.ReadFile("../../testdata/fmt/expr_driver.txt")
				if err != nil {
					return err
				}
			}
			return os.WriteFile(target, data, 0o644)
		})
		if err != nil {
			fmtExprsBuild.err = err
			return
		}
		fmtExprsBuild.exe = filepath.Join(dir, "fmtexprs")
		fmtExprsBuild.err = Build(root, fmtExprsBuild.exe, modes[0].opts)
	})
	if fmtExprsBuild.err != nil {
		t.Fatalf("build fmt-exprs driver: %v", fmtExprsBuild.err)
	}
	return fmtExprsBuild.exe
}

var fmtCaseOptions = regexp.MustCompile(`_w([0-9]+)(_none)?$`)

// fmtCaseArgs reads the options from the case name: `_w40` is a print width
// of 40, and `_none` after it turns trailing commas off.
func fmtCaseArgs(name string) []string {
	match := fmtCaseOptions.FindStringSubmatch(name)
	if match == nil {
		return nil
	}
	args := []string{"--width", match[1]}
	if match[2] != "" {
		args = append(args, "--trailing", "none")
	}
	return args
}

// TestSelfhostFmtExprs formats every testdata/fmt/{expr,type,pattern}_*.nra
// with the driver and compares the output with the .fmt file, then formats the
// output again: it must not change (idempotence).
func TestSelfhostFmtExprs(t *testing.T) {
	exe := buildFmtExprs(t)
	var inputs []string
	for _, prefix := range []string{"expr", "type", "pattern"} {
		found, err := filepath.Glob("../../testdata/fmt/" + prefix + "_*.nra")
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, found...)
	}
	if len(inputs) == 0 {
		t.Fatal("no golden cases")
	}
	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".nra")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if os.Getenv("FMT_UPDATE") != "" {
				// FMT_UPDATE=1 rewrites the goldens from the driver's output; read the diff.
				got, _ := exec.Command(exe, append([]string{input}, fmtCaseArgs(name)...)...).Output()
				if err := os.WriteFile(strings.TrimSuffix(input, ".nra")+".fmt", got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(strings.TrimSuffix(input, ".nra") + ".fmt")
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{input}, fmtCaseArgs(name)...)
			runNarrac(t, exe, args, string(want), "", 0)
			// The second pass reads the first pass's output.
			again := filepath.Join(t.TempDir(), "again.nra")
			if err := os.WriteFile(again, want, 0o644); err != nil {
				t.Fatal(err)
			}
			args = append([]string{again}, fmtCaseArgs(name)...)
			runNarrac(t, exe, args, string(want), "", 0)
		})
	}
}
