---
id: selfhost-modules
title: Self-hosted package loading
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-front
    kind: hard
    why: reads each module's imports from the parsed AST
---

# Self-hosted package loading

## Goal

Port `modules.go` and `modules_toml.go` (about 570 Go lines): read `narra.toml`, walk `src/**`, parse each module, order modules by imports, and pick the binary entry. Compare with the Go side through a dump of the package (manifest fields and module order), as the lexer and parser are compared.

## Notes

- Split from selfhost-front on 2026-10-30.
- No directory test or recursive walk in std: `Fs.listDir` returns Err on a file, so recurse on Ok (spike gap 15).
- Outline. Break into a full brief when `plan.py refine` lists it.
