# Outcome: func-type-effects

## Built

- `selfhost/src/types.nra`: `Type.TFunc` has `effects` and `binder`; `typeKey` compares them, `plainTypeKey` does not.
- `selfhost/src/resolve.nra`: function types resolve their `uses` list (sorted, with the binder, as for generators).
- `selfhost/src/exprs.nra`, `check.nra`, `calls.nra`, `stmts.nra`, `stdcalls.nra`: `expectType` accepts a function with the same signature and records a `FunctionStore`; private-function parameters, builtin callbacks and assignments to locals are loose.
- `selfhost/src/effects.nra`: `checkFunctionStores` after the fixed point; calls through any function value charge the effects of its type (`findTypeEffects`, with the call's bound variables). The written-type machinery of PR #98 is deleted.
- `selfhost/src/cnames.nra`, `emit.nra`: C names and the name table ignore effects; emitted C is unchanged.
- D86 in `DECISIONS.md`; `compiler/README.md` updated.
- Tests: 43 `testdata/errors/func_effects_*` tests from the review cases, `testdata/func_effects.nra` (clean).

## Deviations

- Review cases 14b, 25 and 30 are not valid programs (tuple indexing, an uncalled parameter, a missing `Arrays.get`); they are not tests.
- A type argument inferred from an arrow or a named function takes a type with no effects, so `Holder { value: loud }` with an effectful `loud` is rejected (case 37); the type must be written.
- Messages show `uses ...` on function types in `expected X, found Y`, except variables of another function.

## Stubs left

- None.

## Gotchas

- Instances of one generic (struct, enum, function) with function type arguments that differ only in effects get the same C name. Only the composite types and function type structs are deduplicated; a program with two such instances would emit a duplicate C definition. No sample does it.
- Locals that hold a function keep the union of what they are assigned (as before); they are not checked against their type on assignment.
- Stores in generic instance bodies are not checked (the open body is).

## Decisions

- D86, see `DECISIONS.md`.

## Proposed learnings

- None.

## Verification

- Review cases: of 55 probes, 51 valid ones: 48 rejected with `Function ... uses ...` (calls) or `expected ..., found ... uses ...` (stores), 3 clean (42, 49, 52); 14b, 25, 30 invalid.
- `uses` lists added: 4 (struct fields in `fmtprinter.nra`) in selfhost/src, 0 in std/src, plus 4 in test programs (2 files).
- Emitted C compared with the baseline compiler for the `testdata/*.nra` samples that emit with both: 79 same, 0 different (3 skipped, not emitted by the baseline).
- `sh scripts/test.sh`, `sh selfhost/bootstrap.sh`: see the commit.

## Size

- Estimated M, actual M.
