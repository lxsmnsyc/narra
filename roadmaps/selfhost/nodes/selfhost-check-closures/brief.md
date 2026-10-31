---
id: selfhost-check-closures
title: Self-hosted checker: arrow functions and function values
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostClosures'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/closures.nra
  - compiler/internal/narra/selfhost_closures_test.go
  - compiler/testdata/typesdump/closures/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/testdata/typesdump/**
  - compiler/README.md
  - selfhost/src/calls.nra
  - selfhost/src/stmts.nra
  - selfhost/src/check.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: defines the body context, locals, per-body tables, the dispatch arms to replace and the dump; no frozen interface exists yet
  - on: selfhost-check-calls
    kind: hard
    why: edits the argument loop in calls.nra; the orchestrator serialized it after calls on 2026-10-30 to avoid conflicts
---

# Self-hosted checker: arrow functions and function values

## Goal

Port to Narra the checking of arrow functions, functions used as values, calls of function values, non-escaping parameters, captured locals, return inference and arrow parameter hints. Compare through the `narra types` dump.

## Context

- Read the selfhost-check-bodies outcome. Go source (about 480 lines, 650 Narra): `check_closure.go` (`closureState`, `closureOf`, `funcOf`, `finishFuncTypes`, `nonEscapingParam`, `funcParam`, `funcArg`, `useLocal`, `declaredBelow`, `funcValue`, `moduleFuncValue`, `fieldFunc`, `valueCall`, `valueCallOf`, `checkCalleeExclusive`, `mutableRoots`, `hintArrow`, `arrow`, `returnsOf`, `blockResult`, `inferReturns`). `cloneArrow` is not needed: the Narra AST is immutable and results live in per-body tables.
- Function types in Narra are `TFunc(params, result)` values, so `funcOf` interning and `finishFuncTypes` (listing for the emitter) are not ported; record in the outcome that selfhost-emit collects them.
- Hooks: selfhost-check-calls leaves `unsupported: arrow argument` in the generic argument loop and in `argsLoose`; replace it with `hintArrow`. The call-site `valueCall` is a stub in the call switch.
- Dump: you own `ArrowExpr` lines with attributes `stack` (stack environment), `captures=<local>,<local>`, `Local` lines for arrow parameters, and `func=` on an identifier used as a value. Add them to both printers and the Owners list.
- Closure state in Go is per function, saved and restored in `closureFunc`; the Narra equivalent is a field of the body context that each arrow saves and restores.

## Acceptance

- [ ] `TestSelfhostClosures` compares `narrac types` with `narra types` on earlier inputs plus `testdata/typesdump/closures/`: capturing and non-capturing arrows, nested arrows, arrows returned from functions, stored in structs, passed to generic functions (hint from the other arguments), non-escaping parameters, function names as values, module functions as values, calls of function values and fields holding functions, return inference with several returns, and the errors (escaping non-escaping parameter, capture of a mutable, wrong arrow arity, cannot infer parameter type, exclusive callee).
- [ ] The `stack`/`captures` attributes equal Go's for every sample in `testdata/closures*`.
- [ ] Earlier harnesses pass; `narra emit` output unchanged.
- [ ] The outcome lists what the emitter needs from the closure tables.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 650 Narra lines, one known pattern. Coupling medium: it edits the call argument loop owned by selfhost-check-calls.

## Notes for the worker

- Merge order: if selfhost-check-calls has not merged, build against selfhost-check-bodies and keep your edits in `calls.nra` to the argument-loop arm; the orchestrator rebases.
- Go's captured-local analysis walks scopes with `declaredBelow`; the Narra scope stack carries a depth number per local, so a capture is `local.depth < arrow.depth`.
- From std-option-result: a capturing arrow passed to a builtin helper takes a stack environment (`Stack = true`) when the helper's `callbacks` index lists the argument; a local function-typed parameter is allowed as non-escaping. The check itself is selfhost-check-builtins, but provide `checkArrowArg(mutable c, e, want)` for it and say so in the outcome.
- From std-option-result: a generic function value passed as a callback needs its type arguments from context; `Option.map(o, genericFn)` may report "cannot tell type parameter". Keep Go's behavior.
- Newlines inside call parentheses are ignored by the lexer, so a multi-statement arrow cannot be written inline in a test; define it as a `const` first.
