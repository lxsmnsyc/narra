package langc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CompileToC parses, checks and emits C for one source file.
func CompileToC(path string, tests bool) (string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	name := filepath.Base(path)
	file, err := Parse(string(src))
	if err != nil {
		return "", prefix(name, err)
	}
	prog, err := Check(file)
	if err != nil {
		return "", prefix(name, err)
	}
	if !tests && prog.Entry == nil {
		return "", fmt.Errorf("%s: no entry function; declare `export function run(args: Array<string>): int`", name)
	}
	return Emit(prog, name, tests), nil
}

func prefix(name string, err error) error {
	return fmt.Errorf("%s:%s", name, err.Error())
}

type BuildOptions struct {
	Release bool
	Tests   bool
	KeepC   string // when set, the generated C is also written here
	CC      string
	CFlags  []string
}

// Build compiles a source file to a native executable.
func Build(path, output string, opts BuildOptions) error {
	csrc, err := CompileToC(path, opts.Tests)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "langc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	cfile := filepath.Join(dir, "out.c")
	if opts.KeepC != "" {
		cfile = opts.KeepC
	}
	if err := os.WriteFile(cfile, []byte(csrc), 0o644); err != nil {
		return err
	}
	cc := opts.CC
	if cc == "" {
		cc = os.Getenv("CC")
	}
	if cc == "" {
		cc = "cc"
	}
	args := append([]string{"-std=c11", "-w"}, opts.CFlags...)
	args = append(args, "-o", output, cfile, "-lm")
	if opts.Release {
		args = append([]string{"-O2", "-DLANG_RELEASE"}, args...)
	} else {
		args = append([]string{"-O0", "-g"}, args...)
	}
	cmd := exec.Command(cc, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler failed: %w", err)
	}
	return nil
}
