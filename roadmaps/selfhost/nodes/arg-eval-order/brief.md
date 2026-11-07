---
id: arg-eval-order
title: Evaluate call arguments in written order
status: todo
priority: must
type: fix
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - compiler/testdata/arg_eval_order*
  - roadmaps/selfhost/nodes/arg-eval-order/**
shared:
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/emit_expr.go
  - selfhost/src/emitexprs.nra
  - selfhost/src/emittail.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Evaluate call arguments in written order

## Goal

Fix a bug on main, found in the PR #96 review (inbox, 2026-11-07): call arguments are not evaluated left to right. With `let items = ["a", "b", "c"]`, `show(items, { items = ["z"]; return 0 })` prints 1, not 3, because the block argument runs before the earlier argument is read. Every call evaluates its arguments in the order they are written, in both emitters, and SPEC section 2 says so. D87 (named arguments) depends on this rule.

## Context

- When a later argument can have side effects (a call, a block expression, a `match` or `if` value, anything that assigns), evaluate the earlier arguments into temporaries first. Arguments that are plain reads of constants or literals need no temporary; keep the generated C unchanged for calls whose arguments have no side effects, so most samples keep byte-identical C.
- Self tail calls hoist arguments in index order (`emit_call.go` tailJump, `selfhost/src/emittail.nra`); they follow the same rule.
- `mutable` arguments are borrows, not values: a later argument that changes the borrowed value is already an exclusivity error (check this with a sample) and needs no temporary.
- Builtins with special emit paths (std `Mutable` parameters, `Arrays.push`) must follow the rule too.

## Acceptance

- [ ] The PR #96 example prints 3 in both compilers. Samples `compiler/testdata/arg_eval_order*.nra` cover a block argument, a call with a side effect after a read of the same `let`, a self tail call, and a std builtin call.
- [ ] Generated C is unchanged for calls with side-effect-free arguments; list the samples whose C changed in the outcome.
- [ ] SPEC.md section 2 states that arguments are evaluated left to right; remove the inbox item (2) of the PR #96 entry.
- [ ] Fast check passes, including `TestSelfhostBootstrap`, clean under AddressSanitizer and the leak check.

## Sizing

M. One rule in each emitter, but several call paths (user calls, closures, builtins, tail calls) and a C-output comparison over all samples.
