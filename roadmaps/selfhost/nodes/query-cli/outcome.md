# Outcome: query-cli

## Built

`narrac query` in the self-hosted compiler, for SPEC.md section 14.

- `selfhost/src/query.nra`: `runQuery(arguments)`, dispatched from `run` in `selfhost/src/main.nra` (`narrac query ...`). It parses `[--package DIR] [--json] <command> <operand>`, loads and checks the package as `typesDump` does (`Modules.load`, `Body.checkPackage`, `Effects.checkEffects`), indexes the package's own modules (not `std/` source modules) sorted by file, then answers `type-of`, `def`, `refs`, `callers`, `interface` or `recursion`.
- `selfhost/src/querysites.nra`: the position index. `collectSites(file, tokens)` lists every `Site` (kind `ExpressionSite`, `CalleeSite`, `LocalSite`, `VariantPatternSite`, `DeclarationSite`) with its start and end, node before children. `isBefore(left, right)` orders positions.
- `selfhost/src/querysignature.nra`: `writeType`, `writeFunction`, `writeStruct`, `writeEnum`, `writeTrait`, `writeDistinct`: one-line signatures from the syntax.
- `selfhost/src/querygraph.nra`: `findCycles(edges)` (iterative Tarjan; components with more than one node or a self edge) and `sortStrings`.
- `selfhost/src/queryjson.nra`: `Location`, `quoteJson`, `writeStringList`, `writeLocation`, `writeLocationText`, `writeDiagnostic`.
- `selfhost/src/effects.nra` (append-only): `EffectState.calls: Array<CallEdge>` (`CallEdge { body, callee: Ref, pos }`, one per direct call of a declared or implement block function, recorded in the reporting walk only), `findRefEffects(checker, state, ref)` and `findExpressionEffects(mutable checker, mutable state, bodyIndex, expression)`. The effect sets and errors are unchanged.
- Tests: `compiler/internal/narra/selfhost_query_test.go` (`TestSelfhostQuery`) and the cases `compiler/testdata/query/basic`, `recursion`, `check_error`. Each has `narra.toml`, `src/` and `queries.txt`: `> <arguments>` lines, each followed by the expected stdout then stderr, with `# exit N` first when the exit code is not 0. `NARRA_UPDATE=1` rewrites the expected output.
- `compiler/README.md`: new "Queries" section.

Final shapes for `contracts/query-output.md` (the orchestrator freezes it; the draft's JSON shapes hold as written):

- Command line as in the draft. Options come before the command. `FILE` is the module's file as in diagnostics (`src/text.nra`).
- `type-of`: `{"query": "type-of", "location": L, "type": "...", "effects": [...]}`. The location is the range of the token or name found. The innermost site that covers the position and has a recorded type wins: an expression (the type of its token: `(` of a call, the operator of a binary expression, the name of a field read), a parameter, `const`/`let`/loop variable or pattern binding name, a function name (declaration or call: `(A, B) => R` plus the function's effects as callers see them, effect variables by name) or a constant name. Other expressions get the effects of evaluating them. A type name, keyword, punctuation that starts no expression, or whitespace is the error `no expression at FILE:LINE:COL`.
- `def`: `{"query": "def", "matches": [{"name", "kind", "location"}]}`, every match in file and position order. Kinds produced: `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct` (not `local` or `parameter`). A name matches the qualified name (`app/text.helper`, `app/text.Shape.Circle`) or a bare one (`helper`, `Shape.Circle`, `Circle`).
- `refs`: `{"query": "refs", "name": Q, "uses": [{"location", "in"}]}`. `in` names the body: `app/main.run`, `app/text.Point.Show.show` (implement block function), `app/main.test "name"`, `app/text.LIMIT` (constant value), `app/text.Struct.field` (field default).
- `callers`: `{"query": "callers", "name": Q, "callers": [{"name", "location"}]}`, one entry per call site (a function that calls twice is listed twice); the location is the callee's name.
- `interface`: `{"query": "interface", "module": M, "declarations": [{"name", "kind", "signature", "location"}]}`, `name` bare, signatures like `export function helper(text: string): int uses alloc`, `export struct Point { x: int, y: int }`, `export enum Shape derives Show { Circle(radius: int), Empty }`, `export trait Named { function name(value: Self): string uses alloc }`, `export const LIMIT: int`, `export distinct type UserId = int` (opaque types hide fields and base).
- `recursion`: `{"query": "recursion", "module": M, "groups": [[...]]}` as in the draft.
- Errors: `{"error": "..."}` on stdout, exit 1. Load and check errors: `{"errors": [{"code": "E0000", "severity": "error", "message", "file", "range", "fixes": []}]}`, exit 1. Usage: text on stderr, exit 2, also with `--json`.

## Deviations

- Diagnostic objects use code `E0000`: the self-hosted checker has no diagnostic codes yet.
- `refs` and `callers` with a bare name that finds more than one declaration exit 1 with `"helper" names app/text.helper, app/util.helper; write a qualified name`. `def` lists them all (open question 2).
- `refs` covers functions, constants and enum variants only. A struct, enum, trait or distinct name exits 1 (`refs finds uses of functions, constants and enum variants; app/text.Point is a struct`).
- `type-of` on a type annotation or keyword is an error (open question 3, as proposed). The command stays `narrac query` (open question 1).
- `callers` includes test bodies; `recursion` counts only function and implement block function bodies.
- Text forms (for people): `type-of` prints `<type>` or `<type> uses a, b`; `def` prints `<name> <kind> <file>:<line>:<col>`; `refs` and `callers` print `<file>:<line>:<col> <body name>`, and `callers` ends with `note: calls through function values are not counted`; `interface` prints one signature per line; `recursion` prints one group per line, names joined by `, `. An empty list prints `(none)`.

## Stubs left

- No codes in query diagnostics until the self-hosted checker gets them.
- No `refs` for type names (needs type-expression refs in the checker); no `def` for locals, trait functions or implement block functions.

## Gotchas

- A named call records its ref on the `CallExpr` node, not on the callee `Ident`; the callee Ident has no type. Read a call's target from the call's id.
- `PatCtor` and declarations record the position of their first token (`Shape` of `Shape.Circle`, the `function` keyword); the name comes from the tokens (`findNameStart`).
- `EffectState.calls` is filled only by the last (reporting) walk of `checkEffects`. It is empty when the signature pass reported errors.
- `uses`, `from` and `match` are keywords and cannot name a local.
- Formatting a test package moves positions: rerun `NARRA_UPDATE=1 go test ./internal/narra -run TestSelfhostQuery` and check the diff.

## Proposed learnings

- Read checker results only from bodies without type arguments when reporting on source | instance bodies repeat the same node ids with substituted types | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostQuery|TestSelfhostBootstrap|TestSourcesFormatted'`: pass (145 s).
- `cd compiler && go vet ./... && go test -timeout 20m ./...`: pass (1191 s on a machine shared with other worktrees, close to the 20-minute timeout). A test block was added to `testdata/query/basic/src/util.nra` afterwards; the targeted run above covers it.
- By hand on `selfhost/`: `recursion narrac/check`, `refs narrac/check.typeString`, `callers findBodySlot`, `type-of src/check.nra:673:17`, about 0.5 s each (release build).

## Size

- Estimated L, actual L. About 2000 lines of Narra in 5 files (the brief guessed 800 to 1000), mostly from the position index and the formatter's line breaks.
