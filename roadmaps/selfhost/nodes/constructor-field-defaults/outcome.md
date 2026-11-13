# Outcome: constructor-field-defaults

## Built

- `isConstantValue` in `selfhost/src/exprs.nra` accepts `Expr.FieldExpr` whose node ref is a constructor, so a qualified constructor with no fields (`Kind.None`) is a constant. This covers field defaults (the E0106 site in `body.nra`) and every other caller of `isConstantValue`, parameter defaults of D115 included.
- A qualified constructor call with constant arguments (`Kind.Sized(4)`) already worked; the sample covers it. A non-constant argument (`Kind.Sized(count())`) still reports E0106.
- Sample `testdata/constructor_field_defaults.nra` (with `.out`) and error test `testdata/errors/constructor_field_default_call.nra` (with `.err`).

## Proof

- base: 0bab063f, `narrac types testdata/constructor_field_defaults.nra` reports E0106 at 8:21 for `kind: Kind = Kind.None`.
- branch: the same command reports no error, the program prints the expected two lines under AddressSanitizer, and the error test reports E0106 at 11:26.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- Only the bare qualified form was missing; a call node already carried its constructor ref.
- `--update` looked for `<name>.darwin.out` for a new sample, so the `.out` was written from a run.

## Proposed learnings

- None.

## Verification

- `NARRAC=<release compiler> sh scripts/test.sh --filter constructor_field`: PASS, 10 cases.
- `narrac emit selfhost` from base and branch compilers: identical.
- Full check: see the final report.

## Size

- Estimated S, actual S.
