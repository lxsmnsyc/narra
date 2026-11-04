---
id: for-destructure
title: Tuple and struct destructuring in for-of heads
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/for_destructure*
  - compiler/testdata/errors/for_destructure*
  - roadmaps/selfhost/nodes/for-destructure/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes after bootstrap land in both compilers at once; editing the parser during bootstrap conflicts
---

# Tuple and struct destructuring in for-of heads

## Goal

SPEC.md line 110 promises `for (const #[i, item] of Arrays.entries(items))`, and line 58 allows `const { host, port: serverPort } = config` and `const #[a, b] = t` in declarations. The for-of head accepts only a name today: the parser fails with `expected a name, found "#["` because `ForStmt` (`ast.go`) holds a single `Name`.

Accept in a for-of head every pattern a `const`/`let` declaration accepts: tuple `#[a, b]`, nested tuples, struct `{ x, y: other }`, and `let` as well as `const`.

## Approach

Prefer a desugaring in the parser, done the same way in `parser.go` and `selfhost/src/parser.nra`: `for (const <pattern> of iter) body` becomes `for (const <fresh name> of iter) { const <pattern> = <fresh name>; body }`. The checker and emitter then need no change, and errors from the pattern (wrong arity, missing field) come from the existing declaration checks. Pick the fresh name so it cannot clash with user names (shadowing is an error in Narra) and so the two compilers produce the same AST dump and the same C. If the desugaring gives poor error positions or a wrong effect or closure capture, report it in the outcome and move to a real pattern field on `ForStmt` only if needed (then also `clone.go`, checker and emitter in both compilers; that is past M, so split it).

## Acceptance

- [ ] Sample programs in `compiler/testdata/` cover tuple, nested tuple, struct and `let` heads over arrays, `Arrays.entries` and map entries. They run clean under ASan and the leak check.
- [ ] Compile-error tests in `compiler/testdata/errors/` cover a wrong tuple arity and a missing struct field in a for head, with positions on the pattern.
- [ ] The self-hosted compiler matches Go on the new samples: AST dump, types dump and emitted C (`TestSelfhost*` harnesses pick them up).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`
