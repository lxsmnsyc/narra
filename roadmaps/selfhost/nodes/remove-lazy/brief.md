---
id: remove-lazy
title: "Remove lazy values"
status: done
priority: must
type: feature
story: batch-15
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 125187
owns:
  - roadmaps/selfhost/nodes/remove-lazy/**
  - selfhost/src/lazyvalues.nra
  - selfhost/src/stdlazy.nra
  - testdata/lazy*
  - testdata/errors/lazy*
shared:
  - selfhost/src/*.nra
  - std/src/*.nra
  - docs/guide/*.md
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# Remove lazy values

## Goal

D119: remove `const x = lazy expr`, `std/lazy` and `Lazies.force`. Retire E1032-E1035: keep their rows marked retired the way other retired codes are (check how the `codes` suite treats retired codes), never reuse them. `lazy` stops being a keyword if nothing else uses it. No deprecation path.

## Context

- Files: `lazyvalues.nra`, `stdlazy.nra`, and lazy cases in ast, astdump, consteval, declsdump, emit, effects, emitconsts, emitstd, exprs, fmt, parser, stdcalls, stdlib, stdregex (`grep -rln lazy selfhost/src`). Check `stdregex.nra`: if it uses lazy values for compiled patterns, replace them with what top-level constants already give (D61 getters).
- Remove samples and goldens, SPEC text and the guide section.
- The compiler gets smaller: record narrac size and `emit selfhost` instructions before and after in `outcome.md`.

## Acceptance

- [ ] `grep -rn -i lazy selfhost/src std/src SPEC.md docs/guide` finds only unrelated words.
- [ ] `sh scripts/test.sh` passes, `codes` suite included.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
