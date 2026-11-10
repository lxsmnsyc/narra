---
id: type-of
title: "typeof in type positions"
status: done
priority: should
type: feature
story: batch-6
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - testdata/typeof_*
  - testdata/errors/typeof_*
  - testdata/fmt/typeof_*
  - testdata/astdump/typeof_*
  - testdata/query/typeof*/**
  - roadmaps/selfhost/nodes/type-of/**
shared:
  - selfhost/src/**
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0981, diag E0982]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 checker
---

# typeof in type positions

## Goal

Build D96 (DECISIONS.md): `typeof name` in a type annotation stands for the declared or inferred type of a value in scope: a `const`, `let`, parameter, top-level constant or function. For a function it is the full function type, with effects (D86) and `from` (D94).

- Names only, not field paths or expressions. A field path or expression after `typeof` is a parse error.
- A `typeof` that refers to its own declaration is E0981. A name that is not a value (a type, a module) is E0982 with a hint.
- Parser, AST, checker (`resolve.nra`), `fmt`, `astdump`, `narra query` signatures and hover (which show the resolved type), samples and error tests.
- Drop "Not built yet" from D96. SPEC.md gets the rule and the two samples from D96.

## Acceptance

- [ ] Program samples: `typeof` of a local map, of a parameter, of a top-level constant, and `export type Handler = typeof handle` with effects and `from`.
- [ ] Error samples for E0981 and E0982 and the parse error.
- [ ] fmt, astdump and query samples.
- [ ] `sh scripts/test.sh` passes.
