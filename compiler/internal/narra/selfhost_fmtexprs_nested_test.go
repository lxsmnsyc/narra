package narra

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSelfhostFmtExprsNested formats 30 callbacks nested as last arguments.
// The hugged last argument is in the document twice (once per layout), so a
// printer that copies shared parts per use takes time exponential in the
// depth. The case must stay fast.
func TestSelfhostFmtExprsNested(t *testing.T) {
	exe := buildFmtExprs(t)
	const depth = 30
	source := "const x = " + strings.Repeat("run(firstArgument, (item) => ", depth) + "1" + strings.Repeat(")", depth) + "\n"
	path := filepath.Join(t.TempDir(), "nested.nra")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, width := range []string{"100", "40"} {
		start := time.Now()
		out, err := exec.Command(exe, path, "--width", width).CombinedOutput()
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("width %s: %v\n%s", width, err, out)
		}
		t.Logf("width %s: %v, %d bytes", width, elapsed, len(out))
		if elapsed > time.Second {
			t.Fatalf("width %s: took %v", width, elapsed)
		}
	}
}
