package narra

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every message format passed to an error function has a code.
func TestEveryFormatHasCode(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var fn string
			switch f := ce.Fun.(type) {
			case *ast.SelectorExpr:
				fn = f.Sel.Name
			case *ast.Ident:
				fn = f.Name
			}
			idx := 1
			switch fn {
			case "errorf", "fail", "typeErrorf":
			case "codeFor":
				idx = 0
			case "failFix", "errorfFix":
				idx = 2
			default:
				return true
			}
			if len(ce.Args) <= idx {
				return true
			}
			if bl, ok := ce.Args[idx].(*ast.BasicLit); ok && bl.Kind == token.STRING {
				s, _ := strconv.Unquote(bl.Value)
				if codeFor(s) == "E0000" {
					t.Errorf("%s: no code for %q", fset.Position(ce.Pos()), s)
				}
			}
			return true
		})
	}
}

// Every error test gets a coded diagnostic.
func TestErrorCodes(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(testdata, "errors", "*.nra"))
	pkgs, _ := filepath.Glob(filepath.Join(testdata, "packages", "errors", "*"))
	for _, src := range append(files, pkgs...) {
		t.Run(filepath.Base(src), func(t *testing.T) {
			ds, err := CheckDiagnostics(src)
			if err != nil {
				t.Fatal(err)
			}
			if len(ds) == 0 {
				t.Fatal("no diagnostics")
			}
			for _, d := range ds {
				if d.Code == "E0000" || !strings.HasPrefix(d.Code, "E") {
					t.Errorf("diagnostic without a code: %s", d.Message)
				}
				if !strings.Contains(d.JSON(), `"code":"`+d.Code+`"`) {
					t.Errorf("bad JSON: %s", d.JSON())
				}
			}
		})
	}
}

// applyFirstFix applies the first fix of the first diagnostic that has one and
// returns the new source, or "" when there is none.
func applyFirstFix(src string, ds Diagnostics) string {
	for _, d := range ds {
		if len(d.Fixes) > 0 {
			return ApplyFix(src, d.Fixes[0])
		}
	}
	return ""
}

// The TypeScript habit cases carry a fix that makes the file compile.
func TestHabitFixes(t *testing.T) {
	var files []string
	for _, f := range []string{"semicolon", "strict_eq", "method_call", "string_plus"} {
		files = append(files, filepath.Join(testdata, "errors", f+".nra"))
	}
	more, _ := filepath.Glob(filepath.Join(testdata, "diagnostics", "*.nra"))
	files = append(files, more...)
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			ds, err := CheckDiagnostics(src)
			if err != nil {
				t.Fatal(err)
			}
			text, _ := os.ReadFile(src)
			if want := strings.TrimPrefix(strings.SplitN(string(text), "\n", 2)[0], "// code: "); strings.HasPrefix(string(text), "// code: ") {
				if len(ds) == 0 || ds[0].Code != want {
					t.Fatalf("want code %s, got %v", want, ds)
				}
			}
			fixed := applyFirstFix(string(text), ds)
			if fixed == "" {
				t.Fatalf("no fix in %v", ds)
			}
			dir := t.TempDir()
			out := filepath.Join(dir, filepath.Base(src))
			if err := os.WriteFile(out, []byte(fixed), 0o644); err != nil {
				t.Fatal(err)
			}
			after, err := CheckDiagnostics(out)
			if err != nil {
				t.Fatal(err)
			}
			if len(after) > 0 {
				t.Fatalf("the fix does not make the file compile:\n%s\n%v", fixed, after)
			}
		})
	}
}

// habitCase is a TypeScript habit program. wantFix says whether the first
// diagnostic offers a fix; every fix that is offered must leave a clean file.
type habitCase struct {
	name, src, code string
	wantFix         bool
}

const habitMain = "\nexport function run(args: Array<string>): int {\n  return 0\n}\n"

var habitCases = []habitCase{
	{"class_fields", "class Point {\n  x: int\n}\n" + habitMain, "E0923", true},
	{"class_method", "class Point {\n  x: int\n  get(): int {\n    return 1\n  }\n}\n" + habitMain, "E0923", false},
	{"class_extends", "class Point extends Base {}\n" + habitMain, "E0923", false},
	{"interface_fields", "interface Point {\n  x: int\n}\n" + habitMain, "E0924", true},
	{"interface_method", "interface Shape {\n  area(): int\n}\n" + habitMain, "E0924", false},
	{"new_empty", "struct Point {}\n\nexport function run(args: Array<string>): int {\n  const p = new Point()\n  return 0\n}\n", "E0925", false},
	{"new_fields", "struct Point { x: int }\n\nexport function run(args: Array<string>): int {\n  const p = new Point()\n  return 0\n}\n", "E0925", false},
	{"new_function", "function Point(): int {\n  return 1\n}\n\nexport function run(args: Array<string>): int {\n  const p = new Point()\n  return 0\n}\n", "E0925", false},
	{"throw_result", "function fail(): Result<int, string> {\n  throw \"bad\"\n}\n" + habitMain, "E0927", true},
	{"throw_plain", "function fail(): int {\n  throw \"bad\"\n}\n" + habitMain, "E0927", false},
	{"throw_in_closure", "function fail(): Result<int, string> {\n  const f = (): int => {\n    throw \"bad\"\n  }\n  return Ok(f())\n}\n" + habitMain, "E0927", false},
	{"throw_mid_line", "function fail(): Result<int, string> {\n  const x = 1 + throw \"bad\"\n  return Ok(x)\n}\n" + habitMain, "E0927", false},
	{"throw_call", "function fail(): Result<int, string> {\n  if (true) { throw \"a\" }\n  return Ok(1)\n}\n" + habitMain, "E0927", true},
	{"try", "export function run(args: Array<string>): int {\n  try {\n    return 1\n  } catch (e) {\n    return 2\n  }\n}\n", "E0926", false},
	{"this_use", "export function run(args: Array<string>): int {\n  return this.count\n}\n", "E0928", false},
	{"this_const", "export function run(args: Array<string>): int {\n  const this = 1\n  return 0\n}\n", "E0928", false},
	{"this_param", "function f(this: int): int {\n  return 1\n}\n" + habitMain, "E0928", false},
	{"for_in_array", "export function run(args: Array<string>): int {\n  const xs = [1, 2]\n  for (const k in xs) {\n    print(k)\n  }\n  return 0\n}\n", "E0031", false},
}

// The habit cases give the right code, offer a fix only where it is right,
// and every offered fix leaves a file that checks clean.
func TestHabitCases(t *testing.T) {
	for _, tc := range habitCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "habit.nra")
			if err := os.WriteFile(path, []byte(tc.src), 0o644); err != nil {
				t.Fatal(err)
			}
			ds, err := CheckDiagnostics(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(ds) == 0 || ds[0].Code != tc.code {
				t.Fatalf("want code %s, got %v", tc.code, ds)
			}
			if has := len(ds[0].Fixes) > 0; has != tc.wantFix {
				t.Fatalf("want fix %v, got %s", tc.wantFix, ds[0].JSON())
			}
			for i, fix := range ds[0].Fixes {
				fixed := filepath.Join(dir, "fixed.nra")
				if err := os.WriteFile(fixed, []byte(ApplyFix(tc.src, fix)), 0o644); err != nil {
					t.Fatal(err)
				}
				after, err := CheckDiagnostics(fixed)
				if err != nil || len(after) > 0 {
					t.Fatalf("fix %d leaves errors:\n%s\n%v %v", i, ApplyFix(tc.src, fix), after, err)
				}
			}
		})
	}
}

// A generic function is checked once per instance; its error is reported once.
func TestGenericErrorOnce(t *testing.T) {
	ds, err := CheckDiagnostics(filepath.Join(testdata, "diagnostics", "ts_generic_concat.nra"))
	if err != nil || len(ds) != 1 {
		t.Fatalf("want one diagnostic, got %v, %v", ds, err)
	}
	for _, f := range ds[0].Fixes {
		for _, e := range f.Edits {
			if e.End.Line < e.Start.Line || (e.End.Line == e.Start.Line && e.End.Col < e.Start.Col) {
				t.Fatalf("invalid range in %s", ds[0].JSON())
			}
		}
	}
}

// Each code of the TypeScript habit diagnostics has a long explanation.
func TestHabitExplain(t *testing.T) {
	for _, code := range []string{"E0923", "E0924", "E0925", "E0926", "E0927", "E0928"} {
		text, err := Explain(code)
		if err != nil || !strings.Contains(text, "Narra has no") {
			t.Errorf("explain %s: %q, %v", code, text, err)
		}
	}
}

func TestExplain(t *testing.T) {
	if _, err := Explain("e0003"); err != nil {
		t.Fatal(err)
	}
	text, err := Explain("E0003")
	if err != nil || !strings.Contains(text, "semicolons") {
		t.Fatalf("explain E0003: %q, %v", text, err)
	}
	if _, err := Explain("E9999"); err == nil {
		t.Fatal("expected an error for an unknown code")
	}
}
