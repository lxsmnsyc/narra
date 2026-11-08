# Outcome: goldens-programs

## Built

Seven suites in the Narra test runner (`selfhost/tests/src/`), each a module with `collectCases` and `runCase` and one entry in `registry.nra`. All run through narrac. Frozen inputs and their goldens were made with the Go compiler (last use of Go).

| Suite (module) | Cases | What it checks | Data |
|---|---|---|---|
| edge (`edge.nra`) | 98 | `narrac FILE` (lex dump) on 16 files, `narrac parse FILE` on 82 files; golden on stdout (exit 0) or stderr (exit 1) | `testdata/lex/*.nra` + `.lex`, `testdata/parse/*.nra` + `.ast` |
| habits (`habits.nra`) | 20 `case`, about 30 `fix`, one `codes` per error sample (about 340), `generic-once`, `explain` | `check --json` equals the Go golden; first code and fix or nofix from `.meta`; every offered fix (`apply-fix --fix N`) checks clean; the first fix of the old TestHabitFixes files and of the `ctor_import` package checks clean; every error sample has a real code; a generic error is reported once with valid ranges; `explain` | `testdata/habits/*.nra` + `.json` + `.meta` |
| manifest (`manifest.nra`) | 9 | `narrac pkg .` in a folder: `expected.err` on stderr (exit 1) or `expected.dump` on stdout | `testdata/manifest/<name>/narra.toml` (8 errors, `valid`) |
| std-tests (`stdtests.nra`) | 19 | 9 failing std/test programs, debug and release: exit 1, message of `.want`, the next test still runs; plus a program with test blocks that builds and runs | `testdata/std-tests/*.nra` + `.want`, `builds/` |
| runner-options (`runneropts.nra`) | 13 | `narrac test`: filter and time, golden file steps (missing, update, match, mismatch, longer), package goldens, timeout, forced leak, arguments (exit 2 cases, golden names), Test.golden outside a test | `testdata/runner-cases/*.nra`, `testdata/runner/mixed.nra`, `packages/runner` |
| emit (`emitchecks.nra`) | 5 | 2 emit error messages (`noentry.err`, `unknown_binary.err`), C names ignore the directory, type names stable under edits, xxHash32 vectors through `narrac hash` | `testdata/emit-errors/`, `testdata/cnames/`, `packages/cnames_modules` |
| guide (`guide.nra`) | 278 | the extractor (a Narra port of the fence parser) on a typo text and an unclosed fence; each of the 177 samples in `docs/guide/*.md`: `narra` in debug and release (stdout equals the `text` block), `narra error` (`check` reports the text), `narra fragment` passes | `docs/guide/*.md` |

Other changes: `runNarracWith(context, directory, environment, arguments)` in `context.nra`; the `sources` suite skips the folders `fmt`, `lex`, `parse`, `habits` (invalid or unformatted on purpose).

`sh scripts/test.sh` with `NARRAC` unset (seed bootstrap, then the runner): 4386 sharded cases plus the bootstrap case, all pass, 68 s wall (8 shards, macOS arm64, 8 CPUs; includes the 40 s compiler build).

## Deviations

- Outside the owned paths: `compiler/internal/narra/selfhost_fmt_test.go` (`TestSourcesFormatted` skips `testdata/lex`, `parse`, `habits`; without it the Go test fails on the new invalid inputs). `compiler/testdata/` gained `runner-cases/`, `emit-errors/`, `cnames/`, `std-tests/builds/`, which are not in the owned list; they hold inputs of the new suites only.
- Goldens for `testdata/emit-errors/*.err` were written by hand from the Go message (same text as narrac's; checked against `narra build` and `narrac emit` output by eye, not by a script).
- The std-test `.want` lines keep Go's message with the file name changed from `prog.nra` to the sample name; `narrac fmt` did not move any position.
- `TestCNameCollisionStops` is not ported (Go-internal `registerCName` panic). `selfhost/src/cnames.nra:400` has the same check, with no test that reaches it.
- Habit cases check the first diagnostic only (as Go did) and need `.meta` files (not in the brief).

## Stubs left

- None.

## Gotchas

- A file without a parse error is not enough for `testdata/lex`, `parse`, `habits`: they are excluded from `narrac fmt --check` (Narra `sources` suite and Go `TestSourcesFormatted`). The other corpus suites (fmt-stable, trivia, node-ids) include them and pass.
- C names come from the module file name, so a name-stability test must give both versions the same file name (`prog.nra` in separate folders).
- `apply-fix` without `--diagnostic` takes the first diagnostic that has a fix (as Go's `applyFirstFix`); with `--diagnostic 0` a sample whose first diagnostic has no fix fails (`ts_concat_chain.nra`).
- Narra has no C-style `for`; use `Ranges.between` or `while`. `Str` has no lower-case function.

## Proposed learnings

- Run a test case that copies a file under the same name in separate folders when the compiler derives names from the file name | C names depend on it | selfhost/tests | no

## Verification

- `sh scripts/test.sh`, `NARRAC` unset: PASS, 4386 + 1 cases, 68 s.
- Break tests (a line appended to one golden, each failed and was restored): `lex/bad_u.lex`, `parse/null.ast`, `habits/try.json`, `manifest/valid/expected.dump`, `std-tests/ne.want`, `emit-errors/noentry.err`, and a `BROKEN` line in the first `text` block of `docs/guide/02-values-and-types.md`.
- `cd compiler && go test ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: ok, 62 s.
- `narrac fmt selfhost/tests` applied.

## Final table of the Go tests (`compiler/internal/narra/*_test.go`, 96 tests)

Verdicts: replaced (by suite), dropped (compares Go with narrac, or checks Go internals), uncovered.

| Go test | Verdict |
|---|---|
| TestPrograms | replaced: programs |
| TestPanics, TestStringPanics | replaced: panics |
| TestCompileErrors, TestPackageErrors | replaced: errors |
| TestLiteralTypeArgsOneError | replaced: errors (full-output goldens) |
| TestPackages | replaced: packages |
| TestStage2Programs | replaced: programs (stage2/) |
| TestStage2Errors | replaced: errors (stage2/errors) |
| TestStage2OutputComparison | dropped (helper self-test) |
| TestStdInternalFromStdModule | dropped (Go CheckPackage) |
| TestRunRunnerMixed | replaced: runner |
| TestRunRunnerFilterAndTime, TestRunRunnerGolden, TestRunRunnerGoldenPackage, TestRunRunnerTimeout, TestRunRunnerLeak, TestRunRunnerArguments, TestRunRunnerGoldenOutsideTest | replaced: runner-options |
| TestRunStdTestFailures, TestRunStdTestBuilds | replaced: std-tests |
| TestGuide, TestGuideExtract | replaced: guide |
| TestManifest | replaced: manifest (8 errors + `valid`) |
| TestEntryForms | dropped (Go unit; pkgdump/entry_forms golden covers it) |
| TestErrorCodes | replaced: habits (`codes/`) |
| TestHabitFixes, TestHabitCases, TestGenericErrorOnce, TestHabitExplain, TestExplain | replaced: habits |
| TestCtorImportFix | replaced: habits (`fix/packages/errors/ctor_import`) |
| TestEveryFormatHasCode | uncovered: scans Go source for a code per format; no runner test enforces a code per report site (the `codes/` cases only check the samples). checker-codes' `check-codes.py` is a throwaway script |
| TestXXH32 | replaced: emit (`xxh32`, same vectors through `narrac hash`) |
| TestSelfhostXXH32 | replaced: emit (`xxh32`) |
| TestModuleHash | dropped (Go unit; `cnames/ignore-directory` covers the effect) |
| TestCNameCollisionStops | dropped (Go-internal `registerCName`; narrac's check at cnames.nra:400 has no test) |
| TestCNamesIgnoreDirectory, TestCNamesStableUnderEdits | replaced: emit (`cnames/`) |
| TestSelfhostEmitErrors | replaced: emit (`emit-errors/`) |
| TestSelfhostEmit | dropped (C of Go vs narrac) |
| TestASTDump | replaced: dumps (astdump goldens, node-kind coverage) |
| TestASTDumpDocKinds | dropped (reads `astdump.go`) |
| TestDeclDump, TestDeclDumpImports, TestDeclDumpTraits | replaced: dumps |
| TestDeclDumpSkipsBodies | dropped (Go API check) |
| TestPkgDump | replaced: dumps (pkgdump) |
| TestPkgDumpOrder | dropped (Go API check; the pkgdump goldens hold the order) |
| TestTypesDump, TestTypesDumpCoverage, TestTypesDumpGenerators, TestTypesDumpCalls, TestTypesDumpCallsCoverage, TestTypesDumpClosures, TestTypesDumpClosuresCoverage, TestTypesDumpEffects, TestTypesDumpEffectsCoverage, TestTypesDumpPatterns, TestTypesDumpPatternsCoverage, TestTypesDumpStdCalls, TestTypesDumpStdCallsCoverage | replaced: dumps (typesdump goldens, coverage cases) |
| TestTypesDumpInstanceOrder | dropped (Go API check; the typesdump/calls goldens pin the instance order) |
| TestSelfhostCalls, TestSelfhostClosures, TestSelfhostEffects, TestSelfhostPatterns, TestSelfhostStdCalls, TestSelfhostImports, TestSelfhostTraits, TestSelfhostDecls, TestSelfhostBodies | dropped (Go vs narrac) |
| TestSelfhostPackageIsClean | dropped (implied by the bootstrap suite) |
| TestSelfhostBuiltins, TestSelfhostBuiltinsProbeFails, TestStdModuleOrder, TestSelfhostPrelude | dropped (Go registry vs narrac tables) |
| TestUnportedDeclError, TestUnportedTraitError, TestUnsupportedConstruct | dropped (skip rules of the Go-vs-narrac comparison) |
| TestSelfhostLexer, TestSelfhostParser | dropped (Go vs narrac); their 16 + 82 inline edge cases are replaced by the edge suite |
| TestSelfhostModules | dropped (Go vs narrac); the inputs are covered by dumps, node-ids and manifest |
| TestSelfhostModuleIds, TestSelfhostNodeIds | replaced: node-ids |
| TestSelfhostBootstrap | replaced: bootstrap suite and `selfhost/bootstrap.sh` |
| TestSelfhostFmt | replaced: fmt, fmt-stable |
| TestSourcesFormatted | replaced: sources |
| TestSelfhostFmtDoc, TestSelfhostFmtDocErrors, TestSelfhostFmtDocErrorText, TestSelfhostFmtDocDeep | replaced: fmt-doc |
| TestSelfhostTrivia | replaced: trivia |
| TestSelfhostQuery | replaced: query |
| TestSelfhostRuntimeText | uncovered: writes and checks `selfhost/src/runtimetext.nra` from Go runtime strings; retire-go replaces the generator |
| TestStdSources | uncovered: writes and checks `std_sources.go` and `selfhost/src/stdtext.nra` from `std/src`; retire-go replaces the generator |

Counts: 96 tests. Replaced 63. Dropped 30. Uncovered 3: TestEveryFormatHasCode, TestSelfhostRuntimeText, TestStdSources.

## Size

- Estimated M, actual M.
