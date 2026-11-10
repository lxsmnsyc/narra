# Outcome: shared-write-error

## Built

- **E0978 is an error** (`selfhost/src/sharedwrites.nra`, reported with `reportErrorFix`; the fix is unchanged). The early return on other errors is gone, so E0978 is reported next to them. A write that another error already points at (same file, line and column) is skipped, because `mutable_contents_for` got the same write twice.
- **It runs in every command that checks**: `loadAndCheck` in `driver.nra` (build, run, test, check, apply-fix), `emitter.nra` (emit), and the language server. In `lspcache.nra` it is a fourth per-module phase, `ModuleResult.sharedWriteErrors`: a whole check stores each module's E0978 after the effect pass; an incremental check walks only the checked modules (only they have bodies) and reuses the cached list of the others. Merged order: check errors, unused errors, then shared writes, module by module, which is the whole check's order. The query command and the dumps do not run it.
- **Warnings are gone from the compiler**: `Checker.warnings`, `reportWarningFix`, `findWarnings`, `Diagnostic.warning`, the `warning` parameter of `writeDiagnosticJson` (JSON `severity` is always `"error"`), and the test runner's `// warning:` support (errors and habits suites, codes suite row).
- **Function values share their captures** (`canHoldShared` in sharedwrites.nra). The pair from an arrow function value to what it captured was dropped, because a function type holds no tracked value, so `const read = () => items[0]; let other = items; other[0] = 5; read()` checked clean and copied at run time. The compiler stayed clean.
- **Literal indexes are compared** (second review). A place key writes an integer literal index as `[=N]` (`writeIndexKey`), and `findRest` matches keys bracket by bracket: `[]` (any other index) matches every element, two different literals never match. `rows[0]` and `rows[1]` are different places; `rows[i]` and `rows[j]`, `rows[i]` and `rows[0]`, and map keys stay conflated. The same literal index stays one place, and assigning it does not end its sharing (that false positive stays). Messages still read `rows[0]` (`writeSuffixText`). Samples: `testdata/shared_write_indexes.nra` (clean, different literals, also through a field; the copy check reports nothing) and `testdata/errors/shared_write_indexes.nra` (identifier indexes, the same literal, an identifier against a literal).
- **E0290 hint**: a constant passed as `mutable` or looped over with `for (mutable ...)` now suggests `copy(NAME)`; it suggested `let name = NAME`, which E0978 then rejects. `checkConstWrite` takes an `inPlace` flag. Goldens `const_mutable_arg`, `mutable_for_constant` and the guide's chapter 9 error line changed.
- **Debug copy check on for every program.** `Shell.runtimeCheckEnvironment()` (renamed from `leakCheckEnvironment`) sets `NARRA_CHECK_COPIES=1` for every program the test runner starts, so programs, packages, spec, guide and std test samples all run with it. `Copies.checkCopyRun` handles every program; a `<name>.copies` file works for any sample, and the release-C check stays on `copy_*` samples only.
- **Reports found and what became of them** (first run with the check on: 4 programs):
  - `generators_values`, `generator_effects_composites`, `params_pointer_escape`: a `for ... of` or `yield*` over a generator it does not own resumes its own copy (SPEC section 8). That copy is the language's, so `emitgenerators.nra` (`takeOwnCopy`) now wraps the loop's retain in `NARRA_COPY_GRANT(true/false)`, like `copy(...)`.
  - `mutable_for`: `const team = teams[0]; resetTeam(mutable team)` with `teams` not used afterwards. The static rule accepts a write after the other name's last use, but `teams` holds the array until its scope ends, so the write copies. The sample now takes `copy(teams[0])`. See Gotchas.
  - Left with a report on purpose: `copy_report` (array 2, unchanged: its copies hide behind a `mutable` argument of a user function, which the rule still does not follow) and the new `copy_enum_credit_end` (array 1, below).
- **Credit gap for heap enum payloads** (`emitdecls.nra`). The copy now marks each payload box its retain reached (`narra_copy_credit` on the box, emitted by `buildCreditCode` at the `copy(...)` site). The mark has one credit per copy. The heap enum's release, in debug builds only, compares the mark with the copies that can still hold the box after this release (count minus 2): it calls `uncredit_<enum>` once for each copy past that and lowers or forgets the mark. `uncredit_<enum>` takes one credit back from each counted part (decrements, and forgets at the last one), walking nested payload boxes without marking them. `credit_<enum>` walks the parts only. Renamed: `creditFunction` to `buildCreditFunction` (writes `credit_` or `uncredit_` by `operation`), `creditCode` to `buildCreditCode`, `creditPartCode` to `buildPartCreditCode`, plus `buildPayloadCall`; their comments are inside. The runtime text did not change (the generated code uses `narra_copy_credit`, `narra_copy_find`, `narra_copy_forget` and `narra_copy_slot`, which exist), so no seed update is needed.
  - `testdata/copy_enum_credit_end.nra` with `.copies` (exit 103): a later hidden copy after a dead `copy(tree)`. Before the fix it exited 0. The review's p2 probe is now an E0978 error itself, so the sample hides its sharing behind a `mutable` argument of a user function.
  - `testdata/copy_enum_credit_twice.nra` with `.copies` (exit 103): two overlapping copies of one enum. The second review found that the first mark cleared at the first release, so one uncredit ran for two credits and the later hidden copy went unreported (exit 0).
  - `testdata/copy_enum_credit_kept.nra` (reports nothing): the fresh review found that the first version marked nested boxes too, so a short-lived reference to a nested payload ended the copy's credits, and that forgetting all credits of a part erased an unrelated `copy(part)` credit. Both cases are in this sample; both exited 103 with the first version.
- **Samples.**
  - The 15 `testdata/errors/shared_write_*` warning samples are error samples (`// error:`, exit 1). `shared_write_clean` moved to `testdata/shared_write_clean.nra` (a program with `.out`, run under the copy check).
  - 14 new error samples `testdata/errors/shared_write_scalar_*`, one per D2 form with an all-scalar struct `Pos`: field, reverse, chain, mutable_whole, mutable_field, closure, parameter, pattern, for_const, for_mutable, loop_held, captured, cell, call_result.
  - `mutable_for_generator.err` gains an E0978 line (its `const copy = players` is a real shared write).
  - Five guide samples (chapters 2, 4, 8, 9, 11) failed to build; they take `copy(...)`, and chapter 8's `addWord` loops with `for (mutable tally of tallies)`.
- **LSP test** `testdata/lsp/incremental_shared_writes` (the lsp suite compares it with `lsp --full`): a shared write in `text.nra`'s body (only `text.nra` is checked), a body edit in `other.nra` while `text.nra`'s cached E0978 stays, the fix with `copy(...)`, a shared write next to a type error (E0119 then E0978) and next to an unused function (E0947 then E0978).
- **`reportFound`** returns at once when a body found nothing, and looks up `errorPlaces` only for found writes when the map is not empty.
- **Docs.** D94 amendment in DECISIONS.md (error everywhere, which indexes are conflated, function values, copy check, the known gap). `compiler/README.md` line for E0978. SPEC.md section 6, debug copy check: the generator loop copy counts as explained; enum payload credits end with the copy.

## Proof

- Instructions retired (`/usr/bin/time -l`, release builds, median of 3):

| | base 0e1d09b1 | branch |
|---|---|---|
| `narrac emit selfhost` | 2,468,220,886 | first commit 2,919,134,286 (+18.3%); after the second review 3,018,262,542 (+22.3%, +0.55 G) |
| `narrac build testdata/block_expressions.nra` (cc child not counted) | 28,507,912 | 28,812,578 (+1.1%) |
| `narrac check selfhost` (ran the analysis before too) | 2,275,239,798 | 2,298,034,514 (+1.0%, the function-value pairs) |

  The brief expected about +0.33 G per selfhost compile; it is +0.55 G, accepted by the orchestrator. The literal-index comparison adds about +0.1 G: a first version that split keys into parts cost 9.77 G; matching bracket by bracket with `Str.indexOf` brought it back. The `reportFound` skip is inside that number. The branch numbers were taken while a test run loaded the machine (instruction counts barely move with load).
- Credit gap: `copy_enum_credit_end` exits 0 with the base emitter and 103 (`array 1`) with the branch. Without the dead `copy(tree)` it exits 103 on both.
- Captured form: `const read = () => items[0]; let other = items; other[0] = 5; return read() + other[0]` checks clean on base and reports E0978 on the branch; the copy check reported its copy (exit 103) on base.

## Deviations

- **Indexes**: different literal indexes are compared (second review); the same literal index and identifier indexes stay conflated (D94 amendment).
- `copy_report` is unchanged: this node does not make the rule follow `mutable` arguments of user functions.
- Files outside the brief's shared paths: `docs/guide/{02,04,08,09,11}-*.md` (samples only, the guide suite failed without them; the prose that says the compiler cannot warn about shared copies is value-semantics-docs' to rewrite).
- Two changes beyond the brief, both found by the new samples: function values sharing their captures, and the E0290 hint. The orchestrator kept both, and the guide sample edits, in this node.
- The test runner's `leakCheckEnvironment` is now `runtimeCheckEnvironment` and turns the copy check on in every suite, not only programs.

## Stubs left

- None.

## Gotchas

- **Write after the last use still copies at run time.** E0978 accepts `const team = teams[0]; mutate(mutable team)` when `teams` is not used later, but `teams` holds its array until the end of its scope, so the write copies; the copy check reports it. Either the rule treats a name that is still in scope as live, or the emitter releases a name at its last use. The orchestrator moved it to the roadmap inbox.
- Guide prose in chapters 2, 4 and 11 still says assignment copies and that the compiler cannot warn; the samples now use `copy(...)`.
- E0978 now runs on packages with errors. Side tables of a body with errors can be partial; every error sample checked without a crash, but a new error kind that leaves odd side tables can surface here.
- The lsp incremental path assumes only checked modules have bodies (`Body.checkModules` with `skipped`). If skipped modules ever register bodies, `checkChanged` must select by dirty file.
- The mark on a payload box counts copies, not holders: when the original holder releases first, the credits of one copy end early (a false report at most, never a missed one).
- `uncredit_<enum>` takes back one credit per part. If the copy's own credit on a part was already used by a copy on write, it takes back a credit that another `copy(...)` gave; the result is at most a false report, never a missed one.

## Proposed learnings

- A new kind of runtime copy that the language defines (like a loop's own generator copy) grants credits around its retain | otherwise every program that uses it fails the copy check | selfhost/src/emit*.nra | yes (programs suite runs with the check)
- New error samples for the shared-write rule cover the all-scalar struct form too | the runtime check cannot see inline structs | testdata/errors | no

## Verification

- `NARRAC=out/narrac JOBS=6 sh scripts/test.sh` (full, with the bootstrap job) on the branch before the commit: PASS in 248 s, 6403 cases, 0 failed.
- `out/narrac check selfhost` and `check selfhost/tests`: no output.
- Fresh review of d8ac0293 (separate agent): one bug, the credit marks on nested boxes and the forget-all in `uncredit_<enum>` (fixed, sample `copy_enum_credit_kept`). Nothing serious in lspcache, sharedwrites with errors, `takeOwnCopy` or `checkConstWrite`; the lsp incremental-versus-whole order was checked by reading only.
- First commit's final: `JOBS=3 sh scripts/test.sh` (built from the seed, with the bootstrap job): PASS in 338 s, 6407 cases and the bootstrap job, 0 failed. The seed is older than the source (expected; stage 2 and stage 3 identical, 185,443 lines). `sh selfhost/bootstrap.sh out/final`: stage 2 and stage 3 emit identical C (185,443 lines); the seed is older than the source.
- After the second review: targeted `--suite errors --update` (only the new golden changed) and `--suite lsp --filter incremental_shared` (matches `--full`); `out/narrac check selfhost` and `selfhost/tests` clean. `JOBS=3 sh scripts/test.sh` (built from the seed, with the bootstrap job): PASS in 351 s, 6435 cases and the bootstrap job, 0 failed; stage 2 and stage 3 emit identical C (186,114 lines), the seed is older than the source.

## Size

- Estimated M, actual M to L: the copy check found the generator loop copies, and the scalar samples found the function-value hole.
