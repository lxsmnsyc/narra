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

func TestExplain(t *testing.T) {
	text, err := Explain("E0003")
	if err != nil || !strings.Contains(text, "semicolons") {
		t.Fatalf("explain E0003: %q, %v", text, err)
	}
	if _, err := Explain("E9999"); err == nil {
		t.Fatal("expected an error for an unknown code")
	}
}
