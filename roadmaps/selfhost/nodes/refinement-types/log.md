# Log: refinement-types

## Proof first

- Base: feat/batch-13 at 7df2cd63, compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/refinement-types/base` (release flags).
- `/private/tmp/claude-501/refinement-types/base/narrac check testdata/refinement_types.nra` exits 1:
  `refinement_types.nra:4:46: expected the end of the statement, found "requires"`.
- Branch: the same sample runs and matches `testdata/refinement_types.out` (debug build with AddressSanitizer, leak check and copy check, and release build).

## Notes

- Node types are kept per body, so the refinement items are not resolved with `Walker.resolvePattern`: `refinements.nra` checks them against the base itself and stores plain tests (`RefinementTest`) in `DistinctInfo`. The emitter, the evaluator and the `from` proof read the tests; the hull is stored as `RefinementHull` and turned into a `Fact` in `findRefinedFact` (operands.nra), since types.nra cannot import operands.nra.
- `check` is told apart from `from` and the unwrap by the callee name (`Expr.FieldExpr(..., "check")`), so no `Ref` variant is added.
- The emitter first held an owned argument in an `ownedTemp`; the statement already releases the owned temporary, so ASan saw a double release (`Method.check(\`PO${"ST"}\`)`). The argument is now a plain `temp`, and `Some` retains it.
- A top-level `const` could not call any distinct `from` (E0192). D116 part 2 assumes it can; `checkConstCall` in exprs.nra now accepts `Ref.Distinct` calls (deviation, recorded in the outcome).
- Divisors and shift counts did not go through `findOperandFact` (`isProvenDivisor`, `isProvenCount`); both now read `findRefinedFact` first.
- After `&&` at a line end the lexer joins the next line, so a trailing `&&` is caught by requiring a comparison or a literal when the next item starts on a later line.
- Rebased on feat/batch-13 0bab063f (threads-process runtime changes) before the checks.
- `emit selfhost` of the base source: base compiler and branch compiler give identical C (223915 lines). Instructions retired, alternating, 3 runs each: base 3.064G, 3.050G, 3.035G; branch 3.065G, 3.022G, 3.022G.
