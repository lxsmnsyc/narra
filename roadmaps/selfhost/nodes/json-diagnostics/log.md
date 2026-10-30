# Log: json-diagnostics

- Started from origin/main after modules, traits, closures, patterns, maps, strings-std and effects merged.
- Codes are keyed by message format string (`diagnostic_codes.go`), so error sites in hot files stay untouched. `codeFor(format)` is called where a Diagnostic is built.
- Generated the first table with a script that walks the Go AST for `errorf`, `fail`, `failFix`, `errorfFix`, `typeErrorf` calls; grouped by source file into ranges of 100.
- Lexer and parser errors now carry a `Diagnostic` (code and fixes) instead of text; `syntaxDiagnostic` passes it through.
- Added `Token.End` and `StringLit.End` so the string concatenation fix knows where a literal ends.
- `narra check` needed the entry-function check that only `CompileBinary` had, so `CheckDiagnostics` repeats it (E0419, E0420).
- Single quotes to double quotes (D42 example) has no fix yet.
