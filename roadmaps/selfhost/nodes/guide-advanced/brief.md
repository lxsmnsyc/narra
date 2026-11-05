---
id: guide-advanced
title: "Language guide: chapters 6 to 10"
status: active
priority:
type: docs
story: language-guide
early: false
outline: false
milestone:
target:
ship: auto
from:
model: opus
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run TestGuide
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - docs/guide/06-*.md
  - docs/guide/07-*.md
  - docs/guide/08-*.md
  - docs/guide/09-*.md
  - docs/guide/10-*.md
  - roadmaps/selfhost/nodes/guide-advanced/**
shared: []
reserves: []
provides: []
depends_on:
  - on: guide-basics
    kind: behavior
    via: guide-samples
---

# Language guide: chapters 6 to 10

## Goal

Write chapters 6 to 10 of `docs/guide/`: Option and Result, traits, modules and packages, memory and `mutable`, effects and testing. Every sample follows `contracts/guide-samples.md`. The chapter list and audience are in the story brief, `nodes/language-guide/brief.md`.

## Context

- `SPEC.md` sections 3 (modules), 5 (traits), 6 (memory and `mutable`), 7 (effects) and the `std/test` part of `compiler/README.md`; `DECISIONS.md` for the reasons (read them, do not cite them). `compiler/testdata/` has working programs for each feature (`std_test*.nra` and `tests.nra` for testing).
- Chapter 8 needs a multi-module example. A guide sample is one file, so show the package layout as a `narra fragment` and keep the tested programs single-file, or note in `outcome.md` that the harness needs a multi-file sample form (then guide-basics or a follow-up adds it).
- Source: GOAL.md decision 2026-11-01 (m9-adoption).

## Acceptance

- [ ] Chapters `06-` to `10-`, each with several tested programs. Write for a reader who is new to Narra and may not know TypeScript or Rust: every chapter must make sense on its own. Do not cite `SPEC.md` sections, D-numbers or `DECISIONS.md`, and do not frame rules as the author's decisions; state the rule and, where it helps, the practical reason. Comparisons with TypeScript or Rust ("unlike TypeScript...") go in chapter 11, not in chapters 1 to 10. List each comparison you leave out in `outcome.md` under "For chapter 11". (Changed 2026-11-05 by the maintainer.)
- [ ] Chapter 9 explains why `let` cannot be captured by a closure and what a `mutable` parameter allows, with error samples for both. The comparison with TypeScript references goes to chapter 11.
- [ ] Chapter 10 shows a `std/test` test and how `narra test` runs it (as a fragment, since `TestGuide` runs `run`).
- [ ] `TestGuide` passes on the story branch once guide-basics has landed.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty low, coupling low (one behavior dependency), blast radius low (new files only), verifiability low (samples tested, prose reviewed in the story pull request), size medium (5 files, mostly prose), context medium (SPEC sections 3, 5, 6, 7).

## Notes for the worker

- closure-borrow and cell-capture (m11) will change what chapter 9 says about closures and `let`. Describe today's rules and do not promise the change.
- Do not edit `SPEC.md` or the chapters 1 to 5.
