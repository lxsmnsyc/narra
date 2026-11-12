# Log: contract-logic-design

## 2026-11-12

- Branch `spike/contract-logic-design`, reset to `feat/batch-11` (20f9a487). D102 did not exist. narrac built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/contract-logic-design/out` (fixed point, seed current).
- Today's behavior, checked with samples in `/private/tmp/claude-501/contract-logic-design/samples/`: `requires x < 0 || x > 10`, `requires !(x < 0)` and `requires low <= high` are E0984; `ensures Some(>= 0 && < 64) | None` is a parse error (`expected ")", found "&&"`).
- Two `ensures` lines (`Some(>= 0) | None`, `Some(<= 9) | None`) give the caller both bounds: the bound name is a valid shift count (`lines.nra` runs, prints 32). Lines with two alternatives of the same constructor accept `return Some(50)` against `Some(>= 0) | Some(>= 100)` and `Some(< 10) | Some(< 110)` (`mixed.nra`): the one case where lines are weaker than `&&`.
- `selfhost/src` and `std/src` have no `requires` or `ensures` line. Wrote `count.py` (text heuristics) for checks that read integer parameters: 70 checks, 42 entry guards, 7 two-parameter comparisons (2 reject input), 1 `||` guard (negates to an interval), 12 `&&` guards (none rejects input, read by hand), 0 `!` over a comparison, 24 integer `Option`/`Result` results (2 bounded on both ends).
- Prover facts: one `low`, `high`, `nonzero` per local in `checker.localNames` (`operands.nra` `readFact`, `writeFact`). Narrowing gives nothing for a true `||` (`narrowByBinary` returns), so a caller cannot prove a `||` requirement from its own `||` check under any interval design.
- Under D92 wrapping, `low <= high` does not make `high - low` non-negative, so the brief's example for question 2 does not hold. A relation fact would only be passed on to other contracts.
- `emit selfhost` base: 2.434G and 2.411G instructions retired (two runs, `/usr/bin/time -l`).
- Wrote D102 (before D103) and rewrote the contract-logic brief (size M, no new codes, no Pattern change).

## 2026-11-12, after the gate

- Maintainer: build `||` with the hull, and the hull in checks, as proposed. Build relation facts now instead of deferring them. Reject `!` with the rewrite. Keep several `ensures` lines.
- Fresh review fixes. The check-path hull built each `u64` side as an int, so `x == 9223372036854775813 || x == 1` would join to `x <= 1`. Each side is now built as `refineFact` builds it (a `u64` bound from 2^63 up gives the unknown fact), and `testdata/errors/divisor_or_u64.nra` is in the acceptance. The hull of `x < 0` and `x > 10` is every int except zero, not every int. The array-patterns dependency now names its reason. The `ensures` gap is for any two alternatives that bound the same place. `count.py` gained `negation` rows: 6 lines put `!(` before a comparison, 3 of them in checks.
- Relation design: a set of possible orders (less, equal, greater) per pair of integer locals in `localNames`, with a partner list per local so `endNarrowing` ends relations. Bounds flow one step across a relation. Calls record the relation set between their argument names. No new error code: E0984 and E0985 get new details.
- contract-logic brief: size L in three parts, `patterns.nra` added to the shared files.
