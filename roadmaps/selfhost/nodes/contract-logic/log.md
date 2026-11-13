# Log: contract-logic

## 2026-11-12

- Branch renamed to feat/contract-logic, reset to feat/batch-11 (c2eb28db). No earlier work: `requires` rejected `||` and relations (E0984).
- Baseline: release compiler from bootstrap, `narrac emit selfhost` on the base source, median of 5 with `/usr/bin/time -l`: 2,467,595,488 and 2,469,948,363 instructions (two runs of the measurement).
- Built all three parts in one pass.
  - `Requirement` gets `clause` and `other`; `ContractCall` gets `relations`.
  - `contracts.nra`: `collectRequirements` splits `&&` into clauses and `||` into sides; `findRelationRequirement`; `groupClauses`, `assumeClause` (hull or relation at body start); `proveCall`, `reportFailedClauses`, `reportClause`, `describeConstantArguments`; `describeUnsupported` gives the `!` rewrite (`rewriteNegation`), the offset, self-comparison and `&&` under `||` details.
  - `operands.nra`: `narrowByHull` for a true `||` or false `&&`, `buildSideFact` (u64 rule), relations (`recordRelation`, `readRelation`, `flipRelation`), the bound step (`passBounds`, `passBoundsToPartners`, run after `refineFact` and `narrowFact`), `findArgumentRelation` for calls.
  - `patterns.nra`: relation keys, and `endNarrowing` ends every relation of the local.
- After: 2,481,740,089 instructions on the base source (+0.48% against 2,469,948,363).
- Soundness checks in scratch (`edge.nra`): a relation ends at an assignment in a loop, at an arrow that assigns it, and is not read inside an arrow that captures it.
- `shift_relation_u64.nra`: the brief's case (`limit != 0`) gives limit no end at all, so it passes even without the u64 rule. Added a second function with `limit < 0`, which leaves a high end of -1 (2^64 - 1 as a u64), so the test fails without `isPassableEnd`.
- `testdata/packages/errors/contracts_alias_parameter/expected.err` holds the old E0984 text; its golden changed with the message (outside the owned and shared paths).
- Fresh review: rebased on feat/batch-11 c440978d (clean). The constant message for several parameters now names each one in parameter order (`x = 5 and y = 1 do not satisfy`), with a new E0985 row. `isOffsetComparison` counts parameters over both sides, so `x - y >= 0` gets the offset detail; error tests extended.
