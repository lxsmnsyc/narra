---
id: grammar-fixes-two
title: "Highlighters: implements on the struct line, capitals after typeof"
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
  - roadmaps/selfhost/nodes/grammar-fixes-two/**
shared:
  - editors/textmate/**
  - website/src/highlight/**
  - website/scripts/check-grammar.ts
  - docs/guide/04-structs-and-enums.md
  - docs/guide/07-traits.md
reserves: []
provides: []
depends_on: []
---
# Highlighters: `implements` on the struct line, capitals after `typeof`

## Goal

Two more mismatches between the TextMate grammar and the site highlighter, found by guide-types-more after grammar-fixes (check whether they still fail on the current batch first):

- `struct Bundle<T: Label> implements Label {` on one line: TextMate colors `implements` as a plain name, the site as a keyword. Chapter 7 puts `implements Label` on its own line to avoid it, which `narrac fmt` would join.
- An all-caps constant after `typeof` (`typeof PRICES`): the two highlighters color it differently. Chapter 4's samples use `shopPrices` and `standardDiscount` instead.

Fix both so both highlighters agree with the compiler's tokens, then write the chapter 7 header on one line (as `narrac fmt` would) and use all-caps constant names in chapter 4 where they read better.

## Acceptance

- [ ] Both cases highlight the same on both sides; chapters 4 and 7 use the natural forms.
- [ ] `pnpm --dir website build` passes; the guide suite passes.
