# Adversarial review: batch-12

Fresh reviewer, 2026-11-13. All eight findings checked against the code and applied.

1. Generated files missing (`stdtext.nra` from `std/src`, `runtime*.nra` from `selfhost/runtime/*.c`, the `runtimetext.nra` case). Applied: each async node owns its `.c` source and generated module, lists `runtimetext.nra` and `stdtext.nra`, and its brief says to regenerate with `--suite generated --update`.
2. `std/src/async/**` does not match the flat std layout. Applied: `std/src/async__*.nra`.
3. `Async<T>` hooks (`check.nra`, `types.nra`, `resolve.nra`, `traits.nra`, `sharedwrites.nra`, `calls.nra`) unlisted. Applied to async-core; D109 decides builtin type or special-cased std type.
4. Entry path (`emit.nra` `emitMain`, `emitter.nra` `findEntry`) and `emitexprs.nra`, `emitinline.nra`, `inlining.nra` unlisted. Applied; `driver.nra` dropped.
5. std-async-io needs async-api (cancel). Applied as a hard edge.
6. arrow-arg-order and review-nits touch the same functions as async-core. Applied as notes (second to land rebases); no edge, so the fixes still start in wave 1.
7. async-design edits the `diagnostics.nra` header, and SPEC mentions `Async.run`, `Async.spawn` and `Task<T>` outside section 8. Applied.
8. The parser does not accept `async function` (E0930 for nested ones), and the D85 rule in async bodies was not in async-core's acceptance. Applied.

Kept: `diagnostics.nra` has 7 editors. It is a table with one row per code in reserved ranges; rebases are mechanical, as in batch-11.
