# Outcome: pkg-http

## Built

`packages/http`, the first first-party package: an HTTP/1.1 client and server over `std/net`, with its own `narra.toml` (name `http`, version 0.1.0, tag `http/v<version>`) and a README at `packages/http/README.md`. It is a member of the root workspace.

Public modules (through `[exports]`): `http/headers`, `http/error`, `http/body`, `http/request`, `http/response`, `http/server`, `http/client`. Private: `message` (head parsing and writing, framing decisions, pure), `framing` (buffered reads, timeouts, chunked streams, writing), and the test modules `checks`, `checkkit`, `checkmessage`, `checknet`, `checkcurl`.

API as approved, plus small helpers:

- `headers`: `Header`, `Headers`, `empty`, `get`, `getAll`, `set`, `add`, and `remove` (new).
- `error`: `HttpError` and `describe(error)` (new).
- `body`: `Body`, `BodyStream` (alias), `empty`, `fromText`, `fromBytes`, `fromStream`, async `readAll`, async generator `chunks`.
- `request`: `Method`, `Request`, `create`, `methodName`, `methodFromName` (new).
- `response`: `Response`, `Upgrade`, `Upgraded`, `create`, `text`, `reasonPhrase`.
- `server`: `ServerOptions`, `defaultOptions`, `serve`, `serveConnection`.
- `client`: `ClientOptions`, `defaultOptions` (new), `send`, `get`.

Tests: `narrac run -sanitize packages/http` runs 128 checks (parsers, a 4000-round fuzz of random text, server and client over loopback, raw-socket cases for 400/408/413/431/501/505, keep-alive, chunked both ways, redirect, upgrade, timeouts, random bytes over the socket, and curl when installed). `scripts/test.sh` runs it at the end of a full run (no flags), with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`, and prints its last line or its whole output on failure.

## Deviations

- `Body.stream` is `Option<BodyStream>` where `BodyStream = AsyncGenerator<Result<bytes, HttpError>> uses (io, alloc)`. A generator with effects does not convert to one without, and connection streams have them. The early proof that a generator inside `Option` can be stepped after `is Some` passed, so the shape is otherwise as approved.
- `Upgrade.run` is `(connection, leftover) => Async<void> uses (io, alloc)` (parentheses needed in a field type).
- `serve` and `serveConnection` return nothing (no result type); the handler type is `(request: Request) => Async<Response> uses (io, alloc)`, and the parameter is `escaping`.
- `ClientOptions.timeout` is `Option<Duration>`, default `Some(30 s)`; it limits each step (connect, write, read), not the whole exchange.
- Added `Headers.empty/remove`, `HttpError` `describe`, `Request` `methodName/methodFromName`, client `defaultOptions`.
- The server overrides a success status the handler returns when the request body failed while the handler read it (400 for Malformed, 413 for TooLarge, 408 for TimedOut). This makes the chunk-size 400 and the chunked 413 hold even for a handler that ignores errors.
- A request with `Expect: 100-continue` and a body gets `100 Continue` before the handler runs.
- Review round: `ServerOptions` has a fifth field, `bodyTimeout` (default 60 s), a deadline for a whole request body. The server and client set `Content-Length` and `Transfer-Encoding` from the body and drop the caller's (a stream is always chunked, so a caller cannot pair a stream with a known length). The client keeps a caller's `Connection: Upgrade` and otherwise sends `Connection: close`. A 101 without `upgrade`, or a 1xx other than 101, from a handler gives 500. An HTTP/1.1 request needs exactly one `Host` (else 400). A malformed version (`HTTP/abc`) is 400; a well-formed unsupported one (`HTTP/2.0`) stays 505. `parseChunkSize` does not trim. The client skips at most 10 interim responses.
- `Wire.streamLength` and `Wire.streamChunked` take a `deadline` argument after `limit` (private module).
- `scripts/test.sh` runs the checks of every `packages/*` folder with a `narra.toml`.
- Files outside the owned paths: root `narra.toml` (member added) and `scripts/test.sh` (package check step), both listed as shared.

## Limits

- No TLS, numeric hosts only, UTF-8 heads, no compression, no HTTP/2 (see the README).
- No cap on the number of connections or on the run time of a handler.
- The body deadline counts from the first read of the body, and a read that is already buffered is not cut short.
- A refused request closes with its body unread, so some clients see a reset instead of the answer.

## Stubs left

- None.

## Gotchas

- Use `import * as Fields from "http/headers"`: a module alias and a type name (`Headers`) cannot share a name. The same holds for `Framing` (the enum in `message`), so `http/framing` is imported as `Wire`.
- Narrowing does not carry through `&&`/`||` (`x is None || x.value ...` fails); split the `if`.
- `let x = copy(...)` is needed where a value is shared with a parameter or a field you then write.
- A `test` block cannot `await`; package tests that need sockets must be a binary.
- `Process.run` blocks the event loop: a server and a child process in one loop deadlock. Put the server on a thread.
- `Net.Connection` closes when the last copy is dropped. A body stream holds a copy, so a response whose body is never read keeps its connection open until it is dropped.
- A refused request (413, 431) closes with unread bytes, which can reset the connection before the client reads the answer.
- pkg-websocket: take the connection from `Response.upgrade` (server) or `Response.upgraded` (client). `leftover` holds bytes that arrived after the HTTP head. Header names for the handshake go through `Headers.get/add` (case-insensitive). The head must be UTF-8.
- `Message.chooseStatus`, `Wire.within` and the other private functions are not exported through `[exports]`; other packages cannot import them.

## Proposed learnings

- A package that needs socket tests keeps them in a binary run by `scripts/test.sh`, and every wait in it has a timeout | a stuck check must fail, not hang | packages/ | no
- Effects belong in the type of a generator stored in a struct: `AsyncGenerator<T> uses (io, alloc)` | the plain type does not accept it | packages/ | yes

## Verification

- `out/narrac check .` and `out/narrac fmt --check .` (workspace root): clean.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 ASAN_OPTIONS=detect_leaks=0 out/narrac run -sanitize packages/http`: `128 checks, 0 failed`, curl included.
- `sh scripts/test.sh`: `163 checks, 0 failed` for the package, then `test.sh: PASS, 11018 cases in 214s`.

## Size

- Estimated L, actual L.
- Risk estimated medium, measured medium: new package pattern and a shared change to `scripts/test.sh`; the network code is covered by loopback and curl checks.
