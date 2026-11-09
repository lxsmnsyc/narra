---
id: named-args-check
title: Check and lower named arguments
status: done
priority:
type: feature
story: batch-1
early: false
outline: false
milestone:
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent: 230510
owns:
  - testdata/named_args*
  - testdata/errors/named_args*
  - testdata/diagnostics/named_args*
  - roadmaps/selfhost/nodes/named-args-check/**
shared:
  - selfhost/src/calls.nra
  - selfhost/src/emitexprs.nra
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0328, diag E0329, diag E0330, diag E0331, diag E0332]
provides: []
depends_on:
  - on: named-args-syntax
    kind: hard
    why: needs the name on Arg in both compilers
  - on: arg-eval-order
    kind: hard
    why: reordered arguments are evaluated in written order through the temporaries that fix adds
---

# Check and lower named arguments

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

The checker matches named arguments to parameters and the emitter calls the function with its arguments in declaration order, evaluated in written order (D87). Errors: an unknown name (with an exact fix when one parameter name is close, D11), a parameter passed twice, a positional argument after a named one, a missing parameter (the existing count error, naming the parameter), and a named argument to a callee without parameter names (a function value, or a std builtin registered with unnamed `Params`): "this callee has no parameter names, pass arguments by position".

## Context

- One canonical view of a call's arguments for every pass after the checker (effects, closures, exclusivity, tail calls, `useCollector`): decide it at the start and write it in the outcome. The AST keeps the written order for astdump, fmt and query (O7 in the debate).
- A reordered argument with side effects is evaluated into a temporary in written order, using the arg-eval-order rule. A call written in declaration order must emit the same C as the positional call.
- Generic calls infer type arguments from named arguments as from positional ones. `mutable` named arguments go through `checkMutableArg` like positional ones.
- Differential test: Go and selfhost must emit identical C for every named-argument sample (the generators node shipped a bug both compilers shared).

## Acceptance

- [ ] Samples cover reordering, mixing, `mutable`, generic callees, imported callees, a side effect that shows written-order evaluation, and a self tail call with named arguments.
- [ ] Error tests for each error above, with codes E0328-E0332; a `testdata/diagnostics/` case for the close-name fix.
- [ ] SPEC.md section 2 describes named arguments and amends the D12 line; D87 loses "Not built yet"; the Differences from TypeScript table needs no row (TypeScript has no such syntax).
- [ ] Fast check passes, including `TestSelfhostBootstrap`.

Landed on feat/batch-1 as 7559ad8 and b532619.
