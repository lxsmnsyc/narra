# Outcome: compiler-std-sets

## Built

- 150 `Map<K, bool>` declarations in 41 files under `selfhost/src` are now `Set<K>` from `std/set`. That covers locals, parameters, return types and struct fields (`Checker`, `Emitter`, `HelperNotes`, `BodyContext`, `ParseFacts`, `Trivia`, `TypeTraits`, `LspCache` and others). Their uses moved from `x[k] is Some(_)`, `x[k] is None`, `x[k] = true`, `Maps.has`, `Maps.remove`, `Maps.size` and `Maps.keys` to the `Sets` functions. Eleven more files changed only where they read those fields (for example `emitdynamic.nra`, `fmtprinter.nra`, `triviadump.nra`, `simdtypes.nra` and `stdvalues.nra`). That makes 52 files.
- `heldTypeSets` (`check.nra`, `sharedresults.nra`) is now `Map<string, Set<string>>`.
- 27 places where a membership test was followed right away by an insert (`if (seen[k] is Some(_)) { return } seen[k] = true`) are now a single `if (!Sets.add(mutable seen, k)) { return }`. That saves one hash lookup per visit and offsets the extra lookup inside `Sets.add`.
- The parser's template merge loops for `referenced` and `taggedCalls` copied the values (always `true`) with `Maps.entries`. They now add each member with `Sets.values` and `Sets.add`. The insertion order is the same.
- `sharedresults.nra` no longer imports `std/map`.

## Maps kept

| Map | File | Why |
| --- | --- | --- |
| `Checker.busy` | `check/check.nra` | Holds `false` after a `typeof` or constant resolution ends. `resolve.nra` reads `is Some(true)` to tell a cycle in progress from a finished entry. |
| `Checker.foldedConditions` | `check/check.nra` | The value is the folded truth of a build condition. The emitter keeps only that branch. |
| `PartTable.shareAnswers` | `check/sharedwrites.nra` | A memo of yes/no answers, and a `false` answer is cached. |
| `LayoutMemo.heapEnums` | `check/check.nra` | A memo of the heap decision for each enum, and `false` is a decision. |

## Numbers

`bench/run.sh -n 5 compiler`, which runs `narrac emit selfhost` on the tree being measured:

| | time (median of 5) | peak memory | load before | load after |
| --- | --- | --- | --- | --- |
| base (90e5c5a6) | 1.277 s | 207 MB | 4.89 | 13.31 |
| this change | 0.670 s | 209 MB | 9.39 | 11.09 |

Other workers kept the load high and variable during both runs, so the time column mostly reflects load. An interleaved A/B run in one window gives a fairer comparison: the base compiler on the base tree against the new compiler on the new tree, three rounds of median-of-5. Load at the start was 9.39 (1-minute average) and 8.90 at the end.

- round 1: base 0.835 s, new 0.904 s
- round 2: base 0.796 s, new 0.741 s
- round 3: base 0.710 s, new 0.615 s
- peak memory: base 200 MB, new 201 MB

Neither time nor memory changed beyond noise.

## Acceptance

- [x] Every membership-only `Map<K, bool>` in `selfhost/src` is a `Set<K>`, and the kept maps are listed with the reason. `grep -rn 'Map<[^,]*, bool>' selfhost/src` now finds only the four maps above, plus a string inside the generated `stdtext.nra`.
- [x] `narrac emit selfhost` output is identical before and after. The base compiler (built with `sh selfhost/bootstrap.sh` from 90e5c5a6) and the new compiler produce byte-identical C (`cmp`) on the new tree (12930726 bytes). They also produce byte-identical C on the base tree (12970184 bytes, from `git archive 90e5c5a6`), and that output matches the base compiler's earlier output.
- [x] `sh scripts/test.sh` passes: `PASS, 10925 cases in 355s` with `NARRAC` set to the new compiler. The first full run, which built from the seed, failed only `programs/async_api_all_cancel.nra/debug`. That test orders two 1 ms timers, and the run had load near 25. It passed 3 out of 3 with `--filter`, and the full rerun passed.
- [x] `sh selfhost/bootstrap.sh` reaches the fixed point: "stage 1 and stage 2 emit identical C (269341 lines)". The seed is older than the source as expected, and the orchestrator regenerates it.
- [x] The compiler row is not slower or larger beyond noise (see Numbers).

## Inbox

- Looping over a `Set` costs an array: `Sets.values` copies the members, while `for (... of Maps.keys(m))` is emitted as a direct map loop with no allocation. Five loops in the compiler now pay this cost, all small or cold (two template merges in the parser, two trivia dump loops, `haveCommonType`). A `for ... of` over a `Set` (or `Sets.values` lowered like `Map.keys`) would remove it.
- `async_api_all_cancel` depends on timer order under load and can fail on a busy machine.
