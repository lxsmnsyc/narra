---
id: lsp-hover-parity
title: "LSP hover matches narrac query hovers"
status: active
priority: could
type: fix
story: batch-20
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
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/lsp-hover-parity/**
shared:
  - selfhost/src/tools/query.nra
  - selfhost/src/tools/querysignature.nra
  - selfhost/src/tools/lsp.nra
  - testdata/lsp/**
  - testdata/query/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on:
  - on: lsp-dependency-definition
    kind: hard
    why: both edit selfhost/src/tools/query.nra; land after it and rebase
---
# LSP hover matches `narrac query hovers`

## Goal

`narrac query hovers` shows module aliases inside template strings and the struct name and fields of a struct literal; the LSP hover does not, so editors miss them. Move both into the shared hover code (`describeSite` in `selfhost/src/tools/query.nra`) so `narrac lsp` gets them too.

Batch-17 release review nits, same code (inbox, 2026-11-16):

- With `import { Shape as S }`, `S.Circle(1)` shows `enum S` at `S` while the variant shows `Shape.Circle(...)`, because the enum hover uses the source text (`query.nra:313` on that commit). Show one name consistently (the declared name, with the alias where the source wrote one, as type aliases do).
- The local name in `import { area as measure }` has no hover.
- A module alias inside a pattern (`f is Shapes.Shape.Circle(r)`) has no module hover.
- A private function in a package with errors shows no `uses` line, and no golden covers it.
- Builtin hovers in the LSP changed from `void` plus `uses io` to the signature, covered by one golden only; add goldens for a few builtins.

## Acceptance

- [ ] Each item has an LSP golden that fails on the batch base and passes on the branch, or a note saying why it was left.
- [ ] `sh scripts/test.sh` passes.
