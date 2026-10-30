package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each .nra file and each package directory (with narra.toml) in
// testdata/declsdump is a case. Its dump must equal the golden file next to
// it: <name>.dump for <name>.nra or <name>/. A failed check has the error
// lines as its golden text. Subdirectories without narra.toml (traits/,
// imports/) belong to later nodes. Set NARRA_UPDATE=1 to rewrite the goldens.
func TestDeclDump(t *testing.T) {
	cases := declDumpCases(t, filepath.Join(testdata, "declsdump"))
	if len(cases) == 0 {
		t.Fatal("no declsdump cases")
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

// declDumpCases lists the .nra files and package directories in dir.
func declDumpCases(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(path, "narra.toml")); err == nil {
				cases = append(cases, path)
			}
		} else if strings.HasSuffix(e.Name(), ".nra") {
			cases = append(cases, path)
		}
	}
	return cases
}

// The dump of a sample prints each kind of line, and the declaration pass
// does not check bodies: a body error is not reported.
func TestDeclDumpSkipsBodies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.nra")
	src := "struct P {\n  x: int\n}\nconst N = 1\nfunction f(p: P): int {\n  return \"no\"\n}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := DeclDump(path)
	if !ok {
		t.Fatalf("body errors must not be reported:\n%s", got)
	}
	for _, want := range []string{
		"type struct module=\"body\" body.P\n  field x int\n",
		"func module=\"body\" f\n  param p body.P\n  result int\n",
		"const module=\"body\" N ?\n",
		"order module=\"body\" body.P\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dump lacks %q:\n%s", want, got)
		}
	}
}
