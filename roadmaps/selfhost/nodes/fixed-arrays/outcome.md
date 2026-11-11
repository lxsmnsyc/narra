# Outcome: fixed-arrays

## Built

- `Type.TFixedArray(element: Type, length: Type)` (`types.nra`). The length is a `TConst`, or a constant `TParam` in a generic body. Key `X(<element>,<length>)`, source spelling `FixedArray<T, N>`, C id `internedName("X", "FixedArray", [element, length])`. `findFixedLength(type)` in `types.nra` gives the length, or -1 while it is open.
- Resolution: `FixedArray<T, N>` in `resolveType` (`resolve.nra`, `resolveFixedLength`, `parseLengthText`). The length is an integer literal or a constant parameter in scope, never negative (E0995). Types are interned as composites; `finish` (`check.nra`) puts every closed FixedArray composite into `checker.order`, after its element type, so it gets a C definition and reference counting functions like a struct.
- C (`emitfixed.nra`): `struct X_FixedArray_<hash> { T items[N]; }` (one unused item for `N` of 0), stored, returned and passed like a struct; `isPassedByPointer` uses the layout (`N * element size`). Literal (`emitFixedLiteral`, a compound literal), read (`emitFixedRead`), write place (`buildFixedPointer`), `for ... of` over a copy (`buildFixedLengthText`, `buildFixedItemText` in `emitArrayFor`), retain/release loops (`wrapItemsLoop` in `typeRcFunction`), debug copy credits (`buildCreditCode`, `buildPartCreditCode`). No bounds check is written.
- Equality, hashing, `Ordered` and `Show` follow the element type (`traits.nra computeImplements`, `emittraits.nra`). Contents type for `mutable` (`isContentsType`), `copy(...)` (`isCopyTracked`), shared-write tracking (`sharedresults.nra`, `sharedwrites.nra`), counted and layout rules (`check.nra`), `unify`, `substitute`, `containsTypeParameter`, `containsOpen`, declsdump.
- A FixedArray literal charges no `alloc` (`effects.nra`).
- Proofs (`fixedarrays.nra`, called from `checkIndex` in `exprs.nra`): `checkFixedIndex` checks a constant index against `N` and any other index against the facts of a plain `int` local (`findOperandFact`, so a `let` captured by a closure has no facts inside it): it must be `>= 0` and `< N`. An open `N` defers the proof to each instance, where errors name the instance and its call. E0994.
- Fact sources added: the condition of a C-style for loop now narrows its body (`checkCFor` in `stmts.nra`); `seedCounterFact` gives a `let i = c` counter stepped by `i++` or `i += k` the fact `i >= c` when nothing in the condition or body assigns `i` and the condition bounds `i` with `<` (or with a constant bound that `+ k` cannot pass), so the step never wraps; `seedRangeFacts` gives `Ranges.upTo(n)` and `Ranges.between(a, b)` loop variables their bounds from what is known about `n`, `a`, `b`.
- `std/fixedarray` (Narra source, `std/src/fixedarray.nra`): `get<T, const N: int>(values, index): Option<T>` and `length(values): int`. Registry row in `stdlib.nra`; `stdtext.nra` regenerated.
- Codes: E0994 (index out of range or unproven), E0995 (length negative, not a literal or constant parameter, wrong argument count, literal length mismatch, literal with an open length). Rows and explanations in `diagnostics.nra`.
- Docs: SPEC.md section 2 (fixed-size array bullets and a checked sample), D95 amendment in DECISIONS.md, compiler/README.md (supported line and differences).
- Tests: `testdata/fixed_arrays.nra` (literal with expected type, constant indexes, compound assignment, `mutable` parameter, generic `sum<const N: int>` and `count<T, const N: int>`, proofs by `if`, C-style for, `Ranges.upTo(limit)` with a checked limit, `while`, a generator, a `const` captured by a closure, `FixedArrays.get`/`length`, `for ... of`, `copy(...)`, `==`, `Show`, FixedArrays in structs, enums and FixedArrays, an empty one). Errors `testdata/errors/fixed_arrays_{constant_index,unproven,loop_assigned,closure,generic,length,index_type,shared_write}`.

## Soundness cases

Each is an error sample unless marked accepted:

- loop variable assigned in the body, before or after the use, and through a closure passed to a call (`loop_assigned`); `i--` step, unknown start, `i += 2` with a variable bound and `i <= count` (both could wrap past the largest int).
- `N` of 0: every constant index is out of range (`constant_index`); `first<0>` in a generic instance (`generic`).
- negative indexes: `values[-1]`, `if (i < 3)` alone, `Ranges.between(-1, 3)` (`unproven`, `constant_index`).
- sized index types: `u8`, `u64` and `bool` indexes are rejected, so no index wraps (`index_type`); `values[i + 1]` is never proven.
- proof made before the index is reassigned (`unproven reassigned`); reassigning the array keeps `N`, so it needs no rule.
- index proven against a different array's `N` (`unproven otherLength`, `generic mix<3, 2>`).
- proof inside a closure: a captured `let` has no facts inside the arrow; an arrow parameter is unproven (`closure`); a captured `const` keeps its fact (accepted). In a generator, a `while (index < 3)` without the lower bound is rejected; with both ends it is accepted.
- open `N` in a generic body: no fact and no error in the declaration; each instance proves its indexes (`generic`).

## Review round 1

- Blocking hole (also on the base): `narrowByBinary` kept the left side's facts when a closure in the right side of `&&` (or of a negated `||`) assigns the name. Fixed in its own commit before this one, `fix: narrowing ends for names assigned in the rest of a condition` (divisor sample `testdata/errors/divisor_assigned_in_condition`, proof in `log.md`). FixedArray samples: `testdata/errors/fixed_arrays_condition_closure` (if, `||` exit).
- Size cap: one FixedArray value takes at most 1048576 bytes (1 MiB), checked with a saturating multiply in the layout code (`multiplySaturated`, `findFixedArraySize`, `FIXED_ARRAY_SIZE_LIMIT` in `check.nra`). The cap keeps a local well inside a default thread stack, and is far below what clang rejects. E0995 at the written type (`resolve.nra`), or, for a generic struct instance, at the struct holding it (`reportLargeFixedArrays` in `finish`). Samples `fixed_arrays_size`, `fixed_arrays_size_instance`.
- A name whose facts pin it to one value out of range (`const j = 3`, then `a[j]`) now gets "index 3 is out of range".
- `i = i + 1` counters are listed in compiler/README.md's differences.

## Deviations

- Files outside `shared:`: `selfhost/src/effects.nra` (no `alloc` for a FixedArray literal), `selfhost/src/stdlib.nra` (the `std/fixedarray` module row, kept inline there instead of a `stdfixedarray.nra`, which is outside the owned paths).
- The length resolution lives in `resolve.nra`, not `fixedarrays.nra`: `fixedarrays.nra` imports `operands.nra`, which imports `funcvalues.nra`, which imports `resolve.nra`, so `resolve.nra` cannot import `fixedarrays.nra`.
- D95 lists `if (i < 4)` as a proof; an index is also proven not negative (as the brief says), so the check is `i >= 0 && i < 4`. Recorded in DECISIONS.md.
- The C-style for condition now narrows its body for every check (divisors and shift counts too), not only for indexes.

## Stubs left

- None.

## Gotchas

- For unsafe-ptr (same matches): the exhaustive matches that got a case are `buildKey`, `typeString`, `typeId`, `cType`; the `_` ones are listed above (`isCachedType`, `cachedCType`, `layoutValue`, `measureValue`, `isCountedType`, `reachesEnum`, `visitOrder`, `containsTypeParameter`, `containsOpen`, `unify`, `substitute`, `declsdump`).
- A literal argument does not tell a generic call its length (`sum([1.0, 2.0])` is E0220; write `sum<2>(...)`); inference would need a hint in `checkGenericArguments` (`closures.nra`).
- A generic body cannot build `FixedArray<T, N>` with an open `N` from a literal (E0995).
- `for (mutable item of values)` takes an Array only.
- The goldens `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` list the std modules, so they changed with `std/fixedarray`.

## Proposed learnings

- A proof for an access with no runtime check needs both ends and no wrap | `i < N` alone allowed negative indexes, and `i += 2` with a variable bound can wrap to a negative value that passes `i < n` | selfhost/src/fixedarrays.nra, operands.nra | yes (error samples)

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/fixed-arrays/build`: stage 1 and stage 2 emit identical C (181243 lines); seed older than the source, as expected mid-batch.
- `NARRAC=<build> sh scripts/test.sh --filter fixed_arrays`: pass. Every new sample runs under `-sanitize`, the leak check and the debug copy check in the programs suite, and in release.
- `NARRAC=<build> sh scripts/test.sh` (full, log full1.log): two goldens that list the std modules failed; updated.
- `sh scripts/test.sh` before the review (log full-final.log): PASS, exit code 0.
- After review round 1, rebased on feat/batch-8 with unsafe-ptr (conflicts kept both `Type` cases and both code ranges): `sh scripts/test.sh` (log /private/tmp/claude-501/fixed-arrays/full-rebased.log): PASS in 85 s, exit code 0, 8 shards and the bootstrap job; bootstrap fixed point at 183623 lines.

## Size

- Estimated M, actual M to L: the type touches about 25 files, most with one case.
