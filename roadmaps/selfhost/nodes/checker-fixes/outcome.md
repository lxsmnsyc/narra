# Outcome: checker-fixes

## Built

`narrac` diagnostics carry the same fixes as Go's `narra`, in `check --json`, and `narrac apply-fix` applies them.

### Fix model: `selfhost/src/fixes.nra` (module `narrac/fixes`)

- `Edit { start, end, text }`, `Fix { title, edits }`; builders `createFix`, `createEdit`, `insertAt`, `moveColumn`, `buildQuoteFix`, `buildRemoveSemicolonFix`, `buildQualifyFix`, `buildTemplateEscape`, `sliceCharacters`, table `STD_ALIAS_PATHS`, and `applyFix(source, fix)` (a port of Go's `ApplyFix`: columns count characters, edits apply from the end, ties keep their order).
- Builders that need checker results sit at the report site: `buildImportFix` and `findImportAt` (check.nra), `buildArrayPushFix` (calls.nra), `buildLengthFix` and `buildConcatFix` (exprs.nra), `buildRenameUnusedFix` (check.nra), `buildRenameAliasFix` (scopes.nra), `buildHoistFix` (stmts.nra).
- Report with a fix: checker `reportErrorFix(mutable checker, pos, code, fixes, message)` (the same dedupe as Go's `errorfFix`: an earlier error with the same fix title and start loses its fix); parser `failAtFix(mutable parser, pos, code, fixes, message)`; lexer `errorWithFixesAt`. `CheckError`, `parser.Diagnostic`, `LexError`, `LoadError` and the driver's `Diagnostic` have `fixes` (default `[]`).
- `writeDiagnosticJson(file, pos, code, message, fixes)` writes `"fixes":[{"title":..,"edits":[{"file":..,"range":{"start":[l,c],"end":[l,c]},"text":..}]}]`, the bytes of Go's encoder.

### Source facts the fixes read

The Go AST keeps source text, header line, alias and name positions and call ends. Here the parser fills `File.facts` (`SourceFacts`, ast.nra) and `declareModule` copies it to `ModScope.facts`:

- `source`, `header` (line after the leading comment block), `inTemplate`, `dotUses` (names written before a `.`), `namePositions` (by node id: function parameters, `const`/`let` names, `for` variables), `shorthandIds` (`{ host }` patterns), `callEnds` (by call id).
- `ImportDecl.aliasPos` and `ImportName.end`. `LocalInfo` has `namePos`, `shorthand`, `writes`.

### Fix sites (all of Go's)

Lexer: E0002 (quotes), E0003 (semicolon). Parser: E0014 (default import), E0020 (`T[]`), E0033 (`===`), E0038 (`null`), E0923 and E0924 (class, interface; only for a fields-only body), E0927 (`throw` inside a Result function, found by parsing the expression ahead, then restoring the parser). Checker: E0113 (two fixes) and E0114, E0208 (`.length`), E0216 (string `+`), E0308 (unknown module), E0309 (`xs.push(v)`), E0323 (bare constructor, with import when the enum is not visible), E0421 (alias rename), E0425 (constructor import), E0908 and E0914 (missing import), E0921 and E0922 (unused, also renames assignments), E0929 (hoist the mutable call).

### Apply-fix path (for tests; goldens-programs ports the habit tests on it)

`narrac apply-fix [--diagnostic N] [--fix M] path` (path is a file or a package directory).

- Checks the path as `narrac check` does, picks diagnostic N (0-based, report order; default: the first diagnostic that has a fix) and its fix M (default 0), and prints on stdout the full source of the file the fix edits, with the fix applied. For a package it reads `<dir>/<diagnostic file>` (for example `src/main.nra`).
- Exit 0 and the text on success. Exit 1 with `no diagnostics`, `no such diagnostic` or `no such fix` on stderr. Exit 2 for usage.
- The test writes the output to a file and runs `narrac check` on it:
  - TestHabitFixes: `apply-fix path` for `errors/{semicolon,strict_eq,method_call,string_plus}.nra`, `diagnostics/*.nra`, `errors/unused_*.nra`, then the fixed file must check clean (29 files, all do).
  - TestHabitCases: for each of the 20 cases, `check --json` gives `fixes` of diagnostic 0; `--diagnostic 0 --fix M` for every M below the count; each result checks clean; a case wants a fix or not as in `habit_cases.json`.
  - TestCtorImportFix: `apply-fix compiler/testdata/packages/errors/ctor_import`, write to `src/main.nra` of a copy, check the copy.
- `habit_cases.json` in this node folder holds the 20 inline cases of `diagnostic_test.go` (name to source), made from `habitCases` with a throwaway Go test. Use it to write them as `testdata/diagnostics/` files before Go goes.

### Checks (this node folder, throwaway)

- `compare.py <bindir>` (a dir with `narra` and `narrac`): `check --json` byte-equal on the 332 error samples with Go's fixes kept; for every Go fix, `apply-fix` equals the source with Go's edits applied by a Python port of `ApplyFix`; the same on the 20 habit cases; explain 281 cases.
- `check-codes.py`: checker-codes' check with `reportErrorFix`, `failAtFix` and `errorWithFixesAt` added (338 sites, 0 problems).

## Deviations

- Outside the owned paths: `selfhost/src/ast.nra` (`File.facts`, `SourceFacts`, `DotUse`, `createEmptyFacts`, `ImportDecl.aliasPos`, `ImportName.end`), and sites in `exprs.nra`, `calls.nra`, `ctors.nra`, `match.nra`, `resolve.nra`, `scopes.nra`, `stmts.nra`, `declare.nra`, `lexer.nra`, `manifest.nra`, `modules.nra`, `driver.nra` (fix sites, the `fixes` field and the `apply-fix` command; `apply-fix` is in `driver.nra`, not a new file).
- `findEnumName`'s callers: the E0323 pattern message now says the plain enum name when the module cannot name the enum, as Go does (it printed an empty name).
- A struct field default must be a constant, so `facts`, `aliasPos`, `end` and `namePos` have no default; their few constructors pass them.
- `isDeclaredEnum` in scopes.nra repeats `isEnumName` of exprs.nra, which scopes.nra cannot import.

## Stubs left

- None. `queryjson.nra` still writes `"fixes": []` for `narrac query` (Go's `query` output has no fixes either; the goldens pass).

## Gotchas

- Go quirks reproduced: a call inside a template string gets the call end token position + 1 (the fix text is odd but equal to Go's); `errorfFix` does not drop a fix with a reversed range.
- In Narra, `!(x is Some(name))` cannot bind: write the positive `if` and handle the other case in `else`. `from` and `uses` are not usable as names.
- The use collector's `blocked` counter replaces Go's `hoist` parameter: it is raised for the right side of `&&` and `||`, match arms and block statements.

## Proposed learnings

- A fix site passes its fixes to `reportErrorFix` or `failAtFix`, never edits an error after it was reported | the dedupe of nested fixes is in the reporter | selfhost/src | no
- Keep what a fix reads from the source (positions, text) in `File.facts` | the AST has no spacing and the checker has no source | selfhost/src | no

## Verification

- `cd compiler && go test -timeout 45m ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: 110 passed.
- `go test ./internal/narra -run 'TestSelfhostQuery|TestSelfhostFmt|TestSelfhostParser|TestSelfhostLexer|TestSelfhostModules|TestASTDump'`: 4130 passed.
- `python3 roadmaps/selfhost/nodes/checker-fixes/compare.py $BIN` (stage 2 from `build-stage2.sh`, fixed point ok): 332 samples, 0 differences, 66 fixes (65 diagnostics) applied equal; 20 habit cases, 4 fixes equal; explain 281/281.
- `check-codes.py`: 338 sites, 0 problems. `narrac fmt --check selfhost/src`: clean.
- First fix of the 29 TestHabitFixes files applied through `apply-fix`: all check clean.

## Size

- Estimated M, actual M. The parser and AST side tables (name positions, call ends, dot uses) were more than the 44 builders.
