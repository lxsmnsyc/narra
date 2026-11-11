# Outcome: place-ids

## Built

All in `selfhost/src/sharedwrites.nra`; no other source file changed.

- **Places are integers.** `Place` is `{ root, id, text }`, all `int`. `PlaceTable` (one per body) interns each place as a root, or as a parent id plus one part id, with its depth. The locals have the ids 0 to localCount - 1, and constants are added when named. `isUnder` climbs the parent ids and compares part ids (`isSamePart`: `[]` matches a literal index, two different literals never match). It works without strings and returns at once when the roots differ.
- **Message texts are interned too.** The text of a place is an id into a chain of pieces (a name, `.items`, `[i]`, `[0]`). `writeText(checker, places, text)` joins it only when an error is reported. The texts 0 to localCount - 1 are the locals' names, read from the checker at report time. Write sites keep their kind (`SITE_ASSIGN`, `SITE_COMPOUND`, `SITE_CALL`, `SITE_ARROW`) and the place's text id, and `writeSiteText` writes them out when they are reported. As a result `SharePair`, `PendingWrite` and `Place` hold no strings or `Expr`: a pair's origin is an index into `WriteWalk.origins`. Scans and copies of pair arrays no longer retain and release anything.
- **`PartTable`** holds the parts, the pieces and the answers of `canShareType` per enum index (`canShareKnown`). It does not depend on the body and is kept for the whole run.
- **Pairs are a set.** `FlowState.pairKeys` (`makePairKey`, smaller id first) replaces the `hasPair` scan. `mergeStates` became `mergeInto(mutable into, other)`, which merges in place. A dead `into` takes the other state's arrays, and the order of the pairs is kept (left first).
- **One walk per run.** `startBody` resets the walk per body and keeps the tables the body left empty. The scope of a module is looked up only for walked bodies, and kept while the module stays the same.
- **Work skipped without changing results:**
  - `canWriteInBlock` skips a body that has no field or element assignment, no `mutable` argument, no arrow function and no `for (mutable ...)`, because such a body cannot report. It descends where the walk descends.
  - `evaluate` returns at once when the parts share nothing.
  - `evaluateIgnored` handles values nobody keeps: a place read is only a use, with no share array and no type queries.
  - `evaluateCall` returns no shares when the call's type cannot hold any, and keeps argument shares only when the result or a std `mutable` argument can take them.
  - `addPair`, `noteWrite` and `forgetRoot` skip pairs with no side on the root concerned.
  - `noteUse` and `forgetRoot` rebuild the pending list only when something changes.

## Proof

- base: feat/batch-7 7f0e0534, compiler built from the seed (current at that commit), Apple M2, release build, `/usr/bin/time -l`.
  - On the base source: `narrac emit selfhost` median of 3 is 2,441,776,326 instructions with a 135.2 MB peak; `narrac check selfhost` is 1,933,205,959.
  - `sample` over 6 runs of `emit selfhost`: 275 of 1041 samples under `reportSharedWrites` (26%). Self time there was `memcmp` 51, `findRest` 44, `hasPair` 19, `memmove` 18.
  - A build whose `reportSharedWrites` returns at once emits selfhost in about 2.0G, so the analysis cost about 0.45G to 0.5G.
- branch, on the same source (the branch source), base and branch binaries run in turn, 5 runs each, median:
  - `emit selfhost`: base 2,522,357,620, branch 2,150,927,787 (**-0.371G, -14.7%**).
  - `check selfhost`: base 2,029,668,902, branch 1,652,296,591 (**-0.377G**).
  - Peak memory of `emit selfhost`, 9 runs in turn, median: footprint base 139.2 MB, branch 138.8 MB; max RSS base 136.2 MB, branch 134.7 MB.
  - The analysis is now about 0.14G of `emit selfhost`, down from about 0.5G.
- The machine ran other workers' test suites the whole time (load 30 to 50). Base and branch were measured in turn so the load hits both alike, but the absolute counts move by up to 0.1G from one run to the next: base `emit selfhost` read 2.45G to 2.56G on the same source. Measured three times late in the work, the cut was -0.36G, -0.38G and -0.37G.
- Diagnostics: `check --json` of every `testdata/*.nra` and `testdata/errors/*.nra` (669 inputs, 36 E0978 lines, codes, messages, positions and fixes), plus `check selfhost`, `check selfhost/tests` and `std`. The base compiler's output on the base source and the branch compiler's output on the branch source are byte-identical (`cmp`).

## Deviations

- The brief's direction asked for a memoized prefix relation if `findRest` stayed hot. It did not: `isUnder` on ids is a few compares. The remaining cost was elsewhere (allocation, string hashing, per-body setup, side-table lookups), and the steps after the id change went after those costs.
- No file outside the brief's paths changed.

## Stubs left

- None.

## Gotchas

- The scratchpad folder is shared with other workers in the same session. Another worker wrote `scratchpad/new/narrac` and `scratchpad/measure.sh` while I used them, so some of my mid-work numbers came from its binary. Use a folder named after the node.
- E0978 treats every struct as tracked, even one made only of integers: reading `walk.state.pairs[index]` into a local and then passing `mutable walk` is reported, so `addPair` still takes `copy(...)` of the all-integer pair. The copy is a plain copy of integers.
- `canWriteInBlock` must follow the walk's descent. If the walk ever notes a write from a new kind of statement or expression, add it there too, or that body is skipped and its errors are lost.
- `PartTable` lives for the whole run, so the part, piece and text ids assume fewer than 1,048,576 parts, pieces and texts across the package (not per body) and lines and columns below 1,048,576 (`PART_STRIDE`, `POSITION_STRIDE`). Larger values would overflow the combined keys.

## Proposed learnings

- Keep structs that hot loops scan or copy free of strings and enums (use ids, and an index into a side array for an `Expr`) | each scan or copy retains and releases every string and payload, which cost more than the comparisons | selfhost/src | no
- Never write a scratch file under a generic name in the shared scratchpad; use a folder named after the node | other workers in the session write to the same folder | roadmaps (worker briefs) | no

## Verification

- Every step: `check selfhost` with the base compiler (no errors, so the branch source passes its own E0978), a 3-stage build with a fixed point, the diagnostics diff (identical), and the interleaved instruction count.
- Formatted with `narrac fmt selfhost/src/sharedwrites.nra` (the branch compiler).
- Fast check: `sh scripts/test.sh` (built from the seed, 8 shards and the bootstrap job): PASS in 360 s, 6442 cases and the bootstrap job, 0 failed. Stage 1 and stage 2 emit identical C (187,499 lines); the seed is older than the source (expected, the orchestrator regenerates it at the end of the batch).

## Size

- Estimated M, actual L: the id change alone gave -0.14G; reaching -0.35G needed nine more steps, and the shared scratchpad cost several wrong measurements.

- Review fix (orchestrator, after the worker stopped): the text of a constant is no longer cached per constant root (`constantTexts`), so two import aliases of one constant keep their own names in E0978 messages. Sample: `testdata/packages/errors/shared_write_aliased_constant`.
