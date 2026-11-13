---
id: async-core
title: "Async functions, blocks and await"
status: active
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
  - selfhost/runtime/std_async.c
  - selfhost/runtime/std_async_loop.c
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/runtimestdasyncloop.nra
  - roadmaps/selfhost/nodes/async-core/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
  - selfhost/src/funcvalues.nra
  - selfhost/src/closures.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/check.nra
  - selfhost/src/types.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/resolve.nra
  - selfhost/src/generics.nra
  - selfhost/src/cnames.nra
  - selfhost/src/traits.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/calls.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitter.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitinline.nra
  - selfhost/src/inlining.nra
  - selfhost/src/diagnostics.nra
  - compiler/README.md
reserves: [diag E1100-E1129]
provides: []
depends_on:
  - on: async-design
    kind: hard
    why: the design gate (D109) fixes the runtime shape, the error codes and the SPEC text this node builds
---
# Async functions, blocks and await

## Goal

Build the core of D93 as D109 says: `async function`, `async` arrows, `async { ... }` blocks and `await`; the builtin type `Async<T>` as an identity handle; the unused-handle error; effects charged at creation; an `async` entry function whose `Async<int>` gives the exit code. The loop in this node is a ready queue with no I/O and no timers; event-loop adds them.

## Context

- D109 items 1 to 6, 9, 10 and 11 are this node's design; SPEC section 8 "Async functions" and section 11 are the text it builds to. Read D109 first; it overrides this brief where they differ.
- **Proof first:** `async function f(): int { return 1 }` with `const x = await f()` in `export async function run(_args: Array<string>): int` fails to compile on the batch base.
- **Syntax.** `async` and `await` become lexer keywords (no source uses them as names). The parser rejects a nested `async function` with E0930 (`parser.nra` around line 2675); that stays. New: top-level `async function` (and `export async function`), `async` arrows (`async (x) => ...`, `async x => ...` if plain arrows allow it), `async { ... }` in value position, and `await expr`. New AST forms need cases in the checker, the emitter, `fmt` and `astdump` switches, and in `idcheck`.
- **The type.** `Type.TAsync(value: Type)` beside `Type.TGenerator`; `Async` resolves as a builtin name in `resolve.nra` (around line 584, as `Generator` does), hidden by a user type of that name. The files that name `Type.TGenerator` are the list to visit (about 20). `cnames.nra` gives it the C type `narra_task*`. `traits.nra` (around line 1159, the `Cell` case) gives it no automatic Equatable or Hashable.
- **Emitter.** `emitasync.nra` holds the frame, resume, drop and creation functions and `await`. Share with `emitgenerators.nra` what D109 item 1 lists (`recordLive`, `emitStateSwitch`, the frame-field path of `declareVariable`), exported from there; spill operands evaluated before a nested `await` into frame temporaries and lower `&&`, `||` and conditional expressions holding an `await` into branches (D109 item 1); rename `inGenerator` to `inFrame` in `emitstate.nra`, `emitstmts.nra`, `emitmatch.nra` and `emitinline.nra` (which clears it when inlining arrows, around lines 282-320). An `async` arrow is a function value whose call makes a frame from the closure's captures (`emitclosures.nra`).
- **Entry.** `findEntry` (`emitter.nra` around line 162) accepts an async entry; `emitMain` (`emit.nra` around line 605) calls the runtime: set up the loop, call the entry, run until its handle is done, read the code, then the exit sequence of D109 item 6 (cancel, drain, drop), then the existing release and leak report.
- **Runtime.** `selfhost/runtime/std_async.c` (task header `narra_task`, `narra_task_wait`, `narra_task_finish`, release) and `selfhost/runtime/std_async_loop.c` (a `_Thread_local` ready queue, run-until-done, the exit sequence, the deadlock stop with exit code 101). Add one case to `runtimetext.nra`: `std/internal/async` returns both texts. Run `sh scripts/test.sh --suite generated --update` to write `runtimestdasync.nra` and `runtimestdasyncloop.nra`. The checker marks `std/internal/async` used when it checks an async body, an `await` or a type holding `Async<T>`; the emitter notes the helper (`noteHelper`), so programs with no async emit the same C.
- **std/async.** `std/src/async.nra` (source, embedded in `stdtext.nra`) with `AsyncError` (`Cancelled`, `Failed(message: string)`) and `detach`; the rest of its functions are async-api's and event-loop's. Its primitives are the builtin module `std/internal/async`, whose rows go in `stdasync.nra` and are registered in `stdlib.nra`.
- **Checks.** The unused-handle rule (E1103) sits beside the unchecked-`Result` rule E0229 in `stmts.nra` (around line 570). The capture rule (E1110) sits beside E0612 in `funcvalues.nra` (around line 132). The shared-write rule of D109 item 9 extends `sharedwrites.nra`: a local holding a handle holds every argument and capture of its task.
- Seed rule: no async in `selfhost/src` or `std/src`; `std/src/async.nra` may not use `async` syntax itself until the seed is updated at the end of the batch (its functions take and return handles through primitives).

## Codes (E1100-E1129)

E1100 `await` outside an async body (the message names the place: outside, in a `function*`, in a non-async arrow); E1101 left free (dropped with the statement-position limit; not reused); E1102 `await` on a value that is not an `Async<T>`; E1103 an unused `Async<T>`; E1104 a `mutable` parameter on an async function or arrow; E1105 `for (mutable ...)` in an async body; E1106 `==`, `!=` or a map key on an `Async<T>`; E1107 `async function*`; E1108 `async` not followed by `function`, an arrow function or `{`; E1109 the wrong number of type arguments for `Async`; E1110 an `async` arrow or block that captures a `let` or a `mutable` parameter. E1111-E1129 stay free for this node's follow-ups.

## Acceptance

- [ ] Samples under `testdata/async_*`, output checked, clean under AddressSanitizer and the leak check: async functions calling each other with `await`; an `async` arrow, and `Arrays.map(items, async (item) => ...)` giving `Array<Async<T>>` awaited in a loop; an `async { ... }` block with `return`; a handle awaited twice; `Async.detach` of a task that finishes before the entry; an async entry returning a nonzero exit code; a plain entry that starts and detaches a task; a task left waiting at exit (cancelled and dropped, no leak); a task that awaits a handle already finished (no suspension); `await` nested in expressions (`g(await f(), await h())`, `if (await ready())`, `await a + await b`, `a && await b` where `a` is false, an `await` in a template and in a block expression), with output that shows left-to-right order; a task dropped at exit while suspended inside an expression with a spilled heap operand (no leak).
- [ ] Error tests under `testdata/errors/async_*` for each of E1100 and E1102 to E1110, and one for a shared write (E0978) to an array passed to a task whose handle is still held.
- [ ] A program with a cycle of awaits stops with exit code 101 and the D109 message (a `panics` or `exits` case).
- [ ] Code with no async emits byte-identical C (`narrac emit selfhost` unchanged), and the generator and coroutine suites pass.
- [ ] fmt and astdump cover the new syntax; the `sources` suite passes.
- [ ] `compiler/README.md`: "Not supported yet" drops async functions and `async` arrows; the D85 line names E1105.
