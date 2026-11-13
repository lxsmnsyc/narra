# Outcome: contract-logic

## Built

- D102 parts 1 to 3, all in one commit.
- Part 1, `||` clauses. `Requirement` (`selfhost/src/check.nra`) has `clause` and `other`. `collectRequirements` (`selfhost/src/contracts.nra`) splits a condition into clauses at `&&` and each clause into sides at `||`. At the start of the body, `assumeClause` applies each clause: a one-comparison clause refines the fact as before, and a clause over one parameter narrows it by the hull of its sides (`buildSideFact` and `joinFacts`, then `narrowFact`). `proveCall` proves each clause from any one of its comparisons. `reportFailedClauses` reports one E0985 for the failed one-comparison clauses of an argument, joined by ` && ` as before, and one E0985 per failed `||` clause. When every parameter of a clause got a constant, the message names the values: `5 does not satisfy ...`, or each parameter in parameter order, `x = 5 and y = 1 do not satisfy ...` (one new E0985 format row, `%s do not satisfy `requires %s` of %s`).
- Hull in checks (D92 amendment). `narrowByHull` (`selfhost/src/operands.nra`) runs for a true `||` or a false `&&`, which negates each side with `negateComparison` (now exported). It narrows only when every side compares the same integer local with a constant. `buildSideFact` gives the unknown fact for a `u64` bound from 2^63 up, as `refineFact` does.
- Part 2, relations. `findRelationRequirement` accepts two different integer parameters of the same type. The fact is a set of orders (bits 1 less, 2 equal, 4 greater) stored in `localNames` under `relation|<a>|<b>`, with the partner list in `relations|<a>` and `relations|<a>|<k>`, all written through `bindName`. The key helpers live in `selfhost/src/patterns.nra`, and `endNarrowing` ends every relation in the local's list. `narrowByBinary` records a relation for any comparison of two integer locals of one type (`recordRelation`), and `assumeClause` records each one-relation clause. The bound step (`passBounds`, `passBoundsToPartners`) runs when a relation is recorded and after `refineFact` and `narrowFact`. It goes one step, never passes an unbounded end, skips locals an arrow captures, and does not pass a `u64` end from 2^63 up (`isPassableEnd`). For calls, `recordContractCall` stores one set per relation requirement in `ContractCall.relations` (`findArgumentRelation`: constants with the `u64` order, bounded intervals, the same local twice, the relation fact), and the proof is a subset test (`isRelationProven`).
- Part 3, `!`. `describeUnsupported` prints `write <condition>` through `rewriteNegation`, which pushes the negation inward and adds parentheses where `&&` and `||` alternate. The other new E0984 details are: `split it into clauses joined by &&; && inside || is not supported`, `compare two parameters with no arithmetic; + and - wrap`, and `this compares the parameter with itself`.
- E0984 fixed text (`selfhost/src/diagnostics.nra`): `requires compares integer parameters with integer constants or with each other, joined by && and ||; %s`. No new code.
- Samples: `testdata/contracts_logic.nra` and `testdata/contracts_relations.nra`. `testdata/proven_operands.nra` gained `eitherSide`, `falseBoth` and `relationShift`.
- Error tests: `testdata/errors/contracts_logic_{unproven,constant,and_under_or,not,not_or}.nra`, `contracts_relation_{unproven,ended,overlap,offset}.nra`, `divisor_or_u64.nra`, `divisor_or_two_names.nra` and `shift_relation_u64.nra`. In `contracts_unsupported.nra`, the `||` case is now `&&` under `||` and the two-parameter case is now an offset. The goldens `contracts_unsupported.err` and `contracts_names.err` were rewritten for the new text.
- Docs: SPEC section 13 (the condition rule, call proof, body facts, one bound per `ensures` line) and section 4 (the D92 checks paragraph: hull, relations, bound step). Also `compiler/README.md` (the `requires` entry) and a D102 "Built" line.

## Proof

- Not a fix node. Feature proof: on base c2eb28db, `testdata/proven_operands.nra` fails with E0979 at `eitherSide` and `falseBoth` and E0980 at `relationShift`. On the branch it runs.

## Deviations

- `testdata/packages/errors/contracts_alias_parameter/expected.err` is outside the owned and shared paths. It holds the E0984 text, so its golden changed with the message.
- `testdata/errors/shift_relation_u64.nra` has a second function. In the brief's case (`limit != 0`), `limit` has no end at all, so the test passes even without the `u64` rule. The second function uses `if (limit < 0)`, which leaves a high end of -1 (2^64 - 1 as a u64), so this test fails if `isPassableEnd` is removed.
- The relations sample proves a `FixedArray<int, 4>` index (`high < 4`) instead of `FixedArray<int, 64>`, so the array literal stays short. The shift count sample uses `high < 64`.
- The constant message for a clause over two or more parameters names each value with its parameter, in parameter order (fresh review). It needed one new E0985 format row, `%s do not satisfy `requires %s` of %s`.
- The offset detail also covers arithmetic over two parameters on one side (`x - y >= 0`, fresh review nit).

## Stubs left

- None.

## Gotchas

- `endKey` also clears the saved value of a key in the scope undo list, so an ended relation stays ended after the scope closes. A pair key with value -1 means no relation; 0 is the empty set, which proves everything.
- Each comparison between two integer locals in `selfhost/src` now writes a relation. A future change that makes `refineFact` or `narrowFact` heavier is multiplied by the bound step.
- `narrowFact` now passes bounds through relations. Callers that narrow on purpose without that step should use the private `intersectFact`.

## Proposed learnings

- A new path that orders `u64` values must skip ends from 2^63 up, because they are negative as an int | the brief's `u64` test passed without the rule until a case with a negative end was added | selfhost/src/operands.nra | yes, with a `u64` error test whose fact really has a negative end

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/contract-logic/out`: stage 1 and stage 2 emit identical C (195664 lines after the rebase). The seed was not regenerated, per the batch rule.
- Fresh review fixes, after rebasing on feat/batch-11 c440978d: errors PASS 652 cases, codes PASS 1.
- Targeted suites with the built narrac (before the review): errors PASS 642 cases, codes PASS 1, sources PASS 1, programs `--filter contracts` PASS 10, programs `--filter proven` PASS 2.
- Soundness checks in scratch (`/private/tmp/claude-501/contract-logic/edge.nra`): a relation ends at an assignment in a loop and at an arrow that assigns the local, and it is not read inside an arrow that captures the local. Each case is E0985.
- `narrac emit selfhost` instructions (release build, median of 5, `/usr/bin/time -l`, both compilers on the base source c2eb28db): 2,469,948,363 before, 2,481,740,089 after, +0.48%.
- `sh scripts/test.sh` on the final tree (rebased on c440978d, with the review fixes): `test.sh: PASS, 7466 cases in 243s`

## Size

- Estimated L, actual L.
