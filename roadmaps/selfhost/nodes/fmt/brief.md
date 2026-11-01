---
id: fmt
title: narra fmt
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
  - on: selfhost-bootstrap
    kind: hard
    why: written in Narra on the self-hosted parser, which must keep comments
---

# narra fmt

## Goal

A formatter that defines the one valid layout, with `--check`. Written in Narra on the self-hosted parser (decision 2026-11-01), so it is not built in Go and ported.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
