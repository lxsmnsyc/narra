# Outcome: guide-threads

## Built

- `docs/guide/15-threads-and-shared-state.md` replaces the stub. It keeps the `Next:` line to chapter 16. The running example counts primes below 100000. Sections:
  1. Run work on several cores: `Thread.scope`, `Thread.start`, `Thread.join`, a second join, `Thread<void>`, the scope waits for threads nobody joins, the difference from an async task, thread output order, effects, the 8 MiB stack, a panic on any thread ends the program with 101.
  2. What a thread function can use: copied captures, `let` capture (error sample), arrow at the call or a top-level name, not `async`, `Scope` cannot be captured, `Sendable` captures and results, `Cell` capture (error sample).
  3. Map an array on every core: `Parallel.map`.
  4. Pass values through a channel: `Channel.create<T>(capacity)`, `send`, `receive`, `values`, `close`, the `bool` of `send`, `io` effect. A worker pool over copies of one `Receiver`. The reply-channel pattern and the sender-in-its-own-channel cycle in text.
  5. Share a value with a lock: `Shared.create`, `get`, `set`, `update`, `read`, lock held only during a call, no `async` callback, the self-lock stop (exit 101), lock-order deadlocks, `Shared` cycles leak.
  6. Count with an atomic: `Atomic.create`, `load`, `store`, `add`, `exchange`, `compareExchange`, `Atomic<f64>` (error sample), `==` on handles rejected.
  7. Values across threads: the types that are not `Sendable`, `implement Sendable` rejected, `<T: Sendable>` in generic code, a thread's in-place write copies first, plain reference counts without thread modules.
  8. Threads or async tasks: when to use which, one event loop per thread, `Async` not `Sendable`, tasks started on a thread are cancelled when its function returns, blocking calls in async code.
- Audit rows 140 and 148 to 151 are each taught with a runnable example, except the items in Deviations.

## Deviations

- No runnable example for the self-lock stop and a thread panic: the guide suite runs programs that exit with 0, so both are stated in text (`testdata/exits/shared_reentered.nra` and `testdata/panics/threads_panic.nra` cover them).
- No example for a request that carries its reply `Sender` (one sentence; `testdata/channel_reply.nra` covers it), for tasks started on a thread (text only; `testdata/threads_tasks.nra`), or for the thread-function-in-a-variable, `async` thread function and captured-`Scope` errors (listed as rules).
- `Dynamic<Trait>` is left out of the list of non-`Sendable` types: no guide chapter teaches `Dynamic` yet. The list says "a function" instead.
- The prime test writes `divisor != 0 && number % divisor == 0`, since the compiler rejects `%` on a name not checked against zero. One sentence links to chapter 6.

## Stubs left

- None in this chapter. Links to chapters 11, 12 and 16 point at chapters other writers fill in batch-21.

## Gotchas

- Error-sample positions count the `// error:` line as line 1.
- A C-style `for` variable is a `let`, so a thread function in the loop needs `const part = low` first.
- A private function that declares `uses io` must also declare `alloc` when it loops over `Channel.values`.

## Proposed learnings

- none

## Verification

- Every `narra` block: `narrac check`, `narrac run -sanitize` (with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`) and `narrac run -release` with the prebuilt narrac. Stdout matches each `text` block. Every `narra error` block reports its message at its position. The worker pool and atomic samples gave the same output on 5 more release runs each.
- Extra claims checked with the compiler: `implement Sendable`, `==` on `Shared` and `Atomic`, an `async` callback of `Shared.update`, `Thread.scope` inside an async entry.
- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 380 cases.
- `pnpm --dir website build`: passes.
- plain-writing checker on `docs/guide/15-threads-and-shared-state.md`: `prose: ok`.

## Size

- Estimated M, actual M.
