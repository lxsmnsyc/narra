---
id: destructuring-assign
title: "Destructuring assignment"
status: dropped
priority: should
type: feature
story:
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/destructuring-assign/**
  - testdata/destructuring_assign*
  - testdata/errors/destructuring_assign*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - selfhost/src/fmt.nra
  - selfhost/src/stmts.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/diagnostics.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - SPEC.md
  - DECISIONS.md
reserves: []
provides: []
depends_on:
  - on: destructuring-design
    kind: hard
    why: D103 part 3 decides whether this node is built
---

# Destructuring assignment

## Goal

Dropped by D103 (2026-11-12): 4 corpus sites, all with field or element targets, and neither spelling is TypeScript text. The rules below are kept for reopening.

`#[a, b] = #[b, a]` assigns existing places through a tuple pattern. Only the tuple form; `{ x, y } = point` waits for a parenthesized statement form, since `{` starts a block.

- The right side is evaluated whole into temporaries. Then each target's place is evaluated and written, left to right.
- Targets are what `=` accepts: a local `let`, a field or element of a changeable value, the contents of a `mutable` parameter. Another target gets the existing "only variables, fields and array elements can be assigned" error. Nested tuple patterns and `_` are allowed.
- A target whose path reads a name that an earlier target of the same pattern assigns (`#[i, items[i]] = ...`) is a compile error (reserve a code when the node is reopened).
- Each write is checked by E0978 as the separate assignment would be. The swap `#[items[i], items[j]] = #[items[j], items[i]]` is accepted exactly when the three-line form with a temporary is (`std/src/random.nra:57` compiles today).
- No compound form (`#[a, b] += ...` is a parse error).

## Context

- D103 in DECISIONS.md. `AssignStmt` in `ast.nra` takes a target expression; a tuple pattern target needs its own statement kind or a pattern field, with the AST switch cases (roadmap CLAUDE.md).
- Shared writes: `sharedwrites.nra`. Temporaries are last used at their own write.
- Corpus sites: `std/src/random.nra:57` (element swap), `selfhost/src/stmts.nra:669`, `989`, `1115` (field copies from one value; these would need a struct form and stay as they are).

## Acceptance

- [ ] Samples: swap of two `let` names, of two array elements, of two fields, a nested tuple target, a tuple result of a call assigned to existing names; all clean under the leak check and sanitizers.
- [ ] Error tests: a `const` target, a plain parameter target, the earlier-target case, an E0978 case.
- [ ] astdump and fmt goldens; SPEC.md section 2 "Bindings" describes the form; D103 gets a "Built" line.
- [ ] Fast check passes: `sh scripts/test.sh`
