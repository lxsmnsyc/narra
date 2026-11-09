---
id: template-positions
title: "Real positions for tokens inside template interpolations"
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
  - testdata/template_position*
  - testdata/errors/template_position*
  - roadmaps/selfhost/nodes/template-positions/**
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

# Real positions for tokens inside template interpolations

## Goal

Inbox 2026-11-05 (from the PR #88 review): the parser gives every token inside a template literal's `${...}` the template's start position (`selfhost/src/parser.nra`, `parseTemplate`). Diagnostics and `narra query` locations inside templates point at the backtick; `narra query` works around it (`inTemplate` in query). Give those tokens their real line and column, update the AST dump goldens, and drop the query workaround.

Prove it first: an error inside `${...}` on the base reports the backtick's column.

## Acceptance

- [ ] Proof on the base and the branch: an error test with a type error inside `${...}` on a later line and column.
- [ ] `narra query` finds a name inside a template without the workaround; the workaround is removed.
- [ ] AST dump goldens updated; no other golden changes.
- [ ] `sh scripts/test.sh` passes.
