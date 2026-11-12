---
id: std-function-values
title: "Standard functions as values"
status: active
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/std-function-values/**
  - selfhost/src/stdvalues.nra
  - testdata/std_function_values*
  - testdata/errors/std_function_value*
shared:
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/unused.nra
  - selfhost/src/query.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1013]
provides: []
depends_on: []
---

# Standard functions as values

## Goal

A builtin std function with a fixed signature is a function value: `map(names, Str.length)`, `apply(Int.toString, 1)`, `const show: (value: int) => string = Int.toString`, and the same through a named import (`import { toString } from "std/int"`). Its effects follow it, as for a user function (SPEC line 1296: `map(paths, File.readText)` has `alloc, io`). A builtin whose signature is computed per call (generic ones such as `Arrays.length`) is E1013, with a hint to wrap it in an arrow function.

Std functions written in Narra source already work as values: `Options.isSome` and `Int.compare` (from `std/src/ext__int.nra`) pass and run. Only the registry builtins fail. Errors on the current compiler (built from the seed at `57d55087`):

```ts
import * as Int from "std/int"

function apply(f: (value: int) => string, value: int): string {
  return f(value)
}

export function run(_args: Array<string>): int uses io, alloc {
  print(apply(Int.toString, 1))
  return 0
}
```

`fv1.nra:8:15: a module is not a value` (E0202, `checkName` in `exprs.nra:1031`, reached because `checkField` treats `Int` as an expression), then E0209 `void has no fields`. `Str.length` as a value gives E0202, then E0208 `there is no .length`. `import { toString } from "std/int"` then `apply(toString, 1)` is E0203 `unknown name "toString"`.

Plan:

- The checker resolves `Alias.f` and an imported name to a registry row and records `Ref.Builtin(index)` on the value's node.
- The effect pass reads the row's effects for such a value.
- The emitter writes one thunk per builtin used as a value and builds the function value from it, as `emitFuncValue` does for user functions.

## Context

- SPEC: "Functions" and "Closures" in section 2, and "Effect variables" (line 1286; its examples already pass `Str.length` and `File.readText`). README "Not supported yet" line 209: "standard functions (`Int.toString`) as values" comes out of that line; `@inline` and `async` arrow functions stay.
- Registry: `StdFunc` (`builtins.nra:22`). A row has a fixed signature when `check` is `""` or `"fixed"` and `result` is set (`checkBuiltinCall`, `exprs.nra:2421`, reads the types with `stdType`). Rows with another `check` value compute their signature in `checkStdCall` (`stdcalls.nra`) and are E1013. About 80 of the 146 rows have a fixed signature.
- Which fixed rows qualify: not rows with `mutables` (a function value has no `mutable` parameter, E0605 for user functions) or `callbacks`, and not `"fixed"` rows whose check proves something about an argument at the call (read `checkFixedCall` and each module's fixed check). Those are E1013 too, with the reason in the message.
- Checker, in a new module `selfhost/src/stdvalues.nra` (`checkStdFunctionValue(checker, nodeId, pos, builtin, want)`): builds the `Type.TFunc` from the row like `checkFunctionValue` (`funcvalues.nra:243`) does from a declaration, with no effects in the type (SPEC line 1282), and calls `setRef(..., Ref.Builtin(index))`. Call it from `checkField` (`exprs.nra:1197`) when the alias names a std module (`checker.registry.modules[path]`, as `calls.nra:449` does for calls) and the name is a row, before the source module and extension lookups fail. Call it from `checkName` (`exprs.nra:990`) when `checker.scopes[scope].imported[name]` holds the row (set in `scopes.nra:149`).
- Effects: `findFunctionValue` (`effects.nra:926`) goes through `findCallee`, which has no case for `Ref.Builtin`, so a stored `Fs.readText` would carry no effects. Add the row's `effects` there. Then storing it where the type lists fewer effects, and passing it to an effect variable, behave as for user functions (E0803 and D91).
- Emitter: `emitFuncValue` (`emitclosures.nra:174`) writes a thunk `static R thunk(narra_env* env, P0 p0, ...)` that calls the function and returns `((F){.fn = thunk, .env = NULL})`. For a builtin, the thunk body is the builtin's call as the emitter writes it, through `emitNamedCall` (`emitstd.nra`, private today) and the other dispatchers (`emitexprs.nra:565`, `emitBuiltinCall`). These read their arguments as AST `Arg` values through `EmitWalker.emitExpression` and `readNodeType`. Hook `emitexprs.nra:300` (`Ref.Function(_) => emitFuncValue`) for `Ref.Builtin` on a name or field node.
- `Ref.Builtin` today only marks calls. Check each reader that may meet it on a value node: `exprs.nra:390`, `sharedwrites.nra` (1959, 2350 to 2400), `sharedresults.nra:815`, `stdcopy.nra:286`, `pathends.nra:12`, `unused.nra`, `query.nra` and `lspnavigation.nra` (hover and definition give the std name).

## Acceptance

- [ ] Samples (`testdata/std_function_values*`): `Int.toString` passed to a user function and stored in a `const` with a written type; `Str.length` and `Str.trim` through a generic `map<T, U, effect E>` user function; `Fs.readText` passed where the type says `uses io, alloc`; a named import (`import { toString } from "std/int"`) used as a value; a builtin value stored in a struct field and in an `Array`; the same builtin used as a value twice (one thunk); a builtin returning a counted value (`string`) whose result is released by the caller.
- [ ] Every sample is clean under the leak check and sanitizers.
- [ ] Error tests (`testdata/errors/std_function_value*`): `Arrays.length` and `Maps.size` as values (E1013, the hint shows the arrow function); a row with a `mutable` parameter (E1013); `Fs.readText` stored where the type lists no effects (the existing effect error, E0803 or the storing error the user-function case reports); `Int.toString` where `(value: string) => string` is expected (E0119).
- [ ] E1013 is a row in `diagnostics.nra`; the `codes` suite passes.
- [ ] SPEC.md section 2 "Functions" says which std functions are values. README: the "standard functions as values" words leave the "Not supported yet" line; "Differences from the spec" says that std functions with a computed signature are not values yet, if SPEC keeps promising them.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one checker function in a new module with two call sites, one effects case, and one emitter thunk. The risk is the emitter: builtin emitters read AST arguments, and the thunk has only C parameters. If that hook spreads across many std emitters, split the emitter into its own node and land the checker with E1013 for every builtin first.

## Notes for the worker

- Open design questions, with the recommended answer:
  - **How the thunk calls the builtin.** Recommended: synthetic argument nodes. The thunk builder makes one `Expr.Ident` per parameter with fresh node ids, records each id's C text (`p0`) and type in an emitter side table, and passes them as `Arg` values to the builtin's emitter; the identifier emitter and `readNodeType` read that table first. Every builtin then works with no per-module change. Rejected: calling the runtime C function straight from the thunk, which covers only the `FixedCall` rows (about 35) and duplicates what the emitters know about ownership and position arguments.
  - **Ownership in the thunk.** A function value's parameters arrive borrowed and its result leaves owned, as for user thunks. If the builtin's `Code` is not owned and the result type is counted, retain it before returning.
  - **Generic builtins.** Out of scope: their types come from per-call checks. E1013 with the arrow form in the hint (`(items) => Arrays.length(items)`) is the answer for now. Many generic helpers are moving to Narra source (`std/src/`), where they become values for free.
- E1013 message: `%s cannot be used as a value: %s; wrap it in an arrow function, such as \`%s\``.
- A function value is still not equatable, so `Int.toString == Int.toString` stays an error.
- No deprecation paths.

## Plan review (2026-11-12)

- inline lands after this node and changes `emitIdent` and `readNodeType` too. Put the side-table lookup and any emitter context save and restore in one named helper in emitstate.nra, so inline can reuse it.
