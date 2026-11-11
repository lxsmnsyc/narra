# Outcome: tree-shake-types

## Built

- Types: `emitProgram` (`emit.nra`) writes the type sections after the bodies. The values in the name cache (`emitter.typeNames`) are the types the emitted code named. `collectUsedTypes` closes them over struct fields, variant fields, element, key and value types, and function parameters and results (a distinct type counts as its base). Every type section filters on that set: the comment table, the array, map, type and heap typedefs, the function type definitions (and `ENV_RUNTIME`), the struct definitions, and the array, map and box structs. `cachedCType` of a distinct type now takes its base through the cache (`cnames.nra`), so the base gets recorded.
- Helpers, one function at a time: `HelperNotes`, `noteHelper(emitter, name)` and `useHelper(emitter, container, operation)` in `emitstate.nra`. The notes live in a cell, so code that holds the Emitter read-only can note. Notes are made by `refCountCode` (retain/release), `buildPayloadCall` (credit/uncredit), every array and box call site (`_new`, `_get`, `_push`, `_at`, `_filled`, `_unique`) and `emitMain`. `emitdecls.nra`: `HelperSlot`, `indexHelpers` (each possible name, built from uncached `typeId`) and `emitHelper` (the text of one function). A helper's text notes the helpers it calls. `emitRcFunctions`/`emitBoxPrototypes(emitter, helperTexts)` put the texts together in the old order and filter the prototypes the same way.
- Map functions: written per map type, when that map type is named (`EmitMaps.emitMapFunction`, `splitMapTemplate`, `listMapTypes`). The map prototypes, typedefs and structs filter the same way.
- Std runtime sections: `emitNamedCall` (`emitstd.nra`) notes the module whose runtime a call uses (`findRuntimeModule`: `Str.` gives std/string, a sized-integer name gives std/int, and so on). StringBuilder rc notes std/string. A section stays when its module is imported and noted. std/test is kept in test builds, as before.
- String literals were already limited to reached code by tree-shake. The trait pass is now incremental (`emitTraitFunctions` returns whether it wrote a function, and `buildTraitText` joins them; the state is `traitPrototypes`/`traitDefinitions` in the Emitter). Each trait function is written once, so no literals are duplicated.
- One loop in `emitReachableBodies` runs bodies, constant release, helpers, maps and trait functions until a round adds nothing. `ReachableBodies` gained `constantRelease`, `helperTexts` and `mapTexts`.
- Emit case `tree-shake/unused-types` (`selfhost/tests/src/emitchecks.nra`): the C of fib has no `Option<int>`, no `#[int, int]` and no `_get(`, `_push(` or `_filled(`, and it keeps `Array<string>`, `_new` and `_release`.

## Proof

Base: feat/batch-7 at db7a3871, compiler from `sh selfhost/bootstrap.sh` in a clean `git archive` copy. Branch: this commit, compiler from its own bootstrap. Apple M2, shared machine. Warnings are from `cc -std=c11 -Wall -c`. "Helpers" counts the unused functions whose names do not start with `narra_` or `lstr_`; the rest are the core runtime that is always written.

| C of | base lines | branch lines | base `-Wunused-function` (helpers) | branch (helpers) |
|---|---|---|---|---|
| `bench/fib/fib.nra` | 1028 | 975 | 135 (5) | 133 (0) |
| `testdata/tree_shake.nra` | 1064 | 1009 | 135 (5) | 132 (0) |
| `narrac emit selfhost` | 178544 (8202362 B) | 177500 (8121990 B) | 588 (471) | 162 (45) |

- fib is 975 lines, below the 1012 of the batch-5 compiler. The `Option<int>`, `Option<#[int, int]>` and `#[int, int]` types and their rc functions are gone.
- With the same input (the base source), the branch compiler emits 175443 selfhost lines against 178544 (-1.7%).
- `/usr/bin/time -l narrac emit selfhost`, base and branch alternating, 9 runs each. Base 2.184G median, branch 2.220G median (+1.6%) with its own (larger) source. On the same base source: base 2.191G, branch 2.208G (+0.8%). Peak 135.9 MB base, 137.1 MB branch.
- The first version re-ran the whole trait pass whenever needs grew: 2.284G against 2.222G. The incremental trait pass removed that. Scanning the emitted text for names was ruled out: a byte scan of the 8 MB selfhost C costs 156M instructions in Narra.

## Deviations

- Files outside `shared:`, kept to note sites: `cnames.nra` (distinct base through the cache); `emitexprs.nra`, `emitstmts.nra`, `stdbytes.nra`, `stdenv.nra`, `stdfs.nra`, `stdstring.nra` (`useHelper` at the array and box call sites); `emitstd.nra` (runtime module note); `selfhost/tests/src/emitchecks.nra` (the new emit case).
- The output order is the old order, filtered, with one change: trait functions, string literals and map functions follow discovery order, which is deterministic. Selfhost's C has fewer duplicate literals because trait functions are no longer rewritten.
- The rc builders now compute only the count code they write. The old struct and inline-enum builders built the retain code for release too.

## Stubs left

- Map functions are still one unit per map type. In selfhost, 31 or more `M_*_remove` and 10 `M_*_retain` remain unused (the 45 helper warnings). Filtering them needs the map template split per function and notes at the map call sites in `emitmaps.nra`.
- Core runtime functions (`runtime.h`: `narra_wrap_*`, `narra_eq_*`, `lstr_*`) are always written and draw about 130 warnings in every program. They are not generated helpers, so they were left alone.

## Gotchas

- A new site that writes the name of a generated container or rc function must go through `useHelper`, `retainCode`/`releaseCode` or `noteHelper`. Otherwise the function is not written and C fails with an implicit declaration error. The programs suite and the bootstrap catch it.
- A helper is written only if its exact name was noted. A new helper operation needs a row in `ARRAY_OPERATIONS`, `BOX_OPERATIONS` or `TYPE_OPERATIONS` and a case in `arrayFunction`/`boxFunction`/`typeRcFunction`.
- The used-type set comes from the name cache. Code that computes a type id with the cache before the bodies are emitted (for example sort keys) marks that type as used. That is safe, but it shakes less. `indexHelpers` and the map candidates use uncached `typeId` for this reason.
- A helper that calls another must note it while it writes it (`useHelper` inside the text), as `push` does for `unique`.

## Proposed learnings

- Decide what to emit from the names emitted code writes: the name cache for types, `noteHelper` for helpers. Never scan the C text | the cache already sees every type name, and a text scan costs about 7% of an emit | selfhost/src/emit*.nra | yes (bootstrap and the C compile fail on a missed name)

## Verification

- `sh selfhost/bootstrap.sh`: fixed point (stage 2 = stage 3, 177500 lines); the seed is older than the source, as expected mid-batch.
- `tree-shake/unused-types` fails with the base compiler (`the emitted C should not hold "= Option<int> */"`) and passes with the branch compiler.
- `NARRAC=<branch> sh scripts/test.sh --suite <s>` for emit, programs, sources, leaks, panics, packages, edge, runner, guide and spec: pass.
- `sh scripts/test.sh`: PASS in 95 s (8 shards, 834 to 835 cases each, plus the bootstrap job). Log in /private/tmp/claude-501/tree-shake-types/full-test.log.

## Size

- Estimated M, actual M.
