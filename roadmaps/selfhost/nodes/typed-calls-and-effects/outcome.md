# Outcome: typed-calls-and-effects

## Already fulfilled

The gap in the brief does not exist on feat/batch-4 (81638c3). No code change.

## Evidence

Compiler: `sh selfhost/bootstrap.sh <dir>` (seed is current), then `narrac build` on each program.

- The `Box` program of the brief (`export function pure(box: Box): int { return box.run("x") }`, field type `... uses alloc`): rejected, `Function \`pure\` uses \`alloc\` but its signature does not declare it; write \`uses alloc\``.
- One program with an exported function for each form, none declaring `uses alloc`: array element call `items[0]("x")`, map value call (`items["a"] is Some(callback)`, then `callback("x")`), nested field `outer.box.run("x")`, field of a call result `makeBox().run("x")`. Each gets the same error.
- A private function that calls `make()()` (result of a call, effect `io`) makes its exported caller report `uses io (through privateResult)`.
- `narrac query --json type-of` on the `(` of `box.run("x")` answers `"effects": ["alloc"]`; on the field name it answers the function type `(string) => int uses alloc`. `testdata/query/basic/queries.txt` already pins this (lines 20 to 23).
- Existing error tests: `testdata/errors/effects_indirect_{field,paren,index,call_result,capture,field_var,field_unknown}.nra`, `effect_variable_value_*.nra`, `generator_value_call_alloc.nra`. The clean sample is `testdata/effects_indirect.nra`.
- The README "Not supported yet" list has no F20 line, and the code that does this is `findTypeEffects` in `selfhost/src/effects.nra` (D86, node effect-variable-values).

## Deviations

- The acceptance names error tests `testdata/errors/effects_function_value*`. The existing tests carry the `effects_indirect_` prefix and cover struct field, call result and array element. No test calls a map value directly through a `Some` binding; that form passes by the binding's type (D86).

## For the orchestrator

- `roadmaps/selfhost/inbox.md` line 13 (2026-11-05, call through a struct field or call result) is stale; I did not edit it. Remove it when this node lands.

## Stubs left

- None.

## Proposed learnings

- None.

## Verification

- The programs above, run with a compiler built from the seed.

## Size

- Estimated M, actual XS.
