# Outcome: std-net-lookup

## Built

- `std/net` has async `lookup(host): Result<Array<string>, AsyncError>` (`std/src/net.nra`). A numeric host (IPv4 or IPv6) is returned as it is. A name goes to the system resolver (`getaddrinfo`, AF_UNSPEC, stream sockets), so `/etc/hosts` and the system DNS settings apply. Addresses are numeric text, deduplicated, in the system's order. A name with no address gives `Err(AsyncError.Failed(...))`.
- `connect(host, port)` takes a name: it looks the name up and tries each address in turn, giving the last error when none accepts, and `Cancelled` at once. A numeric host skips the lookup. `listen` stays numeric.
- Runtime (`selfhost/runtime/std_async_io.c`): `narra_handle_lookup` starts a detached pthread that runs `getaddrinfo`, writes one byte to a pipe and releases its reference. The handle holds the read end as its descriptor, so the existing `Handles.wait` and poller serve it, and a cancel ends the wait at once. A plain-malloc lookup object with an atomic count of two (thread and handle) owns both pipe ends and the result, and the last release frees it, so a cancelled lookup is freed by the thread when it ends and nothing writes to a closed pipe.
- Registry rows `Handles.lookup`, `Handles.lookupResult`, `Handles.isNumeric` in `selfhost/src/std/stdasyncio.nra`. Generated files regenerated: `selfhost/src/runtime/runtimestdasyncio.nra`, `selfhost/src/std/stdtext.nra`, `docs/reference/std/net.md`.
- `packages/http/src/client.nra`: the numeric-host check is gone; the client passes the host (brackets stripped for IPv6) to `Net.connect`. README limit removed. `checknet.nra`: the "host name is Malformed" check became a fetch of `http://localhost:<port>/hello` and an `.invalid` host giving `HttpError.Io`.
- SPEC section 8 and the async guide chapter describe `lookup` and names in `connect`.
- Tests: `testdata/net-lookup.nra` and `testdata/net-lookup-cancel.nra` (with `.out`), no internet needed.

## Proof

Not a fix node.

## Deviations

- `selfhost/src/driver/driver.nra` (not in the shared list): links `-pthread` also when `std/internal/asyncio` is used and the target OS is not Windows or WASI, since the lookup calls `pthread_create` outside thread mode. No Windows cross build of a `std/net` sample runs in the suite, so the Windows branch was not run.
- Review round: the lookup pipe is made with `pipe2(O_CLOEXEC)` on Linux (fcntl elsewhere), and the helper thread starts with every signal blocked (the caller's mask is restored right after `pthread_create`).
- `docs/guide/14-async.md` (not in the shared list): one sentence said the host is numeric; it now describes names and `lookup`.
- `packages/http/src/checknet.nra` (not in the shared list): its client check expected `Malformed` for a name.
- Windows has no resolver here: a name gives `Failed` ("name lookup is not supported on this platform"). WASI rejects every `std/net` function at compile time (E0991), so `lookup` is rejected there rather than giving `Failed`.
- Cancel and "other tasks keep running" are proven by the cancel test and by design (the helper thread, the loop only polls a pipe). No test can slow the system resolver, so no test shows a slow lookup beside a busy task.

## Stubs left

- None.

## Gotchas

- A std source or runtime text change reaches a compiler only after `sh selfhost/bootstrap.sh out`; `--suite generated --update` must run with `NARRAC=out/narrac` for the std reference to show the new text.
- `a ?? b` is not an operator; use an `if`.
- A lookup of a name that the resolver answers slowly (DNS offline) ends the helper thread late. At process exit the thread is simply abandoned.
- `localhost` may list `::1` before `127.0.0.1`; `connect` falls through to the next address, which matters for an IPv4-only listener.

## Proposed learnings

- A std function that blocks the process runs on a helper thread and wakes the loop through a pipe the poller watches | a cancel then ends the wait with no loop change | selfhost/runtime | no

## Verification

- `out/narrac run -sanitize testdata/net-lookup.nra` and `net-lookup-cancel.nra`: expected output, no ASan error.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 out/narrac run -sanitize packages/http`: `164 checks, 0 failed`.
- `NARRAC=out/narrac sh scripts/test.sh --suite generated --update`: PASS. `sh selfhost/bootstrap.sh out`: fixed point holds.
- `sh scripts/test.sh`: `test.sh: PASS, 11028 cases in 290s`. An earlier run failed one case, `programs/io_fs_more.nra/debug` (a file-system sample that shares a temp folder; other workers ran at the same time). It passed alone and in the rerun.
- Seed updates: the seed must be regenerated twice (runtime text and std text are embedded in it). `selfhost/src` does not use `lookup`, so the compiler source needs no wait; only a compiler source that wants `Net.lookup` or a name in `Net.connect` must wait for both updates.

## Size

- Estimated M, actual M.
- Risk estimated medium, measured medium: a helper thread and a pipe in the runtime, covered by cancel and connect tests.
