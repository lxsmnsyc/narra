# Log: tree-shake-types

## 2026-11-11

- Branch `feat/tree-shake-types`, reset to feat/batch-7 at a8388ff0. feat/batch-7 moved on while I worked (build-targets, contracts-ensures); rebased onto db7a3871 before the measurements below. Scratch: /private/tmp/claude-501/tree-shake-types/.
- Already done? No: on base, every type in `Checker.order` and `composites` gets its typedef, struct, rc functions and array/box functions, and `runtimeFor` runs for every imported std module.

### Base (feat/batch-7 at db7a3871, compiler from `sh selfhost/bootstrap.sh` in a clean `git archive` copy)

| C of | lines | bytes | `-Wunused-function` | of them generated helpers |
|---|---|---|---|---|
| `bench/fib/fib.nra` | 1028 | 35973 | 135 | 5 (`A_Array_*_retain/push/get/at/filled`) |
| `testdata/tree_shake.nra` | 1064 | 37469 | 135 | 5 (same) |
| `selfhost` | 178544 | 8202362 | 588 | 471 |

The other warnings are the core runtime (`runtime.h`, always written: `narra_wrap_*`, `narra_eq_*`, `lstr_*` and so on). Generated helpers = names not starting with `narra_` or `lstr_`. `cc -std=c11 -Wall -c`.

Selfhost generated-helper warnings on base by kind: `A_*_filled` 119, `retain_*` 108, `A_*_at` 84, `A_*_get` 41, `M_*_remove` 30, `credit_*` 28, `A_*_push` 17, `A_*_retain/release` 9 each, `M_*_retain` 4, `release_*` 3, `B_*_at` 3.

### Design

- Types: the name cache (`emitter.typeNames`) already records every type id that emitted code asks for. The type sections now run after the bodies, so the cache values are the named types; `collectUsedTypes` closes them over struct fields, variant fields, element, key/value and function parts (distinct types by their base). `cachedCType` of a distinct type now goes through the cache for its base, so the base is recorded.
- Helpers: per function, not per type, because the warnings are per function (a reached `Array<string>` only uses `_new` and `_release`). `noteHelper` (a cell in the Emitter, so read-only emitter code can note) at `refCountCode`, `buildPayloadCall`, and the array/box call sites through `useHelper`. `indexHelpers` maps each possible helper name to its type and operation (names from uncached `typeId`, so the index does not mark types as named). A helper's text notes the helpers it calls (unique -> copy, push/at -> unique, copy/filled -> new, release -> drop and uncredit).
- Map functions stay one unit per map type, written when the map type is named.
- Runtime sections: `emitNamedCall` notes the std module (`Str.` -> std/string and so on, sized names -> std/int); `narra_sb_*` rc notes std/string.
- One loop: bodies, constant release, helpers, maps, trait functions, until a round adds nothing. The first version re-ran the whole trait pass whenever needs grew: +2.5% instructions (2.284G vs 2.222G). Made the trait pass incremental (`traitPrototypes`/`traitDefinitions` in the Emitter, `buildTraitText` at the end): back to 2.209G on a quick build.
- Noticed: the old inline-enum and struct rc builders computed `retainCode` before replacing it with `releaseCode`; harmless before, but it would note retain helpers that release never calls. The per-operation builder computes only the one it writes.
- Probe: a byte scan of the selfhost C in Narra (8 MB) costs 156M instructions, so scanning the emitted text for names was ruled out.

### Branch

| C of | lines | bytes | `-Wunused-function` | generated helpers |
|---|---|---|---|---|
| `bench/fib/fib.nra` | 975 | 33676 | 133 | 0 |
| `testdata/tree_shake.nra` | 1009 | 35073 | 132 | 0 |
| `selfhost` | 177500 | 8121990 | 162 | 45 (`M_*_remove` 31+, `M_*_retain` 10+, a few others) |

- `sh selfhost/bootstrap.sh`: fixed point, stage 2 = stage 3, 177500 lines; seed older than source (expected mid-batch).
- New emit case `tree-shake/unused-types` (fib has no `Option<int>`, `#[int, int]`, `_get(`, `_push(`, `_filled(`; keeps `Array<string>`, `_new`, `_release`): FAIL with the base compiler ("should not hold \"= Option<int> */\""), PASS with the branch.
- Targeted suites with the branch compiler: emit, programs, sources, leaks, panics, packages, edge, runner, guide, spec: pass.
