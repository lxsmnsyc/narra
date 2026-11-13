# Outcome: threads-design

## Built

- D111 in `DECISIONS.md`, "Proposed, waiting for the maintainer". It answers the seven questions:
  1. Crossing values: an automatic trait `Sendable` (no functions, `Dynamic`, `Ptr`, `Cell`, `Async`, generators, coroutines, `Scope`, `Thread`, sockets). The checker asks for it at the captures and the result of a thread function and at the type argument of `Channel.create`, `Shared.create` and `Atomic.create`. Thread functions must be an arrow at the call or a top-level function name.
  2. Counts: sign-tagged (positive plain, negative shared and atomic, 0 immortal), only in a thread mode that a program enters by importing `std/thread`. A per-type share walker marks values at the boundary.
  3. Allocator: one heap per thread. The owner's number goes in the block header, frees from other threads go onto a lock-free remote stack, and the main thread stays on the static heap while no other thread runs.
  4. `Shared` uses the `Cells` API plus `read`, and its callback returns the new value. `Atomic` takes `bool` and integers, sequentially consistent. Channels are bounded (capacity below 1 is 1), `send` returns `bool`, and copies of a receiver share the queue. `==` on thread handles is an error.
  5. Async: one loop per thread, and the D109 exit rule runs when a thread function returns. `Async` is not Sendable, so no async receive in the first version.
  6. Runtime fixes first: forwarding with a count of active calls, a pid array and `pthread_sigmask` in `std_process.c`, `test_runner.c` and `std_test.c`.
  7. Split into five build nodes (below).
- SPEC: section 6 "Values across threads" rewritten (Sendable, thread mode). A `Sendable` row was added to the traits table. Section 8 Threads, Channels, Cell and Shared state were updated. The Shared sample now uses `(count) => count + 1` with a `std/shared` context. Everything is marked "not built yet". The first lines of the three skipped samples are unchanged, so `spec-skip.txt` still matches.
- `selfhost/src/diagnostics.nra` header has two new range lines: E1190-E1199 for threads and E1400-E1419 for channels, Shared and Atomic.
- Briefs: `threads` (L, now the first child: runtime, thread mode, heap, `std/thread`), plus four new nodes. `threads-process` (S, no codes). `threads-check` (M, E1190-E1195, `std/parallel`). `threads-channels` (M, E1400-E1409, none expected). `threads-shared` (M, E1410-E1411). Order: threads-process, threads, threads-check, then channels and shared side by side.

## Proof

- Measured on `narrac emit selfhost` with load average 5 to 6, median of 5 runs (details in `log.md`). All counts atomic: +26% cycles. Tagged counts: +6.4% instructions, about +5% cycles. `_Thread_local` heap: +6.5% instructions, +11.6% cycles on macOS. Bench programs: under 2%, except particles, which is +13% with atomic counts only. All variants emit identical C.

## Deviations

- The brief listed "lock inside" as a possible `Shared` form. D111 uses callbacks that return the value, the D90 pattern, because D81 does not allow a `mutable` scalar.
- SPEC said "`==` on an identity type compares identity". D111 makes it an error instead (E1195), to match `Cell` and `Async`. This needs the maintainer's yes.

## Stubs left

- None. No code was changed.

## Gotchas

- Immortal counts move from -1 to 0 in both modes (static strings, stack closure environments). The threads node must check that nothing else reads -1.
- `spec-skip.txt` keys on the first lines of the Threads, Channels and Shared samples. The build nodes remove those lines once the samples check.
- A `_Thread_local` access on macOS is a function call. Read the heap pointer once per allocation call.

## Proposed learnings

- Measure runtime cost on the compiler, not on bench | Perceus removed most counts from the bench loops, so they hide count costs | bench/ | no

## Verification

- `NARRAC=<release narrac> sh scripts/test.sh --suite spec`: PASS, 62 cases. `--suite sources`: PASS. `--suite codes`: PASS. `plan.py check`: no errors.

## Size

- Estimated S, actual S.
