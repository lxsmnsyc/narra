---
id: field-shorthand
title: "Field shorthand in struct literals"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: p3-review
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on: []
---

# Field shorthand in struct literals

## Goal

D119: `Point { x, y }` means `Point { x: x, y: y }`. Parser, checker, fmt (keeps what was written), astdump, SPEC and guide; then a sweep of the compiler's 512 `name: name` sites after a seed update.
