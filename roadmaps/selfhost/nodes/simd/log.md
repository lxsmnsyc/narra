# Log: simd

## Proof first

- Base: `feat/batch-13` at 2c58e901, compiler from `sh selfhost/bootstrap.sh /private/tmp/claude-501/simd/base`.
- `base/narrac run testdata/simd/operators.nra`: exit 1, first line `operators.nra:2:1: unknown module "std/simd" (...)`, then `unknown type "F32x4"` and the follow-on errors.

## Work

- `Type.TVector(lane, count)` in `types.nra` with `findLaneSize`, `formatVectorName` and `findIndexedLength` (a sibling of `findFixedLength`, so array patterns keep treating only FixedArrays as fixed). `findFixedLength` is unchanged.
- Logic in `simdtypes.nra` (table, name test, operator rules), `stdsimd.nra` (four calls), `emitsimd.nra` (typedefs, literals, lane reads, operators, calls). Shared files carry rows.
- A std/simd call takes a call type argument as its expected type: one row in `checkBuiltinCall` (`exprs.nra`), since the std checks get no node id.
- `isVectorName` (length 4 to 6, first byte F, I or U, contains `x`) gates the vector lookup in `resolveType`, so other type names pay one length test.
- The typedefs come from `checker.composites`, after the comment table and before the struct typedefs.
- `first<T>(items: Array<T>): Option<T>` needs `from items` (E0976-style shared result rule), so the sample writes it.
- clang `-std=c11 -Wall` on the C of `operators.nra`: 126 unused-function and 1 unused-variable warnings (runtime text), no vector warning.

## Review fixes

- `!v` passed the checker and failed in cc; it now takes the `!` path of numbers (expected bool). Test `simd_not_operator`.
- An import used only through vector names was E0946 (proof: `narrac run testdata/simd/types_only.nra` with the 872b6416 compiler: `"Simd" is imported but never used`). `resolveVectorType` marks the alias referenced.
- SPEC: templates reject vectors, `Show.show` prints them; the E1251 text matches the message. D112 Built note with the Shape ruling; test `simd_shape`.
