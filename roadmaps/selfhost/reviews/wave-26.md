# Wave 26 report: batch-12 children landed

Every batch-12 child landed on `feat/batch-12` as one commit after a worker check and a fresh review: the four fixes, review-nits, async-design (D109), async-core, event-loop, async-api and std-async-io. Two rounds of whole-branch review fixes followed. Last checks on the batch head: `test.sh: PASS, 8551 cases`, `linux-check: PASS`, seed current, debug copy check exit 0. Bench: within 3% of batch-11, `emit selfhost` +0.3% instructions.

Maintainer decisions during the batch (all in D109): `await` in any expression; the quiescent `Async.idle()` rule; a cancel passes down the await chain.

Orchestrator rulings: `Async.idle()` built in async-core; E1111; discarding a value that contains `Async<T>` is E1103; `then` on a finished handle runs on the next loop turn; E1150 for `peek`, `all`, `race` on void handles; `all` and `race` inputs and handles with `then` callbacks stop the cancel walk; async functions and arrows follow the `from` rules.

## Still to do

1. Pull request PR #117 (`feat/batch-12` into `main`) is open for the maintainer.
2. After it merges: set batch-12 `done` on `main`, prune `feat/batch-12` and `../narra-batch-12`.

## Inbox from this batch

Indexed assignment order (right side before target); `from` through stored handles and plain functions returning `Async<T>`.
