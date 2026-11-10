# Log: batch-3 review fixes

- Guard on an or-pattern: `emitGuardedAlternatives` (emitmatch.nra) handles a guarded arm whose pattern is a top-level or-pattern that binds names. Each matching alternative binds, runs the guard and releases its bindings when the guard fails; the first success selects the arm. Test `testdata/or_pattern_guard`. Limit: an or-pattern nested inside another pattern (`Some(a | b)` with bindings) under a guard still binds from the first matching alternative.
- Destructuring: `checkDestructure` (match.nra) reports E0974 for alternatives; this covers `const`, `let` and `for ... of` (the parser desugars the latter). Samples `errors/or_pattern_destructure`, `errors/or_pattern_for_of`.
- E0971 on `+%=`, `-%=`, `*%=` now quotes the compound operator.
- The seed is not regenerated.
