# Outcome: diagnostics-polish

## Built

All four items reproduced on feat/batch-9 (1956faac) and are fixed.

1. Mutual typeof cycle: `Checker.typeofChain` (check.nra) lists the names being resolved; `reportTypeofCycle` (resolve.nra) prints `typeof g1 is part of the cycle g1 -> f1 -> g1; write one of the types out` (E0981, new row). A direct self reference keeps the old message.
2. Constant without a written type in a signature: new code E1003 (next free after E1002); E0982 keeps the other typeof cases. SPEC.md updated.
3. Narrowing cascade: `expectType` (exprs.nra) skips a void mismatch at a position that already has an error (`hasErrorAt` in check.nra). 16 existing `.err` goldens lost only their cascade line "expected X, found void".
4. Long lines: the lexer (`POSITION_LIMIT`, lexer.nra) rejects a token at line or column >= 1048576 with new code E1004. KEY_STRIDE stays (widening only moves the aliasing point).

New codes: E1003, E1004 (orchestrator should record them).

## Proof

- base (1956faac, compiler built from a clean checkout), `narrac check FILE`:
  - cycle.nra (f1/g1): `1:16: typeof g1 refers to its own declaration`
  - constsig.nra: `3:23: the type of constant "limit" is not known here` (E0982)
  - narrowing_shadowed_panic.nra: second line `8:16: expected int, found void`
  - 1048577-line file with a token on the last line, and a token at column 1048591: no error.
- branch: same files give the cycle message `g1 -> f1 -> g1`, E1003, one error only, and `a source line or column must be below 1048576` (E1004) at 1048581:1 and 2:1048591.
- Tests: testdata/errors/typeof_cycle, testdata/lex/position_limit_line and position_limit_column (each about 1 MB of blanks), updated goldens.

## Deviations

- Touched check.nra (new field and helper), lexer.nra, exprs.nra, stmts.nra, resolve.nra, which are outside the listed shared paths, with minimal edits.
- DECISIONS.md (type-of amendments) and compiler/README.md now say E1003 for the constant case (approved by the coordinator).
- The limit test is an edge-suite case (`lex/position_limit`, `selfhost/tests/src/edge.nra`) that writes its inputs into a scratch folder; no large samples are committed.

Review fixes: `hasErrorAt` also compares the file (package test `error_hidden_across_modules`); constant cycles now name the cycle (checked top-level constants in exprs.nra, local constants in stmts.nra; tests `typeof_const_cycle`, `typeof_const_cycle_mixed`); `narrac explain` text for E1003 and E1004; one SPEC line for the position limit; `POSITION_LIMIT` is exported from lexer.nra and used for `KEY_STRIDE`; `CHAIN_SEPARATOR` constant.

## Stubs left

- None.

## Gotchas

- The seed does not know E1003/E1004 until the seed update; `sh scripts/test.sh` builds from the seed through bootstrap stage 3, which is fine.
- The checks of the cycle for a top-level constant run in `checkConstant` (exprs.nra), not in `declareConstTypes`.
- A token past the limit now fails to lex, so any generated file with a 1M-column line is rejected.

- `expectType` drops any void mismatch at a position that already has an error in the same file, not only the narrowing one (accepted trade-off).

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: see the result in the commit message / report.
- `narrac fmt` run on all edited `.nra`.

## Size

- Estimated M, actual M.
