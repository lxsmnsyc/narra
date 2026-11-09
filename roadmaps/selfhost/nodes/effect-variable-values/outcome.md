# Outcome: effect-variable-values

## Built

- `selfhost/src/effects.nra`: `addValueBoundEffects`, called for calls through a function value (`Ref.Value`). The parameter types of the value's type name the effect variables of the function that made it; each variable listed in that function's `uses` takes the effects of the argument passed there (functions and generators). A variable of the function being walked stays a name.
- D91 in `DECISIONS.md` (option 1), SPEC.md section 7 and `compiler/README.md` updated.
- Tests: `testdata/errors/effect_variable_value_{top,local}.nra`, clean sample `testdata/effect_variable_value.nra` (pure and effectful arguments through a top-level and a local constant).

## Proof

- base: `6478acd`, `narrac build -o out testdata/errors/effect_variable_value_top.nra` and `_local.nra`: both build with no diagnostic although `pure` calls `apply(loud, 1)` and `loud` does `io`.
- branch: the same commands report `Function \`pure\` uses \`io\` (through \`print\`) but its signature does not declare it`; the sample runs and prints the expected output.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- A function value's own type lists no effects (D86), so the binding reads the parameter types of the value's type, not its effects. The function's non-variable effects of a private function used as a value are still not charged by this path; the store check on constants covers the constant case.
- A value whose type came through a generic parameter (type parameters erase effects) is out of scope.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh <dir>`: stage 2 and stage 3 identical.
- `sh scripts/test.sh`: see the commit message and reply.

## Size

- Estimated M, actual S.
