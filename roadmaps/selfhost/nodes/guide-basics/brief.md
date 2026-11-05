---
id: guide-basics
title: "Language guide: sample test and chapters 1 to 5"
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
  - compiler/internal/narra/guide_test.go
  - docs/guide/README.md
  - docs/guide/01-*.md
  - docs/guide/02-*.md
  - docs/guide/03-*.md
  - docs/guide/04-*.md
  - docs/guide/05-*.md
  - roadmaps/selfhost/nodes/guide-basics/**
shared: []
reserves: []
provides: [guide-samples]
depends_on: []
---

# Language guide: sample test and chapters 1 to 5

## Goal

Write `TestGuide`, which checks every code sample in `docs/guide/*.md` by the format in `contracts/guide-samples.md`, the guide index, and chapters 1 to 5 (getting started, values and types, functions and closures, structs and enums, pattern matching). The chapter list and audience are in the story brief, `nodes/language-guide/brief.md`.

## Context

- `compiler/internal/narra/run_test.go`: `TestPrograms` (`runProgram`, `modes`, the `.out` comparison) and `TestCompileErrors` are the two checks the harness repeats for extracted samples. Reuse their helpers; do not copy them.
- `SPEC.md` sections 1 to 5 and 9 are the reference for these chapters (read them, do not cite them); `compiler/README.md` "Supported" and "Differences from the spec" say what the compiler implements. Programs under `compiler/testdata/` show working code for every feature.
- Source: GOAL.md decision 2026-11-01 (m9-adoption).

## Acceptance

- [ ] `TestGuide` reads `../../../docs/guide/*.md`, extracts samples, runs programs in every mode and compares stdout with the `text` block, checks error samples against their message, and fails on a `narra` block with no `text` block. Each sample is a subtest named `<file>:<line>`.
- [ ] A broken sample makes `TestGuide` fail with the file and line (checked once by hand, noted in `outcome.md`).
- [ ] `docs/guide/README.md` with the chapter list for all ten chapters (links to 6 to 10 may be dead until guide-advanced lands) and the "Planned" list of parked features.
- [ ] Chapters `01-` to `05-`, each with several tested programs. Write for a reader who is new to Narra and may not know TypeScript or Rust: every chapter must make sense on its own. Do not cite `SPEC.md` sections, D-numbers or `DECISIONS.md`, and do not frame rules as the author's decisions; state the rule and, where it helps, the practical reason. Comparisons with TypeScript or Rust ("unlike TypeScript...") go in chapter 11, not in chapters 1 to 10. List each comparison you leave out in `outcome.md` under "For chapter 11". (Changed 2026-11-05 by the maintainer.)
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty low (test pattern exists in run_test.go), coupling low, blast radius low (new files only), verifiability low (the test checks samples; prose is reviewed in the story pull request), size medium (7 files, mostly prose), context medium (SPEC sections 1 to 5).

## Notes for the worker

- Prefer one small program per idea over one big program per chapter.
- Use the names the maintainer uses (roadmap CLAUDE.md): full descriptive names, action-named functions.
- Do not edit `SPEC.md`. If the spec and the compiler disagree, follow the compiler, and list the disagreement in `outcome.md`.
