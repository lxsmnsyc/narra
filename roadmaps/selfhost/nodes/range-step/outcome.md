# Outcome: range-step

## Built

- `Ranges.stepped(start, end, step)` now returns a `Range` (before: `Option<Range>`, `None` for step 0). End is exclusive; a negative step counts down; a run-time step of 0 gives the empty range; a loop stops before the value would pass the int range (existing `narra_range_advance`).
- Check: `checkRangeStepped` in `selfhost/src/stdrange.nra`. A step that folds to 0 at compile time (literal, negated literal, top-level const, via `findTypedConstant`) reports E1270.
- Emitter: `selfhost/src/emitexprs.nra` builds the range with `narra_range_stepped(low, high, step)`, which now returns the struct (runtime source `selfhost/runtime/runtime.h`; `runtimecore.nra` regenerated). A for loop over the call uses the existing range-value C loop, no generator.
- Evaluator: `Range.stepped` in `selfhost/src/consteval.nra` returns a span, empty for step 0, so `@const` bodies may call it.
- Diagnostics row E1270 in `selfhost/src/diagnostics.nra`.
- Tests: `testdata/range_step.nra` and `.out` (up, down, stride past the end, empty, run-time zero, both ends of int, through `Iterators.map`), `testdata/errors/range_step_zero.nra` and `.err` (literal and constant zero). Updated callers: `std_no_panic_numbers`, `const_functions_language`.
- Docs: SPEC section 3 counting loops and the D92 table row removed, `compiler/README.md`, guide chapter 2 (new example, run by the guide suite).

## Proof

- base: feat/batch-15 at d64eee66 (compiler from a clean `git archive`), `narrac run testdata/range_step.nra`: `expected Range, found Option<Range>` on each `Ranges.stepped` call.
- branch: same command prints the 18 lines in `range_step.out`; `narrac check testdata/errors/range_step_zero.nra` reports E1270 at the literal and at the constant.

## Deviations

- `testdata/ranges_stepped.nra` and `.out` removed: they used the `Option<Range>` form and `range_step` covers every case. `testdata/std_no_panic_numbers.*` and `testdata/const_functions_language.nra` edited for the new return type (outside the brief's owned paths, needed to compile).
- The brief says "Int.min": there is no such constant and `-9223372036854775808` does not fit as a literal, so the sample uses `-9223372036854775807 - 1`.
- The old behavior was already shipped as `Option<Range>`; D119 replaces it with no shim (maintainer rule).

## Stubs left

- None.

## Gotchas

- `runtimecore.nra` changed, so the committed seed's runtime text is older than the source; bootstrap reports "seed older than the source" until the batch seed update. `selfhost/src` itself does not call `Ranges.stepped`.
- E1270 uses `findTypedConstant`: literals, `-`, binary operators and top-level consts fold. Anything else gets the run-time rule (empty range).

## Proposed learnings

- None.

## Verification

- `emit selfhost` of the base source by the base and the branch compiler: byte-identical except the `narra_range_stepped` runtime function.
- Targeted suites and the full `sh scripts/test.sh`: see the final line recorded in the report.

## Size

- Estimated S, actual S.
