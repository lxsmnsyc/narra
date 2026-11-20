# Outcome: nested-void-match

## Built

- `emitMatchValue` in `selfhost/src/emit/emitmatch.nra` skips the result temporary when the match type is void and emits the arms as statements.
- Case `testdata/nested_void_match.nra` and `.out`.

## Proof

- base: 5c2b15ab, `narrac run testdata/nested_void_match.nra`: `error: variable has incomplete type 'void'` (3 errors), C compiler failed.
- branch: same command prints the 11 expected lines; the `programs` suite runs it under the sanitizer and leak check.

## Deviations

- None. `? :`, `if` expressions and blocks as arm bodies with void results are rejected by the checker, so they never reach the emitter.

## Stubs left

- None.

## Gotchas

- A void block arm inside a match in expression position is rejected by the checker ("a block expression needs `return value`"); not changed here.

## Proposed learnings

- A node that declares a result temporary must skip it for void | C has no void variables | selfhost/src/emit | no

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds.
- `sh scripts/test.sh`: full run before commit, all pass.

## Size

- Estimated S, actual S.
