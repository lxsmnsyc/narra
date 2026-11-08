# Outcome: checker-codes

## Built

Every diagnostic of the self-hosted compiler carries the code Go gives it. `narrac check --json` and `narrac explain` match Go's `narra`, except that `fixes` is always `[]` (checker-fixes).

### The code table: `selfhost/src/diagnostics.nra` (module `narrac/diagnostics`)

- `const CODE_ROWS: Array<CodeRow>`: the 289 rows of Go's `diagnostic_codes.go`, in Go's order. `CodeRow { code: string, format: string }`; `format` is the Go format string verbatim (`%q`, `%s`). A code with several formats (E0046, E0934) has several rows.
- `const EXPLANATIONS: Array<Explanation>`: the 22 long texts (`Explanation { code, text }`).
- `getCodeRows()`, `isKnownCode(code)`, `explainCode(input): Result<string, string>` (uppercases ASCII, prints `CODE: format\n` and, if there is one, `\n<long text>\n`; the Err text is Go's `unknown diagnostic code "E9999"`), `writeDiagnosticJson(file, pos, code, message): string`.
- To add a code: append a row at the end of its range (ranges are in the file's header), never renumber or reuse. The formats name and explain codes; they never decide a code.

### How a report site passes a code

The code is a required parameter placed before the message, as a string literal:

- checker: `reportError(mutable checker, pos, "E0102", `function ${quoteGo(name)} is declared twice`)`; `reportErrorIn(mutable checker, file, pos, code, message)`; `reportTypeError(mutable checker, type, code, message)`; `checkLoopExit(..., word, code, message)`. `CheckError` has a new field `code`. The duplicate filter (`errorKeys`) now includes the code.
- parser: `failAt(mutable parser, pos, code, message)`, `failHere(mutable parser, code, message)`, `failFound(mutable parser, code, message)`. `parser.Diagnostic` has a new field `code`.
- lexer: `errorAt(line, column, code, message)` and `fail(...)` build a new `LexError { code, line, column, message }`; `lex` and `lexWithComments` return `Result<_, LexError>`; `formatLexError(error)` gives the old `line:col: message`. The parser turns it into a `Diagnostic` with the lexer's code (the template-string case is E0041 around the formatted lex error).
- loader: `LoadError` (manifest.nra) has `code`; `createError(line, code, message)` for narra.toml; hash collision E0417, import cycle E0418, syntax error: the parser's code. An error that is not a diagnostic (an unreadable file, `stat ...`) has code `""` and prints as text, as in Go.
- driver: `Diagnostic` (driver.nra) has `code`; missing entry functions use E0419 (single file) and E0420 (package).
- To pick the code for a new site, find the Go format in `diagnostic_codes.go` / `CODE_ROWS` and use its code. Three codes are not exact Go formats: `unsupported: ...` (selfhost-only) uses E0213, and the two "unreachable" lexer fallbacks use E0009 and E0010.

### JSON shape (`narrac check --json`)

One line per diagnostic on stdout, in report order, exit 1; no output and exit 0 for a clean program. The bytes of Go's encoder (compact, no spaces):

`{"code":"E0033","severity":"error","message":"...","file":"strict_eq.nra","range":{"start":[3,9],"end":[3,9]},"fixes":[]}`

- keys in this order: code, severity ("error"), message, file, range, fixes.
- `range.start` equals `range.end` (a diagnostic records a position); lines and columns are 1-based, `[0,0]` for an unknown position (a missing entry function).
- `file` is the base name for a single file and `src/x.nra` (relative to the package) for a package; `narra.toml` for manifest errors.
- `fixes` is `[]` always. Go's shape of a fix, for checker-fixes: `{"title": "...", "edits": [{"file": "...", "range": {"start": [l, c], "end": [l, c]}, "text": "..."}]}`. Add it in `writeDiagnosticJson` (diagnostics.nra) and give `CheckError` and `driver.Diagnostic` a fixes field; the `--json` branch is in `runCheck` (driver.nra).
- A problem that is not a diagnostic (code `""`) prints as text on stderr even with `--json`, as Go does.
- `narrac check` also accepts `-json`. `narrac query --json` shows the same code in its own (spaced) shape, `E0000` for code `""`.

### `narrac explain CODE`

Output equals Go's for all 277 distinct codes, a lowercase code, an unknown code (stderr + exit 1). No argument or two: usage on stderr (`narrac explain E0000`), exit 2.

### Checks (in this node's folder, throwaway scripts, not part of the Go suite)

- `check-codes.py`: fails when a report site has no code argument, passes `E0000` or a code that is not in the table, passes something that is not a code literal or a `code` variable, or passes a code whose table format does not fit the message literal. The compiler itself already rejects a call without the code parameter. 337 sites, 0 problems.
- `compare.py <bindir>` (a dir with `narra` and `narrac`): `check --json` on the 332 samples of `errors/`, `diagnostics/`, `packages/errors/` and `stage2/errors/`, and `explain` for 281 cases. Go's `fixes` are emptied before comparing; 65 Go diagnostics have fixes.

## Deviations

- `compiler/testdata/query/check_error/queries.txt` (outside owned paths): `E0000` became `E0119`, regenerated with `NARRA_UPDATE=1 go test -run TestSelfhostQuery`. One line.
- `narrac check` with `--json` compares equal to Go only after `fixes` is emptied (the brief says the same: fixes are the next node).
- Usage text says `narrac`, not `narra`.
- The lexer error type changed from `string` to `LexError`; callers in `fmtexprs.nra` (ignore the value), `main.nra`, `query.nra` and `parser.nra` were adapted. `main.nra` is outside the owned list (one call).

## Stubs left

- `fixes` is `[]` (checker-fixes).

## Gotchas

- Codes cannot come from the message text (113 of 410 are ambiguous): `E0409` (`%s: %v`) matches every message, and `%s` vs `%q` pairs such as E0204/E0206 and E0921/E0922 differ only in quoting. Look the Go site up.
- The duplicate filter keys on code, file, position and message, as Go's `report`.
- Go's JSON escapes only what `encoding/json` escapes; `quoteJson` (queryjson.nra) gives the same bytes on all current samples, but U+2028 and invalid UTF-8 are not handled.

## Proposed learnings

- Report a diagnostic with its code at the call site | the text alone maps to several codes | selfhost/src | yes, `roadmaps/selfhost/nodes/checker-codes/check-codes.py`

## Verification

- `cd compiler && go test -timeout 45m ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: 110 passed.
- `go test ./internal/narra -run 'TestSelfhostQuery|TestSelfhostFmt|TestSelfhostParser|TestSelfhostLexer|TestSelfhostModules'`: 3406 passed after the golden update.
- `python3 roadmaps/selfhost/nodes/checker-codes/compare.py $BIN`: check --json 332/332 equal (fixes aside), explain 281/281.
- `python3 roadmaps/selfhost/nodes/checker-codes/check-codes.py`: 337 sites, 0 problems.
- `narrac fmt --check selfhost/src`: clean.

## Size

- Estimated M, actual M. A codemod placed most codes; the format table matched about 230 of 256 checker sites uniquely.

Note: `check-codes.py` was deleted. The `codes` runner suite (`selfhost/tests/src/codes.nra`) replaced it.
