package narra

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// portedImportErrors are the error formats of the import passes that the
// Narra checker reports (selfhost/src/scopes.nra, resolve.nra's
// qualifiedType). TestSelfhostImports accepts them on top of the
// declaration and trait formats.
var portedImportErrors = []string{
	"unknown module %q (this compiler supports %s)",
	"unknown module %q; package %s has no file src/%s.nra",
	"module %q has no export %q",
	"%q is not exported by module %q; add `export` to its declaration",
	"%s is not a type",
	"two imports use the alias %s; give one of them another alias",
	"the entry function must be `export function %s(args: Array<string>): int`",
	"the import alias %s is also a type name; rename it, for example `import * as %s from %q`",
	"imported name %q is already declared in this module; rename the import with `as`",
	"imported name %q is already imported from another module; rename one with `as`",
	// A declaration error that selfhost-check-bodies ported with them.
	"enum %s declares the constructor %q twice",
}

var portedImportRes = func() []*regexp.Regexp {
	var out []*regexp.Regexp
	all := append(append(append([]string{}, portedDeclErrors...), portedTraitErrors...), portedImportErrors...)
	for _, f := range all {
		re := regexp.QuoteMeta(f)
		re = strings.NewReplacer("%q", `"(?:[^"\\]|\\.)*"`, "%s", `.*`, "%d", `-?[0-9]+`).Replace(re)
		out = append(out, regexp.MustCompile("^"+re+"$"))
	}
	return out
}()

// unportedImportError returns the first error line of a Go dump that the
// Narra declaration passes do not report, or "".
func unportedImportError(dump string) string {
	for _, line := range strings.Split(strings.TrimSuffix(dump, "\n"), "\n") {
		m := declErrorLine.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		ported := false
		for _, re := range portedImportRes {
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

// importInputs are declInputs plus the package cases of
// testdata/declsdump/imports.
func importInputs(t *testing.T) []string {
	inputs := declInputs(t)
	for _, p := range declDumpCases(t, filepath.Join(testdata, "declsdump", "imports")) {
		if !strings.HasSuffix(p, ".nra") {
			inputs = append(inputs, p)
		}
	}
	return inputs
}

// TestSelfhostImports checks that `narrac decls <path>` prints the same
// dump as `narra decls`, every line kind included, or the same errors, on
// every input of importInputs: single files, every package under
// testdata/packages (errors included), selfhost/ and the import cases.
// Skipped inputs are logged with the reason.
func TestSelfhostImports(t *testing.T) {
	exe := buildSelfhost(t)
	for _, path := range importInputs(t) {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			want, ok := DeclDump(path)
			if _, err := Load(path); err != nil {
				t.Skip("does not load")
			}
			if !ok {
				if line := unportedImportError(want); line != "" {
					t.Skip("unported error: " + line)
				}
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
			if want != out {
				t.Fatalf("dump differs: %s", firstDiff(want, out))
			}
		})
	}
}

// TestDeclDumpImports compares the goldens in testdata/declsdump/imports
// with the Go dump. NARRA_UPDATE=1 rewrites them.
func TestDeclDumpImports(t *testing.T) {
	cases := declDumpCases(t, filepath.Join(testdata, "declsdump", "imports"))
	if len(cases) == 0 {
		t.Fatal("no declsdump/imports cases")
	}
	for _, path := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got, ok := DeclDump(path)
			if !ok && !strings.HasPrefix(got, "error ") {
				t.Fatalf("a failed dump must start with `error `: %q", got)
			}
			golden := strings.TrimSuffix(path, ".nra") + ".dump"
			if os.Getenv("NARRA_UPDATE") != "" {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("dump differs: %s", firstDiff(string(want), got))
			}
		})
	}
}

// stdFixed is one "fixed" builtin of the Narra registry: a Go Check closure
// that calls checkFixed, and the signature the Narra table gives it.
type stdFixed struct {
	module, key, name string
	params            []string
	result            string
}

// splitFixed removes the `fixed param` and `fixed result` lines from the
// Narra registry dump and returns the signatures they give.
func splitFixed(dump string) (string, []stdFixed) {
	var b strings.Builder
	var out []stdFixed
	head := regexp.MustCompile(`^builtin module="([^"]*)" export=(\S+) (\S+)$`)
	var cur *stdFixed
	for _, line := range strings.SplitAfter(dump, "\n") {
		text := strings.TrimSuffix(line, "\n")
		if m := head.FindStringSubmatch(text); m != nil {
			out = append(out, stdFixed{module: m[1], key: m[2], name: m[3]})
			cur = &out[len(out)-1]
		}
		switch {
		case strings.HasPrefix(text, "  fixed param "):
			cur.params = append(cur.params, strings.TrimPrefix(text, "  fixed param "))
		case strings.HasPrefix(text, "  fixed result "):
			cur.result = strings.TrimPrefix(text, "  fixed result ")
		default:
			b.WriteString(line)
		}
	}
	var fixed []stdFixed
	for _, f := range out {
		if f.result != "" {
			fixed = append(fixed, f)
		}
	}
	return b.String(), fixed
}

// TestSelfhostBuiltins checks the Narra std registry (selfhost/src/stdlib.nra)
// against the Go registry. `narrac builtins` must print BuiltinDump, apart
// from the signatures of the "fixed" builtins, which Go keeps inside Check
// closures. Each of those is checked by compiling a call with the Narra
// signature: the arguments have the parameter types and the call's value is
// returned as the result type. A std function added only in Go fails here
// until stdlib.nra lists it.
func TestSelfhostBuiltins(t *testing.T) {
	exe := buildSelfhost(t)
	out, errOut, code := runNarracRaw(t, exe, "builtins")
	if code != 0 {
		t.Fatalf("exit code %d; stderr:\n%s", code, errOut)
	}
	rest, fixed := splitFixed(out)
	if want := BuiltinDump(); rest != want {
		t.Fatalf("registry differs from Go (narra first): %s", firstDiff(want, rest))
	}
	if len(fixed) == 0 {
		t.Fatal("no fixed builtins in the Narra registry")
	}
	for _, f := range fixed {
		t.Run(f.name, func(t *testing.T) {
			var params, args []string
			for i, p := range f.params {
				params = append(params, fmt.Sprintf("a%d: %s", i, p))
				args = append(args, fmt.Sprintf("a%d", i))
			}
			src := fmt.Sprintf("import * as M from %q\n\nfunction probe(%s): %s {\n  return M.%s(%s)\n}\n\nexport function run(args: Array<string>): int {\n  return 0\n}\n",
				f.module, strings.Join(params, ", "), f.result, f.key, strings.Join(args, ", "))
			path := filepath.Join(t.TempDir(), "probe.nra")
			if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
				t.Fatal(err)
			}
			pkg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := CheckPackage(pkg); err != nil {
				t.Fatalf("the Narra signature of %s does not check in Go: %v\n%s", f.name, err, src)
			}
		})
	}
}

// The probe must fail when the Narra signature is wrong.
func TestSelfhostBuiltinsProbeFails(t *testing.T) {
	src := "import * as M from \"std/string\"\n\nfunction probe(a0: string, a1: int): string {\n  return M.join(a0, a1)\n}\n"
	path := filepath.Join(t.TempDir(), "probe.nra")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CheckPackage(pkg); err == nil {
		t.Fatal("a wrong signature was accepted")
	}
}

// stdModuleOrder (decldump.go) must list the modules as moduleList does.
func TestStdModuleOrder(t *testing.T) {
	paths := stdModuleOrder()
	got := strings.Join(paths, "")
	if len(paths) >= 2 {
		got = strings.Join(paths[:len(paths)-1], ", ") + " and " + paths[len(paths)-1]
	}
	if got != moduleList() {
		t.Fatalf("stdModuleOrder %q, moduleList %q", got, moduleList())
	}
}
