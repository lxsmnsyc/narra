# Outcome: thread-scope-void

## Built

- `Thread.scope` in `std/src/thread.nra` keeps the body's result in a `Box<Option<T>>`, as `Thread.start` does, so a body that returns nothing works. `selfhost/src/std/stdtext.nra` regenerated.
- Case `testdata/threads_scope_void.nra` (+ `.out`). SPEC section 8 says a scope gives nothing when its body returns nothing; chapter 15 drops the `return 0` from three scopes whose result was unused.

## Proof

- base: 77674c71 plus batch-21, `narrac run main.nra` with the brief's program: `std/src/thread.nra:42:3: this expression has no value`; `narrac run testdata/threads_scope_void.nra` gives the same error.
- branch: the same commands print `done`, and `total 3` then `done`.

## Deviations

- None. The general void rule was not built.

## Stubs left

- None.

## Gotchas

- Cause: the checker rejects a void local (E0225), and in an instance with a void type argument the error is reported at the std line. Any std generic that holds a `T` in a local and is instantiated with void hits it. `Thread.start` and `Thread.join` already go through a `Box<Option<T>>`. The `Async` functions do not return a body's result (`then`, `all`, `race` take handles; `Async<void>` has its own case). `Parallel.map` builds an `Array<U>`, and an `Array<void>` has no meaning. No other std function needed a change.
- A general fix needs void locals, fields and pattern bindings in the checker and emitter (the 2026-11-14 inbox note); then `isVoidInstance`, `isVoidParameterValue` and `bindThreadVoid` could go. It touches checker, emitter and seed, so it was not done here.
- A user program that hits E0225 inside another std generic instantiated with void would still show a std location.

## Proposed learnings

- A std generic that holds a value of a type parameter in a local stores it in a `Box<Option<T>>` so `T = void` works | the checker rejects void locals | std/src | no

## Verification

- Suites `spec` (65), `guide` (610) and `--filter threads` (189) pass with the built compiler. Full `sh scripts/test.sh`: PASS, 10787 cases.

## Size

- Estimated S, actual S.
