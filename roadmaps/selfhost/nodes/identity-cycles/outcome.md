# Outcome: identity-cycles

## Built

- Decision D80 in `DECISIONS.md`: "Cycles through identity types leak and are caught by the leak check". No cycle collector, no type rule, no `Weak<T>`.
- `SPEC.md` section 6 "Cycles": why plain values cannot form a cycle, the identity types that can, the two leaking patterns, how to break them (`Cell.set(cell, None)` before the last outside handle drops, or `std/arena`), and room for a later collector.
- `SPEC.md` section 8 "Channels": a channel in a cycle stays open; close it with `Channel.close` when senders travel through channels; request and reply channels are fine. "Shared state": identity types can form cycles, `Task<T>` and `Cell<T>` follow the same rule.
- `SPEC.md` section 16 "Deferred": a cycle collector for identity objects.
- Documents only. No compiler, runtime or test change.

## Result

- result: met. The plan holds: plain values cannot form a cycle in the current compiler, so the rule "every cycle passes through an identity object" is true today.
- evidence: three probes run with `NARRA_CHECK_LEAKS=1 out/narrac run -sanitize` (details in `log.md`). A struct pushed into its own array field, a `Box` stored in its own payload, and a closure stored in the struct it captures all run clean with no leak report, because the shared value copies before the write. A `const` closure that names itself and an escaping closure that captures a `let` are compile errors.
- reproduce: `sh selfhost/bootstrap.sh out` on `spike/identity-cycles`, then `NARRA_CHECK_LEAKS=1 out/narrac run -sanitize roadmaps/selfhost/nodes/identity-cycles/probes/<name>.nra`.

## What this means for dependents

### cell-capture

- Build `Cell<T>` with no cycle handling: plain reference counting on the cell object. Do not add a collector, a weak handle, or a type rule against a `Cell` holding a closure that captures it.
- `Cell.set(cell, value)` must drop the old value at the call, so `Cell.set(cell, None)` breaks a cycle at once.
- Add the sample `testdata/cell_cycle.nra`: a `Cell<Option<() => int>>` that holds a closure capturing the same `Cell`. Expected result: the program runs, then the leak check reports (exit 102, `leak check: N allocations still live`). The test runner's `programs` suite expects a clean leak check, so this sample needs a way to expect the report (a golden of the exit code and stderr, or a `narra test` case that expects the `leaked N allocations` failure). Pick the cheapest form the runner supports and add the runner support if it has none.
- Add a second sample that builds the same cycle and breaks it with `Cell.set(cell, None)`; it must run clean.
- `Cell` cannot cross threads, so no atomic counts are needed for it.

### threads

- `Shared<T>`, `Sender<T>`, `Receiver<T>` and `Atomic<T>` are counted handles with no cycle handling. `Task<T>` (async) follows the same rule.
- Add one leak sample: a `Shared<Option<Shared<...>>>` (or a `Sender` sent into its own channel) whose expected result is the leak report, in the same form cell-capture chooses.
- The leak counter `narra_live` in `selfhost/runtime/runtime.h` is a plain `int64_t`. With threads it must become atomic (or per-thread counts summed at exit), or the leak check reports wrong counts.
- A cycle that holds the last `Sender` keeps the channel open, so a receiver waiting for the close blocks. `Thread.scope` then never returns. In tests this is a timeout (`NARRA_TEST_TIMEOUT`, 10 s), not a leak report. The SPEC text tells programs to close channels with `Channel.close` when senders travel through channels; test that `Channel.close` on one copy ends `Channel.values` even when a copy of the sender is inside the channel.
- A future collector (rejected for now) would need to be concurrent for these types; keep the handle layout simple (count plus payload) so one can be added without changing programs.

## Gate decisions

- **Cycles through identity types (D80).** Chosen: documented leaks, caught by the debug leak check. Alternatives: (a) type rule banning identity types in identity payloads, rejected because it breaks request and reply channels and stops a `Cell` holding a closure; (b) trial-deletion collector for identity objects (Nim ORC, CPython), rejected for now for its per-decrement cost and the need for a concurrent collector, and kept addable later without changing programs; (c) `Weak<T>`, rejected because `std/arena` handles cover back edges. Why: zero cost for programs without identity types, deterministic frees and closes, and every leak has a named cause that tests catch.
- **`Task<T>` counts as an identity type in the cycle rule.** The brief lists it; SPEC section 8's identity table does not, because a task does not share data across threads. Chosen: name it in the cycle text of sections 6 and 8 and leave the table as is. Alternative: add it to the table, rejected because the table's sentence ("the only way two threads can see the same changing data") does not fit tasks.
- **Expected-leak samples are left to the nodes that build the types.** Alternative: add a runner feature now for expected leak reports, rejected because no identity type exists to test, and the brief limits this node to documents.

## Deviations

- The brief asks to align the bullet "A `let` variable or `mutable` parameter cannot be captured or stored" with closure-borrow. closure-borrow already rewrote it in section 6 "Cycles" ("only a closure that cannot outlive the call captures a variable by reference"), so the text was kept.
- Added a line to SPEC section 16 "Deferred" for the collector, not named in the brief.

## Stubs left

- None.

## Gotchas

- The leak check reports a count, not the objects. A failing leak sample tells you that something leaked, not what.
- In test builds a coroutine that panics resets the leak counter to its value before the resume (D79), so a cycle built inside it is not reported. Written in D80 as a limit.
- D81 "const fixes the name, not the contents" is not built yet: a `const` struct cannot be passed as `mutable` today. Samples should use `let` until it lands.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh out`: stage 2 and stage 3 identical.
- Probes under `NARRA_CHECK_LEAKS=1 ... run -sanitize`: as in Result.
- `NARRAC=$PWD/out/narrac sh scripts/test.sh`: see the final report.

## Size

- Estimated S, actual S.
