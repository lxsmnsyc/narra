# Wave 22 report: batch-9 (in progress)

Batch-9 runs on the local story branch `feat/batch-9` (not pushed). Every child has landed as a direct commit after a worker check and a fresh review. seed-switch landed as four commits so the seed regenerations stay readable.

## Landed

"Compiler" is `narrac emit selfhost`, instructions retired.

- process-exit: `Process.exit(code)` flushes and calls `_exit`; it ends a path like `panic` (`isPathEndingCall`). New `exits` suite, expected codes in `<name>.code`.
- test-runner-fixes: the final line of `scripts/test.sh` shows the case total; the 0-case PASS sightings came from reading `| tail -2` past empty shards. `io_files` uses `mktemp -d`.
- host-target-builtin: `Build.hostTarget(): Option<Target>` from the C compiler's macros (`std/internal/host`, `std_host.c`). Linux check passed in review.
- ptr-cast: `Ptr.cast` (type from the expected type) and `Ptr.element` (FixedArray index proof). fmt keeps a comment between `from` and the library name.
- hidden-copy-fix: the copies in `loadFile` and `loadPackage`. Narrowed: a debug-built compiler still exits 103 on most runs (checker and emitter copies, inbox). `NARRA_COPY_GRANT` needed no change.
- process-signals: `narrac run` and `narrac test` (runner and `Test.panics`) forward SIGTERM, SIGINT and SIGHUP; the caller survives and `Process.run` reports 128+sig. Test uses pid files and `kill -0`, no `pgrep`. Linux check passed. Thread safety of the global child slot is a note on the threads node.
- diagnostics-polish: typeof cycles named (functions and constants), E1003 for an unknown constant type, cascade suppression per file, E1004 for positions past 2^20 (tested by an edge case that writes its input at run time).
- guide-multifile-samples: `narra file PACKAGE PATH` blocks, guide-samples contract v2; chapters 1, 4 and 8.
- fmt-block-hug: a block or arrow as the last argument stays on the call line, except beside a comment (also fixes the arrow case on main). Operand hug dropped after comparing with oxfmt 0.66.0; the operands of an assignment that ends in a block now sit level after `=`.
- seed-switch: two seed updates, one pass each. The compiler exits with `Process.exit` (full release under `NARRA_CHECK_LEAKS`) and uses `Build.hostTarget()`. Compiler 2.427G to 2.396G on the same source (-1.3%, close to noise).

Diagnostic codes used: E1003, E1004. c-audit-target's unused E1003 reservation was released.

## Checks

On the final children: `sh scripts/test.sh` passes (7107 cases); `sh scripts/linux-check.sh` passed on seed-switch's branch, which is the story branch tip.

## Left before the story pull request

1. A fresh review of the whole `feat/batch-9` branch against `main`.
2. `bench/BASELINE.md` end-of-batch row.
3. Remove the inbox entries this batch closed (Ptr cast, fmt block hug, guide fragments, and the rest named in the batch-9 brief).
4. Curator pass: promote lasting rules from the outcomes (for example, `narrac fmt` follows oxfmt and was checked against oxfmt 0.66.0; read the total on the final line of `scripts/test.sh`).
5. Push `feat/batch-9` and open the story pull request into `main` for the maintainer.

Stale worktree to remove: `.claude/worktrees/agent-aaa8582c137bb0b37` (`fix/batch-8-review`, merged).
