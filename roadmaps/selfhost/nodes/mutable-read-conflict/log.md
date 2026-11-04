# Log: mutable-read-conflict

- Checked first: no existing work. The old rule (E0319, `checkExclusive`) only compares the arguments of one call.
- Message needs "in one call ... rest of the expression", so E0929 (reserved) instead of E0319.
- Go scan runs at the end of `Checker.stmt` (deferred), so locals are resolved. Selfhost does the same at the end of `checkStatement`.
- Uses carry the chain of enclosing calls. A pair is skipped when one use is inside the other's call (E0319 covers it), when the uses are on the two sides of `&&`/`||` (the emitter finishes the left side first) or in different match arms, and when a callee use meets a mutable argument inside it (E0609 covers it).
- Selfhost hit: only `parser.nra` (two `newId(mutable parser)` in one call). Fixed in a first commit.
- Tried moving the narrowing end after the statement so the option case reports only E0929. It broke the four `let_narrowing_same_statement_*` tests of #68, so I kept the order; the option test shows both errors.
- The fix needs source text and the call end: `File.Src`, `CallExpr.End`.
- Selfhost has no fixes, so the fix is Go only.
