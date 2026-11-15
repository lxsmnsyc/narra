---
id: query-hovers-gaps
title: "Hovers inside templates and for builtin functions"
status: active
priority: could
type: fix
story: batch-17
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/query-hovers-gaps/**
  - selfhost/src/tools/query*.nra
  - testdata/query/**
shared:
  - selfhost/src/tools/lsp*.nra
  - seed/narrac.c
  - SPEC.md
  - docs/guide/12-tools.md
reserves: []
provides: []
depends_on:
  - on: query-hovers
    kind: hard
    why: fixes gaps in its command
---
# Hovers inside templates and for builtin functions

## Goal

Found by the orchestrator on the site with the real `narrac query hovers` (2026-11-16):

    import * as Int from "std/int"
    export function run(args: Array<string>): int uses io, alloc {
      const text = Int.toString(4)
      print(text)
      print(`${Int.toString(4)}`)
      return 0
    }

- Outside a template, `Int` shows `module std/int` and `toString` shows `(int) => string uses alloc`. Inside `${...}` of a template string, `Int` gets no hover and `toString` shows `string`, the call's result.
- `print` shows `void`, the call's result, on every line. A builtin function called by name should show its signature, with effects (`(string) => void uses io` or what the checker records).

Fix both at the cause, so the LSP hover improves too. Look for other places where a callee name shows the call result instead of the signature (method-style calls, generic calls, calls through a module inside other nested expressions such as arguments, array literals and match arms) and fix them the same way.

## Acceptance

- [ ] The sample above gives the same hovers for `Int` and `toString` inside and outside the template, and a signature for `print`.
- [ ] Cases added to `testdata/query/hovers_names`.
- [ ] `sh scripts/test.sh` passes; seed regenerated until "seed unchanged".
