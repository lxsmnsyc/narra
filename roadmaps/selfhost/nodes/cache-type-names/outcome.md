# Outcome: cache-type-names

## Built

- `cnames.cachedTypeId` and `cnames.cachedCType` look up a per-type-key cache before calling `typeId` and `cType`. Only struct, enum, distinct, array, box, map, function and generator types are cached; the rest are string literals already.
- `Emitter.typeNames` is now a `Cell<Map<string, string>>`, so the non-mutable `refCountCode` path can fill it. A write takes the map out of the cell first, so it does not copy.
- `temp`, `declareLocalVariable`, `containerName` and `structOrEnumCode` in `emitstate.nra` use the cached calls. Other call sites are unchanged.

## Result

- Instructions retired for `narrac emit selfhost` (base e64faa2, 3 runs each): base 3.692, 3.677, 3.662 billion; branch 3.617, 3.591, 3.596 billion. About 2.1% fewer, outside the 0.8% spread. Memory unchanged.
- Emitted C is byte-identical; bootstrap stages 2 and 3 agree.
- `functionSignature` and the other `emitter.checker` call sites still use the uncached calls; moving them is a follow-up.

## Deviations

- None. The seed is not regenerated.
