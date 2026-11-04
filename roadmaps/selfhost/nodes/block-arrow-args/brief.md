---
id: block-arrow-args
title: "Block-body arrows with several statements as call arguments"
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent: 59925
owns:
  - compiler/testdata/block_arrow_args*
  - roadmaps/selfhost/nodes/block-arrow-args/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/lexer.go
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Block-body arrows with several statements as call arguments

## Goal

SPEC line 64. `apply((x) => { ⏎ print("io") ⏎ return x ⏎ })` fails with `expected the end of the statement, found "return"`; one statement works, and the same arrow bound with `const` works. Line breaks inside parentheses are probably ignored for statement ends, also inside a `{ }` block nested in `( )`. Statements in such a block must end at line breaks again. The README marks block arrows as supported, so this is a bug.

## Acceptance

- [ ] Sample programs pass block-body arrows with several statements as call arguments, nested in another call, and as the last of several arguments. Clean under ASan and the leak check.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on the new samples (AST dump, types dump, emitted C) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
