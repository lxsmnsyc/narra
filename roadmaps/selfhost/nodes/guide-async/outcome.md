# Outcome: guide-async

## Built

- `docs/guide/14-async.md` replaces the stub. It keeps the `Next:` line to chapter 15. Sections:
  1. Wait for several things at once: `async function`, `Async<T>`, `await`, the async entry `export async function run`, `await` outside an async body (error), one thread taking turns, effects of a call.
  2. Every handle is used: E1103 (error sample), `const _ = f()` also rejected, `Async.detach`, `Async<void>`, a first use of `Async.idle`.
  3. Wait for all of them, or the first: `Async.all`, `Async.race`.
  4. Errors and cancel: `AsyncError.Failed` and `AsyncError.Cancelled`, `Async.cancel`, cooperative cancel, a timeout with `race` and `sleep`, cancel passing down the await chain.
  5. Look at a handle without waiting: `Async.peek`, `Async.then`, `Async.idle`, the `then` callback of an `Async<void>`, E1150 on `peek`, `all` and `race`, `==` on handles rejected.
  6. Start tasks from arrows and blocks: async arrows with `Arrays.map`, `async { }` with `return`, capture rules (let capture error sample), no `mutable` parameter, function-type parameters must be `escaping`.
  7. Read files and talk to the network: `std/async/file` (three reads with `Async.all`, a missing file as `AsyncError.Failed`), `std/async/net` (a server and a client on loopback, port 0, `localPort`), blocking calls block the loop.
  8. When the program ends: exit cancels running tasks, `await Async.idle()` to let them finish, a plain entry that starts tasks, the await-cycle stop with exit code 101.
  9. Streams of values: `async function*`, `AsyncGenerator<T>`, `for await`, laziness, `break` drops the stream, `yield*`, `for ... of` over a stream (error sample), `Async.next`, shared position of copies, `Net.chunks`.
- Audit rows 152 to 158 are each taught with a runnable example, except the items in Deviations.

## Deviations

- No runnable example for: the await-cycle stop (the guide suite runs programs that exit with 0, so the message and exit code 101 are stated in text; `testdata/exits/async_await_cycle.nra` covers it), `Async.next` and `Net.chunks` (named in text after the `for await` example), `yield*` (one sentence). These keep the chapter near the length of the others.
- Not taught: the `from` rules for async results and E0978 on values passed to a task (chapter 9 territory, niche here), E1105 (`for (mutable item ...)` in an async body), the per-connection one-reader rule of `std/async/net`, the `kqueue`/`epoll` and Windows notes.
- The capture error message suggests `Cell` from `std/cell`. The guide does not teach `Cell` yet, so the chapter names it in one sentence only.
- The file example makes its own folder under `TMPDIR` with a random suffix (`std/random`, `std/env`), since the debug and release runs of one sample can run at the same time. Those modules are chapter 11 and 13 material, not explained here.

## Stubs left

- None in this chapter. The link to chapter 12 (generators) and chapter 11 (files) points at chapters other writers fill in batch-21.

## Gotchas

- `narrac fmt` puts a short `match` on one line. The chapter uses the formatted layout in its one `match` with two arms. Chapter 6 keeps multi-line matches. The guide suite does not run `fmt`.
- Error-sample positions count the `// error:` line as line 1.

## Proposed learnings

- none

## Verification

- Every `narra` block: `narrac check`, `narrac run -sanitize` (with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`) and `narrac run -release` with the prebuilt narrac. Stdout matches each `text` block. Every `narra error` block reports its message at its position.
- Extra claims checked with the compiler: `const _ = f()` on a handle (E1103), `==` on handles, `Async.peek` on an `Async<void>`, `mutable` and non-`escaping` function parameters of an async function.
- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 358 cases.
- `pnpm --dir website build`: passes.
- plain-writing checker on `docs/guide/14-async.md`: `prose: ok`.

## Size

- Estimated L, actual M. The testdata programs already showed what compiles, so few examples needed rework.
