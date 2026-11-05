# Log: query-cli

- 2026-11-05: started on origin/main 8a8059c, branch feat/query-cli. No query code existed.
- The effect pass keeps no call graph, only effect sets per function. Added `EffectState.calls` (`CallEdge`), filled in `walkCall` during the one reporting walk only, so what the pass computes is unchanged. Also exported two readers: `findRefEffects` and `findExpressionEffects` (walks one expression again with reporting off).
- Position index: every expression position is a token (FieldExpr at the name, CallExpr at `(`, BinaryExpr at the operator), so a site's range is the lexer token at its position. Declarations record their keyword; the name is the first identifier token with that text a few tokens later. Inside a template string no token starts at a part, so an identifier spans its name and a string piece its `end`.
- Innermost: sites are listed node before children; type-of tries the covering sites from last to first and takes the first the checker recorded something for. A callee name gets its own site after the callee Ident, because a named call records its ref on the call node and nothing on the Ident.
- Open bodies only (no type arguments): a generic function shows `T`. Each module's open bodies have disjoint id ranges, so the first body whose slot covers a node id is the one.
- Narra keywords that bit as local names: `uses`, `from`, `match`.
- Decided: refs and callers with a bare name that finds several declarations is an error listing the qualified names (def lists them all). callers lists each call site, not each function once; test bodies count as callers (`app/main.test "name"`). recursion counts only func and impl bodies.
- Decided: refs covers functions, constants and enum variants; type names in annotations are not in the refs tables.
- Diagnostic JSON uses code E0000: the self-hosted checker has no codes.
- `narrac fmt` collapsed a match in a test package onto one line; reran NARRA_UPDATE=1 after formatting.
- On selfhost itself: `refs narrac/check.typeString` and `callers findBodySlot` take about 0.5 s each (release build).
- Review round 1: the parser puts every node inside `${...}` at the template's position, so sites inside templates are marked `inTemplate` and span the template; type-of there errors, refs/callers flag them. `CallEdge` gained `callId` so callers can find the call's site.
- type-of rule changed: names evaluate with no effects, function types carry `uses` (from the declaration, the annotation via `Signature.writeType`, or the arrow body via `findArrowEffects`). The text form now puts effects on a second line (`effects: ...`), because `<type> uses ...` became ambiguous.
- The call node holds the callee's ref too, so the `(` of `helper("ab")` answered with the function type until describeCallable was limited to name sites.
- `queries.txt` now separates stderr with a `# stderr` line.
