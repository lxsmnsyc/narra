# Log: nested-or-guards

- Proof on feat/batch-4 (81638c3): `NARRAC=<narrac built from the seed> sh scripts/test.sh --filter or_patterns_nested_guard` failed. Output differed at line 1: `small 1` instead of `big 10` (the first matching nested alternative failed the guard and the arm was skipped).
- Fix: a guarded arm with bindings and an or-pattern anywhere in the pattern is expanded into the or-free patterns it stands for (`expandOrPatterns`, `combinePatternOptions`). `emitGuardedAlternatives` then tries each in order, as it did for a top-level or-pattern.
- On the branch the same filter passes. The expansion is a product, so k nested lists of n alternatives give n^k patterns.
