# Log: threads-design

## 2026-11-13

- Branch `spike/threads-design`, reset to `feat/batch-13` (b23bbd4f). DECISIONS.md has no thread decision past D28 (scoped threads) and the notes in D61, D90 and D109; D111 is new.
- Read SPEC section 6 "Values across threads" and section 8, D28, D61, D80, D90, D94, D104, D109, the old threads outline note (process-signals review), `selfhost/runtime/runtime.h`, `std_process.c`, `test_runner.c`, `std/src/cell.nra`, `std/src/async__net.nra`.
- Count operations in the emitted compiler (`narrac emit selfhost`, 223,852 lines): 197 `x->rc++`, 187 `x->rc--`, 155 `--a->rc == 0`, 142 `a->rc == 1`, 45 `--v.u->rc == 0`, 42 map ones, plus strings and closure environments, which use -1 for immortal (static strings, stack environments, `.h = {-1, NULL}` in `emitclosures.nra`). All counts are `int64_t`. The release paths are emitted per type in `emitdecls.nra`.
- Global state that threads touch: `narra_small` (heap, `NARRA_HEAP_LOCAL` is empty with a comment to make it `_Thread_local`), `narra_live`, `narra_panic_jmp` and `narra_co_panicked` (test builds), the copy-check tables (debug), the first-use constant getters (D61 says not thread safe), `narra_proc_child`, `narra_runner_child`, `narra_panics_child` (one child slot each, `sigprocmask`). Already `_Thread_local`: the async loop, `narra_io_error`, the I/O buffer, the pending-drop queue.

### Measurement

Machine: Apple M2, 8 GB, eight spike agents running; load average 5 to 6. Scripts in `/private/tmp/claude-501/threads-design/scripts/`. Each variant is the same emitted C rewritten by a Perl script, built with `cc -O2 -DNARRA_RELEASE`. All three compiler variants emit byte-identical C for `selfhost` (shasum d7be13e2).

- A: every count operation atomic (`__atomic_fetch_add` relaxed, `__atomic_sub_fetch` acq_rel, acquire loads for `rc == 1`).
- B: sign-tagged counts (Lean 4 style): `rc > 0` plain, `rc < 0` shared and atomic, `rc == 0` immortal. Strings and stack environments move from -1 to 0.
- C: plain counts, `_Thread_local` heap.
- BC: B and C together.

`narrac emit selfhost`, median of 5 interleaved runs (`/usr/bin/time -l`):

| variant | instructions | vs plain | cycles | vs plain |
| --- | --- | --- | --- | --- |
| plain | 3.072G | | 1.133G | |
| A atomic | 3.036G | -1% | 1.431G | +26% |
| B tagged | 3.268G | +6.4% | 1.185G | +4.6% |
| C TLS heap | 3.271G | +6.5% | 1.264G | +11.6% |
| BC | 3.491G | +13.6% | 1.391G | +22.8% |

Cycles are noisy at this load (plain ranged 1.06G to 2.00G); instructions are stable within 2%. The TLS heap costs as much as the tagged counts: on macOS each `_Thread_local` access outside the async loop goes through a TLV thunk.

bench programs (median of 5, same variants): fib, sieve, matmul, strings, interp change by less than 2% in cycles for A and B (Perceus removed the counts from their loops); particles A +13% cycles, B unchanged; strings BC +16% cycles (allocation-heavy, TLS heap).

Conclusion: atomic counts everywhere cost a quarter of the compiler's time; tagged counts cost about 5%; the TLS heap costs another 6 to 12% on macOS. So threads get their own runtime mode: a program that does not import `std/thread` emits the same C as today and pays nothing; a program that does pays the tagged counts, and the heap reads its thread-local only once threads have started.
