package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// typesDumpDir holds the goldens of the types dump: <case>.nra or a
// package directory <case>/, with <case>.dump.
var typesDumpDir = filepath.Join(testdata, "typesdump", "bodies")

// TestTypesDump compares the goldens in testdata/typesdump/bodies with the
// Go dump. NARRA_UPDATE=1 rewrites them.
func TestTypesDump(t *testing.T) {
	cases := declDumpCases(t, typesDumpDir)
	if len(cases) == 0 {
		t.Fatal("no typesdump cases")
	}
	for _, path := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got, ok := TypesDump(path)
			if !ok && !strings.HasPrefix(got, "error ") {
				t.Fatalf("a failed dump must start with `error `: %q", got)
			}
			if strings.HasPrefix(filepath.Base(path), "err_") == ok {
				t.Fatalf("a case named err_ must fail and no other may (ok=%v):\n%s", ok, got)
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

// The goldens together print every line kind and attribute of the format.
func TestTypesDumpCoverage(t *testing.T) {
	var all strings.Builder
	for _, path := range declDumpCases(t, typesDumpDir) {
		got, _ := TypesDump(path)
		all.WriteString(got)
	}
	text := all.String()
	for _, want := range []string{
		" func ", " test \"", " const ", " default ", " impl ", " targs=",
		" local ", " local=", " const=", " func=", " builtin=", " module ",
		" int ", " float ", " string ", " char ", " bool ", " template ", " ident ",
		" field-expr ", " index ", " unary ", " binary ", " call ", " structlit ", " array ",
		"\nerror ",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("no golden prints %q", want)
		}
	}
}
