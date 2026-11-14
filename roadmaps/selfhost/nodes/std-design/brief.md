---
id: std-design
title: "Design: std APIs for the tool layer"
status: done
priority: should
type: spike
story: batch-14
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
spike: true
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent: 175745
owns:
  - roadmaps/selfhost/nodes/std-iterator/**
  - roadmaps/selfhost/nodes/std-arena/**
  - roadmaps/selfhost/nodes/std-unicode/**
  - roadmaps/selfhost/nodes/std-args/**
  - roadmaps/selfhost/nodes/std-toml/**
  - roadmaps/selfhost/nodes/std-regex/**
  - roadmaps/selfhost/nodes/std-terminal/**
  - roadmaps/selfhost/nodes/std-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D118, diag E1300-E1399]
provides: []
depends_on: []
---
# Design: std APIs for the tool layer

## Goal

Write D118: the public API of seven std modules, TypeScript-like names, errors as `Result` values, no hidden calls (D12), no panics outside D92's list. Then refine the seven build briefs.

## Context

- Existing std shape: `std/src/*.nra` and the registry modules in `selfhost/src/std*.nra`; `std/json` and `std/shape` (D83, D84) as a model for parse and stringify APIs; `std/fs`, `std/process`, `std/env`.
- Each module's outline brief has its open points (read `nodes/<module>/brief.md`).
- What the tool layer needs: narrac's own argument parsing (`selfhost/src/main.nra`, `driver.nra`), `narra.toml` reading (`selfhost/src/manifest.nra`), diagnostics colors, lint patterns.

## Questions

1. std/iterator: the operator list, callbacks escaping, cost per stage (measure), whether operators consume their source.
2. std/arena: `Arena<T>`, `Handle<T>` with generations (SPEC section 6, around line 1273).
3. std/unicode: graphemes, case mapping, normalization; tables as generated Narra source or C data; the Unicode version; size cost.
4. std/args: declarative spec or builder; subcommands; help text; how narrac would use it.
5. std/toml: TOML 1.0 parse and stringify through `Shape` like `std/json`; dates and times; how `manifest.nra` would move onto it.
6. std/regex: syntax flavor, a linear-time engine (no backtracking), `@const`-checked patterns, captures API.
7. std/terminal: styles and colors, TTY detection, width, NO_COLOR; std tags for tagged templates if D114 adds them (coordinate through the outcome, do not wait).
8. Which modules need runtime C (unicode tables, terminal ioctl, regex engine) and which are pure Narra source.

- Reserved codes: E1300-E1399 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build briefs to refine: `nodes/std-iterator/brief.md`, `nodes/std-arena/brief.md`, `nodes/std-unicode/brief.md`, `nodes/std-args/brief.md`, `nodes/std-toml/brief.md`, `nodes/std-regex/brief.md`, `nodes/std-terminal/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D118: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build briefs listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D118 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D118 answers every question and is approved by the maintainer.
- [ ] The build briefs follow D118, or the outcome says why the feature is dropped and which SPEC lines change.
