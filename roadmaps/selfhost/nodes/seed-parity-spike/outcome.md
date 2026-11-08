# Outcome: seed-parity-spike

## Built

- A report, no code. Throwaway scripts in this folder, each runnable from the repo root with a bin dir that holds `narra`, `narrac1` and `narrac2`:
  - `build-stage2.sh <dir>`: builds Go `narra`, stage 1 and stage 2 (release), and checks the fixed point.
  - `survey-tests.sh`: prints each Go test with its doc comment and helpers.
  - `diagdiff.py <dir> <out.json>`: diagnostics of Go and stage 2 on errors/, diagnostics/, packages/errors/, stage2/errors/.
  - `habitdiff.py <dir> <work>`: the inline habit cases.
  - `codematch.py <out.json>`: whether a code can be found from the message text.
  - `rundiff.py <dir> <work>`: programs, panics, packages, runner through stage 2 in debug and release.
  - `golddiff.py <dir> <work>`: fmt, fmt-doc, trivia, query, guide through stage 2.
  - `dumpdiff.py <dir>`: astdump, declsdump, typesdump, pkgdump goldens through stage 2.

## Result

- result: met. Every language-level suite already passes on stage 2. What only Go has is the diagnostic codes, fixes, JSON and explain, the user commands, and the generators of two selfhost source files.
- numbers (stage 2 at b569948):

| Suite | Stage 2 result |
|---|---|
| errors/*.nra (TestCompileErrors) | 286/286: `// error:` text found; every diagnostic line equals Go's (file:line:col: message, same order) |
| diagnostics/*.nra | 18/18 lines equal Go's; 0/18 codes (stage 2 has none); 0/18 fixes |
| packages/errors/* (TestPackageErrors) | 26/26 contain the `# error:` text and equal Go's lines |
| stage2/errors/*.nra | 2/2 |
| habitCases (inline, 20) | 20/20 text equal to Go's |
| All diagnostics together | 330 samples, 410 Go lines, 0 text or position mismatches; 142 distinct codes and 65 samples with a fix that stage 2 cannot produce |
| testdata/*.nra (TestPrograms) | 81/81, debug (ASan, UBSan, leak check) and release; C byte-equal to Go's |
| stage2/*.nra | 1/1 |
| panics (TestPanics 3 + TestStringPanics 6) | 9/9 exit 101 with the message |
| packages with run.out (TestPackages) | 8/8 debug, release and tests mode |
| runner/mixed.nra | 1/1 both modes |
| fmt goldens / config / fmt-doc / trivia | 55/55, 10/10, 14/14, 3/3 |
| `fmt --check selfhost/src` | clean |
| query (3 cases) | 72/72 commands |
| guide samples (TestGuide) | 177/177 (89 run in both modes, 88 errors; 11 fragments skipped) |
| dump goldens | astdump 8/8, declsdump 53/53, typesdump 56/56, pkgdump 28/28 |

- Code lookup from text: of the 410 Go diagnostics, 289 match exactly one format with the right code, 113 match several formats, 3 match only a wrong one, 5 match none. Examples: "expected int, found void" fits E0011 (parser `expected %q, found %s`) and E0119 (checker); `call_type_args_builtin.nra` E0326 reads as E0115; `effects_indirect_*.nra` E0800 messages start with "Function" and match no format. Codes must be passed where the error is reported.
- reproduce: commit b569948, `sh roadmaps/selfhost/nodes/seed-parity-spike/build-stage2.sh $DIR`, then the scripts above with `$DIR`. macOS arm64, Apple clang, Python 3.9.

### Go test files

Port: the test checks the language or narrac and must run on the Narra runner. Drop: it compares Go with narrac or checks Go internals, and goes with Go. Parity: needs a feature only Go has (checker-parity).

| File | What it checks | Testdata | Verdict |
|---|---|---|---|
| run_test.go | TestPrograms (run, compare .out, debug + release), TestPanics (3), TestCompileErrors (`// error:` substring) | *.nra, panics/, errors/ | Port (core) |
| stage2_test.go | TestStage2Programs, TestStage2Errors; TestStage2OutputComparison (helper self-test); TestStdInternalFromStdModule (Go CheckPackage) | stage2/ | Port the first two (core, merge into programs and errors); drop the other two |
| std_string_test.go | TestStringPanics (6 panic messages) | panics/strings2_* | Port (core) |
| std_test.go | TestRunStdTestFailures (9 inline programs, exit 101 + message), TestRunStdTestBuilds | inline | Port (goldens): move the inline sources into testdata files |
| modules_test.go | TestPackages, TestPackageErrors; TestManifest (manifest parser, 8 inline errors); TestEntryForms (Go unit) | packages/ | Port the first two (core); TestManifest as pkgdump cases; drop TestEntryForms (pkgdump/entry_forms covers it) |
| runner_test.go | 8 tests of the test binary: mixed, --filter/--time, golden match/mismatch/missing/update, package goldens, timeout, leak, arguments, Test.golden outside a test | runner/, packages/, inline | Port (goldens); needs env vars NARRA_UPDATE_GOLDEN, NARRA_TEST_TIMEOUT, NARRA_TEST_FORCE_LEAK and temp dirs |
| guide_test.go | TestGuide (177 samples), TestGuideExtract (fence parser) | docs/guide | Port (goldens); extractor rewritten in Narra |
| diagnostic_test.go | TestEveryFormatHasCode (Go source scan), TestErrorCodes (every error has a code, JSON), TestHabitFixes (fix makes file compile), TestHabitCases (20 inline), TestGenericErrorOnce, TestHabitExplain, TestExplain | errors/, diagnostics/, packages/errors/ | Parity: all need codes, fixes or explain. TestHabitCases becomes 20 testdata/diagnostics files. TestEveryFormatHasCode is replaced by whatever enforces a code per report site |
| literal_type_args_test.go | one diagnostic per sample | errors/literal_type_args_* | Port (covered once error goldens compare full output) |
| ctor_import_test.go | the import fix of a constructor import, applied, compiles | packages/ | Parity (fixes) |
| cnames_test.go | TestXXH32, TestModuleHash (Go units); collision stop, names ignore directory, names stable under edits (C output properties) | packages/ | Drop the two units (TestSelfhostXXH32 covers the hash); port the three properties as `narrac emit` comparisons (goldens) |
| astdump_test.go | TestASTDump (Go dump deterministic, goldens), TestASTDumpDocKinds (Go doc) | astdump/, all *.nra | Port goldens with `narrac parse` (8/8 already equal); drop DocKinds |
| decldump_test.go | goldens, bodies skipped | declsdump/ | Port with `narrac decls` |
| pkgdump_test.go | goldens, load order | pkgdump/, packages/ | Port with `narrac pkg` |
| typesdump_test.go, generators_test.go, selfhost_calls/closures/effects/patterns/stdcalls_test.go | TestTypesDump* goldens (Go dump) and *Coverage (goldens hold every line kind); TestTypesDumpInstanceOrder; TestSelfhost{Calls,Closures,Effects,Patterns,StdCalls} (Go vs narrac on all inputs); TestSelfhostPackageIsClean | typesdump/*/ | Port goldens with `narrac types` and the coverage greps; drop the Go-vs-narrac tests; PackageIsClean is implied by the bootstrap |
| selfhost_imports_test.go, selfhost_traits_test.go, selfhost_decls_test.go | TestDeclDumpImports/Traits goldens; TestSelfhostImports/Traits/Decls (Go vs narrac); TestSelfhostBuiltins (+ProbeFails, Go registry vs stdlib.nra); TestStdModuleOrder; TestSelfhostPrelude; TestUnported*Error (skip rules); TestSelfhostNodeIds (`pkg --ids`, `--ids-dup`) | declsdump/ | Port goldens and NodeIds; drop the rest |
| selfhost_test.go | TestSelfhostLexer/Parser/Modules (Go vs narrac on 719 testdata + 63 src files + 16 lex and 82 parse inline edge cases); TestSelfhostModuleIds | all | Drop the comparisons; port ModuleIds. Freeze the 98 inline edge cases as testdata files with goldens from Go before it goes |
| selfhost_bodies_test.go | TestSelfhostBodies (Go vs narrac), TestUnsupportedConstruct | all | Drop |
| selfhost_emit_test.go | TestSelfhostEmit (C equal to Go on every emit input), TestSelfhostEmitErrors (2 messages), TestSelfhostXXH32 | all | Drop emit; EmitErrors as 2 error samples; port XXH32 via `narrac hash` |
| selfhost_bootstrap_test.go | stage 1 = Go, fixed point, stage 2 = Go on emit inputs | selfhost/ | Replaced by `bootstrap.sh` fixed point (seed-file); Go half dropped |
| selfhost_fmt_test.go | TestSelfhostFmt (goldens, config, --check, in place, lost comment, PWD, directories, idempotence over the corpus), TestSourcesFormatted | fmt/, all | Port (goldens); narrac-only today |
| selfhost_fmtdoc_test.go, selfhost_trivia_test.go, selfhost_query_test.go | fmt-doc goldens and errors and 50000-deep doc; trivia goldens and `--check` over the corpus; queries.txt | fmt/, query/ | Port (goldens); narrac-only today |
| selfhost_runtime_test.go | writes/checks `selfhost/src/runtimetext.nra` from Go runtime strings | none | Replace: see Gotchas |
| std_sources_test.go | writes/checks `std_sources.go` and `selfhost/src/stdtext.nra` from `std/src` | std/src | Replace: see Gotchas |

Counts: of 96 tests, about 48 port (most are already narrac-only or golden-based), 8 need checker-parity (7 in diagnostic_test.go and TestCtorImportFix), about 38 drop, 2 are generators to replace.

### Commands Go has and narrac lacks

- `build [-release] [-bin name] [-o out] [-emit-c file.c] path`: narrac has only `emit [--tests] [--bin name] path`. Package build makes every `[[binary]]`; `-o` with several binaries is an error; output is named after the file or binary.
- `run [-release] [-bin name] path [args...]`: builds into a temp dir, passes stdin, stdout, stderr and args, exits with the program's code.
- `test [-filter text] [-time] path [args...]`: builds with Tests, runs with `--filter`, `--time` and `--golden-dir <abs dir of file, or <pkg>/testdata>`.
- `check [--json|-json] path`: exit 1 on errors, JSON one object per line (needs codes, so it belongs with checker-parity; plain `check` can land in selfhost-driver).
- `explain CODE` (case-insensitive, 289 formats, 17 long texts).
- `lex path`: narrac's bare `narrac <file>` already prints the same dump. `parse`, `pkg`, `decls`, `types`: narrac has them.
- Flags: Go's `flag` package accepts `-x` and `--x`; narrac emit takes `--x` only.
- cc call (driver.go `Build`): `$CC` or `cc`; debug `-O0 -g -std=c11 -w <CFlags> -o out file.c -lm`; release `-O2 -DNARRA_RELEASE -std=c11 -w -o out file.c -lm`. The C file goes to a temp dir unless `-emit-c`. cc output goes to stderr; failure is `C compiler failed: ...`.
- Sanitizers and leaks are not a mode of `narra build`: the tests add `-fsanitize=address,undefined -fno-omit-frame-pointer` through `CFlags` and run with `NARRA_CHECK_LEAKS=1 ASAN_OPTIONS=detect_leaks=0` (a runtime env var, not a build flag). `selfhost/bootstrap.sh` builds stage 2 with release flags; `TestSelfhostBootstrap` builds it with the debug sanitizer flags.

### Files outside compiler/ that use Go

- `selfhost/bootstrap.sh`: stage 0 `go build ./cmd/narra`, stage 1 `narra build -release`, compares with `narra emit`.
- `bench/run.sh` line 37: `go build -o $build/narra ./cmd/narra` (line 63 builds the Go benchmarks; that stays). `bench/README.md` line 22 cites `compiler/internal/narra/driver.go`; `bench/BASELINE.md` line 8 records `narra build -release`.
- `scripts/linux-check.sh`: runs `go vet ./... && go test` in the container. `scripts/linux/Dockerfile`: `FROM golang:1.23-bookworm`, needs a plain image with clang or gcc.
- `selfhost/CLAUDE.md` lines 7, 18, 19: `narra build selfhost`, `selfhost_test.go`, the `go run ./cmd/narra build` format recipe.
- `compiler/README.md` (whole usage and test sections), `compiler/internal/narra/CLAUDE.md` (all Go rules), root `README.md` line 7.
- `docs/guide/01-getting-started.md` line 12 (`go build -o narra ./cmd/narra`) and the `narra run/build/check/test` command table; `08-modules-and-packages.md`, `10-effects-and-testing.md`, `11-narra-for-typescript-programmers.md`, `docs/guide/README.md` use `narra <command>` (fine if the self-hosted binary is installed as `narra`).
- `DECISIONS.md` line 457 cites `compiler/internal/narra/cnames.go`; `SPEC.md` names `narra check/test/explain` (language commands, not Go).
- `selfhost/src`: 282 comment mentions of "Go" or Go files (`check_types.go`, `std_map.go` ...) in 49 files. Harmless but stale after Go goes.
- No `.github/` workflows exist.

## Proposals for the later nodes

- checker-parity, L, split in two:
  - checker-codes (M): a code at every report site (the 289-row table of `diagnostic_codes.go`, ported as data), `narrac check --json` per `contracts/diagnostic.md`, `narrac explain`, the E0000 placeholder in `queryjson.nra` replaced, and a check that every report site has a code. Owns `selfhost/src/check.nra` (CheckError gains `code`), `parser.nra`, `lexer.nra`, `modules.nra`/`manifest.nra` error sites, a new `selfhost/src/diagnostics.nra`, `queryjson.nra`. About 250 checker report calls plus parser, lexer and loader sites. Text and positions need no work: 0 mismatches.
  - checker-fixes (M): the fix builders of `diagnostic_fixes.go` (importFix, concatFix, pushFix, quoteFix, lengthFix and the parser fixes, 44 fix sites), fix edits in the JSON, an apply-fix path for tests. Acceptance: TestHabitFixes, TestHabitCases, TestCtorImportFix equivalents (65 samples with a fix today). Depends on checker-codes.
  - If the user accepts losing fixes for now, checker-fixes can be parked (`later`) and only checker-codes stays must.
- test-runner-core, M. Scope: a Narra package (proposed `selfhost/tests/`, its own `narra.toml` and binary) that builds samples with `narrac emit` + `cc` and runs: programs (81 + stage2 1, debug and release, `.in` stdin for 2 files), panics (9), errors (286 + stage2 2 + packages/errors 26), packages (8, three runs each), runner/mixed. One command; takes the testdata root as an argument so it survives a move. Owned: `selfhost/tests/**`, a `scripts/test.sh` wrapper. Gaps it must work around: `Process.run` blocks, closes stdin, cannot set env and has no timeout; std/fs has no remove or temp dir. Proposed workaround: run every child through `sh -c` (env, `< file`, `mkdir -p`, `rm -rf`), and get parallelism by running N shards of the runner from the wrapper script (`--shard i/N`), or by `sh -c '... & ... & wait'` batches. Proposed upgrade: compare the full stderr of error samples with a `.err` golden (all 330 already equal Go's full output), generated once from Go while it exists; keep the `// error:` line as documentation.
- test-runner-goldens, L as listed; split in two M:
  - goldens-tools (M): dumps (145 goldens), fmt (55 goldens, 10 configs, the in-place, --check, PWD and directory cases, corpus idempotence), fmt-doc (14 + errors + deep), trivia (3 + corpus `--check`), query (72), sources formatted, node ids, the coverage greps. All narrac-only today, mechanical.
  - goldens-programs (M): guide (177 + extractor), runner tests (8, env vars and temp dirs), std/test failures (9 inline programs to testdata files), cnames properties (3), the 2 emit-error messages, and freezing Go-only inline cases into testdata with Go-made goldens: 16 lex and 82 parse edge cases, 20 habit cases, 8 manifest errors. Freezing must land before retire-go.
- selfhost-driver (running): note that Go's debug build has no sanitizer flags. If narrac's debug build adds ASan and UBSan, as its brief says, that differs from Go; the runner then needs a way to build without them, or the brief should match Go.
- retire-go, M. Scope: delete `compiler/cmd`, `compiler/internal`, `compiler/go.mod`; move testdata (below); switch the files in the list above; replace the two generators (below); update `scripts/linux/Dockerfile`. Owned: `compiler/**`, `bench/run.sh`, `bench/README.md`, `scripts/**`, `selfhost/CLAUDE.md`, `README.md`, `docs/guide/01-getting-started.md`, `selfhost/src/runtimetext.nra`, `selfhost/src/stdtext.nra`, a new `selfhost/runtime/` folder. Keep the gate. Optional follow-up (S, later): reword the 282 "Go" comments in `selfhost/src`.
- Generators, S, may merge into retire-go: `runtimetext.nra` (77 KB of one-line strings) is the only copy of the C runtime once Go goes, and nobody can review a diff of it. Proposed: move `runtime*.h` and the per-module `Module.Runtime` strings to `selfhost/runtime/*.h` (or `.c`), and add `narrac gen-text` or a shell script that writes `runtimetext.nra` and `stdtext.nra`, checked by the runner. Must land before or with retire-go.

### Where testdata should live

- `git mv compiler/testdata testdata` (repo root), done in retire-go, last. The layout inside stays; only 5 files outside `roadmaps/` name `compiler/testdata` (3 Go tests, `selfhost/CLAUDE.md`, `DECISIONS.md`), plus 195 mentions in roadmap files that need no change. Until then the runner takes the root as an argument with `compiler/testdata` as the default, so no node waits on the move. `selfhost/tests/` for the runner keeps code and data apart; the root `testdata/` holds no compiler code and suits other test tools later (the LSP, lint).

## Deviations

- The brief asked to compare diagnostics with "the expected output". The samples hold only a `// error:` substring (and `// code:`), so the comparison is against Go's full output, which is stricter. Both are reported.
- Stage 2 used is the release build of `bootstrap.sh`, not the debug sanitizer build of `TestSelfhostBootstrap`. Text output does not depend on it.

## Stubs left

- None.

## Gotchas

- `narrac fmt` finds `narra.toml` through `$PWD`, not the working directory. A runner that only sets cwd gets the default options, and 30 fmt cases fail.
- Error lines differ only by narrac's `error ` prefix. Go prints `file:line:col: message`; narrac prints `error file:line:col: message`. A single file shows its base name.
- `Process.run` closes stdin: `io_stdin.nra` and `io_stdin_all.nra` need `sh -c 'prog < file'`.
- Codes cannot come from matching message text (113 of 410 ambiguous). Pass them at the report site.
- The Go-vs-narrac tests are today the only check of lexer, parser, loader, checker and emitter on all 782 corpus files and the 98 inline edge cases. After Go goes, only the golden suites remain. Freezing the inline cases keeps the error paths covered.
- `TestSelfhostBootstrap` uses sanitizer stage 1 and stage 2. The seed bootstrap should keep one sanitizer self-build somewhere (the runner), or memory bugs in the compiler go unseen.

## Proposed learnings

- A test harness that runs `narrac fmt` sets `PWD` to the working directory | the config lookup reads `$PWD` | selfhost/tests | yes
- Report a diagnostic with its code at the call site | the text alone maps to several codes | selfhost/src | yes, a test that every report call passes a code from the table

## Verification

- `sh build-stage2.sh <dir>`: fixed point ok.
- `diagdiff.py`, `habitdiff.py`, `codematch.py`, `rundiff.py`, `golddiff.py`, `dumpdiff.py`: results in the table above.
- No compiler code changed, so the fast check was not run.

## Size

- Estimated M, actual M.
