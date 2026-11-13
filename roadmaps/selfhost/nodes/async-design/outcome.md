# Outcome: async-design

## Built

- D109 in `DECISIONS.md`, after D108, marked "Proposed, waiting for the maintainer". It answers the eleven questions of the brief.
- SPEC.md: section 8 "Async functions" rewritten to D93, D95 and D109 (one `ts` sample, an echo server with an async entry). Also changed: section 2 (async blocks, escaping parameters, the TypeScript differences table with a new `await` row), section 3 (the std module list), section 6 and 8 identity-type lists (`Async<T>` for `Task<T>`), section 7 (the entry may be `async`), section 11 (how a task frame compiles), section 15 (async entry, exit), section 16 (thread cancellation only). No `Async.run`, `Async.spawn`, `Task<T>`, `Loop`, `Async.sleep`, `Async.join`, `Async.map` or `Async.forEach` is left.
- `selfhost/src/diagnostics.nra` header: the range line `E1100-E1199  concurrency: async, the event loop and async I/O`.
- Briefs refined, status unchanged: async-core (E1100-E1129, L), event-loop (E1130-E1149, M, now also `std/async/time`), async-api (E1150-E1169, M), std-async-io (E1170-E1189, M to L, now net and file only). No brief was split.

## Deviations

- Runtime file names: `selfhost/runtime/std_async.c`, `std_async_loop.c`, `std_async_io.c` (generated `runtimestdasync.nra`, `runtimestdasyncloop.nra`, `runtimestdasyncio.nra`) instead of the `async.c`, `async_loop.c`, `async_io.c` and `runtimeasync.nra` of the brief and the batch-12 brief. The generated suite (`selfhost/tests/src/generated.nra`, `runtimeTarget`) makes a case only for `std_NAME.c` and four fixed names, and no async node owns the test runner. The batch-12 brief's line "`runtimecore.nra` gets only the hook that includes it" no longer holds: the text is picked in `runtimetext.nra` like every std runtime. The orchestrator may want to update the batch-12 brief.
- `Async.sleep(milliseconds)` is dropped. Sleep is `std/async/time` `sleep(duration: Duration): Result<Duration, AsyncError>`, built by event-loop instead of std-async-io.
- The maintainer approved D109 on 2026-11-13 with one change: `await` is allowed in any expression of an async body. D109 item 1 now says how the emitter spills earlier operands into frame temporaries; E1101 is left free.
- A cycle of awaits stops the program with exit code 101: a runtime stop that D92's list does not have.

## Stubs left

- None. The SPEC sample stays in `selfhost/tests/spec-skip.txt` (line "Async functions | import * as Async ..."), which already matched it; std-async-io removes the line.

## Gotchas

- The `spec` suite checks every `ts` block of SPEC.md, and the skip list matches by section and first line. The Async functions section therefore has one sample, starting with `import * as Async from "std/async"`. Another sample there would fail the suite until async lands.
- `Async` is a builtin type name (like `Generator`), so `import * as Async from "std/async"` is not E0421. A later change that puts builtin names in the scope type tables would break that import.
- Async sockets need a counted OS handle type (`Handle` of `std/internal/asyncio`), since SPEC section 6 closes sockets on the last release and std has no destructor. That is why std-async-io grew to L.

## Proposed learnings

- A new `selfhost/runtime` C file is named `std_NAME.c` unless the generated suite gets a new row | `runtimeTarget` ignores other names, so the generated module is never checked | selfhost/runtime | yes (the generated suite could fail on an unmapped file)

## Verification

- `grep -n "Async\.run\|Async\.spawn\|Task<\|\bLoop\b" SPEC.md`: no match.
- After the approval change and the rebase on feat/batch-12 (94d1fb5b): `sh scripts/test.sh --suite spec` PASS, 62 cases (the sample is still matched by its skip line); `--suite sources` PASS, 1 case.

## Size

- Estimated S, actual S.
