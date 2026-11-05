package narra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The sample format is in roadmaps/selfhost/contracts/guide-samples.md.

// guideFence is one fenced block of a guide chapter.
type guideFence struct {
	info string
	line int // line of the opening fence, 1-based
	end  int // line of the closing fence, 1-based
	body string
}

// guideSample is a fenced narra block of a chapter.
type guideSample struct {
	file string
	line int
	kind string
	code string
	// want is the stdout from the text block that directly follows a
	// narra block; hasWant is false when there is none.
	want    string
	hasWant bool
}

func readGuideFences(text string) []guideFence {
	var fences []guideFence
	lines := strings.Split(text, "\n")
	for index := 0; index < len(lines); index++ {
		if !strings.HasPrefix(lines[index], "```") {
			continue
		}
		start := index
		info := strings.TrimSpace(strings.TrimLeft(lines[index], "`"))
		var body []string
		index++
		for index < len(lines) && !strings.HasPrefix(lines[index], "```") {
			body = append(body, lines[index])
			index++
		}
		fences = append(fences, guideFence{info, start + 1, index + 1, strings.Join(body, "\n") + "\n"})
	}
	return fences
}

func extractGuideSamples(file string, text string) []guideSample {
	fences := readGuideFences(text)
	lines := strings.Split(text, "\n")
	var samples []guideSample
	for index, fence := range fences {
		if fence.info != "narra" && fence.info != "narra error" && fence.info != "narra fragment" {
			continue
		}
		sample := guideSample{file: file, line: fence.line, kind: fence.info, code: fence.body}
		if fence.info == "narra" && index+1 < len(fences) && fences[index+1].info == "text" {
			// Only blank lines may sit between the two blocks.
			between := strings.Join(lines[fence.end:fences[index+1].line-1], "")
			if strings.TrimSpace(between) == "" {
				sample.want = fences[index+1].body
				sample.hasWant = true
			}
		}
		samples = append(samples, sample)
	}
	return samples
}

func TestGuide(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "..", "..", "docs", "guide", "*.md"))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, sample := range extractGuideSamples(filepath.Base(file), string(data)) {
			t.Run(fmt.Sprintf("%s:%d", sample.file, sample.line), func(t *testing.T) {
				runGuideSample(t, sample)
			})
		}
	}
}

func runGuideSample(t *testing.T, sample guideSample) {
	if sample.kind == "narra fragment" {
		return
	}
	src := filepath.Join(t.TempDir(), "sample.nra")
	if err := os.WriteFile(src, []byte(sample.code), 0o644); err != nil {
		t.Fatal(err)
	}
	if sample.kind == "narra error" {
		first := strings.SplitN(sample.code, "\n", 2)[0]
		if !strings.HasPrefix(first, "// error:") {
			t.Fatalf("an error sample starts with a `// error:` comment")
		}
		want := strings.TrimSpace(strings.TrimPrefix(first, "// error:"))
		_, err := CompileToC(src, false)
		if err == nil {
			t.Fatalf("expected an error containing %q", want)
		}
		message := strings.ReplaceAll(err.Error(), "sample.nra:", "")
		if !strings.Contains(message, want) {
			t.Fatalf("want error containing %q, got:\n%v", want, err)
		}
		return
	}
	if !sample.hasWant {
		t.Fatalf("a `narra` block needs a `text` block with its output right after it")
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			out, stderr, code := runProgram(t, src, mode.opts)
			if code != 0 {
				t.Fatalf("exit code %d\nstderr:\n%s", code, stderr)
			}
			if out != sample.want {
				t.Fatalf("output mismatch\nwant:\n%s\ngot:\n%s", sample.want, out)
			}
		})
	}
}
