package narra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CompileToC parses, checks and emits C for a source file or a package
// directory with one binary.
func CompileToC(path string, tests bool) (string, error) {
	return CompileBinary(path, "", tests)
}

// CompileBinary emits C for the binary bin of the package at path. An empty
// bin selects the only binary. In test mode the C runs every test block.
func CompileBinary(path, bin string, tests bool) (string, error) {
	pkg, err := Load(path)
	if err != nil {
		return "", FormatError(err)
	}
	prog, err := CheckPackage(pkg)
	if err != nil {
		return "", FormatError(err)
	}
	if tests {
		return Emit(prog, nil, true), nil
	}
	b, err := pkg.binary(bin)
	if err != nil {
		return "", err
	}
	entry := prog.Entries[b.Name]
	if entry == nil {
		if pkg.single {
			return "", fmt.Errorf("%s: no entry function; declare `export function run(args: Array<string>): int`", pkg.Modules[0].Files[0].Path)
		}
		mod, fn := pkg.entryModule(b.Entry)
		return "", fmt.Errorf("narra.toml: binary %q: no entry function %q; declare `export function %s(args: Array<string>): int` in module %q", b.Name, b.Entry, fn, mod)
	}
	return Emit(prog, entry, false), nil
}

type BuildOptions struct {
	Release bool
	Tests   bool
	Binary  string // the binary to build from a package; empty for the only one
	KeepC   string // when set, the generated C is also written here
	CC      string
	CFlags  []string
}

// Build compiles a source file or a package binary to a native executable.
func Build(path, output string, opts BuildOptions) error {
	csrc, err := CompileBinary(path, opts.Binary, opts.Tests)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "narra")
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
		args = append([]string{"-O2", "-DNARRA_RELEASE"}, args...)
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
