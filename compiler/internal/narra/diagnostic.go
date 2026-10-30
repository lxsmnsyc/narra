package narra

import (
	"fmt"
	"strconv"
	"strings"
)

// Diagnostic is a structured compile error. See
// roadmaps/selfhost/contracts/diagnostic.md.
type Diagnostic struct {
	Code    string // "E0000" until json-diagnostics assigns codes
	Pos     Pos    // line and column, 1-based; Pos{} when unknown
	File    string // filled by the driver
	Message string
	Fixes   []Fix
}

// Fix is a suggested change that resolves a diagnostic.
type Fix struct {
	Title string
	Edits []Edit
}

// Edit replaces the text from Start to End with Text. End is exclusive.
type Edit struct {
	Start, End Pos
	Text       string
}

// Error prints `line:col: message`. An unknown position prints as 0:0, as before.
func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s", d.Pos, d.Message)
}

// Diagnostics is a list of diagnostics, in the order they were reported.
type Diagnostics []Diagnostic

// Error joins the diagnostics, one per line.
func (ds Diagnostics) Error() string {
	lines := make([]string, len(ds))
	for i, d := range ds {
		lines[i] = d.Error()
	}
	return strings.Join(lines, "\n")
}

// report records a diagnostic. Checking continues, so one run can report several.
func (c *Checker) report(d Diagnostic) {
	// A generic function is checked once per instance and repeats its errors.
	for _, old := range c.errs {
		if old.Code == d.Code && old.File == d.File && old.Pos == d.Pos && old.Message == d.Message {
			return
		}
	}
	// Drop a fix with a range that ends before it starts.
	var fixes []Fix
	for _, f := range d.Fixes {
		ok := true
		for _, e := range f.Edits {
			if e.End.Line < e.Start.Line || (e.End.Line == e.Start.Line && e.End.Col < e.Start.Col) {
				ok = false
			}
		}
		if ok {
			fixes = append(fixes, f)
		}
	}
	d.Fixes = fixes
	c.errs = append(c.errs, d)
}

// syntaxDiagnostic wraps an error from the lexer or the parser, which stop at
// the first error and format it as `line:col: message`.
func syntaxDiagnostic(err error) Diagnostic {
	if d, ok := err.(Diagnostic); ok {
		return d
	}
	text := err.Error()
	d := Diagnostic{Code: "E0000", Message: text}
	parts := strings.SplitN(text, ": ", 2)
	if len(parts) == 2 {
		if lc := strings.SplitN(parts[0], ":", 2); len(lc) == 2 {
			line, err1 := strconv.Atoi(lc[0])
			col, err2 := strconv.Atoi(lc[1])
			if err1 == nil && err2 == nil {
				d.Pos = Pos{Line: line, Col: col}
				d.Message = parts[1]
			}
		}
	}
	return d
}
