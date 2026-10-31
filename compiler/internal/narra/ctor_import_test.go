package narra

import (
	"os"
	"path/filepath"
	"testing"
)

// The fix of `import { Teacher }` imports the enum, and the package then
// compiles.
func TestCtorImportFix(t *testing.T) {
	src := filepath.Join(testdata, "packages", "errors", "ctor_import")
	dir := t.TempDir()
	for _, f := range []string{"narra.toml", "src/people.nra", "src/main.nra"} {
		text, err := os.ReadFile(filepath.Join(src, f))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), text, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ds, err := CheckDiagnostics(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ds) == 0 || ds[0].Code != "E0425" || len(ds[0].Fixes) == 0 {
		t.Fatalf("want E0425 with a fix, got %v", ds)
	}
	main := filepath.Join(dir, "src", "main.nra")
	text, _ := os.ReadFile(main)
	fixed := ApplyFix(string(text), ds[0].Fixes[0])
	if err := os.WriteFile(main, []byte(fixed), 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := CheckDiagnostics(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) > 0 {
		t.Fatalf("the fix does not make the package compile:\n%s\n%v", fixed, after)
	}
}
