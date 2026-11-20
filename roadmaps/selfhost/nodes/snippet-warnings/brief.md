---
id: snippet-warnings
title: "Doc samples print no unrelated warnings"
status: todo
priority: could
type: docs
story: batch-22
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/snippet-warnings/**
shared:
  - docs/**
  - SPEC.md
  - README.md
  - website/src/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on:
  - on: async-thread-functions
    kind: hard
    why: rewrites thread and async samples
  - on: entry-args-unused
    kind: hard
    why: most warnings are the args warning
  - on: std-async-fold
    kind: hard
    why: rewrites many samples
---
# Doc samples print no unrelated warnings

## Goal

Maintainer, 2026-11-20: some samples show warnings unrelated to the example. After entry-args-unused, collect every warning and error the website shows inline on samples (the twoslash plugin, `website/plugins/twoslash.ts`, runs `narrac query --json hovers` on each sample in a package it writes; the built pages under `website/dist` show the diagnostics) and every warning the samples of `docs/guide/**`, `docs/reference/**`, `docs/install.md`, `README.md`, `SPEC.md` and the home page (`website/src/routes/index.brm`) produce with `narrac check`, classify each (related to what the sample teaches, or noise), and remove the noise by fixing the sample. A sample that teaches a warning keeps it and says so in its fence (find how the guide suite marks expected output; add a form for an expected warning if none exists). Then make the website build (or a check it runs) and the guide suite (and the spec suite, if it runs samples) fail when a sample prints a warning it does not expect, so it stays clean.

## Acceptance

- [ ] No sample prints an unexpected warning; the suites enforce it.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
