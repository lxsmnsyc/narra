---
id: std-header-comments
title: "std headers without SPEC and decision citations"
status: done
priority: should
type: chore
story: batch-23
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-header-comments/**
shared:
  - std/src/**
  - selfhost/src/std/stdtext.nra
  - generated/**
  - docs/reference/**
reserves: []
provides: []
depends_on: []
---
# std headers without SPEC and decision citations

## Goal

From the inbox (async-thread-functions review): 15 files in `std/src` cite SPEC sections and decision numbers in their header comments (for example `std/thread`: "SPEC.md section 8, D111", "D111 item 2"; `std/toml`: "D118 item 5"). The code-comment rule forbids D-numbers in code, and the headers show up in the generated std reference and in hovers.

Rewrite each citation as plain words that state the fact the citation pointed at, or drop it when the sentence already says it. Keep the rest of the header. Then regenerate the embedded std text and the generated pages: `sh scripts/test.sh --suite generated --update`.

## Acceptance

- [ ] `grep -nE 'D[0-9]{2,3}|SPEC' std/src/*.nra` finds no citation in comments.
- [ ] `selfhost/src/std/stdtext.nra` and the generated reference are regenerated.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
