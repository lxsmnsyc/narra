---
id: for-await
title: "for await over AsyncGenerator<T>, and Net.chunks"
status: done
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
spent: 230124
owns:
  - roadmaps/selfhost/nodes/for-await/**
  - selfhost/src/forawait.nra
  - selfhost/src/emitforawait.nra
  - testdata/for_await_*
  - testdata/errors/for_await_*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/stmts.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/contracts.nra
  - selfhost/src/*.nra
  - std/src/async__net.nra
  - selfhost/src/stdtext.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1122-E1129]
provides: []
depends_on:
  - on: async-generators-design
    kind: hard
    why: the design gate (D110) fixes the loop rules, the codes and the SPEC text
  - on: async-generators
    kind: hard
    why: the loop reads AsyncGenerator<T> through the step protocol and Async.next that node builds
---
# for await over AsyncGenerator<T>, and Net.chunks

## Goal

Build D110 items 2 and 7: the statement `for await (const x of stream) { ... }` over an `AsyncGenerator<T>` in any async body, and `Net.chunks(connection, maxBytes): AsyncGenerator<Result<bytes, AsyncError>>` in `std/async/net`.

## Context

- Read D110 and the async-generators outcome first; they override this brief where they differ. SPEC section 2 (Control flow) and section 8 (Async generators) have the text.
- Proof first: `for await (const x of s) {}` is a parse error on the batch base.
- AST: one field on `ForStmt` (for example `awaits: bool`), so the binding, destructuring heads, labels, `pathends`, effects and borrow logic of `for ... of` apply unchanged. That touches the 32 `Stmt.ForStmt(` patterns in 22 files; most get one more `_`. `await` between `for` and `(` can carry a comment on either side, so it needs its own trivia node (`selfhost/CLAUDE.md`).
- Checker (`forawait.nra`): E1122 outside an async body, naming the place as `reportAwaitPlace` in `asynccheck.nra` does; E1123 for a source that is not an `AsyncGenerator<T>`, with the hints of D110 item 2 for a `Generator<T>` and an `Array<Async<T>>`; E1124 for `for ... of` over an `AsyncGenerator<T>`. `for await (mutable x of s)` goes through the existing path and reports E1105. Charge the type's effects plus `alloc` per loop, as `Async.next` does. `seedLoopFacts` (`contracts.nra`) gives the binding the `ensures` facts of an async generator call, as for a generator.
- Emitter (`emitforawait.nra`): hold the stream in a frame field; each turn makes a step (the primitive `Async.next` uses), awaits it as `emitAwait` does (`emitasync.nra`), breaks on `None` and binds the value. The stream and the step are released on `break`, `return` and labeled exits through the usual unwind.
- `Net.chunks` is an `async function*` in `std/src/async__net.nra` over `Net.read`: `Ok(chunk)` per non-empty read, end at an empty read, one `Err` then end on an error. Then `sh scripts/test.sh --suite generated --update`. batch-14 edits std sources and `stdtext.nra` on its own branch; the batch that merges second regenerates.

## Codes (E1122-E1129)

- E1122 `for await` outside an async body (outside an async body, in a `function*` that is not async, in an arrow that is not async).
- E1123 a `for await` source that is not an `AsyncGenerator<T>`.
- E1124 `for ... of` over an `AsyncGenerator<T>`; hint `for await`.
- E1125-E1129 stay free. Reused: E1105.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check:
  - `for_await_basic`: `for await` over an async generator that sleeps between yields, in an async function, an async arrow, an `async { ... }` block and an async generator.
  - `for_await_break`: `break`, `return`, `continue` and a labeled `break` out of nested `for await` loops; the stream is freed at the break (a counted value held at its `yield` is released before the next line prints).
  - `for_await_destructure`: a destructuring head over a stream of tuples.
  - `for_await_cancel`: cancelling the task that runs a `for await` ends the stream's wait with `Err(AsyncError.Cancelled)`.
  - `for_await_chunks`: `Net.chunks` over a local connection (`Net.listen("127.0.0.1", 0)`, `Net.localPort`) gives every byte written, then ends; a test that starts a server kills it on every failure path.
- [ ] Error tests in `testdata/errors/for_await_*` for E1122 (each place), E1123 (a `Generator<T>`, an `Array<Async<T>>`, an `int`), E1124 and E1105.
- [ ] fmt (`for await (const x of s) {`), astdump, the language service and `narrac query` cover the new form; the `sources` suite passes.
- [ ] SPEC: no "not built yet" left in section 8 "Async generators", in the Control flow bullet and in the section 2 table row; the async generator sample becomes a checked `ts` block.
- [ ] `compiler/README.md` updated.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

M: one statement form on an existing loop node, one emitter path that reuses `emitAwait`, one std function.

## Notes for the worker

- Seed rule: nothing in `selfhost/src` or seed-compiled std uses `for await`.
- The stream is shared, not copied: do not use `takeOwnCopy`.
- Keep feature logic in `forawait.nra` and `emitforawait.nra`; shared files carry dispatch rows only.
