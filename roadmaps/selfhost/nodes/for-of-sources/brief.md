---
id: for-of-sources
title: "for-of over any Range, Array or Str.chars value"
status: active
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/for_of_sources*
  - compiler/testdata/errors/for_of_sources_*
  - roadmaps/selfhost/nodes/for-of-sources/**
shared:
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/std_array.go
  - compiler/internal/narra/std_range.go
  - compiler/internal/narra/std_string.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/stmts.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
reserves: [diag E0920]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# for-of over any Range, Array or Str.chars value

## Goal

SPEC lines 110-111 (F6). `for ... of` accepts only direct calls to `Ranges.upTo`, `Ranges.between`, `Arrays.values`, `Arrays.entries`, `Maps.keys` and `Maps.entries`. `const r = Ranges.upTo(3) ⏎ for (const i of r)`, `for (const item of ["a"])` and `for (const c of Str.chars("ab"))` fail with E0232. Also, `Arrays.values` and `Arrays.entries` used outside a for-of panic the Go emitter (`emit: unknown builtin`) instead of giving a compile error.

## Approach

Accept as a for-of source: any expression of type `Range` (iterate its bounds), any expression of type `Array<T>` (as `Arrays.values`), and `Str.chars(s)` (characters; check SPEC for whether `Str.chars` returns an array or is an iterator source only, and follow it). Evaluate the source once, before the loop, and hold it for the loop's length. `Arrays.values` and `Arrays.entries` outside a for-of head become a compile error (or, if SPEC gives them a value meaning, that value). Update the E0232 message to the new list. Generators and `std/iterator` are separate nodes; do not add a general iterator protocol.

## Acceptance

- [ ] Sample programs cover a `Range` in a `const`, a struct field and a function result; an array literal, an array local and an array returned by a call; `Str.chars` over ASCII and multi-byte text; nested loops and `break` (clean under ASan and the leak check, including an array source freed after the loop).
- [ ] `Arrays.values(xs)` outside a for-of is a compile error with a position, not an emitter panic; error test included.
- [ ] E0232 lists the accepted sources.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
