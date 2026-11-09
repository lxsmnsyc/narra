---
id: lsp-navigation-parse
title: "LSP navigation reuses the parse cache"
status: active
priority: should
type: fix
story: batch-3
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - testdata/lsp/**
  - roadmaps/selfhost/nodes/lsp-navigation-parse/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# LSP navigation reuses the parse cache

## Goal

Inbox 2026-11-09 (from lsp-navigation and lsp-incremental): each navigation request (hover, definition, references, document symbols) checks the whole package again. The incremental cache cannot serve it, because it never builds a checker with every body, but navigation can reuse parsing through `LspCache.loadPackage`, about 37% of a whole check. Do that; a cached whole checker on `Server` is out of scope.

Prove it first: time a hover on a large package (selfhost/src itself) on the base, then on the branch, several runs each.

## Acceptance

- [ ] Timings on the base and the branch, with the gain beyond the noise.
- [ ] Every LSP session test still passes; answers are unchanged.
- [ ] `sh scripts/test.sh` passes.
