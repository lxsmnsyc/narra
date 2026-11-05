---
id: website
title: Narra website
status: todo
priority: later
type: feature
story:
early: false
outline: true
milestone: m9-adoption
target:
ship: auto
from:
size: L
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
  - on: language-guide
    kind: hard
    why: the site publishes the guide chapters and their samples
---

# Narra website

## Goal

A website for Narra that publishes the language guide, the installation guide and `SPEC.md`, with syntax highlighting for `narra` code fences. Requested by the maintainer on 2026-11-05. It is the reason guide samples use the `narra` fence, not `ts`.

## Notes

Outline, parked (`later`). Decide when refined: the site generator, hosting, the highlighter (a TextMate grammar that editors can reuse too), and whether `SPEC.md` samples switch from `ts` to `narra` fences.
