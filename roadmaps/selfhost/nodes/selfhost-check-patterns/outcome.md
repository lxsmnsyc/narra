# Outcome: selfhost-check-patterns

## Built

Patterns, `is`, `match`, destructuring declarations, narrowing, and Map, tuple and spread literals in the Narra checker. About 1,000 new Narra lines.

| module | ports |
| --- | --- |
| `narrac/patterns` (patterns.nra, below exprs) | narrowing helpers (`findNarrowable`, `findNarrowed`, `setNarrowed`, `findOtherVariant`), `irrefutable`, `areArgsIrrefutable`, `isWild`, `collectPatternLocals`, `bindPatternLocals`, the exhaustiveness algorithm (`findMissingCases`, `specialize`, `describeConstructor`) |
| `narrac/match` (match.nra, above exprs) | `resolvePattern` (resolvePat, bindRest), `requireQualified`, `checkIs`, `checkMatch`, `checkDestructure`, `checkPatternExpression` |
| `narrac/literals` (literals.nra, above exprs) | `checkMapLiteral` (with type arguments), `checkMapIndex`, `checkTuple`, `checkSpread` |

Edits in shared files: `Walker` gained `checkPatterns`, `checkMapLiteral`, `checkTuple`, `checkSpread`, `checkMapIndex` (implemented by `BodyWalker` in body.nra); exprs.nra arms call them. `checkField` reads the narrowing. stmts.nra: `checkIf` (binding, narrowing, early-return narrowing), match statements, destructuring, `assigningId` around an assignment target. typesdump.nra and typesdump.go print the new attributes.

### Narrowing representation

No field on the local. Narrowing is a key `narrow|<local index>` with the variant index as value, in the same per-scope maps as local names (`Checker.localScopes`). `pushScope`/`popScope` therefore save and restore it. `setNarrowed` writes to the innermost scope, so the early-return case (`if (x is A) { return }` narrows `x` to the other variant for the rest of the block) lands in the enclosing scope. Only `const` locals narrow. A narrowed field read records `Ref.Narrow(enum, variant)` on the FieldExpr node.

Other tables: a pattern naming a variant records `Ref.Constructor(enum, variant)` (the ref of selfhost-check-calls, which prints the same `ctor=`); a binding is the local its PatName node declares (`BodyInfo.declared`), created at resolve time and put in scope by the caller, as in Go.

### Dump (Go and Narra)

- Pattern lines carry `ctor=<Enum>.<Variant>` when they name a variant.
- A binding pat-name line is followed by its `local` line (`mutable` for `let` destructuring).
- `narrow=<Variant>` on narrowed field reads (already in the Go format).
- Goldens: `compiler/testdata/typesdump/patterns/` (match_values, destructure, narrowing, literals, err_match, err_patterns, err_literals, err_narrowing: leaks past if and else, mutable local, wrong variant field, the 10-case limit); `TestTypesDumpPatterns` and `TestTypesDumpPatternsCoverage` in selfhost_patterns_test.go.

### Tests

- `TestSelfhostPatterns`: same comparison as `TestSelfhostBodies` over the patterns cases only (TestSelfhostBodies compares the earlier inputs): 8 compared, 0 skipped. No compared input differs. The missing-case text (order, limit 10) equals Go's on the error goldens.
- After merging origin/main (selfhost-check-calls, merge commit instead of a rebase because both commits conflicted): `TestSelfhostBodies` compared 245, skipped 122; `TestSelfhostCalls` compared 247, skipped 122; `TestSelfhostPatterns` (patterns cases only) compared 8, skipped 0. No compared input differs, including the inputs that need both patterns and calls (before the merge: bodies 166/187).

### Skip reasons now (TestSelfhostBodies)

| count | reason | node |
| --- | --- | --- |
| 30 | does not load | (TestSelfhostModules) |
| 16 + 8 + 2 + 1 | an arrow function; a call of a function value; a non-escaping parameter; a function used as a value | closures |
| 12, 10, 8, 5, 4, 4, 3, 3, 2 and 8 more | builtins with a named check: Array.length, Array.values, Box.new, Array.push, Map.size, Test.eq, Option.map, Range.upTo, Result.andThen, ... | builtins |
| 6 | a constant declared twice | (bodies stub) |
| 1 | an import of a constructor (E0425) | calls |

By node: builtins 54, closures 27, does not load 30, constant declared twice 6, calls 1, patterns 0.

## Deviations

- The brief says to put the module "below exprs.nra". Only `patterns.nra` is. match.nra and literals.nra need `expectType`, `describeType`, `findField` from exprs.nra, so they sit above it and exprs reaches them through five new `Walker` trait functions.
- Shared files touched beyond new arms: check.nra (`Ref.Narrow`, the trait functions, `Checker.assigningId`), body.nra (trait implementation), stmts.nra (`checkVariable` takes the `Option<Pattern>` instead of a bool).
- "Unreachable arm": Go has no such error, so none is ported or tested. The other listed errors (non-exhaustive, unknown variant, arity, wrong type, duplicate binding, refutable declaration, float map key, map type arguments) are in the goldens.
- No new diagnostic codes.

## Stubs left

- None for this node. Inputs that need constructors, calls of generic functions, closures or builtins still skip as `unsupported:` (calls, closures, builtins).
- `checkConstField` (exprs.nra) treats any resolved FieldExpr as fine, which would also accept `Ref.Narrow`; a constant has no locals, so it cannot happen.

## Gotchas

- `x is Pat(a, b) && cond` is not an allowed shape in Narra; an `is` with bindings must be the whole condition of an `if`. This applies to selfhost code too.
- `f(mutable checker, ... typeString(checker, ...) ...)` is rejected; compute the message into a `const message` first.
- Pattern ids are `-1` for the synthetic wildcard used by exhaustiveness; `getRef` ignores ids outside the body.
- `namesEnum` and `findEnumName` now come from ctors.nra (exported there, `namesEnum` declared `uses alloc`); match.nra no longer has its own copies.

## Proposed learnings

- Keep scope-bound facts such as narrowing in the scope stack under a reserved key | push/pop restores them with no extra bookkeeping | selfhost/ | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 117 s).
- `go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostPatterns'`: pass.
- Go emit and check are unchanged (only typesdump.go and tests changed on the Go side).

## Size

- Estimated M, actual M. About 1,000 Narra lines as estimated.
