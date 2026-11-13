# Outcome: refinement-types

## Built

- D116 is built. `distinct type Port = int derives Show requires >= 1 && <= 65535` and `distinct type Method = string requires "GET" | "POST"`. The list may also start on the next line.
- Syntax: `parseRefinement` in `selfhost/src/parser.nra` stores the items in `DistinctDecl.refinement: Array<Pattern>` (`ast.nra`). Items are parsed with `contractPattern` set and joined by `&&`. The parser reports E1220 for: a missing item, `&&` with nothing after it, a second `requires`, and `implements` or `derives` after `requires`.
- Declaration check: `selfhost/src/refinements.nra`, `checkModuleRefinements`, called from `declareModule` (`declare.nra`) after `declareConsts`.
  - A base that is not an integer type or `string` is E1221. Another distinct type counts as such a base.
  - An item that does not fit the base is E1222.
  - Valid items are stored in `DistinctInfo` (`types.nra`):
    - `refined`: the type has a `requires` list.
    - `tests: Array<Array<RefinementTest>>`: the alternatives of each item.
    - `described`: each item as written, from `describePattern`.
    - `hull: Option<RefinementHull>`: integer bases only.
  - `tests` stays empty when an item failed.
- `T.from(x)`: `proveRefinedFrom` runs from `checkDistinctCall` (`calls.nra`). It proves each item with `findOperandFact` and `isComparisonProven`; a string item needs an equal string literal.
  - A failed proof is one E1223 at the argument, naming the failed items: `70000 does not satisfy Port (<= 65535)`, `"PUT" does not satisfy Method ("GET" | "POST")`, or `Port.from needs raw to satisfy <= 65535; check it first or call Port.check`.
  - A top-level `const` and a struct field default may hold a distinct `from` call, proven the same way.
  - A non-name argument is named as written (`holder.raw`, `Wide.toInt(wide)`), and string items and values are written escaped (`"a\"b"`).
- `T.check(x): Option<T>` exists only when `refined` is true. Its type comes from `findCheckType`.
  - `emitRefinedCheck` (`emitexprs.nra`) writes one C condition: comparisons, `||` between alternatives, `lstr_eq` for strings. It builds `Some` or `None`.
  - `evaluateDistinctCall` (`consteval.nra`) runs the same tests in `@const` code.
  - E0321 now names the function (`%s.%s is not available ...`). E0322 lists `check` only for a refined type.
- Facts: `findRefinedFact` (`operands.nra`) gives the hull for `T.toInt(e)`, or any unwrap of a refined integer type, whatever `e` is.
  - `findOperandFact`, `isProvenDivisor` and `isProvenCount` read it.
  - `seedResultLocal` (`contracts.nra`) seeds `const n = Port.toInt(port)` with it.
- `derives Shape`: `writeDistinctFromValue` (`deriveshape.nra`) calls `check` for a refined type. On failure it returns `Err` with `<Show of the value> does not satisfy T (<items>)`; the item text comes from `writeRefinementItems` (`querysignature.nra`).
- Tools:
  - `astdump` prints a `.refinement` slot, only when the list is not empty.
  - `fmt` prints ` requires a && b`, or a line break and an indent when it was written on the next line. A comment before `requires` stays: the list is noted as the `n<first item id>` node.
  - `trivia.nra` gains `getPatternStart`.
  - `narrac query` signatures print the list; an opaque type prints neither its base nor the list.
- Docs:
  - SPEC "Refinements": the "not built yet" line is gone, and there are two checked `ts` blocks, one with `<!-- error: E1223 -->`.
  - DECISIONS.md: a "Built" paragraph under D116.
  - `compiler/README.md`: a feature bullet.
- Tests:
  - Samples: `testdata/refinement_types.nra`/`.out` (with the `@const` counts) and `testdata/refinement_shape.nra`/`.out`.
  - Errors: `testdata/errors/refinement_{syntax,syntax_and,syntax_order,syntax_twice,base,pattern,from,check_plain}.nra`/`.err` and `testdata/packages/errors/refinement_opaque/`.
  - Tools: `testdata/astdump/refinement_types.nra`/`.ast`, `testdata/fmt/refinement_types.nra`/`.fmt` and `testdata/query/refinement_types/`.

## Proof

- base: feat/batch-13 7df2cd63, `narrac check testdata/refinement_types.nra` exits 1 with `refinement_types.nra:4:46: expected the end of the statement, found "requires"`.
- branch: the sample matches its `.out` in the debug build (AddressSanitizer, leak check, copy check) and in the release build.

## Deviations

- Files outside the owned paths:
  - `selfhost/src/exprs.nra`: `checkConstCall` now accepts `Ref.Distinct` calls. Without it, `const DEFAULT_PORT = Port.from(8080)` (the brief's own sample, and D116 part 2) was E0192 for every distinct type. `isConstantValue` (`isConstantCallee`) now accepts a distinct `from` call with constant arguments, so a field default `port: Port = Port.from(8)` is no longer E0106 (fresh review, should-fix 1).
  - Tests in other suites' folders: `testdata/astdump/refinement_types.*`, `testdata/fmt/refinement_types.*`, `testdata/query/refinement_types/` and `testdata/packages/errors/refinement_opaque/`. E0321 needs two modules, so the opaque test is a package test, not a single file.
- E1220 is split over four single-file tests, `refinement_syntax*.nra`. The parser stops at its first error, so one file can show only one case.
- Hover shows type-of answers, not declarations. It is covered by `query type-of` of `Port.check(...)` (`Option<Port>`), the same engine as hover. The `requires` list is in the `interface` signatures.
- `!= c` with `c != 0` adds no fact, as in `ensures`.
- After `&&` at the end of a line, the list goes on only when the next line starts with a comparison or a literal; anything else is E1220. The lexer joins the line after `&&`.

## Notes from the review, left as they are

- A bound written as a constant name prints the name in messages (`<= HIGHEST`), not its value (review nit 3).
- A template string without interpolation (`` Method.from(`GET`) ``) is not proven: only a string literal proves a string item (review nit 5).

## Stubs left

- None.

## Gotchas

- Node types are kept per body. The items are not checked with `Walker.resolvePattern`; read `DistinctInfo.tests`, not the AST patterns, in later checker or emitter code.
- `check` is detected by the callee name (`Expr.FieldExpr(_, _, _, "check")`) together with `DistinctInfo.refined`. No new `Ref` variant was added.
- `types.nra` cannot import `operands.nra`, so the hull is the struct `RefinementHull`, converted to `Fact` in `findRefinedFact`.
- `refinements.nra` imports `contracts.nra` and `operands.nra`, so those two cannot import it.

## Proposed learnings

- A value-level fact added to `findOperandFact` must also be added to `isProvenDivisor` and `isProvenCount` | they read local facts directly and skip `findOperandFact` | selfhost/src | no

## Verification

- `NARRAC=<branch narrac, release build> sh scripts/test.sh` on the branch rebased on feat/batch-13 0bab063f: `PASS, 8625 cases in 282s`.
- Final `sh scripts/test.sh` (seed build) on the final commit, rebased on feat/batch-13 14af8001: `FAIL after 332s`. 8791 cases passed and 1 failed: `programs/async_api_all_observed.nra/release`. That is the known timing flake under load: its timer lines (`quick slept`) came out in another order. Rerun alone, `--filter async_api_all_observed`: `PASS, 5 cases`.
- `--suite spec`: PASS, 63 cases. `--filter refinement`: PASS, 73 cases.
- `narrac emit selfhost` of the feat/batch-13 source (ef0a8023, after the review fixes): the base and branch compilers emit identical C (226310 lines).
- Instructions retired, 3 runs each, alternating: base 3.064G, 3.050G, 3.035G; branch 3.065G, 3.022G, 3.022G. The change is within the noise.

## Size

- Estimated L, actual L.
