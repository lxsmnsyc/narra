package narra

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSelfhostTrivia runs two checks of `narrac trivia`:
//   - every testdata/fmt/trivia_*.nra prints the side tables in its .out
//     golden (NARRA_UPDATE_GOLDEN=1 rewrites the goldens);
//   - `narrac trivia --check` passes on every .nra file under testdata/ and
//     selfhost/src: parseWithTrivia gives the same tokens, tree dump, next id
//     (or error) as parse, and each comment is attached exactly once.
func TestSelfhostTrivia(t *testing.T) {
	exe := buildSelfhost(t)

	goldens, _ := filepath.Glob(filepath.Join(testdata, "fmt", "trivia_*.nra"))
	if len(goldens) == 0 {
		t.Fatal("no files in testdata/fmt")
	}
	for _, path := range goldens {
		t.Run("golden/"+filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			golden := strings.TrimSuffix(path, ".nra") + ".out"
			if os.Getenv("NARRA_UPDATE_GOLDEN") != "" {
				out, err := exec.Command(exe, "trivia", path).Output()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, out, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			runNarrac(t, exe, []string{"trivia", path}, string(want), "", 0)
		})
	}

	files, names := selfhostInputs(t, map[string]string{
		"only_comments.nra": "// a\n/* b */\n",
		"crlf.nra":          "// a\r\nfunction f() {\r\n  // b\r\n}\r\n",
		"template.nra":      "const a = `x ${1 + (2)} y` // c\n",
	})
	for _, path := range files {
		t.Run("check/"+names[path], func(t *testing.T) {
			t.Parallel()
			runNarrac(t, exe, []string{"trivia", "--check", path}, "ok\n", "", 0)
		})
	}
}
