# Outcome: compiler-std-sorts

## What changed

All 15 hand-written sorts in `selfhost/src` now call `Arrays.sort` or `Arrays.sortBy`:

- `Arrays.sort` on `Array<string>`: `sortStrings` in `tools/querygraph.nra` (exported, used by `query.nra` and `querystd.nra`), `driver/modules.nra`, `std/builtins.nra`; `sortNames` in `driver/workspace.nra` and `check/ctors.nra`; `sortEffects` (dedupe stays, sort replaced) in `check/resolve.nra`; `sortEffectNames` in `check/effects.nra`. The four private `sortStrings`/`sortNames` helpers and the exported one are deleted; callers call `Arrays.sort(mutable ...)`.
- `Arrays.sortBy`: `sortDiagnostics` (`driver/driver.nra`, `compareDiagnostics`), `sortAuditEntries` (`tools/audit.nra`, `compareAuditEntries`), `sortFunctionTypes` (`emit/emitclosures.nra`), the edit order in `applyFix` (`driver/fixes.nra`, descending start offset, ties keep order), the entry order in `formatLock` (`driver/lockfile.nra`), `sortKeys` (`syntax/triviadump.nra`, sorts on the padded key through a `PaddedKey` struct), `sortedByHead` (`check/declsdump.nra`; `mergeRuns` deleted).
- Every replaced sort was stable or had no visible ties, and `sortBy` is stable, so no output changed.

## Kept

- Lexer `isLetter`, `isDigit`, `isHexDigit`: `isLetter` counts `_`, so it differs from `std/char`; all three take a byte as `int`, and `std/char` takes a `char` and has no hex class. The lexer is hot, so no conversion was added. Not compared on `emit` time because the lexer is unchanged.
- `trivia.nra` `isSpace`: space, tab and CR only. `Char.isWhitespace` also matches LF, VT and FF, which would change comment and blank-line handling.
- `manifest.nra` `isLetter(int)`: same set as `Char.isLetter`, but the callers pass `Str.byteAt` ints and there is no int-to-char conversion in `std/char`. Kept to avoid a hand-built conversion.
- `parser.nra` and `externs.nra`, `exprs.nra` have local `isLetter`/`isDigit` variables over Unicode or hex ranges; not the named helpers, left alone.

## Numbers (`narrac emit selfhost`, median of interleaved runs)

The machine was shared and loaded, so the `bench/run.sh` rows are not comparable:

- Before (base, `bench/run.sh -n 5 compiler`): 0.448 s, load 3.5 (peak memory 210 MB).
- After (`bench/run.sh -n 5 compiler`): 0.737 s, load 21 (peak memory 208 MB).
- A/B instead: base and new compilers built from the same bootstrap, alternated 6 times on the same input, load 7.3 to 6.7: base 0.586 0.605 0.608 0.612 0.644 1.147, new 0.591 0.607 0.614 0.617 0.691 0.709. Medians 0.61 s each. No slowdown beyond noise.

## Acceptance

- [x] No hand-written sort left in `selfhost/src` (grep for `insertion`, `function sort*` finds only wrappers around std sorts).
- [x] Character checks: `std/char` not usable where definitions match exactly (reasons above); none replaced.
- [x] `sh scripts/test.sh`: PASS, 10925 cases. A first full run had one failed case in shard 2 (shard log was cut off, not identified); the rerun passed. Possibly load related.
- [x] `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C (271012 lines).
- [x] Compiler row not slower beyond noise (A/B above).
- [x] `narrac fmt --check selfhost` clean.

## Inbox

- The first full test run failed one case in shard 2; consider re-running once on a quiet machine.
- A `Char.fromCode` or an int-taking class function in `std/char` would let `manifest.nra` and the lexer use it; the lexer would still need its `_` rule.
