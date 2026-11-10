# Outcome: body-id-ranges

## Built

- The parser records the ids it hands out: `FuncDecl`, `ConstDecl` and `TestDecl` (`selfhost/src/ast.nra`) have `firstId` and `endId` (default 0), set in `parseFuncDecl`, `parseConstDecl` and the `test` branch of the declaration parser (`selfhost/src/parser.nra`). The range is a superset of the body's node ids and may have gaps.
- `check.startBody` (`selfhost/src/check.nra`) takes `firstId` and `endId` and sizes the body's tables from them. The three callers pass them (`body.nra` for functions and tests, `exprs.nra` for constants).
- Field defaults are bodies without a parsed range (the checker's `Field` copies only the expression). They pass `endId = -1`, which makes `startBody` call the new `scanBodyIds`, the old walk. `narrac pkg --ids` still uses `idcheck.collectIds`.

## Proof

- base: 91233c52, `/usr/bin/time -l narrac emit selfhost`, instructions retired, both compilers built with `sh selfhost/bootstrap.sh <dir>` and emitting the same input. Load average 8 to 16. Spread within one side about 4% (base 1739 to 1850 M).
- branch: lower than base in 12 of 18 complete pairs. Medians: round 1 (8 pairs) 1773 M vs 1764 M; round 2 (15 base, 11 branch runs) 1758 M vs 1745 M; round 3 (10 pairs) mean difference 18 M. About 0.5% to 1% fewer instructions, well under the 2.8% the profile placed on the walk.
- The gain is below the loaded run-to-run spread. It is positive in most pairs, but the brief's rule (gain must exceed the spread) is not met on this machine. The orchestrator decides whether to keep it; the change is small and has no behavior change.

## Deviations

- Gain is smaller than the proposal's estimate of 3%: the profile counted time, and the walk is cheap in instructions.
- Field defaults keep the walk (see Built).

## Stubs left

- None.

## Gotchas

- A `FuncDecl`, `ConstDecl` or `TestDecl` built outside the parser has range 0..0, so its body would get empty tables. Only the parser builds ones that reach `startBody`.
- Emitted C of `selfhost` changes with the branch source (the compiler's own structs gained fields); the check is that the base and branch compilers emit identical C for the same input.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds (stage 2 and stage 3 identical).
- The base compiler and the branch compiler emit identical C for `selfhost` (cmp).
- `sh scripts/test.sh`: PASS (full run at the end).

## Size

- Estimated S, actual S.

## Orchestrator decision (2026-11-10)

Not landed. The branch won 12 of 18 alternating pairs, about 1% fewer instructions against a 4% spread under load, so the gain is within noise. The change adds two AST fields per declaration and keeps the old walk for field defaults, which is more code for no proven gain. The code stays on the local branch `feat/body-id-ranges` (ee250253) if a quiet-machine measurement later shows the 2.8% the profile found in time.
