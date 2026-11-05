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
	info string
	code string
	// want is the stdout from the text block that directly follows a
	// narra block; hasWant is false when there is none.
	want    string
	hasWant bool
}

func readGuideFences(text string) ([]guideFence, error) {
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
		if index >= len(lines) {
			return nil, fmt.Errorf("line %d: the fence is never closed", start+1)
		}
		fences = append(fences, guideFence{info, start + 1, index + 1, strings.Join(body, "\n") + "\n"})
	}
	return fences, nil
}

// extractGuideSamples returns the narra blocks of a chapter. An info string
// that starts with "narra" (any case) but is not one of the three kinds is
// returned with the kind "invalid", so a typo fails instead of going untested.
func extractGuideSamples(file string, text string) ([]guideSample, error) {
	fences, err := readGuideFences(text)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(text, "\n")
	var samples []guideSample
	for index, fence := range fences {
		kind := fence.info
		if kind != "narra" && kind != "narra error" && kind != "narra fragment" {
			if !strings.HasPrefix(strings.ToLower(kind), "narra") {
				continue
			}
			kind = "invalid"
		}
		sample := guideSample{file: file, line: fence.line, kind: kind, code: fence.body, info: fence.info}
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
	return samples, nil
}

func TestGuide(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "docs", "guide", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no guide files match docs/guide/*.md")
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		samples, err := extractGuideSamples(filepath.Base(file), string(data))
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(file), err)
			continue
		}
		for _, sample := range samples {
			t.Run(fmt.Sprintf("%s:%d", sample.file, sample.line), func(t *testing.T) {
				runGuideSample(t, sample)
			})
		}
	}
}

func runGuideSample(t *testing.T, sample guideSample) {
	if sample.kind == "invalid" {
		t.Fatalf("the info string %q is not narra, narra error or narra fragment", sample.info)
	}
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

func TestGuideExtract(t *testing.T) {
	bad := "```narra errror\nx\n```\n\n```Narra\nx\n```\n\n```narra,\nx\n```\n\n```sh\nx\n```\n"
	samples, err := extractGuideSamples("bad.md", bad)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 3 {
		t.Fatalf("want 3 samples, got %d", len(samples))
	}
	for _, sample := range samples {
		if sample.kind != "invalid" {
			t.Fatalf("line %d: want kind invalid, got %q", sample.line, sample.kind)
		}
	}
	if _, err := extractGuideSamples("open.md", "text\n\n```narra\nx\n"); err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("want an unclosed fence error for line 3, got %v", err)
	}
}
