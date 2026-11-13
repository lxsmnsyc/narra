---
id: arrow-arg-order
title: "Arguments before an arrow that writes a captured let are not held"
status: done
priority: should
type: fix
story: batch-12
early: false
outline: false
milestone: 
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
added: 2026-11-13
spent:
owns:
  - testdata/arg_order_arrow_*
  - roadmaps/selfhost/nodes/arrow-arg-order/**
shared:
  - selfhost/src/emitexprs.nra
reserves: []
provides: []
depends_on: []
---
# Arguments before an arrow that writes a captured let are not held

## Goal

`pair(counter, runIt(5, (n) => { counter = counter + n; return counter }))` prints `6,6`, not `1,6`, because `mayWrite` (`selfhost/src/emitexprs.nra` around line 637) treats an arrow argument as non-writing, so the earlier argument is not held in a temporary. Count a call with an arrow argument that writes a captured `let` as a writer.

## Context

- From the inline review, 2026-11-12. On the base too.
- async-core adds async arrows and `async { ... }` blocks to the same `mayWrite`; whichever lands second makes the rule cover them too (an async block that writes a captured `let` is a writer).
- Proof first: the sample prints `6,6` on the base.

## Acceptance

- [ ] `testdata/arg_order_arrow_*` prints `1,6`, plus a block-bodied arrow that only reads (no temporary needed). Clean under AddressSanitizer and the leak check.
- [ ] `narrac emit selfhost` output does not grow by more than a few temporaries.
