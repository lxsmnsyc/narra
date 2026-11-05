# Outcome: query-cli

## Built

`narrac query` in the self-hosted compiler, for SPEC.md section 14.

- `selfhost/src/query.nra`: `runQuery(arguments)`, dispatched from `run` in `selfhost/src/main.nra` (`narrac query ...`). It parses `[--package DIR] [--json] <command> <operand>`, loads and checks the package as `typesDump` does (`Modules.load`, `Body.checkPackage`, `Effects.checkEffects`), indexes the package's own modules (not `std/` source modules) sorted by file, then answers `type-of`, `def`, `refs`, `callers`, `interface` or `recursion`.
- `selfhost/src/querysites.nra`: the position index. `collectSites(file, tokens)` lists every `Site` (kind `ExpressionSite`, `CalleeSite`, `LocalSite`, `VariantPatternSite`, `DeclarationSite`) with its start and end, node before children. `isBefore(left, right)` orders positions.
- `selfhost/src/querysignature.nra`: `writeType`, `writeFunction`, `writeStruct`, `writeEnum`, `writeTrait`, `writeDistinct`: one-line signatures from the syntax.
- `selfhost/src/querygraph.nra`: `findCycles(edges)` (iterative Tarjan; components with more than one node or a self edge) and `sortStrings`.
- `selfhost/src/queryjson.nra`: `Location`, `quoteJson`, `writeStringList`, `writeLocation`, `writeLocationText`, `writeDiagnostic`.
- `selfhost/src/effects.nra` (append-only): `EffectState.calls: Array<CallEdge>` (`CallEdge { body, callId, callee: Ref, pos }`, one per direct call of a declared or implement block function, recorded in the reporting walk only), `findRefEffects(checker, state, ref)`, `findArrowEffects(state, nodeId)` and `findExpressionEffects(mutable checker, mutable state, bodyIndex, expression)`. The effect sets and errors are unchanged.
- Tests: `compiler/internal/narra/selfhost_query_test.go` (`TestSelfhostQuery`) and the cases `compiler/testdata/query/basic`, `recursion`, `check_error`. Each has `narra.toml`, `src/` and `queries.txt`: `> <arguments>` lines, each followed by `# exit N` when the exit code is not 0, the expected stdout, and `# stderr` with the expected stderr when there is any. `NARRA_UPDATE=1` rewrites the expected output.
- `compiler/README.md`: new "Queries" section.

Final shapes for `contracts/query-output.md`. The draft's JSON shapes hold as written; the exact text changes are in `contracts/query-output.change-request.md`, which the orchestrator takes to the maintainer.

- Command line as in the draft. Options come before the command. `FILE` must be the module's file exactly as diagnostics print it (`src/text.nra`).
- `type-of`: `{"query": "type-of", "location": L, "type": "...", "effects": [...]}`. The location is the range of the token or name found. The innermost site that covers the position and has a recorded type wins: an expression (the type of its token: `(` of a call, the operator of a binary expression, the name of a field read), a parameter, `const`/`let`/loop variable or pattern binding name, a function name or a constant name.
  - `type` is the type of the value. Function types carry their `uses` list: from the declaration for a named function (declared effects of an exported one, inferred ones of a private one), from the annotation for an annotated parameter or local, from the body for an arrow function. An unannotated local holding a function shows none.
  - `effects` is what evaluating the expression does. A name (local or declared function) has none; a call has the call's effects. With `return f(text)`: `f` gives `(string) => int uses alloc` with `effects: []`, the `(` gives `int` with `["alloc"]`. A builtin or constructor name in a call answers for the call, because the checker records nothing for that name.
  - A position inside a template string with `${...}` parts is the error `positions inside template strings are not supported yet`; the backtick answers for the template. A type name, keyword, punctuation that starts no expression, or whitespace is the error `no expression at FILE:LINE:COL`.
- `def`: `{"query": "def", "matches": [{"name", "kind", "location"}]}`, every match in file and position order. Kinds produced: `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct` (not `local` or `parameter`). A name matches the qualified name (`app/text.helper`, `app/text.Shape.Circle`) or a bare one (`helper`, `Shape.Circle`, `Circle`).
- `refs`: `{"query": "refs", "name": Q, "uses": [{"location", "in"}]}`. A use inside a template string has the whole template as its location and `"inTemplate": true` (also on `callers` entries); the field is absent otherwise. `in` names the body: `app/main.run`, `app/text.Point.Show.show` (implement block function), `app/main.test "name"`, `app/text.LIMIT` (constant value), `app/text.Struct.field` (field default).
- `callers`: `{"query": "callers", "name": Q, "callers": [{"name", "location"}]}`, one entry per call site (a function that calls twice is listed twice); the location is the callee's name.
- `interface`: `{"query": "interface", "module": M, "declarations": [{"name", "kind", "signature", "location"}]}`, `name` bare, signatures like `export function helper(text: string): int uses alloc`, `export struct Point { x: int, y: int }`, `export enum Shape derives Show { Circle(radius: int), Empty }`, `export trait Named { function name(value: Self): string uses alloc }`, `export const LIMIT: int`, `export distinct type UserId = int` (opaque types hide fields and base).
- `recursion`: `{"query": "recursion", "module": M, "groups": [[...]]}` as in the draft.
- Errors: `{"error": "..."}` on stdout, exit 1. Load and check errors: `{"errors": [{"code": "E0000", "severity": "error", "message", "file", "range", "fixes": []}]}`, exit 1; the range is empty at the error's position, and a load error with no file (a missing package directory) leaves out `file` and `range`. Usage: text on stderr, exit 2, also with `--json`.

## Deviations

- Diagnostic objects use code `E0000`: the self-hosted checker has no diagnostic codes yet.
- `refs` and `callers` with a bare name that finds more than one declaration exit 1 with `"helper" names app/text.helper, app/util.helper; write a qualified name`. `def` lists them all (open question 2).
- `refs` covers functions, constants and enum variants only. A struct, enum, trait or distinct name exits 1 (`refs finds uses of functions, constants and enum variants; app/text.Point is a struct`).
- `type-of` on a type annotation or keyword is an error (open question 3, as proposed). The command stays `narrac query` (open question 1).
- `callers` includes test bodies; `recursion` counts only function and implement block function bodies.
- Text forms (for people): `type-of` prints `<type>`, then `effects: a, b` on a second line when there are any; `def` prints `<name> <kind> <file>:<line>:<col>`; `refs` and `callers` print `<file>:<line>:<col> <body name>`, with ` (in a template string)` after a use inside one, and `callers` ends with `note: calls through function values are not counted`; `interface` prints one signature per line; `recursion` prints one group per line, names joined by `, `. An empty list prints `(none)`.

## Stubs left

- No codes in query diagnostics until the self-hosted checker gets them.
- Template strings: the parser gives every node inside `${...}` the template's position. Until that is fixed (the orchestrator put it in the inbox), `type-of` there is an error and `refs`/`callers` report the whole template with `inTemplate`. When positions exist, drop the `inTemplate` override in `addSite` (`querysites.nra`) and the error in `queryTypeOf`.
- No `uses` in the type of an unannotated local that holds a function: the checker's `TFunc` has no effects.
- No `refs` for type names (needs type-expression refs in the checker); no `def` for locals, trait functions or implement block functions.

## Gotchas

- A named call records its ref on the `CallExpr` node, not on the callee `Ident`; the callee Ident has no type. Read a call's target from the call's id.
- `PatCtor` and declarations record the position of their first token (`Shape` of `Shape.Circle`, the `function` keyword); the name comes from the tokens (`findNameStart`).
- `EffectState.calls` is filled only by the last (reporting) walk of `checkEffects`. It is empty when the signature pass reported errors.
- `uses`, `from` and `match` are keywords and cannot name a local.
- A named call records the callee's ref on the call node too, so `describeCallable` must run only for name sites, or the `(` of a call answers with the function type.
- Formatting a test package moves positions: rerun `NARRA_UPDATE=1 go test ./internal/narra -run TestSelfhostQuery` and check the diff.

## Proposed learnings

- Read checker results only from bodies without type arguments when reporting on source | instance bodies repeat the same node ids with substituted types | selfhost/src | no

## Verification

Review round 1 (fresh review of PR #88, four should-fix findings and nits 5 to 8), fixed in the second commit:
- Template positions: `inTemplate` sites, the type-of error, `inTemplate` in refs and callers, cases at `basic/src/main.nra:19`.
- Function types keep `uses` (declaration, annotation, arrow body); cases at `main.nra:5:16`, `6:10`, `12:20`, `12:25`.
- One rule for `type` and `effects`; a local callee (`6:10`, `6:11`) and a declared callee (`11:27`, `11:33`).
- `queries.txt` checks stdout and stderr apart (`# stderr`).
- A load error with no file leaves out `file` and `range` (`check_error`: `--package missing`).
- Nits 6 to 8 and the decisions above are in `contracts/query-output.change-request.md`.


- After round 1: `go test -count=1 -timeout 40m ./internal/narra -run 'TestSelfhostQuery|TestSourcesFormatted|TestSelfhostBootstrap|TestSelfhostEmit'`: pass (75 s).
- After round 1: `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass (791 s).
- Before round 1: the full check passed in 1191 s on a machine shared with other worktrees.
- By hand on `selfhost/`: `recursion narrac/check`, `refs narrac/check.typeString`, `callers findBodySlot`, `type-of src/check.nra:673:17`, about 0.5 s each (release build).

## Size

- Estimated L, actual L. About 2000 lines of Narra in 5 files (the brief guessed 800 to 1000), mostly from the position index and the formatter's line breaks.
