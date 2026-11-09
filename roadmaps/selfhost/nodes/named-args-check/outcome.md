# Outcome: named-args-check

Named arguments check, reorder and emit (D87). The E0213 placeholder from named-args-syntax is gone.

## Built

- Canonical view of a call's arguments: the AST (`CallExpr.arguments`) keeps the written order, which is the evaluation order. When named arguments are not in parameter order, the checker records `Checker.argumentOrders[callNodeId]`: entry `i` is the written index of the argument for parameter `i`. A call in parameter order (all positional, or named in declaration order) has no entry and emits the same C as the positional call.
- Helpers in `selfhost/src/check.nra`: `hasNamedArgument(arguments)`, `orderArguments(checker, nodeId, arguments)` (the arguments in parameter order) and `findParameterIndexes(checker, nodeId, count)` (the parameter of each written argument).
- Matching: `matchNamedArguments` in `selfhost/src/calls.nra`, called from `checkFunctionCall` (plain, generic and imported module functions, including std modules written in Narra such as `std/option`) and `checkTraitCall`. Generic inference, `checkArguments`, `checkMutableArgument` and the exclusivity checks run on the ordered list.
- Callees without parameter names (`Ref.Value`, `Ref.Builtin`, `Ref.Constructor`, `Ref.Distinct`) are rejected after the check by `checkNamedArgumentsTaken` in `exprs.nra`.
- Passes that pair arguments with parameters use the ordered view: effects `walkCall` and `findCallResultType`, `emitCall` (plain, instance and impl calls), `emitTraitCall`, `isSelfCall`, `emitTailJump`. Order-insensitive passes (use collector, closures, id check, const checks, typesdump) read the written order.
- Emitter: `holdReorderedArguments` (`emitexprs.nra`) evaluates into a temporary each argument that a later-written argument is passed before, plus the arg-eval-order rule (earlier arguments before one that may write). The self tail jump evaluates in written order and assigns in parameter order.
- SPEC.md section 2 (Functions) describes named arguments. DECISIONS.md D12 no longer rules out labeled arguments (points to D87); D87 lost "Not built yet". No Differences-from-TypeScript row and no compiler/README.md difference.

## New diagnostic codes

| Code | Message |
| --- | --- |
| E0328 | `%s has no parameter %q`, and with a fix `...; did you mean %q?` (rename fix when exactly one unpassed parameter is within 2 edits and under half the name length) |
| E0329 | `parameter %q is passed twice` |
| E0330 | `a positional argument cannot follow a named one; name it, or move it before the named arguments` |
| E0331 | `this call passes no argument for parameter %q` (one per missing parameter, only in calls with a named argument; all-positional calls keep E0316) |
| E0332 | `%s has no parameter names; pass its arguments by position` |

Too many arguments in a call with named ones reuses E0316, unless a duplicate was already reported.

## For named-args-query

- The written order is in the AST; the parameter for each argument is `findParameterIndexes(checker, callId, count)[writtenIndex]`, and the callee is `getRef(checker, callId)` (`Ref.Function`, `Ref.Implementation` or `Ref.TraitFunction`). The parameter list is `checker.functions[index].declaration.parameters` (or the trait or impl function declaration).
- An `Arg` with `name: Some(n)` and `pos` at the name token is what a rename-parameter query must rewrite; the name spans `pos` to `pos.column + charCount(n)`.

## Tests

- `testdata/named_args.nra` (+ `.out`): reordering, mixing, `mutable`, generic callee, imported generic callees (`Options.withDefault`, `Options.map` with an effect variable), a trait function, printed side effects in written order, a block that writes a variable read by an earlier argument (int and string), a self tail call with named arguments.
- `testdata/errors/named_args_{unknown,twice,positional_after,missing,no_names,mutable}.{nra,err}`.
- `testdata/diagnostics/named_args_close_name.nra` (E0328 fix, applied by the habits suite).

## Deviations

- Edited `check.nra`, `effects.nra`, `emittail.nra`, `emittraits.nra`, `exprs.nra` and `diagnostics.nra` besides the shared `calls.nra` and `emitexprs.nra`; each edit is small (see the diff).
- No package sample for a user module callee; `std/option` covers the imported path (same `checkModuleCall` code).
- No Go differential test (Go is retired).

## Gotchas

- `holdReorderedArguments` is conservative: a reordered non-literal argument is held even when nothing later could change it (an extra temporary, plus a retain/release for a reference-counted name).
- A new pass that pairs call arguments with parameters must go through `orderArguments`.
