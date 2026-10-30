package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each directory in testdata/pkgdump is a package (or a broken one). Its dump
// must equal the golden file next to it, `<name>.dump`. A failed load has the
// error lines as its golden text. Set NARRA_UPDATE=1 to rewrite the goldens.
func TestPkgDump(t *testing.T) {
	dirs, _ := filepath.Glob(filepath.Join(testdata, "pkgdump", "*"))
	count := 0
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		count++
		t.Run(filepath.Base(dir), func(t *testing.T) {
			got, ok := PkgDump(dir)
			if !ok && !strings.HasPrefix(got, "error ") {
				t.Fatalf("a failed dump must start with `error `: %q", got)
			}
			golden := dir + ".dump"
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
	if count == 0 {
		t.Fatal("no pkgdump cases")
	}
}

// The sample packages dump their modules in load order, imports first.
func TestPkgDumpOrder(t *testing.T) {
	got, ok := PkgDump(filepath.Join(testdata, "packages", "app"))
	if !ok {
		t.Fatal(got)
	}
	list := strings.Index(got, `module "app/util/list"`)
	text := strings.Index(got, `module "app/text"`)
	cli := strings.Index(got, `module "app/cli"`)
	if !(list >= 0 && list < text && text < cli) {
		t.Fatalf("modules are not in load order:\n%s", got)
	}
}
