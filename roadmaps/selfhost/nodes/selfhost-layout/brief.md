---
id: selfhost-layout
title: Group the self-hosted compiler's modules into directories
status: todo
priority: should
type: refactor
story:
early: false
outline: true
milestone: m9-adoption
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-source
    kind: hard
    why: its estimate for a checked-in C seed decides whether the Go compiler goes away, and so whether `selfhost/` is the compiler's final home or moves (for example to `compiler/`)
---

# Group the self-hosted compiler's modules into directories

## Goal

`selfhost/src/` holds about 50 modules in one flat directory, and fmt, query, the language service and lint will add more. Group them by area (for example lexer and parser, checker, emitter, tools), so a reader finds the code for a feature in one place. Behavior-preserving: the emitted C, every dump and the bootstrap fixed point stay byte-identical. Requested by the maintainer on 2026-11-05.

## Notes

Outline. Decide when refined:

- Whether `selfhost/` is final. If the std-source gate chooses to retire the Go compiler, decide the final location first (for example `selfhost/` becomes `compiler/`), and do the move and the regrouping in one change.
- Whether the module system resolves nested paths (`src/check/body.nra` as `"narrac/check/body"`). SPEC section 15 only shows flat files. If it does not, adding nested paths to both compilers is a separate early node.
- Hashed C names include module paths, so check whether the emitted C stays identical or every golden C file changes.
- The move touches every file in `selfhost/src/` and conflicts with every open node that edits it. Run it in a quiet window with no open selfhost nodes, next to the "reformat the repo" item in `inbox.md`, and land it before the next wide wave starts.
