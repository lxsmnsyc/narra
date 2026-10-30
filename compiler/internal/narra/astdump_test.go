package narra

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// astKinds is every node kind the dump can print. The goldens together must
// contain each of them.
var astKinds = []string{
	"file", "import", "func", "const", "struct", "enum", "distinct", "test",
	"trait", "impl", "variant", "field", "param", "type", "block", "var",
	"assign", "exprstmt", "return", "if", "while", "for", "break", "continue",
	"int", "float", "string", "char", "bool", "template", "ident", "field-expr",
	"is", "index", "map", "call", "arg", "unary", "binary", "structlit", "init",
	"array", "match", "arm", "arrow", "pat-wildcard", "pat-name", "pat-ctor",
	"pat-struct", "fieldpat", "pat-tuple", "pat-lit",
}

// TestASTDump dumps every .nra file under testdata/ (no panic, same output
// twice) and compares the files in testdata/astdump/ with their .ast goldens.
// NARRA_UPDATE_GOLDEN=1 rewrites the goldens.
func TestASTDump(t *testing.T) {
	err := filepath.WalkDir(testdata, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".nra") {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		t.Run(path, func(t *testing.T) {
			first, ok1 := ASTDump(string(src))
			second, ok2 := ASTDump(string(src))
			if first != second || ok1 != ok2 {
				t.Fatalf("dump differs between two runs")
			}
			if !ok1 && !strings.HasPrefix(first, "error ") {
				t.Fatalf("failed dump does not start with `error `: %q", first)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	files, _ := filepath.Glob(filepath.Join(testdata, "astdump", "*.nra"))
	if len(files) == 0 {
		t.Fatal("no files in testdata/astdump")
	}
	seen := map[string]bool{}
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := ASTDump(string(src))
		golden := strings.TrimSuffix(path, ".nra") + ".ast"
		if os.Getenv("NARRA_UPDATE_GOLDEN") != "" {
			if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Errorf("%s: dump differs from %s (rerun with NARRA_UPDATE_GOLDEN=1 to update)", path, golden)
		}
		for _, line := range strings.Split(got, "\n") {
			line = strings.TrimLeft(line, " ")
			if line == "" || line[0] == '.' {
				continue
			}
			seen[strings.Fields(line)[0]] = true
		}
	}
	for _, k := range astKinds {
		if !seen[k] {
			t.Errorf("no golden contains a %q node", k)
		}
	}
}
