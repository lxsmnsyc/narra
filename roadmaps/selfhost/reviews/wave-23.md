# Wave 23 report: batch-10

Batch-10 runs on the local story branch `feat/batch-10` (not pushed). Every child has landed as one direct commit after a worker check and a fresh review. Two seed updates at the end; `bootstrap.sh` reports the seed current.

## Landed

- destructuring-design: D103. Array patterns are built next (array-patterns, L, refined), with `...rest` as a copy of the tail. Destructured parameters are dropped because they cannot be passed by name (D87). Destructuring assignment is rejected for now. param-destructuring and destructuring-assign are dropped. Gate answered by the maintainer in session on 2026-11-12.
- seed-switch-cleanup: every early exit of `narrac run` goes through `exitWithoutRelease`; the host target matches by `Os` and `Arch`.
- fmt-std-sources: the `sources` suite covers `std/src`; `endsWithBlock` levels a block anywhere in a same-precedence chain.
- copy-check-holders: heap enum values carry a debug-only `copy_held` id (in padding), so credits follow holders. Credits end only when the box is freed.
- checker-emitter-copies: D104. A `const` or `let` holds its value until its block ends, also an enum built from a place, an `is` or `match` binding of a temporary, and a `copy(...)` initializer. 171 sites in `selfhost/src` fixed; unexplained copies in a debug `emit selfhost` went from 1927 to 0. New `copies` suite case runs a debug compiler on `selfhost`. Shared array elements stay a documented gap.

## Checks at the end

- `sh scripts/test.sh`: PASS, 7172 cases. `sh scripts/linux-check.sh`: PASS.
- Debug build of the new seed: `check testdata/arrays.nra` and `emit selfhost` with `NARRA_CHECK_COPIES=1 NARRA_CHECK_LEAKS=1` exit 0.

## Inbox

Closed: the nine entries batch-10 covered, including the two stale ones (`TestSelfhostBuiltins`, the `@unique` sample). Added: `runBuild` early exits, the store into a `mutable` parameter that shares another parameter, and a part taken out of a temporary `copy(...)` of a heap enum (needs a decision).
