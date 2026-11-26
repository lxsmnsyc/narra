---
id: pkg-websocket
title: "websocket package over http"
status: active
priority: could
type: feature
story: batch-25
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: L
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-websocket/**
  - packages/websocket/**
shared:
  - narra.toml
  - scripts/test.sh
  - examples/**
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: the handshake runs through the http server's upgrade hand-off
  - on: std-digest
    kind: hard
    why: the handshake key is SHA-1 then base64
---

# websocket package over http

## Goal

RFC 6455 WebSocket server and client on top of `packages/http`: the upgrade handshake, framing, masking, fragmented messages, ping and pong, close codes, and limits on message size. No extensions (permessage-deflate waits for a `compress` package).

## Context

- Motivating user: porting a game server whose browser client talks to it over WebSocket. The Autobahn test suite is the reference for conformance.
- Read `roadmaps/selfhost/nodes/pkg-http/outcome.md` and the outcomes of `pkg-http-forms`, `pkg-http-cookies` and `std-net-lookup`: `Response.upgrade` (server, `Upgrade.run(connection, leftover)`) and `Response.upgraded` (client, `Upgraded { connection, leftover }`), the package test-binary pattern (`narrac run -sanitize packages/<name>`, run by `scripts/test.sh` for every `packages/*`), and `examples/*` (built by `scripts/test.sh` as outside programs).
- `std/digest` has `sha1`; `std/bytes` has `toBase64`.

## API (maintainer approved, 2026-11-26)

Package `packages/websocket` (name `websocket`), a workspace member that depends on `http = { workspace = true }`. Names are the approved shape; small signature details (effects, helper names) are the worker's call, recorded in the outcome.

- `websocket/message`: `enum Message { Text(text: string), Binary(data: bytes) }`; `CloseInfo { code, reason }`; `enum WebSocketError { Http(error: HttpError), Io(error: AsyncError), Protocol(code, message), TooLarge, Closed(info: CloseInfo) }`; `Options { maxMessageBytes (16 MiB), maxFrameBytes (16 MiB), maxQueuedMessages (16), pingInterval: Option<Duration> (None), closeTimeout (5 s), protocols: Array<string> }`; `defaultOptions()`.
- `websocket/socket`: opaque `Socket`; async `receive(socket): Result<Message, WebSocketError>`; `messages(socket)`: an async stream of the same results, ending after the close; async `send(socket, message): Option<WebSocketError>`; async `close(socket, code, reason): Option<WebSocketError>`; `protocol(socket): Option<string>`.
- `websocket/server`: `accept(request, options, onOpen: (socket: Socket) => Async<void> uses (io, alloc)): Response`: a 101 with `upgrade` set, or 400 (bad handshake) or 426 (`Sec-WebSocket-Version` not 13, with the version header) for a bad request. Picks the first protocol of the client's offer that is in `options.protocols`.
- `websocket/client`: async `connect(url, options): Result<Socket, WebSocketError>` for `ws://` URLs; checks `Sec-WebSocket-Accept`; `wss://` is an error until a `tls` package exists.

Behavior:

- Each socket runs a reader task of its own. It answers pings with pongs and the peer's close with a close reply at once, whatever the handler is doing, reassembles fragmented messages, and puts data messages in a queue of at most `maxQueuedMessages`. When the queue is full the reader stops reading the connection, so TCP back-pressure holds and memory stays bounded. `receive` and `messages` take from the queue.
- `send` and `close` are safe from many tasks at once: frames of different messages never interleave; a control frame may go between the fragments of a message.
- A client masks every frame with a random key; a server refuses unmasked frames (1002). Reserved bits set, unknown opcodes, fragmented or oversized (over 125 bytes) control frames, and a continuation with no message in progress close with 1002. Text that is not UTF-8 closes with 1007. A message over `maxMessageBytes` closes with 1009 and gives `TooLarge`. Close codes that may not be sent (1005, 1006, 1015, out of range) are refused on the wire and in `close`.
- `close` sends a close frame, waits up to `closeTimeout` for the reply, then ends the connection. After the close every call gives `Closed(info)`.
- `pingInterval`, when set, sends pings and closes with 1006-style `Closed` when no pong comes within the interval.
- An extension offer (`Sec-WebSocket-Extensions`) is declined, not an error.

## Acceptance

- [ ] The root `narra.toml` lists `packages/websocket`, and `scripts/test.sh` runs its checks.
- [ ] Echo over loopback in both directions with the package's own client and server: text, binary, an empty message, a message fragmented by the peer, a 1 MiB message.
- [ ] Raw-socket checks: unmasked client frame, reserved bits, unknown opcode, oversized control frame, bad UTF-8 and a message over the limit each close with the right code and never panic; a fuzz check feeds random bytes to the frame parser.
- [ ] A handler that sleeps without calling `receive` still answers pings (checked with raw ping frames), and a full queue stops reading rather than growing memory.
- [ ] Many tasks sending at once on one socket give whole, unmixed messages at the peer.
- [ ] `examples/websocket_echo` (an echo server) builds as an outside program.
- [ ] `packages/websocket/autobahn.sh` runs the Autobahn fuzzing client in Docker against the echo server for cases 1 to 10 when Docker is available (not part of `test.sh`). Record the result in the outcome.
- [ ] `packages/websocket/README.md` documents the package, its limits (no `wss://`, no compression) and the reader-task guarantee.
- [ ] `sh scripts/test.sh` passes.

## Sizing

L: handshake (low), frame codec and validation (medium), the reader task with its queue and the shared writer (medium), client (low), Autobahn harness (low).
