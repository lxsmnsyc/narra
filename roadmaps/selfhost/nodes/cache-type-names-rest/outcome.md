# Outcome: cache-type-names-rest

## Built

- `selfhost/src/cnames.nra`: `cachedModuleHash`, `cachedDeclName` (module hash cached per module), `cachedFunctionName` (per program function), `storeName` (the cache write shared with `cachedTypeId`). `functionSignature` and `parameterCType` take the cache and use `cachedCType` and `cachedFunctionName`. `constName` and `constGetterName` are gone; `cachedDeclName` replaces them.
- Every `cType`, `typeId`, `functionName`, `functionSignature` and `parameterCType` call in `emit*.nra` and `std*.nra` now goes through `emitter.typeNames` (the one cache map from cache-type-names; new keys start with `module|`, `plain|`, `instance|` and `impl|`, which no type key uses). Uncached calls remain only in `dropDuplicateTypes`, which runs before an Emitter exists.
- Edited files formatted with `narrac fmt`. The seed is not regenerated.

## Proof

- base: 1ee1a90a, `/usr/bin/time -l narrac emit selfhost`, instructions retired, median of 9 pairs 1.746 G (range 1.741 to 1.774 G, spread 1.9%), load average about 10 on 8 cores.
- branch: same command, median 1.686 G (range 1.681 to 1.708 G). Branch won 9 of 9 pairs, about 3.4% fewer instructions, above the spread. Emitted C of `selfhost` is byte-identical (`cmp`).

## Deviations

- None.

## Gotchas

- The cache is keyed by strings and unbounded. `functionSignature` now needs a cache, so any later caller outside the emitter must pass one.
- feat/batch-5 moved during the work; compare against the commit this branch was cut from, not the branch tip.

## Verification

- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C (166575 lines).
- `sh scripts/test.sh`: passed (see the hand-off report).

## Size

- Estimated S, actual S.
