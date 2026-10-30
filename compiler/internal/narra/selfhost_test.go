package narra

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// lexEdgeCases are inputs the testdata files do not cover: escapes, control
// characters and every lexer error.
var lexEdgeCases = map[string]string{
	"escapes.nra":        "const a = \"x\\u{41}\\t\\0\\$\\\\\"\nconst b = '\\''\n",
	"unexpected.nra":     "x @ y\n",
	"unterminated.nra":   "a = \"abc\n",
	"block_comment.nra":  "/* never ends",
	"numbers.nra":        "f(\n1,\n2\n)\n0x1F_F 1.5e-3 1_000 0b102 3.x 7e\n",
	"control.nra":        "a\x01b",
	"unknown_escape.nra": "\"\\q\"",
	"template.nra":       "`raw ${x} \\` more\nline`",
	"char_long.nra":      "'ab'",
	"non_ascii.nra":      "\"café\" x é",
	"quote_forms.nra":    "x ? \\ \x7f",
	"tuple_semi.nra":     "\"\\u{7f}\\u{1}\" #[a,\nb] ; ",
	"bad_u.nra":          "\"\\u{zz}\"",
	"eof_escape.nra":     "\"\\",
	"nul_char.nra":       "'\\0' '\x00'",
	// Non-ASCII escapes; a surrogate, a value above 0x10FFFF and a value
	// that does not fit 32 bits become U+FFFD, as Go's string(rune(v)) does.
	"unicode_escapes.nra": "\"\\u{e9}\\u{3bb}\\u{1F600}\\u{80}\" '\\u{10FFFF}' \"\\u{D800}\\u{110000}\\u{100000041}\\u{FFFFFFFF}\"",
}

// parseEdgeCases are inputs for the parser that the testdata files do not
// cover: one per parser error, and syntax the samples do not use.
var parseEdgeCases = map[string]string{
	"empty.nra":             "",
	"default_import.nra":    "import x from \"a\"\n",
	"import_path.nra":       "import * as A from b\n",
	"import_names.nra":      "import {\n  a,\n  b,\n} from \"m\"\nimport { } from \"n\"\n",
	"import_unclosed.nra":   "import { a, b\n",
	"export_impl.nra":       "export implement T for X {}\n",
	"export_test.nra":       "export test \"a\" {}\n",
	"test_name.nra":         "test x {}\n",
	"no_decl.nra":           "42\n",
	"opaque_private.nra":    "opaque struct A {}\n",
	"opaque_enum.nra":       "export opaque enum A {}\n",
	"distinct_type.nra":     "distinct A = int\n",
	"distinct_ok.nra":       "export opaque distinct type A = Array<int>\ndistinct type B = A.C\n",
	"array_suffix.nra":      "function f(): int[] {}\n",
	"struct_bound.nra":      "struct S<T: Show> {}\n",
	"generator.nra":         "function* g() {}\n",
	"trait_params.nra":      "trait T<A> {}\n",
	"trait_field.nra":       "trait T { x }\n",
	"trait_body.nra":        "trait T { function f() {} }\n",
	"impl_export.nra":       "implement T for X { export function f() {} }\n",
	"impl_field.nra":        "implement T for X { x }\n",
	"arrow_mutable.nra":     "function f() { const g = (mutable x) => x }\n",
	"block_eof.nra":         "function f() {\n",
	"for_const.nra":         "function f() { for (x of y) {} }\n",
	"for_in.nra":            "function f() { for (const x in y) {} }\n",
	"bare_block.nra":        "function f() { { } }\n",
	"strict_eq.nra":         "function f() { a === b }\n",
	"strict_ne.nra":         "function f() { a !== b }\n",
	"compound.nra":          "function f() { x += 1 }\n",
	"compound_spaced.nra":   "function f() { x + = 1 }\n",
	"map_key.nra":           "function f() { Map { a: 1 } }\n",
	"map_ok.nra":            "function f() { Map { -1: 2, [k]: 3, \"s\": 4,\n'c': 5, true: 6, 1.5: 7 } }\n",
	"two_spreads.nra":       "function f() { S { ...a, ...b } }\n",
	"shorthand.nra":         "function f() { S { a } }\n",
	"null.nra":              "function f() { return null }\n",
	"undefined.nra":         "function f() { g(undefined) }\n",
	"template_open.nra":     "function f() { `a ${x` }\n",
	"template_lex.nra":      "function f() { `${\"a}` }\n",
	"template_parse.nra":    "function f() {\n  `x ${)}`\n}\n",
	"template_nested.nra":   "function f() { `a\\n\\t\\${b} ${g(`${h(\"s\")}`)} c` }\n",
	"pattern_minus.nra":     "function f() { match (x) { -a => 1 } }\n",
	"statement_end.nra":     "function f() { a b }\n",
	"statement_end2.nra":    "function f() { const x = 1 const y = 2 }\n",
	"describe.nra":          "function f() { a \"é\\u{80}\\u{a0}\\t\\u{1}\\\"\\\\\" }\n",
	"describe_eol.nra":      "function f() { a = }\n",
	"type_bad.nra":          "function f(x: 1) {}\n",
	"var_patterns.nra":      "function f() {\n  const Some(x) = y\n  let #[a, b]: #[int, int] = t\n  const { p, q: Some(_) } = s\n  const Shape.Circle = c\n}\n",
	"patterns.nra":          "function f() {\n  match (x) {\n    -1 => 0\n    -2.5 => 0\n    'c' => 0\n    \"s\" => 0\n    true => 0\n    Shape.Circle => 0\n    Circle() => 0\n    Circle => 0\n    _ if (g) => {\n      return 1\n    }\n  }\n}\n",
	"func_types.nra":        "function f<T: A & B, effect E>(g: (x: int, y) => T uses E, mutable out: Array<Array<int>>, escaping h: () => #[int, string] uses io, alloc): Map<string, Array<int>> uses E {}\n",
	"escaping_name.nra":     "function f(escaping: int) {}\n",
	"if_else.nra":           "function f() {\n  if (a) {\n  }\n  else if (b) {\n  } else {\n  }\n  if (c) {\n  }\n\n  x = 1\n}\n",
	"while_return.nra":      "function f() {\n  while (a is Some(x) && b) {\n    return\n  }\n  return }\n",
	"unary.nra":             "function f() { x = -~!a.b(mutable c)[0] }\n",
	"arrow_forms.nra":       "function f() {\n  const g = (a: int, b): int => {\n    return a\n  }\n  const h = () => 1\n  const k = (x)\n}\n",
	"struct_ctx.nra":        "function f() { g({ a: 1, ...b }) }\n",
	"member_from.nra":       "function f() { UserId.from(1) }\n",
	"enum_fields.nra":       "export enum E<T> implements A, B derives C {\n  X(a: int, b: T),\n  Y\n}\nstruct S implements A {\n  a: int = 1,\n  b: Option<int>\n}\n",
	"export_const.nra":      "export const A: int = 1 << 2 >> 1 | 3 ^ 4 & 5 +% 6 -% 7 *% 8 % 9\n",
	"decl_eof.nra":          "export\n",
	"match_value.nra":       "function f() {\n  const x = match (y) {\n    A => 1,\n    B => 2,\n  }\n  match (y) {\n    _ => g()\n  } h()\n}\n",
	"eof_in_call.nra":       "function f() { g(",
	"eof_in_type.nra":       "function f(x: Array<",
	"eof_in_params.nra":     "function f(a: int",
	"eof_in_struct.nra":     "struct S { a: int",
	"eof_in_enum.nra":       "enum E { A(x: int",
	"eof_in_match.nra":      "function f() { match (x) { A => ",
	"eof_in_pattern.nra":    "function f() { match (x) { A(b, ",
	"eof_in_array.nra":      "function f() { [1, 2",
	"eof_in_struct_lit.nra": "function f() { S { a: 1",
	"eof_in_map.nra":        "function f() { Map { 1: 2",
	"eof_in_uses.nra":       "function f() uses a,",
	"eof_in_tparams.nra":    "function f<A: B &",
	"eof_in_trait.nra":      "trait T { function f()\n",
	"eof_in_impl.nra":       "implement T for X { function f() {}\n",
	"eof_in_tuple_type.nra": "function f(x: #[int, ",
	"eof_in_func_type.nra":  "function f(x: (a: int, ",
}


// selfhostInputs lists every .nra file under testdata/ and selfhost/src,
// plus the edge cases written to a temporary directory. It maps each path to
// the subtest name.
func selfhostInputs(t *testing.T, edges map[string]string) ([]string, map[string]string) {
	t.Helper()
	var files []string
	names := map[string]string{}
	for _, root := range []string{testdata, "../../../selfhost/src"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".nra") {
				files = append(files, path)
				names[path] = path
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	edgeDir := t.TempDir()
	var edgeNames []string
	for name := range edges {
		edgeNames = append(edgeNames, name)
	}
	sort.Strings(edgeNames)
	for _, name := range edgeNames {
		path := filepath.Join(edgeDir, name)
		if err := os.WriteFile(path, []byte(edges[name]), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, path)
		names[path] = "edge/" + name
	}
	return files, names
}

// buildSelfhost builds selfhost/ in debug mode (AddressSanitizer, UBSan and
// the leak check) and returns the executable. The build is shared by the
// tests of one run.
var selfhostBuild struct {
	once sync.Once
	dir  string
	exe  string
	err  error
}

func buildSelfhost(t *testing.T) string {
	t.Helper()
	selfhostBuild.once.Do(func() {
		selfhostBuild.dir, selfhostBuild.err = os.MkdirTemp("", "narrac")
		if selfhostBuild.err != nil {
			return
		}
		selfhostBuild.exe = filepath.Join(selfhostBuild.dir, "narrac")
		selfhostBuild.err = Build("../../../selfhost", selfhostBuild.exe, modes[0].opts)
	})
	if selfhostBuild.err != nil {
		t.Fatalf("build selfhost: %v", selfhostBuild.err)
	}
	return selfhostBuild.exe
}

func TestMain(m *testing.M) {
	code := m.Run()
	if selfhostBuild.dir != "" {
		os.RemoveAll(selfhostBuild.dir)
	}
	os.Exit(code)
}

// runNarrac runs narrac and checks stdout, stderr and the exit code.
func runNarrac(t *testing.T, exe string, args []string, wantOut, wantErr string, wantCode int) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	if code != wantCode {
		t.Fatalf("exit code %d, want %d\nstderr:\n%s", code, wantCode, stderr.String())
	}
	if got := stdout.String(); got != wantOut {
		t.Fatalf("stdout differs: %s", firstDiff(wantOut, got))
	}
	if got := stderr.String(); got != wantErr {
		t.Fatalf("stderr differs: %s", firstDiff(wantErr, got))
	}
}

// compareSelfhost runs narrac with args on every input and compares it with
// the Go dump: on success the dump is stdout, and on failure the error line
// is stderr (the Go commands print it to stdout) with exit code 1.
func compareSelfhost(t *testing.T, args []string, edges map[string]string, dump func(string) (string, bool)) {
	exe := buildSelfhost(t)
	files, names := selfhostInputs(t, edges)
	for _, path := range files {
		t.Run(names[path], func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, ok := dump(string(src))
			wantOut, wantErr, wantCode := want, "", 0
			if !ok {
				wantOut, wantErr, wantCode = "", want, 1
			}
			runNarrac(t, exe, append(append([]string{}, args...), path), wantOut, wantErr, wantCode)
		})
	}
}

// TestSelfhostLexer checks that `narrac <file>` prints the same token dump
// as `narra lex` (LexDump) on every .nra file under testdata/, on the
// selfhost sources, and on lexEdgeCases. Like TestPrograms, it needs a C
// compiler (`cc`, or $CC) that supports -fsanitize=address,undefined.
func TestSelfhostLexer(t *testing.T) {
	compareSelfhost(t, nil, lexEdgeCases, LexDump)
}

// TestSelfhostParser checks that `narrac parse <file>` prints the same AST
// dump as `narra parse` (ASTDump), including the first error, on the same
// files and on parseEdgeCases.
func TestSelfhostParser(t *testing.T) {
	compareSelfhost(t, []string{"parse"}, parseEdgeCases, ASTDump)
}

// TestSelfhostModules checks that `narrac pkg <path>` prints the same package
// dump as `narra pkg` (PkgDump), including errors, on every package under
// testdata/packages (and its errors/), the manifest cases in testdata/pkgdump,
// selfhost/ itself, and every single .nra file under testdata/ and
// selfhost/src.
func TestSelfhostModules(t *testing.T) {
	exe := buildSelfhost(t)
	inputs := pkgInputs(t)
	for _, path := range inputs {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			want, ok := PkgDump(path)
			wantOut, wantErr, wantCode := want, "", 0
			if !ok {
				wantOut, wantErr, wantCode = "", want, 1
			}
			runNarrac(t, exe, []string{"pkg", path}, wantOut, wantErr, wantCode)
		})
	}
}

// TestSelfhostModuleIds runs `narrac pkg --ids` on every package that loads:
// the id ranges of the files must be disjoint, contiguous from 0, and end at
// Package.nextId.
func TestSelfhostModuleIds(t *testing.T) {
	exe := buildSelfhost(t)
	for _, path := range pkgInputs(t) {
		if _, ok := PkgDump(path); !ok {
			continue
		}
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(exe, "pkg", "--ids", path)
			out, err := cmd.CombinedOutput()
			if err != nil || !strings.HasPrefix(string(out), "ids ok ") {
				t.Fatalf("%v\n%s", err, out)
			}
		})
	}
}

// pkgInputs lists the package directories and single files that the package
// tests run on.
func pkgInputs(t *testing.T) []string {
	t.Helper()
	var inputs []string
	for _, pattern := range []string{
		filepath.Join(testdata, "packages", "*"),
		filepath.Join(testdata, "packages", "errors", "*"),
		filepath.Join(testdata, "pkgdump", "*"),
	} {
		dirs, _ := filepath.Glob(pattern)
		for _, d := range dirs {
			if info, err := os.Stat(filepath.Join(d, "narra.toml")); err == nil && !info.IsDir() {
				inputs = append(inputs, d)
			}
		}
	}
	inputs = append(inputs, "../../../selfhost")
	files, _ := selfhostInputs(t, nil)
	inputs = append(inputs, files...)
	return inputs
}

// firstDiff describes the first line where want and got differ.
func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var a, b string
		if i < len(w) {
			a = w[i]
		}
		if i < len(g) {
			b = g[i]
		}
		if a != b {
			return "line " + strconv.Itoa(i+1) + "\nwant: " + a + "\ngot:  " + b
		}
	}
	return "same lines"
}
