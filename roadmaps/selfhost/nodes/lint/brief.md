---
id: lint
title: narra lint
status: dropped
priority: could
type: feature
outline: false
milestone: p6-docs
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: query-cli
    kind: hard
    why: lint rules read the same resolved program and types as narra query
---

# narra lint

## Goal

A linter for problems the checker accepts on purpose: unused imports, unreachable code, a `let` that is never reassigned (should be `const`), and style rules from the guide. JSON output like the compiler diagnostics, and the language service shows the findings.

## Dropped (2026-11-14, batch-16 refinement)

The rules the selfhost compiler needs and the checker does not enforce are two: unreachable statements and a `let` that is never reassigned. Unused imports, locals, parameters and private functions are already errors, and shadowing is already an error. Style is `narrac fmt`. Two rules do not justify a separate tool, so they become checker errors in nodes dead-code-check and let-never-reassigned. Reopen this node if a rule appears that the checker should not reject.

## Notes

Outline. Before refining, list the rules that real Narra code (the selfhost compiler) needs and that the checker does not already enforce. If the list is short, fold the rules into the checker as warnings and drop this node.
- Unused locals and parameters are checker errors (node unused-names, maintainer decision 2026-11-04, SPEC.md line 57), not lint findings.
