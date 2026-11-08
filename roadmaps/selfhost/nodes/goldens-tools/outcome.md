# Outcome: goldens-tools

## Built

Eight suites in the Narra test runner (`selfhost/tests/src/`), all through narrac, no Go call. Each is a module with `collectCases` and `runCase` and one entry in `registry.nra`.

| Suite (module) | Cases | What it checks | Go tests it replaces |
|---|---|---|---|
| dumps (`dumps.nra`) | 147 | `parse` 8, `decls` 53 (top level, imports/, traits/), `types` 56 (bodies, calls, closures, effects, generators, patterns, stdcalls), `pkg` 28; exit 0 means stdout equals the golden, exit 1 means stderr equals it; `err_*` types samples must fail and others must not; 2 coverage cases (every AST node kind in the `.ast` goldens, every line kind in typesdump/bodies) | TestASTDump (goldens and kind coverage), TestDeclDump, TestDeclDumpImports, TestDeclDumpTraits, TestPkgDump, TestTypesDump, TestTypesDumpGenerators, TestTypesDumpCoverage |
| fmt (`fmtsuite.nra`) | 69 | 55 goldens (--stdin, golden stable, --check, in place, `_w<N>` and `_none` options), 10 configs (stdin and in place, `expected.err` for bad configs), syntax_error, lost_comment, no_pwd, directory | TestSelfhostFmt (golden, config, syntax_error, lost_comment, no_pwd, directory subtests) |
| fmt-stable (`fmtstable.nra`) | 786 | every .nra under testdata and selfhost/src plus 3 edge files: formats to a stable text, keeps tokens, comments and tree; a file that does not parse is rejected | TestSelfhostFmt `stable/` subtests |
| fmt-doc (`fmtdoc.nra`) | 29 | 14 goldens, 10 malformed documents (exit 1), the pinned error line, 4 documents nested 50000 deep | TestSelfhostFmtDoc, TestSelfhostFmtDocErrors, TestSelfhostFmtDocErrorText, TestSelfhostFmtDocDeep |
| trivia (`trivia.nra`) | 789 | 3 goldens, `trivia --check` on every corpus file and 3 edge files | TestSelfhostTrivia |
| query (`query.nra`) | 3 folders (72 commands) | every command of `queries.txt`, byte for byte (`# exit`, `# stderr`) | TestSelfhostQuery |
| sources (`sources.nra`) | 1 | `fmt --check` over selfhost/src, selfhost/tests and testdata (except fmt/ and `ternary_parens.nra`); only error or diagnostics samples may fail to parse | TestSourcesFormatted |
| node-ids (`nodeids.nra`) | 848 | `pkg --ids-dup` must fail with "is used twice"; `pkg --ids` prints `ids ok` on every package and file that loads | TestSelfhostNodeIds, TestSelfhostModuleIds |

Helpers: `scratch.nra` (temp folder with a work dir and a stdin file), `fmtrun.nra` (`narrac fmt` with directory, `PWD` set or cleared, stdin), and `files.nra` gained `findSources` and `listCorpus`.

`scripts/test.sh`: when `NARRAC` is unset it runs `sh selfhost/bootstrap.sh <workdir>` and uses that narrac (cc only). `NARRAC` set behaves as before.

Total with the earlier suites: 3185 cases plus the bootstrap case.

## Deviations

- Not ported: `TestASTDump` determinism over every sample (Go dumping twice), `TestASTDumpDocKinds` (reads `astdump.go`), `TestDeclDumpSkipsBodies`, `TestPkgDumpOrder` (Go API checks; the goldens cover the formats), and every Go-vs-narrac comparison test.
- node-ids decides "the input loads" with `narrac pkg PATH` exiting 0 (Go used its own loader). A regression that makes a loadable input fail to load is caught by the dumps suite, not here.
- The sources suite also checks `selfhost/tests` (Go checked selfhost/src and testdata only).
- Query is one case per folder (as in Go), not per command.
- No file outside the owned paths changed.

## Stubs left

- None. `--update` rewrites the goldens of dumps, fmt, fmt-doc, trivia and query.

## Gotchas

- `Str.slice` and `Str.sliceFrom` panic in the middle of a character; dump lines hold non-ASCII text. Compare bytes (`Str.byteAt`) and step with `Str.nextChar`.
- `find -name '*.nra'` lists a directory called `dir.nra` (pkgdump/nested_order/src); use `-type f`.
- A case that formats "the current directory" must keep its stdin file outside the work dir (`Scratch`).
- `PWD=` in `Command.environment` clears `PWD` for a child (the no_pwd case).
- Cases per shard are equal by count; fmt-stable, trivia and node-ids dominate (about 2400 of 3200).

## Proposed learnings

- Keep a case's stdin file outside the folder the command reads | `fmt` with no path formats the current directory | selfhost/tests | no

## Verification

- `sh scripts/test.sh` with `NARRAC` unset on a clean tree: PASS, 3185 cases in 8 shards of 398 plus the bootstrap job. Runner wall time 47 s after the 37 s bootstrap build (8 CPUs, macOS arm64).
- Break tests: a line appended to one golden of dumps (astdump, declsdump, typesdump, pkgdump), fmt, fmt-doc, trivia and query. Each suite failed with a diff and passed again after restoring. The sources suite failed on unformatted files before `narrac fmt selfhost/tests` was run.
- `narrac fmt selfhost/tests` applied.
- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: ok, 68 s.

## Size

- Estimated M, actual M.
