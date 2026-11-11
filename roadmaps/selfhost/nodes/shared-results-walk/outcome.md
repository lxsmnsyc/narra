# Outcome: shared-results-walk

## Built

All in `selfhost/src/sharedresults.nra`; no other file changed.

- **A body that cannot report is not walked.** `checkSharedResults` returns at once when the body has no arrow function and either its result type cannot hold a tracked value (`canShareType`) or none of its parameters can (`hasSharingParameter`). `checkSharedResultsIn` (tests) returns when the body has no arrow function. Without arrows only a parameter can be shared, and `collectShared` already stops at any value whose type holds nothing tracked, so these bodies never reported.
- **`given` is a linked list, not copied arrays.** `addGiven` copied the whole array of a local on every push, so a parameter such as `mutable checker` passed to many calls cost a quadratic copy. `SharingFacts` now has `givenFirst`, `givenLast` and a `givens` array of `GivenValue { value, next }`; the order of the values is kept.
- **`noteMutableArguments` allocates less.** `findHeldKey` caches the held-type set per type key and returns the key; `haveCommonHeldTypes` reads both cached sets without `copy`. `canHoldTracked` skips scalar types before any key is built.
- `findSharedArgument` and `findExpressionStart` keep their signatures.

## Proof

- Base: feat/batch-7 8eba5970, compiler built from the seed, Apple M2, release build (`-O2 -DNARRA_RELEASE`), `/usr/bin/time -l` instructions retired of `narrac emit selfhost`, runs of the binaries alternated, machine loaded by other workers.
- Cost of the check on the base: the base binary against a binary whose `checkSharedResults` and `checkSharedResultsIn` return at once, 5 runs each, median: 2,199,337,809 against 2,119,540,643 (**+0.080G, +3.8%**), above the 2% line.
- Where it went (3 runs each, disabling one part of the walk in turn; each step moves by up to 0.03G with the load): `noteMutableArguments` about 0.04G, the walk and its tables about 0.04G, `checkFunctionReturns` about 0.02G, `addGiven` about 0.015G.
- After the change, 3 runs each, median, same input, binaries run in turn: base 2,147,686,812, branch 2,101,351,139, check off 2,073,442,287. The check costs **0.028G (1.3%)**, down from 0.080G; `emit selfhost` is about 0.046G (2.1%) cheaper than the base.
  - Steps (medians of one 3-run set each): givens as a list -0.013G to -0.024G, held-type keys -0.005G, the two body gates -0.03G and -0.03G (the numbers move with the load).
- No `sample` was used; the parts were located by instruction counts.
- Diagnostics: `check --json` of every `testdata/*.nra` and `testdata/errors/*.nra` (681 inputs; 11 E0976 lines in 6 of them), plus `selfhost`, `selfhost/tests` and `std`: base compiler and branch compiler give byte-identical output (`cmp`, 0 differences). The branch compiler emits the same C for `selfhost` as the base compiler (fixed point holds).
- `sh scripts/test.sh`: PASS (146s, 0 failed).

## Deviations

- The brief proposed `sample`; instruction counts with parts disabled were enough.
- Memory not re-measured: the change only removes copies and skips bodies.

## Stubs left

- About 0.028G (1.3%) of the walk remains, mostly the walk of bodies that can report and `noteMutableArguments`. It is under the 2% line, so no further change.

Landed on feat/batch-7 as 9c2733aa.
