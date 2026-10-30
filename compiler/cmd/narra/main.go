// Command narra is the bootstrap compiler for Narra. It compiles a source
// file or a package directory to C and builds it with the system C compiler.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"narra/internal/narra"
)

const usage = `usage:
  narra build [-release] [-bin name] [-o output] [-emit-c file.c] file.nra|dir
  narra run [-release] [-bin name] file.nra|dir [args...]
  narra test file.nra|dir
  narra check [--json] file.nra|dir        report errors; --json prints one object per line
  narra explain E0000                      explain a diagnostic code
  narra emit [-bin name] file.nra|dir      print the generated C

A dir is a package: a directory with narra.toml and src/. Without -bin,
build makes every [[binary]] of the package; run and emit need the package
to have one binary.`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, rest := os.Args[1], os.Args[2:]
	switch cmd {
	case "build", "run", "test":
		fs := flag.NewFlagSet(cmd, flag.ExitOnError)
		release := fs.Bool("release", false, "optimize and wrap integer overflow")
		output := fs.String("o", "", "output executable")
		emitC := fs.String("emit-c", "", "also write the generated C to this file")
		bin := fs.String("bin", "", "the package binary to build or run")
		fs.Parse(rest)
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		src := fs.Arg(0)
		opts := narra.BuildOptions{Release: *release, Tests: cmd == "test", KeepC: *emitC, Binary: *bin}
		if cmd == "build" {
			build(src, *output, opts)
			return
		}
		out := *output
		tmpDir := ""
		if out == "" {
			d, err := os.MkdirTemp("", "narra-run")
			check(err)
			tmpDir = d
			out = filepath.Join(d, "program")
		}
		err := narra.Build(src, out, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		run := exec.Command(out, fs.Args()[1:]...)
		run.Stdin, run.Stdout, run.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = run.Run()
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		check(err)
	case "emit":
		fs := flag.NewFlagSet(cmd, flag.ExitOnError)
		bin := fs.String("bin", "", "the package binary to emit")
		fs.Parse(rest)
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		c, err := narra.CompileBinary(fs.Arg(0), *bin, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(c)
	case "check":
		os.Exit(narra.CheckCommand(rest))
	case "explain":
		os.Exit(narra.ExplainCommand(rest))
	case "lex":
		// Hidden: print the tokens of a file, for the self-hosted lexer test.
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr, "usage: narra lex file.nra")
			os.Exit(2)
		}
		src, err := os.ReadFile(rest[0])
		check(err)
		dump, ok := narra.LexDump(string(src))
		fmt.Print(dump)
		if !ok {
			os.Exit(1)
		}
	case "parse":
		// Hidden: print the syntax tree of a file, for the self-hosted parser test.
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr, "usage: narra parse file.nra")
			os.Exit(2)
		}
		src, err := os.ReadFile(rest[0])
		check(err)
		dump, ok := narra.ASTDump(string(src))
		fmt.Print(dump)
		if !ok {
			os.Exit(1)
		}
	case "pkg":
		// Hidden: print the loaded package, for the self-hosted loader test.
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr, "usage: narra pkg path")
			os.Exit(2)
		}
		dump, ok := narra.PkgDump(rest[0])
		fmt.Print(dump)
		if !ok {
			os.Exit(1)
		}
	case "decls":
		// Hidden: print the checked declarations, for the self-hosted checker test.
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr, "usage: narra decls path")
			os.Exit(2)
		}
		dump, ok := narra.DeclDump(rest[0])
		fmt.Print(dump)
		if !ok {
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
}

// build makes one executable for a file, or one per selected binary for a
// package. Executables are named after the file or the binary.
func build(src, output string, opts narra.BuildOptions) {
	pkg, err := narra.Load(src)
	check(narra.FormatError(err))
	var names []string
	if opts.Binary != "" {
		names = []string{opts.Binary}
	} else {
		for _, b := range pkg.Binaries {
			names = append(names, b.Name)
		}
	}
	info, err := os.Stat(src)
	check(err)
	if !info.IsDir() {
		out := output
		if out == "" {
			out = strings.TrimSuffix(filepath.Base(src), ".nra")
		}
		check(narra.Build(src, out, opts))
		return
	}
	if len(names) == 0 {
		check(fmt.Errorf("narra.toml: package %s has no [[binary]]; it is a library", pkg.Name))
	}
	if output != "" && len(names) > 1 {
		check(fmt.Errorf("-o needs one binary; the package has %s, choose one with -bin", strings.Join(names, ", ")))
	}
	for _, name := range names {
		out := output
		if out == "" {
			out = name
		}
		opts.Binary = name
		check(narra.Build(src, out, opts))
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
