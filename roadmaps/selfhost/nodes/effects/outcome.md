# Outcome: effects

## Built

All in `compiler/internal/narra/`. Logic: `check_effects.go`, run from `CheckPackage` (`modules_check.go`) after `finish`, only when there are no other errors.

Syntax:
- `<T, effect E>` type parameters: `FuncDecl.EffectParams` (`typeParamsEffects` in `parser.go`). `effect` is contextual.
- `uses a, b` on function types: `TypeExpr.Uses`. A list inside a parameter list ends at `, name:` (`paramAfterComma`).
- `FuncDecl.Origin` is set on generic instances (`cloneFunc`); the pass reads the template.

Rules:
- Effect set of a function = effects of builtins it calls (`Builtin.Effects`), of declared functions it calls, and of function values it calls. Also `alloc` from array literals (not tuples), template strings with parts, string `+`, and enum constructors with fields.
- Exported functions: body effects must be inside the declared `uses`, else "Function `f` uses `io` (through `print`) but its signature does not declare it; write `uses io`". Callers use the declared set. The entry function (`Program.Entries`) is exempt and has every effect.
- Private functions and arrows: inferred (fixed point over all bodies, so recursion works). An arrow's set is added to its enclosing function only when called.
- Effect variables: a call substitutes the variable by the effects of the arguments passed to the parameters whose function type names it. A parameter whose function type has no variable accepts only arguments within its listed `uses`; otherwise an error at the argument.
- Signature checks: unknown effect names, and an effect variable in `uses` that no parameter's function type names.
- Function-value effects are known for: a function-typed parameter (its `uses`), a local bound to an arrow or named function, an argument written at the call.

Tests: `testdata/effects.nra` (+ `.out`: pure and io callbacks with one `map`, private inference), `testdata/errors/effects_*.nra` (5: io undeclared, io through a private chain, io callback to a pure parameter, unknown effect, unbound variable).

## Deviations

- Existing samples had to change because exported functions and callbacks now need effects: `testdata/closures.nra`, `closures_nested.nra`, `packages/closures/src/fn.nra`, `packages/app/src/text.nra`, `packages/app/src/util/list.nra`, `packages/traits/src/report.nra` (added `uses` / `<effect E>`).
- `compiler/README.md`: Supported and Not supported lines about effects.
- Edits outside owned paths: `modules_check.go` (one call), `ast.go`, `parser.go`, `clone.go`.
- No registry entries needed changes: io-std already declared effects.

## Stubs left

- Function types carry no effects in the type checker. Calls through struct fields, call results and arrays of functions count as pure; passing such values to parameters is not checked.
- Trait function calls resolved to user `implement` bodies are not followed.
- `unsafe` has no source yet (node c-interop). `uses` on arrow functions' own result types and on `async` is not parsed.
- A `uses` list in a function-type return position (`(): () => int uses io`) binds to the inner type.
- `uses E` before a `mutable x` parameter: `paramAfterComma` does not recognise the `mutable` keyword token.

## Gotchas

- Newlines inside call parentheses are ignored by the lexer, so multi-statement arrows cannot be call arguments (existing limitation). Bind them to a const.
- New exported functions in samples need `uses alloc` when they build strings or arrays.
- Later nodes that add builtins must set `Effects`; the pass reads `builtinsByName`.

## Proposed learnings

- Set `Builtin.Effects` on every std builtin | the effect pass reads it; a missing entry makes the builtin pure | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (all samples under ASan/UBSan and release).

## Size

- Estimated M, actual M.
