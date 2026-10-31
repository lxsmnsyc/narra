# Outcome: selfhost-check-closures

## Built

Arrow functions, functions used as values, calls of function values and fields, non-escaping parameters, captures, return inference and arrow hints in the Narra checker. About 650 new Narra lines.

| module | ports |
| --- | --- |
| `narrac/funcvalues` (funcvalues.nra, below exprs) | `findClosure` (Go's `closureOf`), `markArgument` (`funcArg`), `isNonEscaping` (`nonEscapingParam`), `useLocal` with `addCapture` and `isDeclaredBelow` (`declaredBelow`), `checkFunctionValue` (`funcValue`, also `moduleFuncValue`), `findModuleFunction` (moved from calls.nra) |
| `narrac/closures` (closures.nra, above exprs) | `checkArrow` (`arrow`), `hintArrow`, `checkGenericArguments`, `checkValueCall` (`valueCall`), `checkValueCallOf` (`valueCallOf`), `checkCalleeExclusive` and `collectMutableRoots` (`mutableRoots`), `listReturns` (`returnsOf`), `findBlockResult` (`blockResult`), `inferReturns`, `checkArrowArgument` |

Edits in shared files: check.nra (`ClosureInfo`, `ArrowContext`, `Ref.Value`, `BodyInfo` closure fields, `Checker.inferring`, `Walker.checkArrow`), exprs.nra (arrow arm, `checkName`, `checkField` for `Alias.f`, `checkArgument` extracted from `checkArguments`, `checkExclusive` and `usesLocal` exported), calls.nra (value calls, field calls, `Alias.CONST(...)`, generic arguments through `checkGenericArguments`), stmts.nra (`checkReturn` when the result is inferred), body.nra (`BodyWalker.checkArrow`), typesdump.nra and typesdump.go.

### Layout

closures.nra needs `expectType` and the argument checks of exprs.nra, so it sits above exprs and exprs reaches `checkArrow` through `Walker.checkArrow`. What exprs itself needs (the non-escaping rule, captures, named functions as values) cannot import closures.nra, so it is in the extra file funcvalues.nra below exprs. calls.nra imports closures.nra directly.

### Closure tables (per body, D70)

`BodyInfo.closures: Array<ClosureInfo>` with `closureIndex` (arrow node id to index): `stack`, `captures` (`BodyInfo.locals` indexes in order of first use), and the hint (`hasHint`, `hintParams: Array<Option<Type>>`, `hintResult`). `BodyInfo.arrows` is the stack of arrows being checked (`outerDepth` is the scope depth the arrow appears in, popped slots are reused) and `allowedId` is the one Ident that may name a non-escaping parameter. A local is declared below an arrow when `localScopes[depth][name]` holds it for `depth >= outerDepth`. The type of an arrow is the node type (`Type.TFunc`); its parameters are locals declared by their `Param.id` and recorded in `paramTypes`. A function value is `Ref.Function` on the Ident or FieldExpr (an instance is queued for a closed generic one); a call through a value is `Ref.Value` on the CallExpr. `Checker.inferring` makes `checkReturn` skip the expected type while a block arrow's result is inferred; the returns are compared afterwards in `inferReturns`.

### What the emitter needs (selfhost-emit)

- For each arrow node: `BodyInfo.closures[closureIndex[nodeId]]` for `stack` and `captures`; the node type for the signature; the parameter locals through `BodyInfo.declared`.
- Function types are plain `Type.TFunc`, so Go's `funcOf` interning and `finishFuncTypes` are not ported: selfhost-emit collects the closed function types of the program (walk node types, struct fields, locals) and sorts them as Go does.
- A function value is `Ref.Function` on an Ident or FieldExpr whose node type is a `TFunc`; the instance of a generic one is found through `funcInstances` (`instances.nra`).
- A call through a value has `Ref.Value`; the callee type is the callee node type.
- The arrow body is the syntax itself: Go builds a `FuncDecl` and a `ReturnStmt`/`ExprStmt` for an expression body, the emitter must do the same from `ArrowExpr.body`.

### For selfhost-check-builtins

`checkArrowArgument(mutable checker, walker, argument, hintParams: Array<Type>, hintResult: Option<Type>): Type` in closures.nra ports Go's `checkCallback` prelude: an arrow argument gets `stack = true` and the hint, a name may be a non-escaping parameter, and the expected type is `TFunc(hintParams, result)` when the result is known. The builtin checks it replaces (callback arity, `needs a function as argument`) stay in builtins. `useLocal` already allows a function-typed parameter as a callback through `allowedId`.

### Tests

- `selfhost_closures_test.go`: `TestTypesDumpClosures` (goldens in `compiler/testdata/typesdump/closures/`, `NARRA_UPDATE=1` rewrites them), `TestTypesDumpClosuresCoverage`, and `TestSelfhostClosures` (narrac against Go on the closures cases, `testdata/closures*.nra` and `testdata/errors/closures*.nra`; a case in the closures folder may never be skipped).
- Cases: `arrows` (capturing and non-capturing arrows, stack and heap, nested arrows, returned arrows, struct fields, function names as values, calls of values and fields), `inference` (hints from earlier arguments in generic calls, inferred and void block results, bare returns, a generic function as a value from context), `modules/` (package: imported function, `Alias.f`, generic `Alias.f`, `Alias.CONST` holding a function, an arrow passed to a module function), `err_closures` (arity, untyped parameter, captures of a `let`, a mutable parameter and a non-escaping parameter, escaping a non-escaping parameter, calling a non-function, call arity, mutable argument, wrong argument, callee exclusive for a local and a field, inconsistent returns, a function with a mutable parameter as a value, bad `escaping`, a generic function without context, generic arity).
- The dump: `arrow` lines carry `stack` and `captures=<local>,<local>`; calls through values carry `value`; function values carry `func=`. Added to typesdump.go and typesdump.nra, with the Owners list.
- `testdata/closures*.nra` and `errors/closures_*`: the error cases and `closures_traits.nra` are compared and equal; `closures.nra` and `closures_nested.nra` use `Array.values` and `Array.push`, so they skip until selfhost-check-builtins. Deferred: the acceptance item "attributes equal Go's for every sample in `testdata/closures*`" is not met for these two samples; it moves to selfhost-check-builtins.

### Compared and skipped (TestSelfhostBodies)

Compared 271, skipped 101 (main: 245 and 122). TestSelfhostClosures: compared 17, skipped 2 (both builtins). No compared input differs.

| count | reason | node |
| --- | --- | --- |
| 30 | does not load | (TestSelfhostModules) |
| 14, 12, 8, 5, 4, 4, 4, 3, 2, 2 and 6 more | builtins with a named check: Array.values, Array.length, Box.new, Array.push, Map.size, Range.upTo, Test.eq, Option.map, Result.andThen, Result.map, Option.andThen, Option.isSome, Result.mapErr, Test.golden, Test.ne, Test.panicsWith | builtins |
| 6 | a constant declared twice | (bodies stub) |
| 1 | an import of a constructor (E0425) | calls |

By node: builtins 64, closures 0, does not load 30, constant declared twice 6, calls 1.

## Deviations

- One new file outside the brief's owned paths: `selfhost/src/funcvalues.nra` (see Layout).
- `findModuleFunction` moved from calls.nra to funcvalues.nra and is imported by calls.nra, exprs.nra.
- `checkArguments` was split: the loop body is the exported `checkArgument`, so `checkGenericArguments` (closures.nra) can set an arrow hint after the earlier arguments are checked, as Go's lazy `paramType` callback does. `checkGenericCall` no longer builds its `wants` array first.
- Go's `typesdump.go` prints `stack` and `captures=` on arrow lines (new in Go); no other Go change.
- No change requests, no new diagnostic codes.

## Stubs left

- None for this node. Inputs with builtin helpers still skip as `unsupported:` (builtins).

## Gotchas

- `useLocal` is only reached from `checkName`; a non-escaping parameter used as a callee works because `checkValueCall` sets `allowedId` before checking the callee.
- A top-level constant may hold only a function name, not an arrow (`constant cannot use a closure`), so `Alias.CONST(...)` tests use `export const triple = double`.
- `f(mutable checker, g(checker))` and a template that reads `checker` inside such a call are rejected; read into a local first (hit three times here).
- Hints and `stack` live in the body of the arrow, so an instance body (checked again over the same syntax) sets them again; nothing carries over from the template.

## Proposed learnings

- Split a checker feature into a part below exprs (state and rules exprs calls) and a part above it (what needs exprs' helpers) | modules cannot import in a cycle | selfhost/ | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 98 s).
- `go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostClosures'`: pass.
- `narra emit` and `narra check` are unchanged: the only Go change is in typesdump.go.

## Size

- Estimated M, actual M. About 650 Narra lines as estimated, plus about 100 changed in shared files.
