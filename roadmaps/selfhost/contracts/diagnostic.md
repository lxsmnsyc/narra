---
name: diagnostic
version: 1
status: draft
defined_by: split-compiler
---

# Contract: diagnostic

## Purpose

The structured compile error every checker, parser and lexer error becomes. Later nodes report errors through it, and json-diagnostics adds codes, fixes and `--json` output.

## Interface

```go
type Diagnostic struct {
	Code    string // "E0000" until json-diagnostics assigns codes
	Pos     Pos    // line and column, 1-based; Pos{} when unknown
	File    string // filled by the driver
	Message string
	Fixes   []Fix
}

type Fix struct {
	Title string
	Edits []Edit
}

// Start and End are line and column positions in the same file. End is exclusive.
type Edit struct {
	Start, End Pos
	Text       string
}

func (c *Checker) report(d Diagnostic) // errorf builds a Diagnostic and calls report
```

## Behavior

- Diagnostics keep the order they were reported in. They are not sorted.
- Printed form is `file:line:col: message`, the same text as today.
- In v1 the lexer and parser still stop at the first error. That error is wrapped in one `Diagnostic`.
- A node that knows an exact fix fills `Fixes`. Others leave it empty.

## Contract tests

`compiler/testdata/errors/` cases keep matching their messages.

## Changes

- v1: initial.
