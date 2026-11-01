---
id: query-cli
title: narra query
status: todo
priority: should
type: feature
outline: true
milestone: m9-adoption
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
  - on: modules
    kind: behavior
    via: module-graph
  - on: selfhost-bootstrap
    kind: hard
    why: written in Narra on the self-hosted checker; the language service, lint and skills build on it
---

# narra query

## Goal

`narra query type-of|def|refs|callers|interface|recursion`, with `--json`. Written in Narra on the self-hosted checker (decision 2026-11-01).

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
