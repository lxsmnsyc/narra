# Outcome: std-async-io

## Built

- Type `Handle` (D109 item 8): `Type.THandle` in `selfhost/src/types.nra`. It is a builtin type, and only modules that import `std/internal/asyncio` can name it (`scopes.nra`, built the way `StringBuilder` is). It is counted (`isCountedType`, `check.nra`). It is not a contents type, so it is never `mutable`. It is not copy tracked, and it has no Equatable or Hashable (`traits.nra`). Its C type is `narra_handle*` (`cnames.nra`). Retain and release are `narra_handle_retain`/`narra_handle_release`, and both note the `std/internal/asyncio` helper (`emitstate.nra`). Type strings are in `types.nra`, `check.nra`, `cnames.nra` and `stdtypes.nra`, and `exprs.nra` has a `hasNoFields` arm.
- Runtime `selfhost/runtime/std_async_io.c` (generated into `runtimestdasyncio.nra`, picked by `runtimeFor("std/internal/asyncio")`):
  - `struct narra_handle { rc; descriptor; }`. The last release closes the descriptor and frees the struct with `narra_free`, so the leak check counts an open socket as a live allocation.
  - `narra_handle_listen`, `_connect`, `_connected` (SO_ERROR), `_accept`, `_local_port`, `_read`, `_write` (`send` with `MSG_NOSIGNAL`, so a closed peer gives EPIPE and no SIGPIPE), `_error`, `_retry` (EAGAIN, EWOULDBLOCK, EINTR), `_pending` (EINPROGRESS) and `_message` (strerror, or a fixed text for a host that is not numeric).
  - Each call stores errno in `_Thread_local narra_io_error`. Sockets are nonblocking and FD_CLOEXEC. The host goes through `inet_pton` for IPv4, then IPv6. There is no name lookup.
- Primitives: the module `std/internal/asyncio` (rows `Handles.*` in `selfhost/src/stdasyncio.nra`, registered in `stdlib.nra`, dispatched by the `Handles.` case in `emitstd.nra`).
  - `Handles.wait(handle, writable): Async<int>` emits `narra_loop_wait_io((handle)->descriptor, writable)`.
  - `Handles.cancelled()` emits `narra_loop_cancelled_now()`.
  - Both calls go into the program text, so the I/O runtime needs nothing from the task runtime.
- `std/async/net` (`std/src/async__net.nra`):
  - `opaque struct Listener` and `opaque struct Connection`, each over a `Handle`.
  - `listen(host, port): Result<Listener, AsyncError>` and `localPort(listener): int` do not suspend.
  - `async accept`, `connect`, `read(connection, maxBytes): Result<bytes, AsyncError>` (empty bytes at the end of the stream) and `write(connection, data): Result<int, AsyncError>`, which writes every byte.
  - Each suspending function checks `Handles.cancelled()` first. It then tries the nonblocking call and awaits `Handles.wait` itself when the call would block. A wait value of -1 gives `Err(AsyncError.Cancelled)`, and a value above 0 gives `Err(AsyncError.Failed(strerror))`.
- `std/async/file` (`std/src/async__file.nra`): `readText`, `readBytes`, `writeText` and `writeBytes`, with the names and arguments of `std/fs`, returning `Result<T, AsyncError>`.
  - In a cancelled task a call gives `Cancelled` before it touches the file.
  - Otherwise it makes the `std/fs` call, then awaits `Tasks.sleep(0)` once, so the ready tasks get one turn.
- Module rows `std/async/net` and `std/async/file` (`ASYNC_IO_FUNCS`, check `source`).
- Samples (`.nra` and `.out`):
  - `testdata/async_io_echo`: an echo server and two clients in one program, with listen on port 0 and `localPort`.
  - `async_io_file`: write then read, for text and for bytes, a missing file, a ticker that runs during a file call's turn, and a cancelled write that makes no file.
  - `async_io_cancel`: a cancelled read that waits gives `Err(AsyncError.Cancelled)`, the next read gives it at once, and another task still reads the data.
  - `async_io_close`: a connection closes when its last copy (a parameter and a struct field) is dropped, and the peer reads the end of the stream. A read from a peer that has closed gives the data, then the end, then the end again.
  - `async_io_exit`: a detached accept and a detached read are left waiting at exit. Each gets `Cancelled`, a later read gets it at once, and every socket closes.
- Cancel down the await chain (maintainer ruling, amends D109 item 8). `narra_loop_cancel_wait` (`std_async_loop.c`) walks `task->awaited`:
  - An idle handle at the end finishes. A timer or descriptor wait at the end finishes with -1.
  - The task of a std function passes the cancel on, as before.
  - A user task is cancelled too (`narra_task_mark_cancelled`) when the waiter is its only waiter, and the walk goes on from it.
  - A handle that other tasks also await stops the walk, and so does a bare task (a race winner).
  - `narra_task_wait` (`std_async.c`) calls the walk when the waiting task is already cancelled, so a task cancelled while it ran passes the cancel when it suspends.
  - D109 item 8 and its status line ("Amended 2026-11-13 (maintainer): cancel passes down the await chain.") and the SPEC section 8 cancel text are updated.
- Review fixes:
  - `accept` takes the next connection after ECONNABORTED, and on Linux also after ENETDOWN, EPROTO, ENOPROTOOPT, EHOSTDOWN, ENONET, EHOSTUNREACH, EOPNOTSUPP and ENETUNREACH (`Handles.acceptAgain`).
  - `read` reads into a 64 KiB thread-local buffer and returns only the bytes that came, and a `maxBytes` of zero or less gives empty bytes at once.
  - A second reader or writer on one connection gets `Failed("another task is already reading this connection")` (or writing, accepting on this listener), through `Handles.busy`. SPEC section 8 says one reader and one writer at a time, and that a read gives at most 64 KiB.
  - `listen` uses SOMAXCONN.
- Samples `testdata/async_io_cancel_chain` and `async_io_limits`:
  - `async_io_cancel_chain`: a cancel reaches a read two user helpers deep. It also reaches a helper that the task awaits only after the cancel. A handle that two tasks await is not cancelled, and a detached child keeps running.
  - `async_io_limits`: `read(connection, 0)`, `read(connection, 1099511627776)`, and the second-reader message.
- Error tests: `testdata/errors/async_io_internal` (E0426, importing `std/internal/asyncio`) and `async_io_equal` (`==` on a `Listener`: "does not implement Equatable").
- The SPEC section 8 echo-server sample now passes the `spec` suite, and its line is gone from `selfhost/tests/spec-skip.txt`.
- `compiler/README.md`: in "Not supported yet", the async I/O line now lists the first-build limits: a numeric host and no name lookup, file calls that block the loop for their length, and TCP only.
- `std/async/time` is event-loop's and is unchanged here.

## Proof

- The cancel ruling. base: 03e53291. A compiler built with that commit's async runtime text runs `testdata/async_io_cancel_chain.nra` with no output until an 8 s alarm kills it (exit 142): the cancel of `outer` stops at `middle`, and the read waits forever.
- branch: the same program prints `async_io_cancel_chain.out` and exits 0 at once.

## Deviations

- Cancel ruling: `selfhost/runtime/std_async.c` (`narra_task_wait`) and `std_async_loop.c` (`narra_loop_cancel_wait`), async-api's and event-loop's code, plus their generated modules. Also `DECISIONS.md` (D109 item 8 and status) and `SPEC.md` (section 8 cancel text, the one-reader rule, the 64 KiB read limit).

- Files outside the brief's paths:
  - `selfhost/src/scopes.nra`: the `Handle` type name, beside `StringBuilder`.
  - `selfhost/src/emitstd.nra`: the `Handles.` dispatch.
  - `selfhost/tests/src/spec.nra`: the spec wrapper added a second `run` when a sample's entry is `export async function run(`, and now it checks both forms.
  - Goldens `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump`: the std module list now has `std/async/net` and `std/async/file`.
- No codes from E1170-E1189 are used. A misuse is an ordinary error: an import of `std/internal/asyncio` outside std is E0426, and `==` on a `Listener` or `Connection` is the Equatable error.
- Net and file are both built, so there was no split.

## Stubs left

- None. File calls block the loop while they run. That is the first build D109 item 7 asks for, and a worker thread can come later behind the same functions.

## Gotchas

- No sample covers the accept retry errors: the test suites cannot make a connection abort between the poller event and `accept`.
- A suspending std function must await the loop's wait primitive in its own body. If it awaits through a helper `async function`, the helper's task gets `forwards_cancel` and the awaited task does not, so `Async.cancel` on the user's task stops at it and the read waits forever. `checkWait` in `async__net.nra` is therefore a plain function over the wait's value.
- `Handles.error()` must be read right after the primitive call, before any `await`, because the errno slot is per thread and shared by every task.
- A task started by a cancelled user task is not cancelled itself (event-loop rule). A std call inside it does its work. Only the std function's own task reads `started_cancelled`.
- `Listener` and `Connection` hold a `Handle`, so a struct or tuple that holds one is counted. A `Result<#[Connection, Connection], ...>` local keeps both sockets open until its block ends, so drop it, or return from the function, to close them.
- The leak check sees an open socket as one live allocation, the `narra_handle` struct.

## Proposed learnings

- A suspending std function awaits the loop's wait primitive directly in its own async body | `forwards_cancel` is set on the running task, and a helper task breaks the cancel chain | std/src/async__* | no

## Verification

- `NARRAC=<release stage-1 narrac from the branch> sh scripts/test.sh --filter async`: PASS, 317 cases, covering the programs suite under AddressSanitizer, the leak check and the copy check, in debug and release.
- `NARRAC=<same> sh scripts/test.sh --suite spec`: PASS, 62 cases.
- `sh scripts/test.sh` (narrac from the seed, macOS arm64, bootstrap job included): `test.sh: PASS, 8486 cases in 110s`.
- `sh scripts/linux-check.sh` (colima, arm64 Linux, epoll path): `test.sh: PASS, 8486 cases in 92s`, `linux-check: PASS in 206s`.

## Size

- Estimated L, actual M to L. The `Handle` type needed fewer cases than the 13 `TStringBuilder` files suggested, because most `match` arms that list `TStringBuilder` are about copy tracking, which a handle does not take part in.
