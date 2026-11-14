# Outcome: coroutine-send

## Built

- `Co.send(mutable co, input)` runs a `Coroutine<In, Out>`; `Co.resume(mutable g)` runs a `Generator<T>`. The checker picks the function by name (`checkModuleCall` in `selfhost/src/calls.nra`: `resume` to `resumeGenerator`, `send` to `resumeCoroutine`), not by argument count. `std/src/coroutine.nra` keeps both functions; only its header comment changed.
- A two-argument `Co.resume` and a one-argument `Co.send` get the usual wrong-argument-count error. No migration error (no deprecation path).
- The effect label of the coroutine resume is `Co.send` (`describeFunctionLabel` in `effects.nra`). The E0293 texts for a non-mutable argument and for `for ... of` over a coroutine name `Co.send`.
- `SPEC.md` (section 8) and `compiler/README.md` describe both names. The guide uses only the one-argument form, so it did not change.
- Samples use `Co.send`: `testdata/coroutine.nra`, `coroutine_heap.nra`, `coroutine_tests.nra`, `params_pointer_escape.nra`, and the coroutine error tests. New: `testdata/errors/coroutine_resume_input.nra`.
- `stdtext.nra` regenerated with `sh scripts/test.sh --suite generated --update`.

## Proof

- base: d64eee66, `narrac run testdata/coroutine.nra` (with `Co.send`): `module "std/coroutine" has no export "send"`.
- branch: same command runs and prints the steps; `Co.resume(mutable g, 1)` is `this call passes 2 arguments, the function takes 1`.

## Deviations

- None. No new diagnostic codes; two E0293 texts changed.

## Stubs left

- None.

## Gotchas

- `import { send }` reports E0101 (no export); only `resume` has the by-name hint (E0197).
- The seed was not regenerated. Code the seed compiles must not use `Co.send`; after the batch seed update, any two-argument `Co.resume` in std or compiler source must become `Co.send`.

## Proposed learnings

- None.

## Verification

- `emit selfhost` of the base source with the base and the branch compiler: byte-identical (11,851,618 bytes).
- Suites errors (890 cases), generated, codes, sources, guide, programs, and the full check: see the report.

## Size

- Estimated S, actual S.
