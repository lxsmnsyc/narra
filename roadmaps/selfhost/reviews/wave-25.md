# Wave 25 report: batch-12, paused

Batch-12 runs on the local story branch `feat/batch-12` (worktree `../narra-batch-12`, not pushed). Paused by the maintainer on 2026-11-13.

## Landed

- arrow-arg-order, poly-recursion-hang, shape-names-private, bound-trait-shared-results, review-nits. Full check on the batch head with all five: `test.sh: PASS, 8178 cases`.
- async-design: D109 approved by the maintainer on 2026-11-13, with one change: `await` is allowed in any expression of an async body (E1101 left free). Accepted as written: exit 101 on a cycle of awaits, `sleep(Duration)` in `std/async/time`, `Async.race` gives `Async<Option<T>>`, file I/O blocks the loop, a counted `Handle` for sockets (std-async-io is L).

## In progress

- async-core: built on the local branch `feat/async-core` (agent worktree under `.claude/worktrees/`, branch `feat/async-core`). Full check passed before review (8325 cases). A fresh review found three should-fix items and one nit; the worker was fixing them when the session paused. Its `nodes/async-core/log.md` on that branch says which are done.
  1. An expression statement whose value is `Async<T>` (`await outer()` where `outer` returns a handle) must be E1103 (`stmts.nra` around 574).
  2. Discarding a value whose type contains `Async<T>` through `_`, `const _ =`, `let _ =`, or a `_`-prefixed local is E1103 (orchestrator ruling; `AsyncCheck.isHandleType`, `stmts.nra` around 365).
  3. A map key before an `await` in the value must be held in a frame temporary (`emitmaps.nra` around 296, as `holdBeforeAwait` does for struct literals); remove the README gap entry.
  4. Add `await` and `async` to `AST_KINDS` in `selfhost/tests/src/dumps.nra`.
  Accepted deviations: `Async.idle()` built here with a "no task ready" rule (event-loop makes it follow D109 item 7), E1111, no `from` check for async functions.

## Still to do

1. On `feat/async-core`: squash `542bf01b` into `97a04089`, rebase on `feat/batch-12`, run the full check, land it, set async-core `done`.
2. Start event-loop and async-api together (both after async-core), then std-async-io (after both).
3. End of batch: seed updates, full check and Linux check, debug `emit selfhost` with `NARRA_CHECK_COPIES=1`, curator pass, fresh review of the whole branch, bench row, one pull request.

## Inbox from this batch

Indexed assignment evaluates its right side before its target (on the base too); needs a fix node.
