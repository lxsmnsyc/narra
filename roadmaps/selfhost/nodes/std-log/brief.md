---
id: std-log
title: "std/log: leveled logging to stderr"
status: todo
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-log/**
  - std/src/log.nra
  - testdata/log*
shared:
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
reserves: []
provides: []
depends_on: []
---

# std/log: leveled logging to stderr

## Goal

A new `std/log` module writes log lines with a level (`debug`, `info`, `warn`, `error`) and key-value fields to standard error, in a text form and a JSON-lines form. A minimum level filters lines out.

## Context

- Source: std-gap-survey rank 22; `inbox.md` 2026-11-20 (Crystal comparison); `GOAL.md` Decisions, 2026-11-21 (log stays in std).
- Go `log/slog` is the reference for the field model. Keep it smaller: no handler interface, no context, no groups. A logger is a value with a level, a format and fields added with `with`.
- Narra has no implicit calls or operator overloading, so the fields are an explicit list of pairs, not varargs.

## Acceptance

- [ ] A sample logs at each level and its stderr matches a golden for both formats, with the level filter applied.
- [ ] JSON output escapes field values with the same rules as `std/json`.
- [ ] `sh scripts/test.sh` passes.

## Sizing

S: one module over `std/json` and stderr writes.
