// Command langc is the bootstrap compiler for Lang. It compiles one source
// file to C and builds it with the system C compiler.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"langc/internal/langc"
)

const usage = `usage:
  langc build [-release] [-o output] [-emit-c file.c] file.lang
  langc run [-release] file.lang [args...]
  langc test file.lang
  langc emit file.lang      print the generated C`

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
		fs.Parse(rest)
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		src := fs.Arg(0)
		out := *output
		tmpDir := ""
		if out == "" {
			if cmd == "build" {
				out = strings.TrimSuffix(filepath.Base(src), ".lang")
			} else {
				d, err := os.MkdirTemp("", "langc-run")
				check(err)
				tmpDir = d
				out = filepath.Join(d, "program")
			}
		}
		err := langc.Build(src, out, langc.BuildOptions{Release: *release, Tests: cmd == "test", KeepC: *emitC})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if cmd == "build" {
			return
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
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		c, err := langc.CompileToC(rest[0], false)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(c)
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
