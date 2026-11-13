---
id: lint
title: narra lint
status: todo
priority: could
type: feature
outline: true
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

## Notes

Outline. Before refining, list the rules that real Narra code (the selfhost compiler) needs and that the checker does not already enforce. If the list is short, fold the rules into the checker as warnings and drop this node.
- Unused locals and parameters are checker errors (node unused-names, maintainer decision 2026-11-04, SPEC.md line 57), not lint findings.
