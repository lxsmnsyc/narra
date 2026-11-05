---
name: query-output
version: 2
status: frozen
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
- `FILE` must be the package-relative path exactly as diagnostics print it (`src/text.nra`). Other spellings (`./src/text.nra`, an absolute path) find no module.

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

- `kind` is one of `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct`. `local` and `parameter` are reserved for a later position-based `def`; no command prints them yet.
- `type-of` reports the innermost expression, local name (parameter, `const`, `let`, loop variable, pattern binding), or function or constant name at the position. `location` is the range of the token or name it found.
- `type` is the type of the value at the position. In a function type, `uses` comes from the declaration of a named function (the declared effects of an exported function, the inferred ones of a private one), from the annotation of a parameter or local, or from the body of an arrow function. Any other function-typed value (an unannotated local, a struct field, an array element, a call's result) shows no `uses`, and calling it reports only the effects the effect pass records.
- `effects` is what evaluating the expression at the position does, sorted. A name (a local or a declared function) evaluates with no effects. A call has its call effects. On the name of a called local or declared function, the answer is about the name: with `return f(text)`, `type-of` on `f` gives `(string) => int uses alloc` with `effects: []`, and on the `(` it gives `int` with the call's effects.
- Provisional: a builtin or constructor name in a call answers for the call (its type and effects), because the checker records nothing for that name. A later version may answer with the name's function type instead. Consumers must not rely on the call's answer there.
- A position inside a template string that has `${...}` parts is an error: `positions inside template strings are not supported yet`. The template's opening backtick answers for the whole template.
- Lists are in file then position order; recursion groups are sorted by their first name.
- A `refs` use or a `callers` entry inside a template string has the template's whole range as its location and an extra field `"inTemplate": true`. The field is absent otherwise.
- `refs` and `callers` need a name that finds one declaration. A bare name that finds several exits 1 with `"<name>" names <qualified>, <qualified>; write a qualified name`. `def` lists every match.
- `refs` finds functions, constants and enum variants. Any other kind exits 1.
- `callers` has one entry per call site: a function that calls twice is listed twice. The location is the callee's name in the call, or the whole template when `inTemplate` is set. Test bodies count as callers, named `<module>.test "<test name>"`. `recursion` counts only functions and implement block functions.
- `interface` prints each signature from the source as written: type names keep the spelling they have there (aliases and qualified names included), field defaults and bodies are left out, opaque types hide their fields or base, and implement blocks show only through a type's `implements` list.
- Nested function types with `uses` are ambiguous as text (in `(() => int uses io) => int uses alloc`, a reader cannot tell from the text alone which function each `uses` belongs to). A consumer that needs the structure must not parse `type` or `signature`.

Errors with `--json`: exit 1 and one object `{"error": "<message>"}`, or for load and check errors `{"errors": [<diagnostic objects as narra check --json>]}`. Until the self-hosted checker has diagnostic codes, `code` is the placeholder `E0000` and `range` is empty (start equals end) at the error's position. A load error with no file, such as a package directory that does not exist, leaves out `file` and `range`. A usage error prints the usage text on stderr and exits 2, also with `--json`.

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
- v2 (from the query-cli change request): the type-of rule for `type` and `effects`, `inTemplate`, diagnostics without file and range, the `E0000` and empty-range placeholders, `local` and `parameter` reserved, the ambiguity, call site and test caller rules, and notes on `interface` text, nested `uses` text and the FILE spelling.
