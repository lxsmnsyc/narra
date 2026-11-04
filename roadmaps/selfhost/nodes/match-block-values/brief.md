---
id: match-block-values
title: "match arms with blocks produce values"
status: done
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent: 136983
owns:
  - compiler/testdata/match_block_values*
  - compiler/testdata/errors/match_block_values_*
  - roadmaps/selfhost/nodes/match-block-values/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/check_match.go
  - compiler/internal/narra/emit_match.go
  - selfhost/src/**
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: block-expressions
    kind: hard
    why: a match arm in value position reuses the block expression checker and emitter
---

# match arms with blocks produce values

## Goal

DECISIONS.md D75. SPEC.md line 434 says `match` produces a value only when every arm is a single expression. In value position, an arm `pattern => { ... }` becomes a block expression: its value comes from `return v`, and a `return` inside it ends the arm, not the function. A `match` in statement position keeps block-statement arms, where `return` leaves the function, as today.

## Acceptance

- [ ] Sample programs: a `match` as an initializer and as an argument with a mix of expression arms and block arms, guards, payload bindings, an early `return` inside an `if` in an arm, heap locals in arms (clean under ASan and the leak check).
- [ ] Error tests: a block arm path without `return`, arms of different types, `break` out of an arm in value position. Positions in the `// error:` headers.
- [ ] A statement-position `match` with `{ return x }` arms still returns from the function (existing samples pass unchanged).
- [ ] SPEC.md line 434 and the match section describe the rule.
- [ ] The Go compiler and `selfhost/` change together and match on the new tests; fast check passes including `TestSelfhostBootstrap`.
