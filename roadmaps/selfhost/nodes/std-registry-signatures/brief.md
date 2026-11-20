---
id: std-registry-signatures
title: "Readable signatures for every builtin std function"
status: active
priority: could
type: feature
story: batch-21
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-registry-signatures/**
shared:
  - selfhost/src/check/registry*.nra
  - selfhost/src/tools/querystd.nra
  - selfhost/src/tools/query*.nra
  - testdata/query/**
  - testdata/lsp/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on:
  - on: std-interface-query
    kind: hard
    why: extends its output
---
# Readable signatures for every builtin std function

## Goal

std-interface-query (outcome) prints builtin signatures from the registry, which has no parameter names, and 135 builtins with a check function instead of a type list print `(...)` (for example `export function push(...) uses alloc`: Array.push, Map.remove, Range.upTo, std/ptr, std/simd, sized from/tryFrom/wrap). A generated std reference built on that is not readable.

Give every registry entry parameter names, and give each check-function builtin a written signature (generic form, for example `push<T>(mutable array: Array<T>, item: T): void uses alloc`) used only for display: `query interface`, the LSP hover and signature help. Also list the types that builtin modules define without source (for example `StringBuilder` in `std/string`), with their methods. The checker keeps its check functions. A test fails when a public builtin prints `(...)` or an unnamed parameter.

## Acceptance

- [ ] `narrac query interface std/<module>` prints named parameters for every public builtin and no `(...)`; the test above exists.
- [ ] `sh scripts/test.sh` passes.
