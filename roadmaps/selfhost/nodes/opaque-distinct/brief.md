---
id: opaque-distinct
title: Opaque structs and distinct types
status: done
spent: 340264
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
  - compiler/internal/narra/check_opaque.go
  - compiler/testdata/packages/opaque*/**
shared:
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/emit_types.go
  - compiler/README.md
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_match.go
  - compiler/internal/narra/check_patterns.go
  - compiler/internal/narra/check_traits.go
  - compiler/internal/narra/cnames.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_traits.go
  - compiler/internal/narra/modules.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/std_map.go
  - compiler/testdata/packages/errors/**
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

- From patterns: reject struct patterns, destructuring and spread of another module's opaque struct, in `resolvePat` (PatStruct) and `spreadLit`.

- From hashed-cnames: name opaque and distinct types through the `cnames.go` helper as their own nominal types. Never fold them into the type they wrap, so `distinct UserId = int` gets its own C name.

- From box-recursive: an opaque struct that holds a boxed enum or a `Box` field is still reached through `.u->`; do not read fields by hand in generated accessors. Allocate enum values with `e.enumAlloc`.

## Acceptance

- [x] A sample package uses an opaque struct from another module through its functions.
- [x] Error tests: building, reading or spreading an opaque struct from another module; mixing two distinct types.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: checker rules, little emitter work.

PR: [#18](https://github.com/lxsmnsyc/narra/pull/18)
