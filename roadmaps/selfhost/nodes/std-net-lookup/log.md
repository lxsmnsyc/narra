# Log: std-net-lookup

- Checked first: no `lookup` in std, runtime or the http package. Nothing to reuse.
- Design: `Handles.lookup(host)` starts a detached pthread that runs getaddrinfo, writes one byte to a pipe and releases. The handle holds the read end as its descriptor, so the existing `Handles.wait` and the poller serve it, and a cancel ends the wait at once (wait finishes with -1).
- The lookup object (plain malloc, atomic count of 2: thread and handle) owns both pipe ends and closes them on the last release. The thread never writes to a closed pipe, and a cancelled lookup is freed by the thread when it ends.
- Result is addresses joined by newline, deduplicated, system order. Failure text travels in the same string with error code -3.
- Windows and WASI: `Handles.lookup` fails with -4 (no resolver). WASI already rejects all `std/net` functions (E0991).
- `driver.nra` links `-pthread` also when `std/internal/asyncio` is used (pthread_create outside thread mode).
- Editing std/src or runtime text needs bootstrap before the tool sees it (text is embedded); the docs of the generated suite need `NARRAC=out/narrac`.
- Gotcha: `a ?? b` is not an operator here (the lexer reads `?` twice).
