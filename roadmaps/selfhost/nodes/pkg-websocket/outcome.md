# Outcome: pkg-websocket

## Built

`packages/websocket` (package `websocket`, version 0.1.0, tag `websocket/v<version>`): an RFC 6455 server and client on top of `packages/http`. Nothing of it existed before. It is a member of the root workspace and depends on `http = { workspace = true }`.

Public modules (`[exports]`):

- `websocket/message`: `enum Message { Text(text), Binary(data) }`, `CloseInfo { code, reason }`, `enum WebSocketError { Http(error: HttpError), Io(error: AsyncError), Protocol(code, message), TooLarge, Closed(info: CloseInfo) }`, `Options { maxMessageBytes (16 MiB), maxFrameBytes (16 MiB), maxQueuedMessages (16), pingInterval: Option<Duration> (None), closeTimeout (5 s), protocols: Array<string> }`, `defaultOptions()` and `describe(error)` (new).
- `websocket/socket`: opaque `Socket`; async `receive(socket): Result<Message, WebSocketError>`; async generator `messages(socket)` (ends after the first error); async `send(socket, message): Option<WebSocketError>`; async `close(socket, code, reason): Option<WebSocketError>`; `protocol(socket): Option<string>`. The state, `open` and the tasks live in the private module `session.nra` (not in `[exports]`); `websocket/socket` shows `Socket` as a type alias and the functions as aliases and one wrapper (`messages`: a const alias of an async generator gives a C name clash).
- `websocket/server`: `accept(request, options, escaping onOpen: (socket: Socket) => Async<void> uses (io, alloc)): Response`. 101 with `upgrade` set, 400 for a bad handshake (not GET, no `Upgrade: websocket`, no `Connection: upgrade` token, no or bad key, no version), 426 with `Sec-WebSocket-Version: 13` for another version. It picks the first protocol of the client's offer that is in `options.protocols`, and declines extensions. When `onOpen` returns, it closes the socket with 1000 if it is still open.
- `websocket/client`: async `connect(url, options): Result<Socket, WebSocketError>` for `ws://`. It sends the handshake through `http/client` (scheme rewritten to `http`), checks status 101, `Upgrade`, `Connection`, `Sec-WebSocket-Accept`, that no extension and no protocol that was not offered came back, and masks every frame. `wss://` and any other scheme give `Http(Malformed(...))`.

Private modules: `session` (socket state, `open`, reader, writers), `frame` (head parse and write, masking, incremental UTF-8, close payloads), `handshake` (accept key, key check, token lists), `wakeup` (tickets, waiters, lock), and the check modules `checks`, `checkkit`, `checkserver`, `checkecho`, `checkraw`.

Also added: `examples/websocket_echo` (an echo server; arguments are the address and the port, defaults 127.0.0.1 and 9001; built by `scripts/test.sh` as an outside program), `packages/websocket/autobahn.sh`, `packages/websocket/README.md`, and the member entry in the root `narra.toml`. `scripts/test.sh` needed no change: its package loop already runs every `packages/*` with a `narra.toml`.

## Reader and writer design

Narra has one event loop per thread and `std/net` allows one waiting reader and one waiting writer per connection. `std/channel` blocks the thread, and there is no lock or condition for tasks, so the package builds one small primitive (`wakeup.nra`): a ticket is a task that sleeps for a day; a waker removes the ticket from a list and cancels the task, which ends the sleep at once. A ticket that is still in the list when the waiter resumes means the waiter itself was cancelled, so `wait` gives false and the caller returns `Io(Cancelled)` instead of spinning. `Waiters` (wake one or all) and `Lock` (first come, first served, handed over on release) sit on top.

Per socket:

- One reader task, started by `open`. It reads frames, unmasks into an array, answers pings, answers the peer's close, reassembles fragments, validates text, and pushes whole messages on a queue. Before it reads a header it waits while the queue holds `maxQueuedMessages`, and `receive` wakes it after a take. While it waits it reads nothing, so TCP back-pressure holds. Pings are not answered in that state.
- The reader does not wait for pong writes: pongs go into a list of 64 (oldest dropped) that a separate task drains (its handle is kept so the socket can cancel it). The reader does write one frame itself, the close reply or the failure close frame, just before it ends; a peer that stopped reading can hold it there until the write fails or `close` times out. The reason the socket ended is settled before that write, so a failed write cannot change it.
- A reader that waits for room stops waiting once `closeSent` is set (`close` wakes it), so `close` with a full queue returns when the peer answers. A write that fails or is cancelled ends the socket at once (`abandon`: finish, cancel reader, pong task and pinger), since half a frame may be on the wire; so does a cancelled `send` that was in the middle of a fragmented message. The close write-failure path abandons too.
- The server ends the socket (`drop`) after the handler and the close, so a cancelled handler does not leave the socket and reader alive.
- Text is checked per fragment with a UTF-8 state machine kept between fragments; the first bad sequence closes with 1007, and a text that ends inside a character is 1007.
- Writers: `messageLock` is held for a whole data message and `wireLock` for one frame. A message over 64 KiB is sent as 64 KiB frames, so a pong or a close frame can go between them but two messages never interleave. A control frame takes only `wireLock`.
- `close` sets `closeSent`, writes the close frame, then waits on the end of the reader. A timer task ends the socket and cancels the reader when the peer does not answer in `closeTimeout`.
- `pingInterval` runs a ping task. No pong before the next ping ends the socket (`Closed` with 1006).
- The socket cell that holds the reader task is cleared when the reader ends, so the task and the socket do not form a cycle (the run is leak-clean).

## Deviations

- `describe` in `websocket/message` is public on purpose (an error needs a one-line text for logs); it is not in the approved list.
- A cancelled `send` ends the socket even when it was only waiting for its turn, because the package cannot tell whether a byte was written.
- `close` and `send` give `Option<WebSocketError>` as approved; `close` also gives `Protocol(code, ...)` for a code or a reason that may not be sent (1005, 1006, 1015, out of range, reason over 123 bytes).
- A failed receive gives its error once (`Protocol`, `TooLarge` or `Io`) and `Closed(info)` after that, so `Closed` is what every later call gives, as the brief says.
- The ping timeout fires after 1x to 2x the interval (the check is at the next ping, so a peer that went silent just after a pong is caught at the second tick). It sends no close frame, because the peer is silent.
- `std/net` has no close, so "ends the connection" means the socket and every task that holds the connection are gone. A handler that keeps a copy of the `Socket` keeps the connection open.
- Async arrows with a struct parameter or capture do not compile in C today (see Gotchas), so `accept` and the checks use a plain arrow that calls a named async function.
- Shared files: root `narra.toml` (member). `scripts/test.sh` and `examples/**` (new folder only). No file outside the owned and shared paths.

## Autobahn

`sh packages/websocket/autobahn.sh` (Docker through colima, image `crossbario/autobahn-testsuite`, amd64 under emulation) ran cases 1.* to 10.* against `examples/websocket_echo`, rerun after the review round: 301 cases, 291 OK, 7 NON-STRICT, 3 INFORMATIONAL, 0 FAILED. 6.4.1 to 6.4.4 are OK now (first run: NON-STRICT, before UTF-8 was checked per fragment). NON-STRICT: 3.2, 3.3, 4.1.3, 4.1.4, 4.2.3, 4.2.4, 5.15 (the socket closes the moment a rule breaks, where Autobahn also accepts echoing the earlier frames). INFORMATIONAL: 7.1.6, 7.13.1, 7.13.2. The container and the work folder are removed on every path. Cases 12 and 13 (compression) were not run.

## Limits

- No `wss://`, no extensions, no compression.
- Memory of a socket is bounded by `maxQueuedMessages` messages of `maxMessageBytes` (16 x 16 MiB by default) plus one message being assembled.
- While the queue is full the reader answers no pings.
- A message and each frame are held in memory whole; there is no streaming API.
- Masking and unmasking go byte by byte through an `Array<u8>`. The 1 MiB checks and Autobahn 9.x (up to 16 MiB) pass, but a very large message blocks the loop for the time of that loop.
- A cancel of the task that waits in `receive` or `close` gives `Io(Cancelled)` and leaves the socket open.
- A cancelled or failed `send` ends the socket (no close frame); later calls give `Closed`.
- `accept` does not check `Origin`; the handler must (README).

## Stubs left

- None.

## Gotchas

- An `async` arrow with a struct parameter, or one that captures a struct, fails in the generated C (`f->v_x = *v_x`). Use a plain arrow that calls an async function.
- A `from` marker names one parameter; a function that shares two returns a `copy(...)` of one.
- `Cells.create<Array<bytes>>([])` does not parse (`>>`); use a typed local.
- A function type stored in a struct field and an `Upgrade.run` closure need `escaping` on the parameter they came from.
- `Net.read` gives at most 64 KiB, so a head and the first frame often arrive in one read: a raw client must cut the head on bytes, not decode the whole read as text.
- Timing checks under `-sanitize` are slow (a masking loop over 20 MiB takes seconds); a check should wait on a flag or a count, not a fixed sleep.
- `Time.sleep` for a day plus `Async.cancel` is a working wake-up; it keeps no timer after the cancel.
- Autobahn in a Docker VM needs its folders inside the home folder.

## Proposed learnings

- Wake a waiting task by cancelling a ticket task, and tell a wake from a cancel of the waiter by whether the ticket is still queued | no condition primitive for tasks exists in std | packages/ | no
- An `async` arrow over struct values does not compile; call a named async function from a plain arrow | generated C dereferences a struct value | packages/ | no (a compiler fix and test belong to a node of their own)

## Verification

- `out/narrac check .` and `out/narrac fmt --check .` (workspace root): clean.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 ASAN_OPTIONS=detect_leaks=0 out/narrac run -sanitize packages/websocket`, three runs after the review round: `210 checks, 0 failed` each. `packages/http`: `342 checks, 0 failed`.
- `out/narrac build -o <out> examples/websocket_echo`: builds as an outside program (and as part of `scripts/test.sh`).
- `sh packages/websocket/autobahn.sh`: see above.
- `sh scripts/test.sh`: after the review round the package step printed `210 checks, 0 failed` (http: `342 checks, 0 failed`), both examples built, and the last line is `test.sh: PASS, 11028 cases in 221s`.

## Size

- Estimated L, actual L.
- Risk estimated medium, measured medium: the task coordination had to be built from sleep and cancel; loopback checks, raw-frame checks and Autobahn cover it.
