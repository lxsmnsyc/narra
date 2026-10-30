package narra

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// lexEdgeCases are inputs the testdata files do not cover: escapes, control
// characters and every lexer error. Non-ASCII \u{...} escapes are left out;
// the self-hosted lexer cannot build those characters yet (no Char.fromCode).
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
}

// TestSelfhostLexer builds the self-hosted compiler in selfhost/ (debug,
// with AddressSanitizer and the leak check) and checks that its token dump
// equals `narra lex` (LexDump) on every .nra file under testdata/, on the
// selfhost sources, and on lexEdgeCases. Like TestPrograms, it needs a C
// compiler (`cc`, or $CC) that supports -fsanitize=address,undefined.
func TestSelfhostLexer(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "narrac")
	if err := Build("../../../selfhost", exe, modes[0].opts); err != nil {
		t.Fatalf("build selfhost: %v", err)
	}
	var files []string
	for _, root := range []string{testdata, "../../../selfhost/src"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".nra") {
				files = append(files, path)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	edgeDir := t.TempDir()
	for name, src := range lexEdgeCases {
		path := filepath.Join(edgeDir, name)
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, path)
	}
	for _, path := range files {
		name := path
		if strings.HasPrefix(path, edgeDir) {
			name = "edge/" + filepath.Base(path)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, ok := LexDump(string(src))
			wantCode := 0
			if !ok {
				wantCode = 1
			}
			cmd := exec.Command(exe, path)
			cmd.Env = append(os.Environ(), "NARRA_CHECK_LEAKS=1", "ASAN_OPTIONS=detect_leaks=0")
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err = cmd.Run()
			code := 0
			if exit, ok := err.(*exec.ExitError); ok {
				code = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			if code != wantCode {
				t.Fatalf("exit code %d, want %d\nstderr:\n%s", code, wantCode, stderr.String())
			}
			if got := stdout.String(); got != want {
				t.Fatalf("token dump differs: %s", firstDiff(want, got))
			}
		})
	}
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
