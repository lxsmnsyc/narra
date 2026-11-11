# Log: contracts-ensures

- 2026-11-11: branch feat/contracts-ensures reset to feat/batch-7 (dff84e7d). Nothing of ensures exists yet.
- AST: `FuncDecl.ensures: Array<Pattern>`, new `Pattern.PatCompare(id, pos, operator, bound)`. The parser makes a comparison pattern only while `parser.contractPattern` is set (ensures lines), so `match` and `is` report a plain parse error.
- The new variant needs cases outside `shared:`: trivia (getPatternId), emitmatch (two switches, nothing to emit), querysites, typesdump. Recorded as deviations.
- Import cycle: match.nra calls the caller-side seeding in contracts.nra, so contracts.nra cannot import match.nra for resolvePattern. Added `Walker.resolvePattern` (check.nra trait, body.nra impl).
- Trivia key `e<id>` collides with getIdKey (`e<id>` for every expression and pattern), so the ensures line key is `n<pattern id>`. Found by a trailing comment that fmt dropped.
- Requires and ensures lines each skipped the trailing newlines, so an `ensures` after `requires` was not seen; the final skip moved to parseFuncDecl.
- Soundness choices: holder (`const r = f(); r is P`) only for a `const` whose type is not a contents type (a const struct can be passed `mutable`, checked with a probe); `let` results get integer facts that the store ends on assignment; generator results seed only `for` loop names; caller or-patterns seed nothing; comparisons never proven from the unbounded ends (isComparisonProven); `> max` and `< min` add no fact instead of wrapping (refineBounds).
- Rebased onto feat/batch-7 after return-trailer (bce21638): no conflicts; targeted suites pass.
- Maintainer: add `!=`. Only the parser changed (isComparisonToken); isComparisonProven and refineBounds already treat `!=` as asked (non-zero for 0, a range that excludes the bound, never the unbounded ends). Samples added to contracts_ensures and contracts_ensures_unproven.
- Fresh review fixes (after rebasing onto place-ids): u64 `!= 0` keeps non-zero for callers (narrowFact drops only the negative ends), E0987 for a u64 literal pattern from 2^63 up, the bound is parsed with contractPattern cleared, `requires`/`ensures` on a trait signature is E0987 "needs a function body", findHolderKey renamed createHolderKey.
- Rebased onto build-targets (c96227f7): one conflict in the Checker fields of check.nra, both kept. Review round 2: E0987 also for a literal pattern equal to an unbounded end (isUnboundedEnd in operands.nra); `requires` on a trait signature is E0984.
- Dev build: seed compiler emits selfhost, `cc -O0` (9 s). Targeted suites errors, dumps, fmt, codes, node-ids, trivia, fmt-stable: pass. spec, programs, packages, lsp, query: pass.
