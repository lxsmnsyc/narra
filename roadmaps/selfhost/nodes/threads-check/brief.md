---
id: threads-check
title: "Sendable, Scope escape rules, thread function rules and std/parallel"
status: active
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/threads-check/**
  - selfhost/src/threadcheck.nra
  - std/src/parallel.nra
  - testdata/threads_check_*
  - testdata/parallel_*
  - testdata/errors/threads_*
shared:
  - selfhost/src/traits.nra
  - selfhost/src/types.nra
  - selfhost/src/stdthread.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/diagnostics.nra
  - std/src/thread.nra
  - compiler/README.md
  - SPEC.md
reserves: [diag E1190-E1199]
provides: []
depends_on:
  - on: threads-design
    kind: hard
    why: D111 item 1 fixes the Sendable rule and the codes
  - on: threads
    kind: hard
    why: the checks apply to the std/thread functions and types that node builds
---
# Sendable, Scope escape rules, thread function rules and std/parallel

## Goal

Build D111 item 1 and the `Scope` and equality rules of item 4: the automatic trait `Sendable`, the checks where a value crosses a thread, the `Scope` escape rule, the form of thread functions, and `==` on thread handle types. Add `std/parallel` with `Parallel.map(items, f)`.

## Context

- Read D111 and the outcomes of threads-design and threads first.
- `Sendable` is automatic like `Equatable` and `Hashable` (`nominalImplements` in `traits.nra`, where D90 makes `Cell` refuse Equatable). Not Sendable at any depth: function types, `Dynamic<Trait>`, `Ptr<T>`, `Cell` of `std/cell`, `Async<T>`, `Generator<T>`, `Coroutine<In, Out>`, `Scope`, `Thread<T>`, the builtin `Handle` (sockets). Put the walker and the checks in a new module `threadcheck.nra`; hot files carry only the call.
- Checked sites: each capture of the function argument of `Thread.start` and `Parallel.map`, the result type of that function, and later the type arguments of `Channel.create`, `Shared.create` and `Atomic.create` (threads-channels and threads-shared call the same helper). Find the functions by module and name. `std/internal/thread` calls are not checked, so std source can pass a function parameter on.
- `Parallel.map(items, f)` is source over `std/internal/thread`: one part per core (a primitive for the core count), results in order.

## Codes (E1190-E1199)

- E1190 a type that is not `Sendable` where a value crosses threads; the message names the type and the part, the hint for a `Cell` names `Shared`.
- E1191 a `Scope` stored in a field, element, tuple or constructor, returned, or captured by an escaping closure (a thread function included).
- E1192 an `async` thread function.
- E1193 a thread function that is not an arrow function at the call or a top-level function name.
- E1194 `implement Sendable`.
- E1195 `==`, `!=` or a map key on `Thread`, `Scope`, `Shared`, `Atomic`, `Sender` or `Receiver` (the last four are added by their nodes, through the same check).
- E1196-E1199 free.

## Acceptance

- [ ] Proof first: `testdata/errors/threads_cell_capture.nra` (a thread function captures a `Cell`) compiles on the base of this node and fails with E1190 on the branch.
- [ ] Error tests, one per code and site: `threads_cell_capture` (E1190), `threads_async_result` (a thread returns an `Async<int>`, E1190), `threads_closure_capture` (captures a function value, E1190), `threads_generic_unbound` (a generic helper sends `T` without `T: Sendable`, E1190), `threads_scope_stored`, `threads_scope_returned`, `threads_scope_captured` (E1191), `threads_async_function` (E1192), `threads_function_value` (a local holding a closure passed to `Thread.start`, E1193), `threads_implement_sendable` (E1194), `threads_join_equal` (E1195).
- [ ] Samples with checked output, clean under AddressSanitizer and the leak check: `threads_check_generic` (`<T: Sendable>` helper over `Thread.start`), `threads_check_struct` (a struct of arrays, maps, options and a `Box` crosses), `threads_check_scope_helper` (a `Scope` passed to a function that starts threads), `parallel_map` (order kept, more items than cores, an empty array).
- [ ] SPEC: the `Sendable` row and section 6 and 8 notes lose "not built yet" for this node; `compiler/README.md` updated.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

M: one trait walker and five checks in a new module, one small std module.

## Notes for the worker

- `Thread.scope`'s callback is non-escaping, so it may capture `let` and `mutable` (D77); only its `Scope` parameter has the escape rule.
- Shadowing is an error in Narra; check the existing E0612 message for the `let` capture case instead of adding a code.
