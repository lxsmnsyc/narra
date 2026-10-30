# Outcome: json-diagnostics

## Built

All in `compiler/internal/narra/` unless noted.

- `diagnostic_codes.go`: `diagnosticCodes`, the table of code and message format, and `codeFor(format)`. Every `errorf`, `fail`, `failFix`, `errorfFix` and `typeErrorf` call passes a literal format, and the format finds the code. A format without a code gives `E0000`. `explanations` holds long texts for `narra explain`.
- `diagnostic_json.go`: `CheckDiagnostics(path)`, `Diagnostic.JSON()`, `ApplyFix(src, fix)`, `Explain(code)`, `CheckCommand`, `ExplainCommand`, `PrintJSON`.
- `diagnostic_fixes.go`: `c.errorfFix`, `concatFix`, `pushFix`.
- `cmd/narra/main.go`: two cases, `check` and `explain`, and two usage lines.
- Lexer and parser errors are now `Diagnostic` values with code and fixes (`parseError{d}`, `p.failFix`, `lx.errorf` returns a `Diagnostic`). Text output is unchanged.
- CLI: `narra check [--json] file|dir` prints `file:line:col: message` lines, or one JSON object per line (SPEC section 14 shape: code, severity, message, file, range, fixes with edits). Exit 1 when there are errors. `narra explain E0003` prints the code, its message format and the long text if any.
- Tests (`diagnostic_test.go`): every format in the source has a code; every `testdata/errors/*.nra` and `testdata/packages/errors/*` gets a coded diagnostic; each TypeScript habit case has a fix that, applied, makes the file compile (`semicolon`, `strict_eq`, `method_call`, `string_plus` in `testdata/errors/`, and `ts_null`, `ts_array_type`, `ts_import_default`, `ts_concat_chain`, `ts_push` in `testdata/diagnostics/`, whose first line `// code: E0NNN` names the expected code).

Fixes:

| Habit | Code | Fix |
| --- | --- | --- |
| `;` | E0003 | delete it |
| `import X from` | E0014 | insert `* as ` |
| `T[]` | E0020 | `Array<T>` (two edits) |
| `===`, `!==` | E0033 | `==`, `!=` |
| `null`, `undefined` | E0038 | `None` |
| `a + b` on strings | E0216 | template string; only when every operand is a string name or literal |
| `xs.push(v)` | E0309 | `Array.push(mutable xs, v)`; only for a name on one line without extra spaces |

## Code ranges

Next free code in each range is the last one plus one. Reserve new codes there.

| Range | Area | Assigned | Next free |
| --- | --- | --- | --- |
| E0001-E0099 | lexer and parser | E0001-E0042 | E0043 |
| E0100-E0199 | declarations and types | E0100-E0119 | E0120 |
| E0200-E0299 | expressions and statements | E0200-E0234 | E0235 |
| E0300-E0399 | calls and constructors | E0300-E0319 | E0320 |
| E0400-E0499 | modules, packages, narra.toml | E0400-E0420 | E0421 |
| E0500-E0599 | traits | E0500-E0528 | E0529 |
| E0600-E0699 | closures | E0600-E0611 | E0612 |
| E0700-E0799 | patterns and match | E0700-E0719 | E0720 |
| E0800-E0899 | effects | E0800-E0803 | E0804 |
| E0900-E0999 | std library, maps | E0900-E0906 | E0907 |

The numbers inside a range follow the order of the error sites at the time of assignment. The table in `diagnostic_codes.go` is the reference (code to message format). Maps errors are E0900-E0906 (`flushMapKeys` message is in the table), the `StringBuilder` unknown type is the ordinary unknown-type code, effects errors are E0800-E0803.

## Rule for adding a code

1. Write the error through `c.errorf`, `c.errorfFix`, `p.fail` or `p.failFix` with a literal format string.
2. Append `{"E0NNN", "<the exact format>"}` to `diagnosticCodes` with the next free number of the range for the area. Never renumber or reuse a code.
3. Add a test in `testdata/errors/` (first line `// error: <text>`). `TestEveryFormatHasCode` fails when the format has no code, `TestErrorCodes` when a test file gets none.
4. For a TypeScript habit, add an exact fix and a case in `testdata/diagnostics/` (`// code: E0NNN` on the first line).
5. Changing the wording of a format string needs the same change in the table. Keep the code.

## Deviations

- The code lookup is by format string instead of an argument at each of the ~230 error sites. This keeps hot files (`check_expr.go`, `parser.go`) almost untouched. Two sites share one code if they share a format.
- Diagnostics have a position, not an extent. JSON `range.end` equals `range.start` for the diagnostic; fix edits have real ranges.
- Edits outside the brief's paths, all a few lines: `lexer.go` (`Token.End`, coded and fixable errors), `ast.go` (`StringLit.End`), `modules.go`, `modules_toml.go`, `std_map.go` (`codeFor` instead of `E0000`), `check.go` (`errorf` sets the code), `compiler/README.md` (usage lines).
- The `Diagnostic.File` for a single file is its base name, as the driver sets it today.

## Stubs left

- No fix for `'abc'` strings (D42 example). The `T[]` fix covers a simple name or generic type only; `Array<int>[]` is not handled.
- Only `E0003`, `E0014`, `E0020`, `E0033`, `E0038`, `E0216`, `E0309` have long `explain` texts. The rest print their message format.
- The lexer and parser still stop at the first error.

## Gotchas

- A message format must be a string literal at the call site, or it gets `E0000`. `typeErrorf` and `errorfFix` pass a format on; the test only checks literals.
- In a string `+` chain, only the outermost diagnostic keeps the fix (the inner ones would leave an error).
- `ApplyFix` counts columns in runes, like the lexer.

## Proposed learnings

- Give each new error format a code in `diagnostic_codes.go` | the code is looked up from the format | compiler/internal/narra | yes, `TestEveryFormatHasCode`
- Pass error messages as literal formats | non-literal formats get no code | compiler/internal/narra | partly, the same test

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- By hand: `narra check --json` on `semicolon`, `strict_eq`, `method_call`, `string_plus` prints coded diagnostics with fixes.

## Size

- Estimated M, actual M.
