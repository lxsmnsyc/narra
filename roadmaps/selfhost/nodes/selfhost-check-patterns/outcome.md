# Outcome: selfhost-check-patterns

## Built

Patterns, `is`, `match`, destructuring declarations, narrowing, and Map, tuple and spread literals in the Narra checker. About 1,000 new Narra lines.

| module | ports |
| --- | --- |
| `narrac/patterns` (patterns.nra, below exprs) | narrowing helpers (`findNarrowable`, `findNarrowed`, `setNarrowed`, `findOtherVariant`), `irrefutable`, `areArgsIrrefutable`, `isWild`, `collectPatternLocals`, `bindPatternLocals`, the exhaustiveness algorithm (`findMissingCases`, `specialize`, `describeConstructor`) |
| `narrac/match` (match.nra, above exprs) | `resolvePattern` (resolvePat, bindRest), `requireQualified` and `findEnumName`, `checkIs`, `checkMatch`, `checkDestructure`, `checkPatternExpression` |
| `narrac/literals` (literals.nra, above exprs) | `checkMapLiteral` (with type arguments), `checkMapIndex`, `checkTuple`, `checkSpread` |

Edits in shared files: `Walker` gained `checkPatterns`, `checkMapLiteral`, `checkTuple`, `checkSpread`, `checkMapIndex` (implemented by `BodyWalker` in body.nra); exprs.nra arms call them. `checkField` reads the narrowing. stmts.nra: `checkIf` (binding, narrowing, early-return narrowing), match statements, destructuring, `assigningId` around an assignment target. typesdump.nra and typesdump.go print the new attributes.

### Narrowing representation

No field on the local. Narrowing is a key `narrow|<local index>` with the variant index as value, in the same per-scope maps as local names (`Checker.localScopes`). `pushScope`/`popScope` therefore save and restore it. `setNarrowed` writes to the innermost scope, so the early-return case (`if (x is A) { return }` narrows `x` to the other variant for the rest of the block) lands in the enclosing scope. Only `const` locals narrow. A narrowed field read records `Ref.Narrow(enum, variant)` on the FieldExpr node.

Other tables: a pattern naming a variant records `Ref.Variant(enum, variant)`; a binding is the local its PatName node declares (`BodyInfo.declared`), created at resolve time and put in scope by the caller, as in Go.

### Dump (Go and Narra)

- Pattern lines carry `ctor=<Enum>.<Variant>` when they name a variant.
- A binding pat-name line is followed by its `local` line (`mutable` for `let` destructuring).
- `narrow=<Variant>` on narrowed field reads (already in the Go format).
- Goldens: `compiler/testdata/typesdump/patterns/` (match_values, destructure, narrowing, literals, err_match, err_patterns, err_literals, err_narrowing: leaks past if and else, mutable local, wrong variant field, the 10-case limit); `TestTypesDumpPatterns` and `TestTypesDumpPatternsCoverage` in selfhost_patterns_test.go.

### Tests

- `TestSelfhostPatterns`: same comparison as `TestSelfhostBodies` over the patterns cases only (TestSelfhostBodies compares the earlier inputs): 8 compared, 0 skipped. No compared input differs. The missing-case text (order, limit 10) equals Go's on the error goldens.
- `TestSelfhostBodies` (which now also walks testdata/typesdump/patterns, since declInputs include all typesdump files): compared 166, skipped 187. Before: 143 compared, 202 skipped. 8 of the new compared inputs are mine; 15 more existing inputs became comparable.

### Skip reasons now (TestSelfhostBodies)

| count | reason | node |
| --- | --- | --- |
| 47 | an enum constructor | calls |
| 30 | does not load | (TestSelfhostModules) |
| 20 | implement block bodies | calls |
| 10 | a distinct type function | calls |
| 7 + 6 + 4 + 3 | trait function call; namespace without its import; mutable argument beside others; generic function call | calls |
| 11 + 8 | an arrow function; a call of a function value (+ non-escaping parameters) | closures |
| 8, 5, 4, 3, 3 and a few more | builtins with a named check: Array.length, Array.values, Test.eq, Array.push, Map.size, ... | builtins |
| 5 | a constant declared twice | (see check-bodies) |

No pattern, match, `is`, Map, tuple or spread construct is skipped any more.

## Deviations

- The brief says to put the module "below exprs.nra". Only `patterns.nra` is. match.nra and literals.nra need `expectType`, `describeType`, `findField` from exprs.nra, so they sit above it and exprs reaches them through five new `Walker` trait functions.
- Shared files touched beyond new arms: check.nra (`Ref.Variant`, `Ref.Narrow`, the trait functions, `Checker.assigningId`), body.nra (trait implementation), stmts.nra (`checkVariable` takes the `Option<Pattern>` instead of a bool).
- "Unreachable arm": Go has no such error, so none is ported or tested. The other listed errors (non-exhaustive, unknown variant, arity, wrong type, duplicate binding, refutable declaration, float map key, map type arguments) are in the goldens.
- No new diagnostic codes.

## Stubs left

- None for this node. Inputs that need constructors, calls of generic functions, closures or builtins still skip as `unsupported:` (calls, closures, builtins).
- `checkConstField` (exprs.nra) treats any resolved FieldExpr as fine, which would also accept `Ref.Narrow`; a constant has no locals, so it cannot happen.

## Gotchas

- `x is Pat(a, b) && cond` is not an allowed shape in Narra; an `is` with bindings must be the whole condition of an `if`. This applies to selfhost code too.
- `f(mutable checker, ... typeString(checker, ...) ...)` is rejected; compute the message into a `const message` first.
- Pattern ids are `-1` for the synthetic wildcard used by exhaustiveness; `getRef` ignores ids outside the body.
- A later node that needs `ctor=` for constructor expressions can reuse `Ref.Variant` (same dump text).

## Proposed learnings

- Keep scope-bound facts such as narrowing in the scope stack under a reserved key | push/pop restores them with no extra bookkeeping | selfhost/ | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 117 s).
- `go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostPatterns'`: pass.
- Go emit and check are unchanged (only typesdump.go and tests changed on the Go side).

## Size

- Estimated M, actual M. About 1,000 Narra lines as estimated.
