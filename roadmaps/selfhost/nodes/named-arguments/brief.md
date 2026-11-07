---
id: named-arguments
title: Named arguments, in any order (D87)
status: todo
priority: should
type: story
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - roadmaps/selfhost/nodes/named-arguments/**
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Named arguments, in any order (D87)

## Goal

A call can name its arguments, `move(x: 1, y: 2)`, as decided in DECISIONS.md D87 after the debate in `roadmaps/selfhost/debate/named-arguments/debate.md`. Positional arguments come first, then named ones in any order; each parameter is passed exactly once; arguments are evaluated in written order; callees without parameter names take positions only. Parameter defaults are not part of this story.

Children: named-args-syntax, named-args-check, named-args-query. They land on `feat/named-arguments`; the story is one pull request, because the syntax without the checker would accept programs it cannot compile.

## Acceptance

- [ ] Every child is done.
- [ ] SPEC.md section 2 (Functions) describes named arguments; the D12 line "no labeled arguments" is amended; D87 loses "Not built yet".
- [ ] Fast check passes on the story branch.
