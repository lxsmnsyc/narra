# Outcome: selfhost-check-calls

## Built

The call half of the body checker in Narra, compared with Go through the types dump. About 860 new Narra lines (calls 410, ctors 397, instances 52) plus about 380 changed lines in exprs, body, check, typesdump and scopes.

### Modules (selfhost/src)

| module | ports |
| --- | --- |
| `narrac/calls` (calls.nra) | Go's `call` dispatch (`checkCall`, `checkNamedCall`, `checkDottedCall`, moved here from exprs.nra), `genericCall` (`checkGenericCall`), `traitCall` (`checkTraitCall`), `moduleCall` (`checkModuleCall`), `distinctCall` and `unwrapName` (`checkDistinctCall`, `findUnwrapName`) |
| `narrac/ctors` (ctors.nra) | `ctor` (`checkConstructor`), `enumByName` (`findEnum`), `bareOwner`, `ctorOf`, `noCtor`, `ctorArgs`, `ctorOwners` (`listConstructorOwners`), `visibleEnums` (`listVisibleEnums`), `enumNameFor` (`findEnumName`), `namesEnum` |
| `narrac/instances` (instances.nra) | `funcInstance` (`addFuncInstance`) |
| `narrac/exprs` (changed) | `checkExclusive` and `usesLocal` (real port, the stub is gone), `missingImport` (`reportMissingImport`, E0908), `moduleEnumName` (`findModuleEnum`, gives `Alias` the type `module <path>`), `checkArguments` now takes one expected type per argument (`Array<Option<Type>>`), constructor names in `checkName`/`checkField`, constructor calls in `checkConstCall`, constructors in `isConstantValue(checker, expression)` |
| `narrac/body` (changed) | `checkPending`, `checkImplBodies` (Go's `implBodies`), `checkFunctionBody` (shared by functions, instances and implement block functions), the two new `BodyWalker` functions |

Import order: check <- exprs <- stmts; exprs <- ctors <- calls; instances (check, resolve, generics) <- calls; body imports calls and ctors.

### Layout: calls sit above exprs

calls.nra and ctors.nra use `expectType`, `checkArguments`, `checkBuiltinCall` and the other argument checks of exprs.nra, so they import exprs, and exprs reaches them through two new `Walker` functions in check.nra:

```
function checkCall(walker: Self, mutable checker: Checker, nodeId: int, pos: Pos, callee: Expr, args: Array<Arg>, want: Option<Type>): Type uses alloc
function checkConstructorName(walker: Self, mutable checker: Checker, nodeId: int, pos: Pos, enumName: string, name: string, want: Option<Type>): Option<Type> uses alloc
```

The `CallExpr` arm of `checkExpressionKind` is now `Walker.checkCall(...)`. `checkName` and `checkField` pass `want` through.

### Refs

`Ref` (check.nra) gained `Constructor(enumIndex, variant)`, `Implementation(block, index)` (Checker.impls[block].funcs[index]), `TraitFunction(traitIndex, name)` and `Distinct(index)`. A generic call records `Ref.Function` of the generic function, both in the open template and in an instance, as Go's dump names an instance by its origin.

### The pending queue and instance keys

- The queue lives in `Checker` (check.nra): `instanceList: Array<FuncInstance>` (every instance in discovery order), `funcInstances: Map<string, int>` (key to instanceList index), `pending: Array<int>` (instances to check) and `pendingNext` (queue head).
- An instance is keyed by `"<Checker.funcs index>|<typeKeys(targs)>"`. Struct and enum instances are deduped by `Checker.instances`, so equal type arguments give equal text, and the key names the same instances as Go's C-name key.
- `FuncInstance { func, scope, env, targs, params, result }`: the syntax stays shared; the signature is resolved again in the declaring module with `env` (as Go's `funcInstance` does), and the body gets its own `BodyInfo` with `targs`.
- `checkPending` (body.nra) runs after the module loop and before `flushMapKeys`, enters the declaring module and checks each instance body; checking one can queue more. Past 10,000 instances the error is reported and the instance is not queued, as in Go.
- The order equals Go's: `TestTypesDumpInstanceOrder` checks it on `ping`/`pong`, two generic functions that call each other's instances, and `TestSelfhostCalls` compares the whole dump.

### Dump

- Go printed the C name of an implement block function on a `Trait.f(...)` call (`func=m.Cat_Named_name_80482e03`). Both printers now print `func=<Type>.<Trait>.<function>` (`func=traits.Cat.Named.name`), the type as in the decls dump. No existing golden had such a line. typesdump.go gained `implDumpNames`; the doc comment lists the attribute and this node's owners.
- `ctor=`, `trait=`, `distinct=`, `func=` and the `impl` / `targs=` headers are printed on the Narra side from the refs.

### Tests

- `selfhost_calls_test.go`: `TestTypesDumpCalls` (goldens in `compiler/testdata/typesdump/calls/`, `NARRA_UPDATE=1` rewrites them), `TestTypesDumpCallsCoverage`, `TestTypesDumpInstanceOrder`, and `TestSelfhostCalls` (narrac against Go on the TestSelfhostBodies inputs plus the call cases; a call case may never be skipped).
- Call cases: `generics` (one and several instances, nested generic calls, a generic call in a generic body, mutual instances), `ctors` (with and without expected types, a name in two enums, generic enums, Option and Result, constructor defaults and constants), `traits` (implement blocks, calls on structs, enums, prelude instances, type parameters with bounds, compiler-implemented Show, Ordered and Equatable), `distinct`, `modules/` (package: module calls, generic module functions with instances in both modules, `Geo.Shape.Circle`, `Geo.Wrapped.One`, imported opaque distinct), and the errors `err_bound`, `err_infer`, `err_ctor` (bare, ambiguous, field count, unknown, mutable), `err_missing_import` (E0908 both messages), `err_exclusive`, `err_distinct`, `err_private/` (private call, missing export, opaque `from`).

### Compared and skipped (TestSelfhostBodies)

Compared 198, skipped 161 (main: 143 and 202). Of the inputs main had, 184 are compared now; the other 14 are the new call files under testdata/typesdump/calls, which the bodies inputs pick up too. No compared input differs. TestSelfhostCalls: compared 200 (the same plus the two package cases), skipped 161.

| count | reason | node |
| --- | --- | --- |
| 30 | does not load | (TestSelfhostModules) |
| 18 | a match expression | patterns |
| 15 | an arrow function | closures |
| 14 | a tuple literal | patterns |
| 11 | a Map literal | patterns |
| 9, 7, 4, 4, 4, 3, 2, 2, 2, 1 x 7 | builtins with a named check: Array.length, Box.new, Array.push, Array.values, Test.eq, Range.upTo, Map.size, Option.map, Result.andThen, Option.andThen, Option.isSome, Result.map, Result.mapErr, Test.golden, Test.panicsWith | builtins |
| 8 | a call of a function value | closures |
| 6 | a constant declared twice | (bodies stub) |
| 4 | a destructuring declaration | patterns |
| 3 each | `is`; a spread in a struct literal | patterns |
| 2 | a non-escaping function parameter | closures |
| 1 each | a Map index, a match statement (patterns); a function used as a value (closures); an import of a constructor (calls, see Stubs) | |

By node: patterns 55, builtins 50, closures 26, does not load 30, constant declared twice 6, calls 1.

## Deviations

- Files outside the owned and shared paths: `selfhost/src/scopes.nra` (`Namespace` exported and a new `findNamespace`, for the E0908 messages). Shared files edited: check.nra, exprs.nra, body.nra, typesdump.nra, typesdump.go.
- The brief says to find a namespace's module from `StdFunc.name` and `StdFunc.module`. The E0908 message also needs the plural alias (`Arrays`), which only scopes.nra's `NAMESPACES` table has, so both come from there.
- No `module=` attribute: Go prints none. As selfhost-check-bodies decided, a resolved alias shows as the type `module <path>` (now also on `Alias` in `Alias.Enum.Variant`).
- The impl function name in `func=` changed in Go's printer (see Dump). `narra emit`, `check` and `check --json` are untouched: only typesdump.go changed in Go.
- The call dispatch moved from exprs.nra to calls.nra instead of staying in exprs with one-line arms, because calls must import exprs (see Layout).
- No change requests.

## Stubs left

- E0425, an import of a constructor (`scopes.nra`, `unsupported: an import of a constructor (selfhost-check-calls)`): Go reports it at the ImportName, and the Narra AST has no ImportName positions. It needs a parser change (`ImportName.pos`). 1 input.
- `an arrow argument of a generic call (selfhost-check-closures)` in `checkGenericCall`: Go's `hintArrow`.
- Unchanged from earlier nodes: function values and calls of them, `Alias.f` / constants holding functions in `checkModuleCall` (closures), named builtin checks (builtins), patterns.

## Gotchas

- A module that needs the expression checks plus its own entry point from exprs sits above exprs and adds a `Walker` function; one below exprs can only reach `Walker.checkExpression` and `checkBlock`.
- `checkArguments` takes `wants: Array<Option<Type>>`, one per parameter. Callers with known types pass `Some(paramTypes[param.id])`.
- `isConstantValue` now takes the checker: it reads `Ref.Constructor` from the current body, so call it while the default's body is current.
- Go's `isConstant` has no FieldExpr case: `shape: Shape = Shape.Empty` is "a field default must be a constant", while `None` and `Some(3)` are fine. Kept.
- Go's `distinctOf` handles `Alias.T.f(...)`, but `call` only reaches it with an Ident target, so `Geo.Meters.toInt(x)` is an error in Go ("a module is not a value"). Kept.
- An instance's parameter types are in `FuncInstance.params`, not in `Checker.paramTypes`, which holds the template's.
- Writing `f(mutable checker, g(checker))` or a template string that reads `checker` inside a call with `mutable checker` is rejected by checkExclusive; read into a local first.

## Proposed learnings

- Put a checker part that needs exprs' helpers above exprs and reach it through a Walker function | modules cannot import in a cycle and the helpers are not walker-generic | selfhost/ | no
- Key per-instance Narra tables by `typeKeys` text of the type arguments | instance types are deduped by index, so the text names the same instances as Go's C names | selfhost/ | yes (TestSelfhostCalls)

## Verification

- `cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostCalls'` (the brief's verify): pass.
- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 100 s).
- `narra emit` is unchanged by construction: the only Go change is in typesdump.go, which emit and check do not use.

## Size

- Estimated L (about 1,080 Narra lines), actual L: about 860 new lines plus about 380 changed. The argument and builtin checks already existed in exprs.nra, so the new modules reuse them.
