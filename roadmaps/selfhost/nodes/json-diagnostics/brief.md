---
id: json-diagnostics
title: Diagnostic codes, fixes and JSON output
status: todo
priority: should
type: feature
outline: false
milestone: m7-tooling
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - compiler/internal/narra/diagnostic*.go
  - compiler/testdata/diagnostics/**
shared:
  - compiler/cmd/narra/main.go
  - compiler/internal/narra/check*.go
  - compiler/internal/narra/parser.go
reserves: []
provides: []
depends_on:
  - on: modules
    kind: hard
    why: both edit cmd/narra/main.go, and diagnostics need the module file name
  - on: traits
    kind: hard
    why: assigns codes to every existing diagnostic, including the ones traits adds
  - on: closures
    kind: hard
    why: assigns codes to every existing diagnostic, including the ones closures adds
  - on: patterns
    kind: hard
    why: assigns codes to every existing diagnostic, including the ones patterns adds
  - on: split-compiler
    kind: contract
    via: diagnostic
---

# Diagnostic codes, fixes and JSON output

## Goal

Assign stable codes to every diagnostic, add exact fixes for the TypeScript habits in `SPEC.md` section 2 (`===`, `null`, `xs.push`, `import X from`, `;`, `+` on strings), and add `narra check --json` and `narra explain <code>`.

## Context

See `SPEC.md` section 14 and decision D42. The JSON shape is in section 14.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] `narra check --json` on each `testdata/errors/` file prints a diagnostic with a code, and the TypeScript habit cases include a fix that, when applied, makes the file compile.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: touches many error sites, but each change is small.
