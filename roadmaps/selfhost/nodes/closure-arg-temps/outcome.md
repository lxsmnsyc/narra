# Outcome: closure-arg-temps

Done. `mayWrite` in `selfhost/src/emitexprs.nra` returns false for every arrow expression, block-bodied or not; `arrowMayWrite` is removed. Building a closure runs none of its body, so earlier arguments no longer get a temporary.

- Sample: `testdata/closure_arg_temps.nra` (output `523`, `3`).
- Emitted C for it: on the base, `base + 1` is copied to a second temporary before the call; on the branch the call takes the first temporary directly.
- The checker rejects a closure that writes a variable another argument reads, so the sample's earlier argument reads a different variable than the closure writes.
- The compiler emits the same C for the self-build apart from fewer temporaries (stage 1 and stage 2 match at 183740 lines).

## Deviations

- Inbox entry not removed (the orchestrator does it). The seed is not regenerated; it is older than the source until the batch checkpoint.
