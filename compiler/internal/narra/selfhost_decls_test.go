package narra

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// declKinds are the line kinds of the declaration dump that
// selfhost-check-decls owns (decldump.go). TestSelfhostDecls compares only
// these; later nodes compare their own kinds.
var declKinds = map[string]bool{
	"type": true, "field": true, "variant": true, "base": true, "generic": true,
	"func": true, "param": true, "result": true, "const": true, "order": true,
	"error": true,
}

// portedDeclErrors are the error formats of the declaration passes that the
// Narra checker reports. An input whose Go dump has any other error comes
// from a stage this node does not port (traits, imports, default values,
// which the body checker checks) and is skipped.
var portedDeclErrors = []string{
	"type %q is declared twice",
	"function %q is declared twice",
	"constant %q is declared twice",
	"field %q is declared twice",
	"distinct type %s cannot be defined in terms of itself",
	"a default for a field whose type uses a type parameter is not supported by this compiler yet",
	"generic type %s contains %s, so each instance needs a larger one without end; every instance needs its own C type, so polymorphic recursion is not supported (use the same type arguments, or an enum or Box)",
	"type %s contains itself; put the recursive field in an enum, Box or array",
	"Box takes one type argument",
	"Map takes two type arguments",
	"Array takes one type argument",
	"there is no `number`; use `int` or `f64`",
	"use `bool`",
	"%s takes %d type arguments",
	"there is no Tuple type; write a tuple type as #[A, B]",
	"type %s is not generic",
	"unknown type %q",
	"a tuple type needs at least two element types, found %d",
	"unknown module %q in type %s; import it with `import * as %s from \"...\"`",
}

var portedDeclRes = func() []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, f := range portedDeclErrors {
		re := regexp.QuoteMeta(f)
		re = strings.NewReplacer("%q", `"(?:[^"\\]|\\.)*"`, "%s", `.*`, "%d", `-?[0-9]+`).Replace(re)
		out = append(out, regexp.MustCompile("^"+re+"$"))
	}
	return out
}()

var declErrorLine = regexp.MustCompile(`^error [^:]*:[0-9]+:[0-9]+: (.*)$`)

// unportedDeclError returns the first error line of a Go dump that this node
// does not port, or "".
func unportedDeclError(dump string) string {
	for _, line := range strings.Split(strings.TrimSuffix(dump, "\n"), "\n") {
		m := declErrorLine.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		ported := false
		for _, re := range portedDeclRes {
			if re.MatchString(m[1]) {
				ported = true
				break
			}
		}
		if !ported {
			return line
		}
	}
	return ""
}

// declLines keeps the lines whose kind (first word) is in declKinds.
func declLines(dump string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(dump, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && declKinds[fields[0]] {
			b.WriteString(line)
		}
	}
	return b.String()
}

// declInputs lists the inputs of TestSelfhostDecls: every .nra file under
// testdata/ outside packages/ and pkgdump/, every package directory, and
// selfhost/.
func declInputs(t *testing.T) []string {
	t.Helper()
	var inputs []string
	err := filepath.WalkDir(testdata, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "packages" || d.Name() == "pkgdump") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(path, ".nra") {
			inputs = append(inputs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return append(inputs, pkgInputsDirs(t)...)
}

// pkgInputsDirs is pkgInputs without its single files.
func pkgInputsDirs(t *testing.T) []string {
	var dirs []string
	for _, p := range pkgInputs(t) {
		if !strings.HasSuffix(p, ".nra") {
			dirs = append(dirs, p)
		}
	}
	return dirs
}

// declSkip says why an input is not compared, or "". Inputs that fail to
// load are TestSelfhostModules' business. A package whose modules import
// each other needs module scopes (selfhost-check-imports).
func declSkip(path, goDump string, goOK bool) string {
	pkg, err := Load(path)
	if err != nil {
		return "does not load"
	}
	for _, m := range pkg.Modules {
		for _, f := range m.Files {
			for _, imp := range f.Imports {
				if strings.HasPrefix(imp.Path, pkg.Name+"/") {
					return "imports a module of its package (selfhost-check-imports)"
				}
			}
		}
	}
	if !goOK {
		if line := unportedDeclError(goDump); line != "" {
			return "unported error: " + line
		}
	}
	return ""
}

// TestSelfhostDecls checks that `narrac decls <path>` prints the same
// declaration lines as `narra decls` (DeclDump), or the same errors, on
// every input of declInputs. Skipped inputs are logged with the reason.
func TestSelfhostDecls(t *testing.T) {
	exe := buildSelfhost(t)
	for _, path := range declInputs(t) {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			want, ok := DeclDump(path)
			if reason := declSkip(path, want, ok); reason != "" {
				t.Skip(reason)
			}
			out, errOut, code := runNarracRaw(t, exe, "decls", path)
			if !ok {
				if code != 1 || out != "" {
					t.Fatalf("exit code %d, want 1; stdout:\n%s\nstderr:\n%s", code, out, errOut)
				}
				if errOut != want {
					t.Fatalf("errors differ: %s", firstDiff(want, errOut))
				}
				return
			}
			if code != 0 {
				t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
			}
			if w, g := declLines(want), declLines(out); w != g {
				t.Fatalf("dump differs: %s", firstDiff(w, g))
			}
		})
	}
}

// runNarracRaw runs narrac under the leak check and returns stdout, stderr
// and the exit code.
func runNarracRaw(t *testing.T, exe string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
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

// The Narra checker parses its own copy of the prelude; it must be the same
// text as preludeSrc.
func TestSelfhostPrelude(t *testing.T) {
	src, err := os.ReadFile("../../../selfhost/src/declare.nra")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "const PRELUDE = `"+preludeSrc+"`") {
		t.Fatal("selfhost/src/declare.nra PRELUDE differs from preludeSrc in check.go")
	}
}

// TestSelfhostNodeIds runs `narrac pkg --ids` on the inputs of
// TestSelfhostModuleIds: besides the file ranges, the walker in
// selfhost/src/idcheck.nra checks that every node id is unique and inside
// its file's range. The dup subtest copies one id over another with
// `--ids-dup` and the check must fail.
func TestSelfhostNodeIds(t *testing.T) {
	exe := buildSelfhost(t)
	t.Run("dup", func(t *testing.T) {
		out, errOut, code := runNarracRaw(t, exe, "pkg", "--ids-dup", filepath.Join(testdata, "generics.nra"))
		if code != 1 || !strings.Contains(errOut, "is used twice") {
			t.Fatalf("a duplicate id was not found: exit %d\n%s%s", code, out, errOut)
		}
	})
	for _, path := range pkgInputs(t) {
		if _, ok := PkgDump(path); !ok {
			continue
		}
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			out, errOut, code := runNarracRaw(t, exe, "pkg", "--ids", path)
			if code != 0 || !strings.HasPrefix(out, "ids ok ") {
				t.Fatalf("exit %d\n%s%s", code, out, errOut)
			}
		})
	}
}

// The skip rule must accept the ported formats and reject others.
func TestUnportedDeclError(t *testing.T) {
	ported := "error a.nra:1:2: unknown type \"X\"\nerror a.nra:3:4: Option takes 1 type arguments\n"
	if got := unportedDeclError(ported); got != "" {
		t.Fatalf("ported errors rejected: %q", got)
	}
	trait := "error a.nra:1:2: unknown trait \"X\"\n"
	if got := unportedDeclError(trait); got == "" {
		t.Fatal("a trait error was accepted")
	}
}
