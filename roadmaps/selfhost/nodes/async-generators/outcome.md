# Outcome: async-generators

## Built

- `async function*` declarations, the builtin type `AsyncGenerator<T> uses E`, `yield`, `await` and `yield*` in one body, and `Async.next(stream): Async<Option<T>>` in `std/async` (D110 items 1, 3, 4, 5 and 6). `for await` and `Net.chunks` are left to node for-await.
- Type: `Type.TAsyncGenerator(element, effects, binder, origin)` (`selfhost/src/types.nra`), key `W(...)`, C type `narra_task*`, type id `W_<element>`. It is a counted identity type with plain task reference counts and no copy-check hooks. It has a case beside `Type.TGenerator` or `Type.TAsync` in check, cnames, declsdump, effects, exprs, generics, calls, resolve, sharedwrites, sharedresults, stdcopy, traits, asyncvoid, contracts and emit. `generatorEffectList` (check.nra) and `generatorEffects` (effects.nra) take both generator types, so a private stream's effects are inferred as D78 does for a `function*`.
- Parser: `async function*` sets `isAsync` and `generator`. The E1107 report is gone, and its row and explanation say "retired". `AsyncGenerator<T>` takes a `uses` list as `Generator<T>` does.
- Checker, in `selfhost/src/asyncgenerators.nra`: `prepareStreamBody` makes the steps' `Option<T>` instance, `checkStreamYield` handles `yield` and `yield*`, and `checkStreamNext` checks the `Tasks.next` primitive. `resolve.nra` has `resolveStreamResult` (E1112) and `resolveStreamType` (E1114), and the E1113 check sits in `resolveGeneratorResult`. Parameters follow the async rules (E1104, E1111), not the generator ones. `for (mutable ...)` in a stream body is E1105 alone. `==` and map keys are E1106, with a stream message. A stream holds its arguments for E0978, as a task handle does. `ensures` is checked at each `yield`, and `yield*` gets the existing E0988.
- Effects: the call charges `alloc` only. `Async.next` charges the stream type's effects plus `alloc`, through `next<T, effect E>(stream: AsyncGenerator<T> uses E): Async<Option<T>> uses E, alloc` in `std/src/async.nra` over `Tasks.next` (`stdasync.nra`).
- Emitter, in `selfhost/src/emitasyncgenerators.nra`: `emitStreamFunction` reuses `emitTaskParts` (now exported, with `TaskNames`) under `emitter.inStream`. The frame gets a `narra_stream_part s` field, and the end of the body (and `return`) calls `narra_stream_end`. The step struct is `V...` (`narra_step s; Option<T> value`), with creation `n...` (it holds None) and release `w...`. `emitStreamYield` fills the first pending step with `Some(value)`, then calls `narra_stream_pause`. `yield*` loops on `narra_stream_next` of the inner stream, awaits each step and stops at the first step that was not filled. `emitStreamNext` handles `Tasks.next`. Dispatch rows: `emitconsts.nra` (`emitFunction`), `emitstmts.nra` (YieldStmt) and `stdasync.nra`. Emitter fields: `inStream` and `streamOption` (`emitstate.nra`); `inStream` is saved in `BodyContext`.
- Runtime, in `selfhost/runtime/std_async.c`: `narra_step`, `narra_stream_part`, `narra_stream_init` (rc 1, parked, out of the live list), `narra_stream_next`, `narra_stream_take`, `narra_stream_give`, `narra_step_filled`, `narra_step_drop`, `narra_stream_pause` (queue when steps are pending, else park and drop the loop's reference) and `narra_stream_end`. `narra_task` has a new field, `is_step`. A pending step's `awaited` is its stream. In `std_async_loop.c`, `narra_loop_cancel_wait` passes a cancel from a step to its stream when that step is the stream's only pending step, and through a step that one task awaits alone.
- `narrac query` / LSP hover: `writeFunction` now prints `async` (for `async function` too), and types print as `AsyncGenerator<T> uses E`.
- SPEC section 8 "Async generators" is no longer marked "Not built yet", except the `for await` and `Net.chunks` bullets. The section 2 table and the "Produce values one at a time" row are updated. `compiler/README.md` lists the feature, and lists `for await` under "Not supported yet".

## Proof

- base: 7df2cd63 (feat/batch-13). Sample `async function* f(): AsyncGenerator<int> { yield 1 }`; `base-narrac run main.nra` exits 1 with `main.nra:1:15: there is no async function*; ...` (E1107).
- branch: all nine `testdata/async_generators_*` samples give their checked output in debug (ASan, leak check, copy check) and release.

## Deviations

- New goldens outside the owned paths: `testdata/astdump/async_generators.{nra,ast}`, `testdata/fmt/decl_async_generators.{nra,fmt}` (the fmt suite only picks `decl_` names) and `testdata/query/async_generators/`. The brief asks for goldens, and these folders are where they live.
- Removed `testdata/errors/async_generator.{nra,err}`, the E1107 sample of async-core. `async_generators_result` covers its case (a `Generator<T>` result) with E1112.
- `querysignature.nra` `writeFunction` prints `async` for every async function, not only for streams. No query golden had an async function before.
- E1112 has two rows: one for any result and one with the `Generator<T>` hint. The brief's effects-on-declaration case cannot happen: the parser reads `uses` after `AsyncGenerator<T>` into the type.
- E1104 and E1111 reuse the async-function messages ("async function f cannot take ..."), so the codes suite needs no new rows.

## Stubs left

- None. `for await`, E1122-E1124 and `Net.chunks` belong to node for-await.

## Gotchas

- A stream frame is never cloned. Copies share it, and `narra_gen_next` and `takeOwnCopy` are not used.
- Invariant: the body runs only while a step is pending, so `narra_stream_take` at a `yield` always finds one. A new step holds None. The end of the body finishes steps without writing to them, and `narra_step_filled` tells a Some from that None (used by `yield*`, and for-await can use it the same way).
- for-await can build `for await (const x of s)` on the `yield*` loop in `emitStreamDelegate`: `narra_stream_next`, an await of the step, a `narra_step_filled` test, then reading `(*(Option<T>*)step->value)`. `break` releases the stream and the step, as `yield*` does.
- `findStreamOption` looks up the `Option<T>` instance that `prepareStreamBody` made. A new path that emits a stream body without going through `checkFunctionBody` must make it first.
- The query hover of a private stream's callee prints `() => AsyncGenerator<int> uses alloc uses alloc` (the type's effects, then the call's `alloc`). This is the existing rule for generator types, which `compiler/README.md` documents.
- Runtime text changed: `narra_task` gained a field and the cancel walk gained a branch, so every async program's C changes in the runtime part (not in its emitted body). Two seed updates are needed at the end of the batch (runtime edit).

## Proposed learnings

- A runtime object that a waiter list may point to is shared, never cloned | its address is in lists the loop walks | selfhost/runtime | no

## Verification

- `NARRAC=<branch release narrac> sh scripts/test.sh` (full, before the rebase): `PASS` apart from the retired E1107 sample, which was then removed. After the rebase on feat/batch-13 (0bab063f): `sh scripts/test.sh` (seed-built), `PASS, 8673 cases in 577s`.
- `narrac emit selfhost` from the base compiler (seed-built, -release) and from the branch compiler, both on the branch source: byte-identical (225575 lines).
- `sh scripts/test.sh --suite generated --update` after the `std/src/async.nra` and runtime edits.

## Size

- Estimated L (about 900 lines), actual L: about 1000 lines in `selfhost/` and `std/` (generated files excluded), plus tests.
