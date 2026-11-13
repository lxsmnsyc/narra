---
id: async-core
title: "Async functions, blocks and await"
status: todo
priority: should
type: feature
story: batch-12
early: false
outline: false
milestone: 
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - testdata/async_*
  - testdata/errors/async_*
  - selfhost/src/async*.nra
  - selfhost/src/emitasync*.nra
  - selfhost/src/stdasync.nra
  - std/src/async.nra
  - selfhost/runtime/async.c
  - selfhost/src/runtimeasync.nra
  - roadmaps/selfhost/nodes/async-core/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/runtimecore.nra
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/check.nra
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/traits.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/calls.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitter.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitinline.nra
  - selfhost/src/inlining.nra
  - selfhost/src/diagnostics.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: async-design
    kind: hard
    why: the design gate (D109) fixes the runtime shape, the error codes and the SPEC text this node builds
---
# Async functions, blocks and await

## Goal

Build the core of D93 as D109 says: `async function`, `async` arrows, `async { ... }` blocks and `await`; the `Async<T>` type as an identity handle; the unused-handle error; effects charged at creation; an `async` entry function whose `Async<int>` gives the exit code. The loop in this node is a ready queue with no I/O and no timers; event-loop adds them.

## Context

- Refined by async-design after the D109 gate. Read D109 and the async-design outcome first; they override this brief where they differ.
- Generated files: `std/src/*.nra` is embedded in `stdtext.nra`, and `selfhost/runtime/*.c` and `runtime.h` in `runtime*.nra` (headers say "Do not edit"). Edit the source (`std/src/async.nra`, `selfhost/runtime/async.c`), add its case in `runtimetext.nra`, then run `sh scripts/test.sh --suite generated --update`.
- Places to hook: `Cell` is special-cased by module name in `sharedwrites.nra` (around line 1055) and `traits.nra` (around line 1159); generators are the builtin `Type.TGenerator` (`check.nra`, `resolve.nra`). The entry call is `emitMain` in `emit.nra` (around line 605) and `findEntry` in `emitter.nra` (around line 162). `emitinline.nra` (around lines 282-290) clears `inGenerator` when inlining arrows.
- `async` is not a lexer keyword and no `async function` parses today: the parser rejects a nested one with E0930 (`parser.nra` around line 2675). Top-level `async function`, `async` arrows and `async { ... }` are new.
- Proof first: `async function f(): int { return 1 }` with `await f()` in an async entry fails to compile on the batch base.

## Acceptance

- [ ] Samples under `testdata/async_*`: async functions calling each other with `await`, an `async` arrow, an `async { ... }` block, an awaited handle awaited twice, an async entry function returning an exit code. Output checked, clean under AddressSanitizer and the leak check.
- [ ] Error tests under `testdata/errors/async_*` for each code D109 gives this node.
- [ ] The D85 rule: `for (mutable item of items)` in an `async` body is an error (see `compiler/README.md`).
- [ ] Code with no async emits byte-identical C (`narrac emit selfhost` unchanged), and the generator and coroutine suites pass.
- [ ] fmt and astdump cover the new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
