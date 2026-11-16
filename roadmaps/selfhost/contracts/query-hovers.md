---
name: query-hovers
status: frozen
version: 1
defined_by: query-hovers
---
# query-hovers

## Purpose

`narrac query hovers` gives, for one source file, the hover text of every identifier and every diagnostic, from one process. A build-time twoslash for Narra reads it.

## Interface

```
narrac query [--package DIR] --json hovers FILE
```

`FILE` is relative to the package directory, as in diagnostics (`src/main.nra`). It prints exactly one JSON object and a newline, on one line:

```json
{"query": "hovers", "file": "src/main.nra", "hovers": [{"line": 3, "column": 7, "length": 5, "text": "int"}], "diagnostics": [{"line": 5, "column": 3, "length": 4, "code": "E0100", "message": "..."}]}
```

- `line` and `column` start at 1. A column counts Unicode characters (code points) from the start of the line, not bytes and not UTF-16 units. `length` counts the same unit. A consumer that indexes JavaScript strings converts code points to UTF-16 units itself.
- `hovers` lists the identifiers the checker resolved: names of declarations, parameters, locals, pattern bindings, fields, called functions and constants. Each hover is one token on one line. The list is sorted by line then column, with one entry per place.
- `text` is the text the language server shows on hover, without the markdown fence: the type, then a line `uses io, alloc` when the name's evaluation has effects. A function name shows its signature. Lines are separated by `\n`.
- `diagnostics` lists the errors reported in `FILE`, in the order `narrac check` prints them. `code` is the diagnostic code. `length` is the length of the token that starts at the position, or 1 when no token starts there.
- A file with errors still answers: `diagnostics` lists them and `hovers` covers what the checker recorded. When effects cannot be computed because the package has errors, a hover leaves out the `uses` line of expressions. When the package does not load (a parse error), `hovers` is empty and `diagnostics` has the errors of `FILE`.
- Exit code 0 when the command answered, including a file with errors. Exit 1 with `{"error": "<message>"}` (or `{"errors": [...]}` in the shape of `type-of` when the package fails to load with no error in `FILE`) when `FILE` is not a module of the package. Exit 2 for a usage error.
- Without `--json` the command prints one line per hover, `FILE:LINE:COL text` with newlines in the text written as ` | `, then one line per diagnostic, `FILE:LINE:COL CODE message`. The text form is for people.
