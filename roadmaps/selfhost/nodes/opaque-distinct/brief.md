---
id: opaque-distinct
title: Opaque structs and distinct types
status: todo
priority: should
type: feature
outline: false
milestone: m3-core
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
  - compiler/internal/langc/check_opaque.go
  - compiler/testdata/packages/opaque*/**
shared:
  - compiler/internal/langc/ast.go
  - compiler/internal/langc/parser.go
  - compiler/internal/langc/clone.go
  - compiler/internal/langc/check_types.go
  - compiler/internal/langc/types.go
  - compiler/internal/langc/check_call.go
  - compiler/internal/langc/emit_call.go
  - compiler/internal/langc/emit_types.go
reserves: []
provides: []
depends_on:
  - on: modules
    kind: behavior
    via: module-graph
---

# Opaque structs and distinct types

## Goal

Add `export opaque struct`, `distinct type` and `export opaque distinct type`, with generated `from` and `to` functions.

## Context

See `SPEC.md` section 4 and decisions D39 and D40. Opacity only matters across modules.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] A sample package uses an opaque struct from another module through its functions.
- [ ] Error tests: building, reading or spreading an opaque struct from another module; mixing two distinct types.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: checker rules, little emitter work.
