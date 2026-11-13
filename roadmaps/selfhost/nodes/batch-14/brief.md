---
id: batch-14
title: "Batch 14: standard library for the tool layer (phase p2)"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: p2-std
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/batch-14/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-12
    kind: hard
    why: std modules build on the batch-12 compiler and seed
---
# Batch 14: standard library for the tool layer (phase p2)

## Goal

Fourteenth one-shot batch, scope set by the maintainer on 2026-11-13: the std modules phase p2 lists, so the tool layer (narrac, lint, package tools) can be written in Narra on std. Runs beside batch-13 on its own branch, `feat/batch-14`, and merges on its own.

Children:

- std-design (S spike, gate D118): the public API of each module below, TypeScript-like, with errors as values; one gate for all.
- std-iterator, std-arena, std-unicode, std-args, std-toml, std-regex, std-terminal: one build node each, refined by std-design.

## Order and shared files

- std-design first; the seven builds start after its gate, four at a time.
- Each module is its own `std/src/<name>.nra` (plus runtime C only when needed), its own `selfhost/src/std<name>.nra` registry rows, and its own samples. `stdlib.nra`, the generated `stdtext.nra`, `diagnostics.nra` and the module-list goldens are shared: append rows, regenerate generated files, and the second to land rebases.
- Codes: E1300-E1399 (standard library for tools), split by std-design. Decision D118.
- batch-13 may change syntax std uses (for example the `uses` list form): whichever batch merges second rebases and fixes std sources.

## Follow-ups after batch-13 merges

- `chunk` and `window` in std/iterator take the D118 signature (an effect variable) once effect-var-nested-result lands.
- Private functions in std sources and samples that write an incomplete `uses` list are E0800 since effect-var-nested-result (batch-13); fix them with one word each when rebasing.
- `Command.rest` in std/args gets `= Rest.None` once constructor-field-defaults lands.

## Seed

`std/src` modules compiled by the seed may use only features the seed compiles. A module the compiler itself will use (args, toml, terminal) must stay seed-compilable.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated if any runtime text changed, and `sh selfhost/bootstrap.sh` reports it current.
- [ ] SPEC lists each new module; the language guide's std chapter covers it.
- [ ] A fresh review of the whole story branch, before the pull request.
- [ ] `bench/BASELINE.md` has an end-of-batch row if a module adds a benchmark.
