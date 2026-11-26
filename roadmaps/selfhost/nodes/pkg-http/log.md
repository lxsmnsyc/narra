# Log: pkg-http

- Checked the worktree and reset to `feat/batch-25`. `packages/http` did not exist.
- Early proof: an `AsyncGenerator` inside `Option` (and inside a struct field) can be stepped with `Async.next` and `for await` after `is Some`. `Body.stream` keeps the approved shape.
- An `AsyncGenerator<T>` with effects is a different type from one without. A generator that reads a connection has `uses (io, alloc)`, so `Body.stream` holds `AsyncGenerator<...> uses (io, alloc)` (alias `BodyStream` in `http/body`). The alias cannot be the return type of an `async function*`; those write the type out.
- A function type in a struct field needs `uses (io, alloc)` in parentheses; without them the parser reads `, alloc` as the next field.
- `await` is not allowed in a `test` block, so the package tests are a binary (`src/checks.nra`) that `scripts/test.sh` runs after the shards.
- `Process.run` blocks the loop, so the curl check runs the server on a thread (`Thread.startAsync`) and curl on the main thread.
- Timeouts race the I/O task against a sleep with `Async.race` and cancel both afterwards. A cancelled read loses no data because the read happens after the wait.
- Shared reader state (bytes read but not parsed yet) lives in a `Cell<bytes>` that the connection loop and the body stream share.
- Full run: see the outcome.
