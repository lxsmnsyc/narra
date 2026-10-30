package narra

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Each directory in testdata/packages with a run.out builds its only binary
// and must print run.out, under the sanitizers and the leak check.
func TestPackages(t *testing.T) {
	dirs, _ := filepath.Glob(filepath.Join(testdata, "packages", "*", "run.out"))
	if len(dirs) == 0 {
		t.Fatal("no sample packages")
	}
	for _, out := range dirs {
		dir := filepath.Dir(out)
		want, _ := os.ReadFile(out)
		for _, m := range modes {
			t.Run(filepath.Base(dir)+"/"+m.name, func(t *testing.T) {
				got, stderr, code := runProgram(t, dir, m.opts)
				if code != 0 {
					t.Fatalf("exit code %d\nstderr:\n%s", code, stderr)
				}
				if got != string(want) {
					t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
				}
			})
		}
		t.Run(filepath.Base(dir)+"/tests", func(t *testing.T) {
			opts := modes[0].opts
			opts.Tests = true
			got, stderr, code := runProgram(t, dir, opts)
			if code != 0 || !strings.Contains(got, " passed, 0 failed") {
				t.Fatalf("exit code %d\nstdout:\n%s\nstderr:\n%s", code, got, stderr)
			}
		})
	}
}

// Each directory in testdata/packages/errors must fail to compile. The first
// line of its narra.toml, `# error: text`, gives text the error must contain.
func TestPackageErrors(t *testing.T) {
	dirs, _ := filepath.Glob(filepath.Join(testdata, "packages", "errors", "*"))
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			text, err := os.ReadFile(filepath.Join(dir, "narra.toml"))
			if err != nil {
				t.Fatal(err)
			}
			first := strings.SplitN(string(text), "\n", 2)[0]
			want := strings.TrimSpace(strings.TrimPrefix(first, "# error:"))
			_, err = CompileToC(dir, false)
			if err == nil {
				t.Fatalf("expected an error containing %q", want)
			}
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("want error containing %q, got:\n%v", want, err)
			}
		})
	}
}

func TestManifest(t *testing.T) {
	pkg, err := parseManifest(`
# comment
[package]
name = "app"   # trailing comment
version = "0.1.0"

[dependencies-later]
x = { git = "https://example.com/x" }

[[binary]]
name = "app"
entry = "cli.run"

[[binary]]
name = 'tool'
entry = "app/tools/main.run"

[link]
libraries = ["sqlite3", "m"]
search = [
  "/opt/homebrew/lib", # one
  "/usr/local/lib",
]
`)
	if err != nil {
		t.Fatal(err)
	}
	want := &Package{
		Name:     "app",
		Binaries: []Binary{{"app", "cli.run"}, {"tool", "app/tools/main.run"}},
		Link:     LinkConfig{Libraries: []string{"sqlite3", "m"}, Search: []string{"/opt/homebrew/lib", "/usr/local/lib"}},
	}
	if !reflect.DeepEqual(pkg, want) {
		t.Fatalf("got %+v\nwant %+v", pkg, want)
	}
	for src, msg := range map[string]string{
		"[package]\n":                           "[package] needs a name",
		"[package]\nname = app\n":               "expected a string",
		"[package]\nname = \"a-b\"\n":           "package name must be",
		"[package]\nname = \"a\"\n[[binary]]\n": "needs a name and an entry",
		"[dependencies]\n":                      "not supported",
		"[package]\nname = \"std\"\n":           "reserved for the standard library",
		"[package]\nname = \"a\"\n[[binary]]\nname = \"x\"\nentry = \"m.run\"\n[[binary]]\nname = \"x\"\nentry = \"n.run\"\n": "two [[binary]] entries are named \"x\"",
		"[package]\nname = \"a\n": "unterminated string",
	} {
		_, err := parseManifest(src)
		if err == nil || !strings.Contains(err.Error(), msg) {
			t.Errorf("%q: want error containing %q, got %v", src, msg, err)
		}
	}
}

func TestEntryForms(t *testing.T) {
	pkg := &Package{Name: "app", Modules: []*SourceModule{{Path: "app/cli"}, {Path: "app/app/cli"}}}
	for entry, want := range map[string]string{"cli.run": "app/cli", "app/cli.run": "app/app/cli", "tools/x.run": "tools/x"} {
		if mod, fn := pkg.entryModule(entry); mod != want || fn != "run" {
			t.Errorf("%s: got %s %s, want %s run", entry, mod, fn, want)
		}
	}
}
