---
id: grammar-fixes
title: "TextMate grammar: labels and function types in type arguments"
status: todo
priority: could
type: fix
story: batch-21
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/grammar-fixes/**
  - editors/textmate/**
shared:
  - website/scripts/check-grammar.ts
  - docs/guide/09-memory-and-mutable.md
reserves: []
provides: []
depends_on: []
---
# TextMate grammar: labels and function types in type arguments

## Goal

The website build checks the TextMate grammar (`editors/textmate`) against the compiler's tokens (`website/scripts/check-grammar.ts`). Two mismatches found while writing the guide (2026-11-20):

- A label before a loop (`outer: for (...)`) is read as a type annotation. guide-basics-gaps added it to the known-differences list in `explainDifference` instead of fixing it.
- A function type inside nested type arguments (`Cell<Option<() => int>>`) fails the build; chapter 9 uses `type Handler = () => int` to avoid it.

Fix both in the grammar, remove the known-difference entry, and let chapter 9 write the type directly if that reads better (keep the alias if it reads better as it is). Editors that use the grammar get the fix too.

## Acceptance

- [ ] Both cases highlight like the compiler's tokens; the known-difference entry for labels is gone.
- [ ] `pnpm --dir website build` passes.
