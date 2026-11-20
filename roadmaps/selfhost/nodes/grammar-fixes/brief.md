---
id: grammar-fixes
title: "TextMate grammar: labels and function types in type arguments"
status: done
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
  - docs/guide/16-unsafe-code-and-c.md
  - website/src/highlight/**
reserves: []
provides: []
depends_on: []
---
# TextMate grammar: labels and function types in type arguments

## Goal

The website build checks the TextMate grammar (`editors/textmate`) against the compiler's tokens (`website/scripts/check-grammar.ts`). Two mismatches found while writing the guide (2026-11-20):

- A label before a loop (`outer: for (...)`) is read as a type annotation. guide-basics-gaps added it to the known-differences list in `explainDifference` instead of fixing it.
- A function type inside nested type arguments (`Cell<Option<() => int>>`) fails the build; chapter 9 uses `type Handler = () => int` to avoid it.

- Names the site highlighter (`website/src/highlight/narra.ts`) and the TextMate grammar color differently, found by guide-c-interop: `I32`, `U8`, `U64` (site: constant, TextMate: class name) and `CStrings`, `CString` (site: class name, TextMate: identifier). Chapter 16 imports `std/cstring` as `Cstrings`, never writes `CString`, and avoids `std/i32` and `std/u8` because of it.

Fix all of them so the grammar and the site agree with the compiler's tokens, remove the known-difference entry, let chapter 9 write the type directly if that reads better (keep the alias if it reads better as it is), and change chapter 16 back to the spec's names (`CStrings`, `CString`, and the sized-integer modules where they read better). Editors that use the grammar get the fix too.

## Acceptance

- [ ] Every case above highlights like the compiler's tokens; the known-difference entry for labels is gone.
- [ ] `pnpm --dir website build` passes.
