package narra

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CheckDiagnostics parses and checks a source file or package directory and
// returns every diagnostic. The error is set for failures that are not
// diagnostics, such as a missing file.
func CheckDiagnostics(path string) (Diagnostics, error) {
	pkg, err := Load(path)
	if err == nil {
		var prog *Program
		prog, err = CheckPackage(pkg)
		if err == nil {
			return entryDiagnostics(pkg, prog), nil
		}
	}
	if ds := diagnosticsOf(err); ds != nil {
		return ds, nil
	}
	return nil, err
}

// entryDiagnostics reports each binary whose entry function is missing, as
// CompileBinary does.
func entryDiagnostics(pkg *Package, prog *Program) Diagnostics {
	var ds Diagnostics
	for _, b := range pkg.Binaries {
		if prog.Entries[b.Name] != nil {
			continue
		}
		if pkg.single {
			f := "no entry function; declare `export function run(args: Array<string>): int`"
			ds = append(ds, Diagnostic{Code: codeFor(f), File: pkg.Modules[0].Files[0].Path, Message: f})
			continue
		}
		mod, fn := pkg.entryModule(b.Entry)
		f := "binary %q: no entry function %q; declare `export function %s(args: Array<string>): int` in module %q"
		ds = append(ds, Diagnostic{Code: codeFor(f), File: "narra.toml", Message: fmt.Sprintf(f, b.Name, b.Entry, fn, mod)})
	}
	return ds
}

type jsonRange struct {
	Start [2]int `json:"start"`
	End   [2]int `json:"end"`
}

type jsonEdit struct {
	File  string    `json:"file"`
	Range jsonRange `json:"range"`
	Text  string    `json:"text"`
}

type jsonFix struct {
	Title string     `json:"title"`
	Edits []jsonEdit `json:"edits"`
}

type jsonDiagnostic struct {
	Code     string    `json:"code"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	File     string    `json:"file"`
	Range    jsonRange `json:"range"`
	Fixes    []jsonFix `json:"fixes"`
}

func posPair(p Pos) [2]int { return [2]int{p.Line, p.Col} }

// JSON prints a diagnostic as one line of JSON (SPEC.md section 14). The
// range is empty (start equals end) because diagnostics record a position,
// not an extent.
func (d Diagnostic) JSON() string {
	out := jsonDiagnostic{
		Code: d.Code, Severity: "error", Message: d.Message, File: d.File,
		Range: jsonRange{Start: posPair(d.Pos), End: posPair(d.Pos)},
		Fixes: []jsonFix{},
	}
	for _, f := range d.Fixes {
		jf := jsonFix{Title: f.Title, Edits: []jsonEdit{}}
		for _, e := range f.Edits {
			jf.Edits = append(jf.Edits, jsonEdit{File: d.File, Range: jsonRange{Start: posPair(e.Start), End: posPair(e.End)}, Text: e.Text})
		}
		out.Fixes = append(out.Fixes, jf)
	}
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.Encode(out)
	return strings.TrimSpace(b.String())
}

// ApplyFix applies the edits of a fix to source text and returns the result.
// Edits must not overlap. Columns count runes, as the lexer does.
func ApplyFix(src string, f Fix) string {
	lines := strings.Split(src, "\n")
	offset := func(p Pos) int {
		n := 0
		for i := 0; i < p.Line-1 && i < len(lines); i++ {
			n += len(lines[i]) + 1
		}
		if p.Line-1 < len(lines) {
			line := []rune(lines[p.Line-1])
			c := min(p.Col-1, len(line))
			n += len(string(line[:c]))
		}
		return n
	}
	edits := append([]Edit(nil), f.Edits...)
	// Apply from the end so earlier offsets stay valid.
	for i := 1; i < len(edits); i++ {
		for j := i; j > 0 && offset(edits[j].Start) > offset(edits[j-1].Start); j-- {
			edits[j], edits[j-1] = edits[j-1], edits[j]
		}
	}
	for _, e := range edits {
		s, t := offset(e.Start), offset(e.End)
		src = src[:s] + e.Text + src[t:]
	}
	return src
}

// Explain returns the text of `narra explain code`.
func Explain(code string) (string, error) {
	code = strings.ToUpper(code)
	var format string
	for _, d := range diagnosticCodes {
		if d.Code == code {
			format = d.Format
		}
	}
	if format == "" {
		return "", fmt.Errorf("unknown diagnostic code %q", code)
	}
	text := code + ": " + format + "\n"
	if long, ok := explanations[code]; ok {
		text += "\n" + long + "\n"
	}
	return text, nil
}

// CheckCommand runs `narra check [--json] path` and returns the exit code.
func CheckCommand(args []string) int {
	asJSON := false
	var paths []string
	for _, a := range args {
		if a == "--json" || a == "-json" {
			asJSON = true
		} else {
			paths = append(paths, a)
		}
	}
	if len(paths) != 1 {
		fmt.Fprintln(os.Stderr, "usage: narra check [--json] file.nra|dir")
		return 2
	}
	ds, err := CheckDiagnostics(paths[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if asJSON {
		PrintJSON(ds)
	} else if len(ds) > 0 {
		fmt.Fprintln(os.Stderr, FormatError(ds))
	}
	if len(ds) > 0 {
		return 1
	}
	return 0
}

// ExplainCommand runs `narra explain code` and returns the exit code.
func ExplainCommand(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: narra explain E0000")
		return 2
	}
	text, err := Explain(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(text)
	return 0
}

// PrintJSON writes one JSON object per diagnostic and reports whether there
// were any.
func PrintJSON(ds Diagnostics) bool {
	for _, d := range ds {
		fmt.Fprintln(os.Stdout, d.JSON())
	}
	return len(ds) > 0
}
