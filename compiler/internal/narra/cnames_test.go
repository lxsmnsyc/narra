package narra

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestXXH32(t *testing.T) {
	alpha := []byte("abcdefghijklmnopqrstuvwxyz")
	cases := []struct {
		in   []byte
		seed uint32
		want uint32
	}{
		{nil, 0, 0x02cc5d05},
		{[]byte("abc"), 0, 0x32d153ff},
		{[]byte("Nobody inspects the spammish repetition"), 0, 0xe2293b2f},
		{alpha[:15], 0, 0xb918a375},
		{alpha[:16], 0, 0x9d2d8b62},
		{alpha[:17], 0, 0xb3b873e1},
		{[]byte("abc"), 0x9e3779b1, 0xa1ae7709},
		{alpha[:16], 0x9e3779b1, 0x3fab7bd8},
	}
	for _, c := range cases {
		if got := xxh32(c.in, c.seed); got != c.want {
			t.Errorf("xxh32(%q, %#x) = %08x, want %08x", c.in, c.seed, got, c.want)
		}
	}
}

func TestModuleHash(t *testing.T) {
	// The package name is not part of the input.
	if moduleHash("app/text") != moduleHash("other/text") {
		t.Error("the hash depends on the package name")
	}
	if got, want := moduleHash("app/text"), hex8(xxh32([]byte("text"), 0)); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if len(moduleHash("app/a")) != 8 {
		t.Error("the hash is not 8 digits")
	}
}

func TestCNameCollisionStops(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "internal error") {
			t.Fatalf("want an internal error, got %v", r)
		}
	}()
	registerCName("f_collision_test", "one")
	registerCName("f_collision_test", "two")
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.Walk(from, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, rel), src, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// The same package in two different directories gives the same C.
func TestCNamesIgnoreDirectory(t *testing.T) {
	src, _ := filepath.Abs(filepath.Join(testdata, "packages", "cnames_modules"))
	var outs []string
	for _, dir := range []string{"one", "two/deeper/still"} {
		dst := filepath.Join(t.TempDir(), dir)
		copyTree(t, src, dst)
		old, _ := os.Getwd()
		if err := os.Chdir(dst); err != nil {
			t.Fatal(err)
		}
		c, err := CompileToC(".", false)
		os.Chdir(old)
		if err != nil {
			t.Fatal(err)
		}
		outs = append(outs, c)
	}
	if outs[0] != outs[1] {
		t.Fatal("the C differs between directories")
	}
	if !strings.Contains(outs[0], "/* S_Tuple2_") {
		t.Error("no comment table for the tuple types")
	}
}

var typeNameRE = regexp.MustCompile(`/\* ([A-Za-z0-9_]+) = ([^*]*) \*/`)

func typeTable(c string) map[string]string {
	m := map[string]string{}
	for _, g := range typeNameRE.FindAllStringSubmatch(c, -1) {
		m[g[2]] = g[1]
	}
	return m
}

// Adding an unrelated function does not change the names of existing types.
func TestCNamesStableUnderEdits(t *testing.T) {
	base := `struct Box<T> {
  item: T
}

export function run(args: Array<string>): int {
  const b: Box<#[int, Array<int>]> = Box { item: #[1, [2]] }
  const f = (x: int) => x + 1
  const #[n, xs] = b.item
  return f(n) - 2 + xs[0] - 2
}
`
	edited := strings.Replace(base, "export function run", `function unrelated(x: Box<int>, y: Array<string>): Box<int> {
  return x
}

export function run`, 1)
	var tables []map[string]string
	for _, s := range []string{base, edited} {
		path := filepath.Join(t.TempDir(), "prog.nra")
		if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
		c, err := CompileToC(path, false)
		if err != nil {
			t.Fatal(err)
		}
		tables = append(tables, typeTable(c))
	}
	if len(tables[0]) < 4 {
		t.Fatalf("too few interned types: %v", tables[0])
	}
	for typ, name := range tables[0] {
		if tables[1][typ] != name {
			t.Errorf("type %s was %s, now %s", typ, name, tables[1][typ])
		}
	}
}
