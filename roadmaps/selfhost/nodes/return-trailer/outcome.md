# Outcome: return-trailer

## Built

- `hasReachableEnd(block)` in `selfhost/src/pathends.nra`. Both trailer sites (`emitFunction` in `emitconsts.nra`, `emitClosureFunction` in `emitclosures.nra`) write the `narra_panic ... return (T){0};` trailer only when it is true.
- A body ends when a statement is a `return`, a `panic` call, an `if`/`else` whose branches both end, a block that ends, or an endless `while true`/`for` without a leaving `break`. The helper walks the AST, so it sees every branch of a build-condition `if`.
- A `match`, and a `return` of a `match`, do not count as ending the body, so the trailer stays after them.
- `narra_panic_str` is now `_Noreturn` (`selfhost/runtime/runtime.h`, `selfhost/src/runtimecore.nra`).

## Proof

- base (`feat/batch-7`, `narrac emit selfhost`): 1878 lines with "function ended without returning" (4 are the compiler's own string constants), 186150 lines of C, 0 `-Wreturn-type` warnings with `cc -Wall -Wreturn-type`.
- branch: 139 such lines (same 4 constants), 182829 lines of C (3321 fewer), 0 `-Wreturn-type` warnings. Stage 2 and stage 3 emit identical C.

## Review fixes

- The panic check resolves the call's ref (the prelude `panic` builtin) in the checker rule (E0969) and in `hasReachableEnd`; a parameter or local named `panic` no longer ends a body. Error samples: `testdata/errors/missing_return_shadowed_panic`, `missing_return_local_panic` (the base compiler accepts both).
- `while (true)` and `for (...; true; ...)` are emitted without a condition test; emit case `endless-loop/return-type` compiles with `-Werror=return-type`.
- After rebase onto the current `feat/batch-7`: 141 trailer lines (4 are constants), 181857 lines, 0 `-Wreturn-type`. The counts under Proof are against the earlier base.

## Deviations

- No Checker record: the emitter calls the helper on the AST. The record the brief suggested would be the same analysis stored.
- Did not reuse `findBlockOpenPath`: the checker treats a final `match` and a `return match` as ending the body, but the emitted C is a chain of tests that `cc` cannot prove exhaustive (31, then 11, new `-Wreturn-type` warnings). The emitter's own helper is stricter.
- Touched `runtime.h` and `runtimecore.nra` (outside `shared:`) for `_Noreturn` on `narra_panic_str`; one word each.
- `seed/narrac.c` is not regenerated (older than the source, as before).

## Gotchas

- 135 trailers remain, mostly after a final `match`. Making them go means emitting the exhaustive `match` as a `switch` with an unreachable default, or `__builtin_unreachable()`.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds (stage 2 = stage 3).
- `cc -Wall -Wreturn-type` on `narrac emit selfhost`: 0 warnings, base and branch.
- `narrac fmt --check` on the changed files.
- `NARRAC=<built narrac> sh scripts/test.sh`: PASS (8 shards, 456s).

## Size

- Estimated S, actual S.
