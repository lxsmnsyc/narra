---
id: async-generators
title: "Async generators: async function*, AsyncGenerator<T> and Async.next"
status: todo
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
  - roadmaps/selfhost/nodes/async-generators/**
  - selfhost/src/asyncgenerators.nra
  - selfhost/src/emitasyncgenerators.nra
  - testdata/async_generators_*
  - testdata/errors/async_generators_*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/stmts.nra
  - selfhost/src/effects.nra
  - selfhost/src/contracts.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/emitasync.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/stdasync.nra
  - selfhost/src/*.nra
  - selfhost/runtime/std_async.c
  - selfhost/runtime/std_async_loop.c
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/runtimestdasyncloop.nra
  - std/src/async.nra
  - selfhost/src/stdtext.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1112-E1121]
provides: []
depends_on:
  - on: async-generators-design
    kind: hard
    why: the design gate (D110) fixes the type, the step protocol, the codes and the SPEC text this node builds
---
# Async generators: async function*, AsyncGenerator<T> and Async.next

## Goal

Build D110 items 1, 3, 4, 5 and 6 without `for await`: `async function*` declarations, the builtin type `AsyncGenerator<T>`, `yield`, `await` and `yield*` in one body, and `Async.next(stream): Async<Option<T>>` in `std/async`. A program can read a stream by hand with `await Async.next(stream)` in a `while` loop. `for await` and `Net.chunks` are node for-await.

## Context

- Read D110 first, then D109 items 1, 2, 8 and 9 (the task frame, the handle, the cancel walk, shared writes) and SPEC section 8 "Async generators". They override this brief where they differ.
- Proof first: `async function* f(): AsyncGenerator<int> { yield 1 }` is E1107 on the batch base (`parser.nra`, `parseAsyncDecl`).
- Type: `Type.TAsyncGenerator(element, effects, binder, origin)` beside `Type.TGenerator` (58 matches in 22 files). Builtin name in `resolve.nra` next to `Generator` and `Async`; effects on the type as D78 does for generators (`makeGenerator`, `effects.nra` around `createsGenerator`). C type `narra_task*`, plain reference counts, no copy-check hooks, identity type (E1106 for `==` and map keys, as `Async`).
- Checker: `FuncDecl` keeps `generator: true` and `isAsync: true`; both `checker.generator` and `checker.asyncBody` are set in the body. `yield` checks in `stmts.nra` (`checkYield`): element against `AsyncGenerator<T>`, E1115 for a binding, E1116 for a `yield*` operand, E0292 for `return value`. `for (mutable ...)` in the body reports E1105 only, not also E0957.
- Emitter: the frame is a task frame (`emitTaskParts` in `emitasync.nra`) plus a FIFO of pending steps and a step creation function that allocates `{ narra_task h; Option<T> value; }`. `yield` stores `Some(value)` in the first pending step, calls `narra_task_finish` on it, records the state (`recordLive`) and returns; the end of the body finishes every pending step with `None`. After a `yield` with steps still pending the stream queues itself (`narra_loop_queue`). `yield*` over an `AsyncGenerator<T>` loops over `Async.next` of the inner stream and yields each value.
- Runtime (`std_async.c`, `std_async_loop.c`): the stream is out of the loop's live list while it is parked at a `yield` with no pending step, so dropping its last copy frees it through its release function. A pending step forwards a cancel to the stream in `narra_loop_cancel_wait`, which cancels the stream when that step is its only pending one (D110 item 4). Each pending step holds a reference to the stream.
- `Async.next` is Narra source in `std/src/async.nra` over a primitive of `std/internal/async` (rows in `stdasync.nra`): `export function next<T, effect E>(stream: AsyncGenerator<T> uses E): Async<Option<T>> uses E, alloc`, as `Co.resume` takes the effects of `Generator<T> uses E` in `std/src/coroutine.nra`. Then `sh scripts/test.sh --suite generated --update`.
- Parser: remove the E1107 report; `async function*` sets both flags. E1107's row and explanation become "retired: ...", as E0022's.
- batch-14 edits std sources and `stdtext.nra` on its own branch; the batch that merges second regenerates.

## Codes (E1112-E1121)

- E1112 the result of `async function*` is not written `AsyncGenerator<T>` (a `Generator<T>` gets the hint `AsyncGenerator<T>`).
- E1113 `AsyncGenerator<T>` as the declared result of a `function*` that is not async; hint `async function*`.
- E1114 `AsyncGenerator` with a number of type arguments other than one.
- E1115 `const x = yield value` in an async generator.
- E1116 a `yield*` operand that is not an `AsyncGenerator<T>` of the same `T`; for a `Generator<T>` the hint is `for (const x of g) { yield x }`.
- E1117-E1121 stay free for this node. Reused: E1104, E1105, E1106, E1111, E0292, E0291, E1100, E1103, E0978.

## Steps

- [ ] Parser, AST flag use, fmt and astdump for `async function*` (fmt: `async function* name`), with goldens.
- [ ] `Type.TAsyncGenerator` through the type switches, type key, C name, typesdump, declsdump.
- [ ] Checker rules and codes above; effects; `ensures` per yielded value; shared writes as a task handle.
- [ ] Emitter and runtime: frame, steps, yield, end, yield*, release, cancel forwarding.
- [ ] `Async.next` in `std/async`; generated suite updated.
- [ ] SPEC section 8 "Async generators": drop "not built yet" for the parts built here (keep it on `for await` and `Net.chunks`); `compiler/README.md` lists the feature.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check:
  - `async_generators_basic`: a stream that awaits `sleep` between yields, read with `await Async.next` until `None`; the values come in order.
  - `async_generators_lazy`: a stream nobody reads prints nothing; `Async.next` runs it to the first `yield`.
  - `async_generators_inline`: a stream whose values need no wait gives them without a loop turn (a `then` callback registered before shows it runs after).
  - `async_generators_shared`: two copies of one stream share the position; two steps asked at once are served in order.
  - `async_generators_delegate`: `yield*` over another async generator, and `yield await f()`.
  - `async_generators_drop`: dropping the last copy of a stream parked at a `yield` frees what it holds (a counted handle in the body is released, the leak check passes).
  - `async_generators_cancel`: `Async.cancel` on a step ends the stream's `sleep` with `Err(AsyncError.Cancelled)`; later steps give `None`; a step that is not the only pending one does not cancel the stream.
  - `async_generators_timeout`: `Async.race` of a step and a `sleep` task; the loser keeps running and nothing leaks at exit.
  - `async_generators_exit`: a stream waiting at exit is cancelled and freed.
- [ ] Error tests in `testdata/errors/async_generators_*` for E1112, E1113, E1114, E1115, E1116, and for E1104, E1105 (alone, no E0957), E1106, E1111 and E0292 in an async generator.
- [ ] `async function*` is no longer E1107; the row and explanation say retired.
- [ ] fmt, astdump, the language service and `narrac query` cover `async function*`; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

L: about 900 lines. Rubric scores above low: files touched (a type variant in about 22 files), runtime C with the cancel walk, new emitter path.

## Notes for the worker

- Seed rule: nothing in `selfhost/src` or seed-compiled std uses async generators. `std/src/async.nra` is not compiled by the seed (see its header); check before adding `next` that this still holds.
- Do not clone a stream frame: copies share it. `narra_gen_next` and `takeOwnCopy` must not be used for it.
- A step frame is created by the stream (it knows `Option<T>`'s C type); the `std/internal/async` primitive calls it through a function pointer in the stream's frame.
- The parser, `ast.nra`, `astdump.nra`, `fmt.nra` and `diagnostics.nra` are shared with other batch-13 builds: keep feature logic in `asyncgenerators.nra` and `emitasyncgenerators.nra`, dispatch rows only in shared files; the second to land rebases and keeps both sides.
