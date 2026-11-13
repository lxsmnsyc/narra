# Outcome: simd

## Built

- D112 items 1 to 4 without the library functions: the twenty vector types, literals, operators, lane reads, `==`, hashing, `Show`, and `Simd.splat`, `Simd.fromFixed`, `Simd.toFixed`, `Simd.withLane`.
- Type: `Type.TVector(lane: Type, count: int)` in `selfhost/src/types.nra`, key `V(lane,count)`. Also there: `findLaneSize(lane)`, `formatVectorName(lane, count)` (`F32x4`) and `findIndexedLength(type)`, which gives a FixedArray's length or a vector's lane count. `findFixedLength` is unchanged, so array patterns still take only FixedArrays.
- `selfhost/src/simdtypes.nra`:
  - `listVectorTypes`, `findVectorType(name)`, `findVectorFor(lane, count)`.
  - `isVectorName`: a cheap test that gates the lookup in `resolveType`.
  - `importsSimd`, `resolveVectorType`: a name is a type only in a module that imports `std/simd`, unless the module declares a type of that name. Type arguments on a vector name are E0117. A resolved vector name marks the `std/simd` alias as referenced, so an import used only through type names, literals and operators is not E0946.
  - `isVector`, `isFloatVector`, `isIntegerVector`, `isUnsignedVector`.
  - `findUnsignedTwin(v)`: the unsigned vector of the same shape.
  - `formatVectorId` (`f32x4`; the C type is `narra_f32x4`).
  - Checks: `checkVectorOperator`, `checkVectorUnary`, `reportLaneCount` (E1253), `reportLaneWrite` (E1252).
- `selfhost/src/stdsimd.nra`: `SIMD_FUNCS` (module `std/simd`, registered last in `stdRegistry`) and `checkSimdCall`.
  - A call type argument (`Simd.splat<F32x4>(1.0)`) becomes the expected type, through a row in `checkBuiltinCall` (`exprs.nra`) for every check kind that starts with `checkSimd`. simd-library's `Simd.load<V>` gets this for free if its kind starts with `checkSimd`.
  - E1256 (a spare code of this node): a `std/simd` function argument that is not a vector.
- `selfhost/src/emitsimd.nra`:
  - `emitVectorTypedefs`: one typedef per vector in `checker.composites`, plus the unsigned twin of each signed integer vector. Written after the type comment table. Empty for a program without vectors.
  - `emitVectorLiteral` (a compound literal), `emitLaneRead` (`(v)[i]`).
  - `emitVectorArithmetic`, `emitVectorUnary`: signed `+ - *` and unary `-` cast to the twin and back.
  - `buildLaneLoop`, used by the eq, hash and show helpers in `emittraits.nra`.
  - `emitSimdCall`: splat is a compound literal of one temp; `fromFixed` and `toFixed` are a `memcpy` between `.items` and the vector; `withLane` copies the vector, then assigns `t[i]`.
- Rules wired through shared switches:
  - Layout: size `lane * count`, align 8 (`layoutValue`); enum payload size `measureValue`.
  - Not counted and not a contents type, so a `mutable` parameter and `copy` give the existing errors.
  - Traits: Equatable and Show for every vector, Hashable for integer vectors, never Ordered.
  - Map keys: a float vector gets "float keys are not allowed", from `mapKeyProblem` going through the lane.
  - `Ptr<V>` is allowed (`findPointeeProblem`).
  - Lane index proof: `checkFixedIndex` through `findIndexedLength`. For a vector, its hint leaves out "or read it with FixedArrays.get".
- Tests:
  - Samples, checked by the programs suite in debug (ASan, leak check, copy check) and release: `testdata/simd/operators.nra`, `testdata/simd/lanes.nra`, `testdata/simd/containers.nra` (with their `.out`). containers also covers a `Ptr<I16x8>`.
  - `testdata/simd/types_only.nra`: the import is used only by a type name, a literal and `+`.
  - Errors: `testdata/errors/simd_*.nra`/`.err`: not_operator (`!v`), shape (derives Shape), operator_undefined (E1250), scalar_operand (E1251), lane_write (E1252), literal_lanes (E1253), splat_type (E1254), from_fixed (E1255, E1256), lane_index (E0994), negate_unsigned (E0296), float_map_key (E0900), not_contents (`mutable` and `copy`), unimported.
  - Query: `testdata/query/simd_vectors/` (`interface`, `type-of`).
- `!` on a vector is the "expected bool, found F32x4" error, as `!n` is on a number. A struct with a vector field does not derive `Shape` ("F32x4 has no Value form"), as with FixedArray (D84, orchestrator ruling).
- Docs: SPEC "Vector types" names what is built and what is left for simd-library; the other "not built yet" marks of the built parts are gone. `compiler/README.md` has a feature bullet.

## Proof

- base: feat/batch-13 2c58e901, `narrac run testdata/simd/operators.nra` exits 1 with `operators.nra:2:1: unknown module "std/simd" (...)`, then `unknown type "F32x4"`.
- branch: the three samples match their `.out` in debug (`-sanitize`, leak check, copy check) and release: `--suite programs --filter simd` PASS, 6 cases.

## Deviations

- Files outside the owned and shared paths, each a dispatch row:
  - `selfhost/src/resolve.nra`: the vector name lookup in `resolveType`.
  - `selfhost/src/stmts.nra`: compound assignment with a vector operand goes to `checkVectorOperator`.
  - `selfhost/src/emit.nra`: one line writes the typedefs.
  - `selfhost/src/emittraits.nra`: eq, hash and show branches for vectors.
  - `selfhost/tests/src/programs.nra`: the programs suite also runs `testdata/simd/*.nra`. Without it the owned folder `testdata/simd/` would not be tested.
  - `testdata/query/simd_vectors/`: the query case the acceptance asks for.
  - Goldens `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump`: the module list now ends with `std/simd`.
- `DECISIONS.md`: a "Built" note under D112, with the Shape ruling.
- E1256 is used (a spare code of this node) for a non-vector argument of `Simd.toFixed` and `Simd.withLane`.
- Different vector types on one operator (`F32x4 == F64x4`) are E1251, without the splat hint.
- The sample's generic function is `first<T>(items: Array<T>): Option<T> from items`: the shared-result rule needs `from items`.
- The E1251 message has the hint `Simd.splat(...)`, not the scalar's own text, because the check sees only the types.

## Stubs left

- None. The comparison, select, any/all, load/store, shuffle, sum, min/max, abs and shift functions are simd-library's.

## Gotchas

- Name every new check kind `checkSimd...`, so `stdcalls.nra` sends it to `checkSimdCall` and `checkBuiltinCall` turns a call type argument into its expected type.
- With an expected vector type, a float literal on the other side of an operator stays `f64`. So `v * 2.0` on an `F32x4` reports `F32x4 and f64`.
- Vectors are not in the emitter's name cache (`isCachedType`): `cType` builds the name each time. The typedefs come from `checker.composites`, not from the `used` map, so every vector type the checker interned gets a typedef, even one only unreached code names.
- Vector `==` lowers to an `eq_<id>` helper that compares lane by lane with C `==`, so a NaN lane makes a vector unequal to itself.
- 32-byte vectors are passed by value, like any value of 16 bytes or less. gcc on x86-64 without AVX may print `-Wpsabi` notes; the driver builds with `-w`. The Linux check ran on arm64 only; x86-64 gcc was not run.

## Proposed learnings

- A new `Type` variant that names a module type goes through `isVectorName`-style cheap gates in `resolveType` | every type resolution pays for the lookup otherwise | selfhost/src | no

## Verification

- `narrac emit selfhost` on the branch source, from the feat/batch-13 (ef53457c) compiler and from the branch compiler: byte-identical C (`cmp`, 231498 lines). Code without vectors emits the same C.
- `sh scripts/test.sh` after the rebase on ef53457c: `PASS, 9010 cases in 590s`.
- A second rebase, on a227553b (for-await), was clean. Then `sh scripts/test.sh`: `PASS, 9191 cases in 372s`. The emit comparison and the Linux check ran on the ef53457c rebase.
- `sh scripts/linux-check.sh` (colima, arm64, gcc): `test.sh: PASS, 9010 cases in 451s`, `linux-check: PASS in 1234s`.
- Suites with the branch compiler, before the rebase: programs (`--filter simd`) 6, errors 799, query 10, dumps 174, codes 1, sources 1, spec 63, generated 20, all PASS.
- clang `-std=c11 -Wall` on the C of `operators.nra`: no vector warnings (126 unused-function warnings and 1 unused-variable warning, all from the runtime text).

## Size

- Estimated M, actual M (about 1400 lines with tests).
