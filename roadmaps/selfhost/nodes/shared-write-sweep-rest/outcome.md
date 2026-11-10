# Outcome: shared-write-sweep-rest

## Built

- **`narrac check selfhost` is clean without the switch: no E0976, no E0978** in any module. The rest modules went from E0978 62 and E0976 103 to 0 and 0. The edits:
  - The fixes E0976 and E0978 offer (`from <parameter>`, `copy(...)` at a return, `const x = copy(...)`), applied by a script from one `check --json` per round, as shared-write-sweep did. E0976 took 9 rounds. The new marks moved E0978 into callers (139 at the peak, mostly `const t = copy(readNodeType(emitter, ...))` before a `mutable emitter` call); 3 more rounds.
  - Results of all-scalar structs (`Pos`, `Code`, `Token`, `LoopInfo`, `Location`) get `copy(...)` at the return instead of `from` (`getDeclPos`, `getStmtPos`, the `get*Pos` helpers in parser, fmtprinter, fmtexprs, lspnavigation, querysites, `currentToken`, `peekToken`, `advance`, `leaveLoop`, the held-argument return in `emitExpression`, `siteLocation`).
  - Loop-held writes: loops over `copy(emitter.checker...)` in emittraits (7), emitmatch, emitexprs and emit.
  - `emitClosure` (emitclosures.nra) now clears frames, scopes, loop frames, loop infos and blocks right after it saves them (nothing in between reads them), so the saves need no copy.
  - Struct literals that the fix wrapped whole now copy the one field that shares: `Parser { tokens: copy(tokens) }` (twice), `Collector { tokens: copy(tokens), templateTokenEnds: copy(...) }`, `Walk { modules: copy(modules) }`.
  - Other hand edits: `Arrays.push(mutable emitter.blocks, copy(info))`, `copy(bodyBlock(...))`, `copy(readNodeType(...))` where an `is` binding of the type outlives a `mutable emitter` call (emitexprs, stdinternal).
  - The fmt printer's function fields `printExpression`, `printPattern` and `printBlock` are `from printer` (E0977 after the marks; the printer is never written).
- **The switch is gone.** `shared-results = "unchecked"` removed from `selfhost/narra.toml`; no other package used it, so it is removed everywhere the from-marker outcome lists: `Manifest.sharedResultsUnchecked` and the `package.shared-results` key (manifest.nra), `Package.sharedResultsUnchecked` (modules.nra), `Checker.sharedResultsUnchecked` (check.nra), its assignment and test in body.nra, SPEC.md section 6, DECISIONS.md D94 (Rollout, now in the past tense) and compiler/README.md. The key is now skipped like any unknown `[package]` key; E0409 stays (other uses).
- **The shared results check got a cache.** With the switch gone the E0976 walk runs on every compiler body, and `collectHeldTypes` walked the whole type graph of `Checker`/`Emitter` for every `mutable` argument: `emit selfhost` cost 3.96 G instructions (+66%). The held-type set of each type is now built once per type key and kept in `Checker.heldTypeSets` (moved into `SharingFacts` for a body and back). That brought it to 2.46 G.
- **`selfhost/tests`**: the 3 `newCommand(...)` results take the fix: `copy(Shell.newCommand(...))` (context.nra x2, lsp.nra).
- **`testdata`**: no program outside the error and fix samples warns. Every sample keeps its `.out`. Per sample:
  - `copy_report`: the report is the point, so the two hidden copies now come through a `mutable` argument of a user function (`keep(mutable bag, items)`), which the static rule does not follow. Same `.out`, same `.copies` line, checks clean.
  - `copy_hoist_empty_loop`, `copy_hoist_guarded`: the saved originals are `copy(...)`; the copy check still passes, and the loops still test where the uniqueness test goes.
  - `loop_unique_hoist`: the six snapshots are `copy(...)`, and the "not hoisted" push case pushes `copy(items)`; the loop shapes are unchanged.
  - `loop_unique_generator`, `coroutine_heap`: `let copied = copy(original)` for the generator copied before resuming (local renamed from `copy`).
  - `perceus_loop_items`: the loop that pushes to the array it iterates runs over `copy(growing)`; comment updated.
  - `maps`: `b = copy(a)`; the loop that writes its own map iterates `Maps.keys(copy(loop))`.
  - `arrays`, `box_assign`, `structs`, `mutable_contents`, `strings2`, `const`: the second name is `copy(...)` (locals named `copy` renamed to `copied`).
  - `arrays`, `mutable_for`, `mutable_for_shared`: array literals that hold one value twice copy their elements (`[copy(words), copy(words)]`, `players: copy(players)`, `[copy(board), copy(board)]`); `saved = copy(counters)`.
  - `params_pointer_tail`, `tail_loop_borrow`, `strings2_builder`: the local taken from a parameter and then written is `copy(...)`; the borrowed parameters are untouched.

## Proof

- Counts (base 25cd3092, measured with the base compiler; E0978 with the switch on, E0976 with it removed locally, because a package with errors shows no warnings):

| | base | branch |
|---|---|---|
| E0978, rest modules | 62 | 0 |
| E0976, rest modules | 103 | 0 |
| E0978 and E0976, checker modules | 0 and 0 | 0 and 0 |
| `selfhost/tests` | E0978 3 | 0 |
| `testdata` programs and packages outside `errors/` | E0978 37, E0976 14 | E0978 1, E0976 14 |

  What is left in testdata is on purpose: `diagnostics/shared_write_fix` (E0978) and the five `diagnostics/from_*_fix` samples (E0976) are fix samples, and nine `declsdump` samples have E0976 next to unused-name, missing-entry or trait errors that they keep on purpose (dump samples, not checked programs).
- `narrac emit selfhost`, instructions retired (`/usr/bin/time -l`, release builds, each compiler on its own source):
  - base: 2,380,458,720 / 2,378,739,849 / 2,378,579,127, median **2,378,739,849** (idle start); again at the end 2,415,347,282 / 2,375,868,559 / 2,376,661,564, median 2,376,661,564.
  - branch: 2,478,661,421 / 2,455,517,939 / 2,454,289,107, median **2,455,517,939**: **+3.3%** (load average 32 at the time; instruction counts barely move with load).
  - Split: the branch compiler on the branch source with the switch put back measured 2.39 G (before the cache), so the sweep's `copy(...)` calls cost about +0.5% and the check the compiler now runs on itself about +3%. Before the cache the check cost +66% (3.96 G).
- No behavior change: base and branch compilers emit identical C for the base source (raw and after `cc -E -DNARRA_RELEASE`, 178,629 lines preprocessed).

## Deviations

- **A compiler change outside the sweep:** the held-type cache in `sharedresults.nra` and `Checker.heldTypeSets`. Without it, removing the switch makes every compiler build cost 66% more instructions. It changes no result (sets depend only on the type key and the declared fields).
- The `declsdump` samples with E0976 are left as they are (see Proof). Marking them changes their dumps and still leaves their other errors.
- `copy_report` was rewritten rather than moved to `testdata/errors/`: once E0978 is an error, a sample that keeps the runtime report needs a hidden copy the static rule misses.

## Stubs left

- None.

## Gotchas

- `from` on a function that takes `mutable emitter` or `mutable parser` moves warnings into every caller that holds the result across another `mutable` call. For all-scalar results (`Code`, `Token`, `Pos`) write `copy(...)` at the return instead.
- The E0976 walk now runs for the compiler's own package on every build. Any new per-call work in `sharedresults.nra` is paid in every `emit selfhost`; profile it (`sample` on a `-g` build of stage2.c works).
- A `mutable` argument of a user function still does not make it keep the other arguments (shared-write-check outcome). `copy_report` depends on that limit; if shared-write-error closes it, the sample needs another hidden copy or moves to `testdata/errors/`.
- The seed is older than the source (expected; no seed update here).

## Proposed learnings

- Return all-scalar structs (`Code`, `Token`, `Pos`) taken from a `mutable` parameter as `copy(...)`, not `from` | `from` makes every caller warn | selfhost/src | no
- Copy the one field that shares when a struct literal takes a parameter, not the whole literal | the fix wraps the whole literal, which reads as a copy of a fresh value | selfhost/src | no
- Measure `emit selfhost` after changing what the checker runs on the compiler's package | the from check alone cost +66% before its cache | selfhost/src/sharedresults.nra | no

## Verification

- `narrac check selfhost` and `narrac check selfhost/tests` with the branch compiler: no output.
- `sh selfhost/bootstrap.sh out/new`: stage 1 and stage 2 emit identical C (184,832 lines); the seed is older than the source.
- `NARRAC=<branch> sh scripts/test.sh --suite errors --suite habits --suite codes --suite sources --suite dumps --suite fmt --suite spec --suite query --suite generated`: PASS, 1416 cases. `--suite programs`: PASS, 321 cases.
- `JOBS=3 sh scripts/test.sh` on 7c21a345: PASS in 341 s, 6327 cases and the bootstrap job, 0 failed.

## Size

- Estimated L, actual L, plus the cache the switch removal needed.
