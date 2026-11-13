# Outcome: threads-check

## Built

- **`Sendable`** (D111 item 1): a prelude trait with no functions (`trait Sendable {}` in `PRELUDE`, `selfhost/src/declare.nra`). `computeImplements` (`traits.nra`) answers it through `ThreadCheck.isSendable`, so `<T: Sendable>` bounds and `checkBounds` work as for any trait. The walker (`findUnsendable`, `selfhost/src/threadcheck.nra`) rejects, at any depth: function types, `Dynamic`, `Ptr`, `Async`, `AsyncGenerator`, `Generator`/`Coroutine`, `Handle`, `Cell` (std/cell), and the thread handles that are not Sendable (`Thread`, `Scope`). Arrays, boxes, fixed arrays, maps, distinct types, structs, enums, tuples and `Option`/`Result` are Sendable when their parts are. A type parameter is Sendable only with a `Sendable` bound.
- **Public helpers of `threadcheck.nra`** for the later thread nodes:
  - `checkSendable(mutable checker, pos, type, site)` reports E1190 as `<site>: <type> is not Sendable[, because of <part> in field "x" of <Type>][; hint]`. The hint names `Shared` for a Cell and `<T: Sendable>` for a type parameter. threads-channels and threads-shared call it for the type argument of `Channel.create`, `Shared.create` and `Atomic.create`.
  - `THREAD_HANDLES` is a table of `{ module, name, sendable }` rows. Add `Sender`, `Receiver`, `Shared` and `Atomic` there with `sendable: true`: they then get E1195 on `==`, `!=` and map keys, and Sendable is decided by their type arguments.
  - `THREAD_FUNCTIONS` is a table of `{ module, name, work }` rows, the std functions whose argument number `work` is a thread function.
  - `isThreadHandle`, `describeHandle`, `isSendableTrait`, `isThreadModule`.
- **Gate.** `Checker.threads` (`check.nra`) is set by `useModule` (`scopes.nra`) when a module imports `std/thread`, `std/parallel` or `std/internal/thread`. Every check except the Sendable trait itself runs only then, so programs without threads pay one bool test per expression.
- **Checks and their call sites:**
  - Thread calls: `calls.nra` `checkFunctionCall` (generic branch) calls `ThreadCheck.checkThreadCall`. It covers `Thread.start` and `Parallel.map`:
    - E1192 for an `async` arrow, or the name of an async function.
    - E1193 for anything other than an arrow at the call, the name of a top-level function, or `Module.function`. Names are resolved through `Ref.Function`, so an async function of another module is E1192.
    - E1190 for each capture (from `ClosureInfo.captures`), for the result type, and for the item type of `Parallel.map`.
    - Calls of `std/internal/thread` are not checked.
  - E1191 has these sites. Each value is reported once: an enclosing literal or call whose part was already reported adds nothing, and instances of generic bodies are skipped.
    - A built value that holds a Scope (array, struct, map or tuple literal, call or constructor, arrow function whose result holds one): `exprs.nra` `checkExpression` calls `checkScopeValue`.
    - A frame that keeps a Scope, from the review: a generator, coroutine, stream or task could run after `Thread.scope` returns. This covers a `function*` or `async function` with a parameter that holds a Scope (reported at the declaration, `checkScopeFunction` from `body.nra` `checkFunction`), a call that gives a frame and takes an argument that holds a Scope (generic generators included), and an `async` arrow with such a parameter.
    - A declared function result that holds a Scope: also `checkScopeFunction`.
    - A Scope captured by a closure that is not on the stack, or by any task (an async arrow or block): `funcvalues.nra` `useLocal` calls `checkScopeCapture`. A captured Scope is not reported again as E1190.
  - E1194 for `implement`, `implements` or `derives Sendable` (`resolveImpl`, `declareListed`, `checkListed` in `traits.nra`).
  - E1195 for `==` or `!=` on a handle (`checkEquatable`) and for a handle as a map key (`nominalKeyProblem` and `flushMapKeys`). `nominalImplements` gives handles no automatic Equatable or Hashable.
- **`std/parallel`** (`std/src/parallel.nra`), `Parallel.map(items, escaping change)`:
  - It runs on `Threads.cores()`, a new primitive: a row in `stdthread.nra` (`INTERNAL_THREAD_FUNCS`), and `narra_thread_cores` in `selfhost/runtime/std_thread.c`, which calls `sysconf(_SC_NPROCESSORS_ONLN)` and gives at least 1.
  - It never makes more parts than items. Each part runs `mapPart` on a thread of `std/internal/thread` and stores its array in a box. The parts are joined and their results put together in order.
  - The registry row is in `THREAD_FUNCS`.
- **Codes** E1190-E1195: rows and explanations in `diagnostics.nra`. E1196-E1199 are still free.
- **Tests:**
  - Error tests in `testdata/errors/`: `threads_cell_capture`, `threads_async_result`, `threads_closure_capture`, `threads_generic_unbound`, `threads_generic_bound`, `threads_scope_stored`, `threads_scope_returned`, `threads_scope_captured`, `threads_scope_generator`, `threads_scope_async`, `threads_async_function`, `threads_function_value`, `threads_implement_sendable` and `threads_join_equal`. Package test: `testdata/packages/errors/threads_async_module`.
  - Samples with `.out` files: `testdata/threads_check_generic`, `threads_check_struct`, `threads_check_scope_helper` and `parallel_map`.
- **Docs.** SPEC: the `Sendable` row and sections 6 and 8 no longer say "not built yet" for this node. `compiler/README.md`: two lines in the Threads section, and "Not supported yet" now lists only channels, `Shared` and `Atomic`.

## Proof

- base: afaaff2e (before the rebase on feat/batch-13 with simd). Built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/threads-check/base`. `narrac check testdata/errors/threads_cell_capture.nra`: exit 0. `narrac run` on it prints `1`.
- branch: the same command exits 1 with `threads_cell_capture.nra:9:40: capture "counter" of the thread function: Cell<int> is not Sendable; keep state that threads share in a Shared from std/shared`.

## Deviations

- **Files edited outside the owned and shared lists.** Each holds only a call or a row:
  - `selfhost/src/calls.nra`, `exprs.nra`, `funcvalues.nra` and `body.nra`: the check calls.
  - `check.nra`: the `Checker.threads` field.
  - `scopes.nra`: setting the field.
  - `declare.nra`: the prelude trait.
  - `selfhost/runtime/std_thread.c`, with its generated `runtimestdthread.nra`: `narra_thread_cores`.
- **Goldens changed.** All 27 `testdata/declsdump/*.dump` files gain the line `trait module="" Sendable`. `testdata/errors/unknown_module.err` now lists `std/parallel`.
- **Map keys.** The brief names `==` and `!=` for E1195. Map keys on a handle are E1195 too, through the existing map-key check.
- **The item type of `Parallel.map` is checked as well** (site "the items of Parallel.map"), because the items cross threads.
- **Extra diagnostics in two error goldens.** In `threads_async_function` and `threads_async_result` the existing "Async handle is not used" error (D109) is also reported next to E1192 and E1190.
- E1191 also covers frames (generators, coroutines, streams and tasks) that keep a Scope, from the fresh review. This is stricter than the brief: such a frame is an error even when it is used up inside the scope.
- `testdata/packages/errors/threads_async_module` (an async function of another module, E1192) is a package test, outside the owned `testdata/errors/threads_*`.

## Stubs left

- None. threads-channels and threads-shared add their handle rows to `THREAD_HANDLES` and call `checkSendable` for `Channel.create`, `Shared.create` and `Atomic.create`.

## Gotchas

- `TraitInfo.std` means "derivable" (`isDerivable`), so `Sendable` has `std: false` and `stdTrait(checker, "Sendable")` returns -1. Use `isSendableTrait` instead (module "" and the name).
- A parameter whose function type ends with `uses E` cannot take a trailing comma. The parser reads `, )` as one more effect name, and the std source then fails to parse at run time (`panic ... std source std/parallel: expected a name`). D117 (uses-list-syntax) fixes this; until then, leave the comma out.
- On macOS, `_SC_NPROCESSORS_ONLN` is hidden under `_POSIX_C_SOURCE`. `std_thread.c` defines it as 58 there, the Darwin value. Linux was not run.
- Captures are reported at the arrow's position, not at the use of the name, because `ClosureInfo.captures` keeps only local indexes.

## Proposed learnings

- Gate checks for one feature on a `Checker` flag set at import | keeps the checker's cost and the emitted C unchanged for programs that do not import the feature | selfhost/src | yes (emit comparison)

## Verification

- Samples (`/private/tmp/claude-501/threads-check/samples.sh`): each new sample was run with `-sanitize` and with `-release`, both with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`, and diffed against its `.out`. All 8 runs are ok.
- `NARRAC=<branch> sh scripts/test.sh --filter threads_`: PASS, 134 cases. `--filter parallel_`: PASS, 5.
- Emit comparison: `narrac emit selfhost` from the base compiler and from the branch compiler is byte-identical (`cmp`).
- Full check after the review fixes, rebased on feat/batch-13 9284717b: `sh scripts/test.sh` gives `PASS, 9389 cases`.
- Emit comparison, repeated after the rebase: `narrac emit selfhost` from the feat/batch-13 compiler and from the branch compiler is byte-identical.

## Size

- Estimated M, actual M.
