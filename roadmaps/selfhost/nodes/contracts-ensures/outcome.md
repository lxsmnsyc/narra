# Outcome: contracts-ensures

## Built

- Syntax: `ensures <pattern>` lines after the `requires` lines of a function with a body (top-level and `implement` members). AST: `FuncDecl.ensures: Array<Pattern>` (default `[]`) and the new pattern kind `Pattern.PatCompare(id, pos, operator, bound: Expr)` for `< c`, `<= c`, `> c`, `>= c` (`ast.nra`). Parser: `parseEnsuresLines` and `isContractLine(parser, word)` (shared with `requires`) in `parser.nra`. The parser makes a comparison pattern only while `parser.contractPattern` is set, so in `match` or `is` it is a plain parse error ("expected a name").
- Checking in `selfhost/src/contracts.nra`:
  - Declaration phase: `collectModuleRequirements` also stores `checker.resultContracts[funcIndex] = ResultContract { patterns, bounds }` (`check.nra`). Each comparison bound is folded with the same `foldConstant` as `requires` (it takes the parameter names, so a parameter named like a module constant or an import alias is not a constant).
  - Body: `checkEnsures(checker, walker, funcDecl, isInstance)`, called by `checkFunctionBody` (`body.nra`) after `checkRequires`. On the declaration's own check it resolves each pattern against the result type (the element type for a generator) through the new `Walker.resolvePattern`. It reports E0987 for a function that returns nothing, a pattern that binds a name, a comparison on a non-integer place (`match.nra` `resolveNode`), and a bound that is not a constant or does not fit the place's type (u64 bounds from 2^63 up included). Without errors, the contract goes in `checker.bodyContracts[body]`.
  - Proof: `proveResult(checker, value)` from `checkReturn` (not in arrows or block expressions) and from `checkYield`; `rejectDelegatedResults` makes `yield*` E0988. `isValueProven` proves structurally: constructor calls and bare constructors with the same `Ref.Constructor`, tuple literals, struct literal fields written in place (not from a spread or default), non-integer literals written the same, integer places from `findOperandFact` and `isComparisonProven` (constants, narrowed names, `requires` facts). An or-pattern holds when one alternative does. Anything else is E0988, which prints the pattern.
  - Caller facts: `seedResultLocal` (`const n = f()`, `let n = f()`; `checkVariable` in `stmts.nra`), `seedPatternFacts` (`if (f() is P)` in `checkIf`, `match` arms in `checkMatch`, destructuring in `checkDestructure`), `seedLoopFacts` (`for (const n of g())` for a generator with `ensures`). A `const r = f()` whose type is not a contents type stores the key `holds|<local>`, so `r is P` and `match (r)` learn the same patterns. A binding learns the join of what every alternative compatible with the caller's pattern says about its place (alternatives of another constructor are dropped); one alternative that says nothing gives no fact.
- `operands.nra`: new `refineBounds` (pure; `refineFact` uses it; `> max` and `< min` now add nothing instead of overflowing), `createUnknownFact`, `joinFacts`, `narrowFact`.
- fmt: each `ensures` line is printed under the signature like `requires`, as its own trivia node (key `n<pattern id>`); `PatCompare` prints `>= bound` (`fmtexprs.nra`). astdump: an `.ensures` slot and `pat-compare op=">="`.
- Codes: E0987 (5 rows), E0988 (2 rows) in `diagnostics.nra`.
- Docs: SPEC.md section 13 (ensures paragraph with a checked example), D95 "Built" amendments for ensures, D98 note, compiler/README.md line.
- Tests: `testdata/contracts_ensures.nra` (every D95 form proven: `> 0`, `>= 0` with `<= LIMIT`, `Some(>= 0) | None`, `Ok(> 0) | Err(_)`, `#[>= 0, _]`, `{ width: > 0 }` with `requires`, `"GET" | "POST"`, a generator; callers using the facts through `const`, `is`, a `const` holder, `match`, tuple and struct destructuring, `for` over a generator, and a function whose return is proven from a callee's pattern), `testdata/errors/contracts_ensures_unproven` (each form unproven, `yield*`, reassigned `let`, a `let` holder), `contracts_ensures_invalid` (E0987 cases, including a parameter named like a module constant and a u8 bound of 256), `contracts_ensures_match` (comparison in `match`), `testdata/astdump/ensures_lines`, `testdata/fmt/comment_ensures`.
- Review round 1: a u64 `ensures != 0` gives callers non-zero (`narrowFact` treats a negative end of a u64 fact as unbounded and keeps `nonzero`; sample `testdata/contracts_ensures_u64`); a u64 literal pattern from 2^63 up is E0987; the bound of a comparison is parsed with `contractPattern` cleared (`testdata/errors/contracts_ensures_bound_match`); `requires` or `ensures` on a trait signature is E0987 "needs a function body" (`contracts_ensures_trait`); `findHolderKey` is `createHolderKey`.
- Review round 2: a literal pattern equal to an end of the int range (`9223372036854775807`, `-9223372036854775807`) is E0987 too, since those ends mean "no fact" (`isUnboundedEnd` in `operands.nra`); `requires` on a trait signature is E0984, `ensures` E0987 (`contracts_ensures_trait_requires`, `contracts_ensures_trait`).

## Deviations

- Files outside `shared:`: `trivia.nra` (getPatternId case), `emitmatch.nra` (two switch cases, nothing to emit), `querysites.nra` and `typesdump.nra` (one case each): the new `Pattern` variant needs a case in every exhaustive `match`. `idcheck.nra` collects the ensures pattern ids (allowed by the task).
- `check.nra`: the `Walker` trait gained `resolvePattern`, implemented in `body.nra`. `match.nra` calls the seeding in `contracts.nra`, so `contracts.nra` cannot import `match.nra`.
- Struct patterns are written `{ width: > 0 }`, not `Rect { width: > 0 }` as in D95's example: Narra patterns carry no type name.
- `!=` added by the maintainer after the first commit: `!= 0` gives callers non-zero, another `!= c` gives no fact. A return is proven by a constant other than the bound or a range that excludes it (`if (value != 5)` is not enough, since the prover keeps no "not 5").
- Async: Narra has no async functions, so there is nothing to reject.
- `ensures` on `implement` members is allowed and proven, but callers get no facts (calls go through the trait). Not decided by the brief.
- Not in the brief, left out: `return g(x)` is not proven from g's own `ensures` (no pattern implication); `let r = f(); r is P` learns nothing (only a `const` holder).
- The guide (`docs/guide/03-functions-and-closures.md`) is not in `shared:` and is not updated.

## Stubs left

- None.

## Gotchas

- `getIdKey` is `e<id>`, so a trivia key starting with `e` collides with every expression and pattern; ensures lines use `n<id>`.
- A `const` struct, array, map, box or generator can still be passed `mutable`, so it is not a result holder.
- E0988 is reported where the `return` is checked, so it comes in body order, unlike E0985.

## Proposed learnings

- None.

## Verification

- Dev compiler (seed emits selfhost, `cc -O0`): `--suite errors --suite dumps --suite fmt --suite codes --suite node-ids --suite trivia --suite fmt-stable`: pass. `--suite spec --suite programs --suite packages --suite lsp --suite query`: pass.
- Probes (scratchpad): a cross-module package (`Lib.positive`, imported `maybe`, `match (Lib.maybe(2))`) learns the facts; int range end bounds (`>= -9223372036854775808`, `> 9223372036854775807`, u64 above 2^63) neither crash nor prove from the unbounded ends; a `const` struct passed `mutable` gives no fact.
- `sh scripts/test.sh` after review round 2, rebased onto build-targets: PASS (8 shards plus bootstrap, 157 s); stage 2 and stage 3 emit identical C, the seed is older than the source as expected.

## Size

- Estimated M, actual M/L. The caller-side seeding and the switch cases in files outside the node made it larger than M.
