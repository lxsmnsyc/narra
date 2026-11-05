---
name: query-output
version: 1
status: draft
defined_by: query-cli
---

# Contract: query-output

## Purpose

The command line and the `--json` output of `narrac query`. The language service (hover, definition, references), lint and the agent skills read it. A consumer parses the JSON form only; the text form is for people and may change.

## Interface

Command line:

```
narrac query [--package DIR] [--json] type-of FILE:LINE:COL
narrac query [--package DIR] [--json] def NAME
narrac query [--package DIR] [--json] refs NAME
narrac query [--package DIR] [--json] callers FUNCTION
narrac query [--package DIR] [--json] interface MODULE
narrac query [--package DIR] [--json] recursion MODULE
```

- `FILE` is relative to the package directory, as in diagnostics (`src/app/config.nra`). `LINE` and `COL` start at 1 and count characters.
- `NAME` and `FUNCTION` are qualified (`app/text.helper`) or bare (`helper`). `MODULE` is a module path (`app/text`).

Shared JSON shapes, the same as `narra check --json` (SPEC section 14):

```json
{"file": "src/text.nra", "range": {"start": [3, 10], "end": [3, 16]}}
```

A location. `end` is the position after the last character.

`--json` prints exactly one JSON object followed by a newline:

```json
{"query": "type-of", "location": {...}, "type": "Array<string>", "effects": ["io", "alloc"]}
{"query": "def", "matches": [{"name": "app/text.helper", "kind": "function", "location": {...}}]}
{"query": "refs", "name": "app/text.helper", "uses": [{"location": {...}, "in": "app/main.run"}]}
{"query": "callers", "name": "app/text.helper", "callers": [{"name": "app/main.run", "location": {...}}]}
{"query": "interface", "module": "app/text", "declarations": [{"name": "helper", "kind": "function", "signature": "export function helper(text: string): int uses alloc", "location": {...}}]}
{"query": "recursion", "module": "app/text", "groups": [["app/text.walk"], ["app/text.even", "app/text.odd"]]}
```

- `kind` is one of `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct`, `local`, `parameter`.
- `effects` is sorted and empty for a pure expression.
- Lists are in file then position order; recursion groups are sorted by their first name.

Errors with `--json`: exit 1 and one object `{"error": "<message>"}`, or for load and check errors `{"errors": [<diagnostic objects as narra check --json>]}`.

## Behavior

- Exit 0 on an answer, including an empty list. Exit 1 on an unknown name, a position with no expression, or a package that does not check. Exit 2 on a usage error.
- Calls through function values are not in `callers` or `recursion`.
- The output does not depend on the order of files on disk.

## Stub

None. Consumers can build against the JSON examples above.

## Contract tests

`compiler/testdata/query/` and `TestSelfhostQuery` (`compiler/internal/narra/selfhost_query_test.go`), added by query-cli.

## Changes

- v1: initial draft, written while refining query-cli.
