# Outcome: review-nits

## Built

- Item 1 (arrow hint): not reproduced, skipped. `writeArrowHint` writes `(items)` only for rows with a computed check and no listed parameters (`Arrays.length`, `Maps.size`), where it is right and tested. A fixed zero-parameter row has no obstacle and never reaches it.
- Item 2: `checkPreludeFunctionValue` in `selfhost/src/stdvalues.nra`, called from the name check in `exprs.nra`. `print`, `assert`, `panic`, `copy` as values report E1013 with an arrow hint.
- Item 3: arrow loop labels use `_a<tempCount>` (`emitinline.nra`); copies keep `_i<inlineCount>`.
- Item 4: `collectInlineEdges` (`inlining.nra`) reads calls by name from the function's scope when a generic function has no checked body. `CallUse` gets `calleeName`.
- Item 5: the load-error, no-binary and `-o` paths of `runBuild` return `exitWithoutRelease(1)`.
- Tests: `testdata/errors/value_prelude_{print,assert_argument,inline_uncalled}`, sample `testdata/inline_arrow_copy_labels`.

## Proof

- base: `narrac check` on `const show = print` gave `unknown name "print"`; an uncalled exported `@inline function countDown<T>` calling itself gave exit 0, no output.
- branch: E1013 `print cannot be used as a value: it is a prelude function; wrap it in an arrow function, such as `(value) => print(value)``; E1030 `@inline function countDown calls itself (countDown -> countDown)`.

## Deviations

- `testdata/inline_arrow_copy_labels.nra` and `.out` are outside the owned `value_prelude_*` glob; the brief's acceptance asks for the sample. The sample passes before and after; a real label clash needs one loop emitted as arrow and as copy with equal counters, which I did not construct.
- Item 1 skipped (see Built).
- The inlined-arrow sample was not given a new test file name under `value_prelude_*`; the E1030 test is named `value_prelude_inline_uncalled` to stay inside the owned glob.

## Stubs left

- None.

## Gotchas

- The uncalled-generic fallback resolves callee names in the function's scope and ignores local shadowing.
- Seed not regenerated; the new error tests need a compiler built from this source.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh out`: fixed point holds.
- `sh scripts/test.sh`: see the final report.

## Size

- Estimated S, actual S.
