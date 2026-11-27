# websocket

A WebSocket server and client (RFC 6455) for Narra. It runs on top of the `http` package: the
server answers the opening handshake of an http request and takes over the connection, and the
client opens the connection with the http client.

It supports text and binary messages, fragmented messages, ping and pong, close codes, masking,
limits on the size of a message, sub-protocols, and a ping interval that detects a dead peer.

It does not support `wss://` (there is no TLS package yet) or any extension, so there is no
permessage-deflate. An `Sec-WebSocket-Extensions` offer from a client is declined, not an error.

## Use

Add both lines to the `[dependencies]` table of your `narra.toml`. The program imports `http` modules too, so it lists `http` itself:

```toml
[dependencies]
http = { git = "https://github.com/lxsmnsyc/narra", version = "0.1.0" }
websocket = { git = "https://github.com/lxsmnsyc/narra", version = "0.1.0" }
```

The next `narrac build` downloads both packages and records them in `narra.lock`.

An echo server. `accept` turns a request into a 101 response, and the handler runs once the
http server has written it:

```narra fragment
import * as Messages from "websocket/message"
import * as Sockets from "websocket/socket"
import * as WebSocketServer from "websocket/server"
import { Request } from "http/request"
import { Response } from "http/response"
import { Socket } from "websocket/socket"

async function echo(socket: Socket) uses io, alloc {
  for await (const result of Sockets.messages(socket)) {
    if (result is Err(_)) {
      return
    }
    const failure = await Sockets.send(socket, result.value)
    if (failure is Some(_)) {
      return
    }
  }
}

async function handle(request: Request): Response uses io, alloc {
  return WebSocketServer.accept(request, Messages.defaultOptions(), echo)
}
```

Pass `handle` to `Server.serve` from `http/server`. A full program is in
`examples/websocket_echo`.

A client:

```narra fragment
import * as Client from "websocket/client"
import * as Messages from "websocket/message"
import * as Sockets from "websocket/socket"
import { Message } from "websocket/message"

async function talk(): bool uses io, alloc {
  const opened = await Client.connect("ws://127.0.0.1:9001/", Messages.defaultOptions())
  if (opened is Err(_)) {
    return false
  }
  const socket = opened.value
  const failure = await Sockets.send(socket, Message.Text("hello"))
  if (failure is Some(_)) {
    return false
  }
  const answer = await Sockets.receive(socket)
  const closed = await Sockets.close(socket, 1000, "bye")
  return answer is Ok(Message.Text("hello")) && closed is None
}
```

## Modules

- `websocket/message`: `Message` (`Text` or `Binary`), `CloseInfo`, `WebSocketError`,
  `Options`, `defaultOptions` and `describe`.
- `websocket/socket`: the opaque `Socket`, and `receive`, `messages`, `send`, `close` and
  `protocol`. Sockets come from `accept` and `connect`; a program cannot make one itself.
- `websocket/server`: `accept(request, options, onOpen)`.
- `websocket/client`: `connect(url, options)`.

### Options

| Field | Default | Meaning |
| --- | --- | --- |
| `maxMessageBytes` | 16 MiB | A larger message closes the socket with 1009. |
| `maxFrameBytes` | 16 MiB | A larger frame closes the socket with 1009. |
| `maxQueuedMessages` | 16 | Whole messages that wait for `receive` before the socket stops reading. |
| `pingInterval` | none | When set, a ping goes out that often. No pong before the next ping ends the socket. |
| `closeTimeout` | 5 s | How long `close` waits for the reply of the peer. |
| `protocols` | empty | The sub-protocols a server accepts, or a client offers, in order of preference. |

### Errors

`receive`, `send` and `close` give a `WebSocketError`:

- `Http`: the handshake failed (client only).
- `Io`: the connection failed, or the task that waited was cancelled.
- `Protocol(code, message)`: the peer broke a rule; the socket sent `code` in a close frame.
- `TooLarge`: a message or frame was over the limit; the socket closed with 1009.
- `Closed(info)`: the socket has ended, with the code and reason of the close. 1005 means the
  close frame had no code, and 1006 means the connection ended with no close frame.

After a failure `receive` gives that error once, and `Closed` on every call after it. `send` and
`close` give `Closed` once the socket has ended.

## Behavior

### The reader task

Each socket runs a reader task of its own, started when the socket opens. It reads the
connection whatever the handler does, so:

- A ping is answered with a pong, and a close frame of the peer with a close frame, at once,
  even while the handler sleeps or awaits something else and never calls `receive`.
- Fragments are joined, and a whole message goes into a queue of at most `maxQueuedMessages`.
  `receive` and `messages` take from the queue.
- When the queue is full the reader stops reading the connection until a message is taken. TCP
  back-pressure then slows the peer, and the memory of a socket stays below `maxQueuedMessages`
  messages of `maxMessageBytes`. While the reader waits for room it answers no pings, so a
  handler that never takes a message from a full queue will be closed by a peer that pings with
  a timeout.
- The reader does not wait for the write of a pong: pongs wait in a list of 64 and are written
  by another task, and the oldest is dropped when a peer pings faster than it reads. The reader
  does write one frame itself, the close reply to the close of the peer (or the close frame of
  a protocol error), just before it ends. A peer that has stopped reading can hold the reader
  in that write, and then the socket ends when the write fails or when `close` times out.
- A `close` that is called while the queue is full stops the reader from waiting for room, so it
  is not held back by messages nobody takes. Messages that arrive after a close are dropped.

### Sending

`send` and `close` are safe from many tasks at once. A message goes out whole before the next
message starts. A message over 64 KiB is sent as several frames, and a ping reply or a close
frame may go between them. A client masks every frame with a random key. A server refuses a frame
that is not masked, and a client refuses one that is.

A write that fails or is cancelled, including a cancel of the task that calls `send` while it
writes, may leave half a frame on the connection. The socket then ends at once, with no close
frame, and every later call gives `Closed`. A cancel while the sender only waits for its turn
ends the socket too, since the package cannot tell the two apart.

### Closing

`close(socket, code, reason)` sends a close frame and waits up to `closeTimeout` for the reply.
The code must be 1000 to 1003, 1007 to 1011 or 3000 to 4999, as it must on the wire; 1005, 1006
and 1015 are refused. When a handler passed to `accept` returns, the server closes the socket
with 1000 if it is still open, and ends it at once when the handler task was cancelled.

`accept` does not check the `Origin` header. A browser sends it on every WebSocket handshake,
and a server that uses cookies or other ambient credentials must compare it with the origins it
trusts, in the handler that calls `accept`, and answer 403 for the others before calling it.

The standard library has no way to close a connection except to drop every copy of it. The
connection ends when the socket and the tasks that use it are gone, which happens right after a
close, a failure or the end of the handler.

### What the socket checks

These close the socket, with the code in parentheses: a frame with a reserved bit set, an
unknown opcode, a fragmented control frame or one over 125 bytes, a continuation with no message
and a new message in the middle of another (1002); an unmasked client frame (1002); text that is
not UTF-8, checked as each fragment arrives so the first bad sequence ends the socket, and a close
reason that is not UTF-8 (1007); a message or frame over the limit (1009); a close frame with one byte, or a code that
may not be sent (1002).

## Tests

```sh
narrac run -sanitize packages/websocket
```

runs 210 checks over loopback: the codec, the handshake (with a frame in the same write), echo in
both directions (text, binary, empty, a message fragmented by the peer, 1 MiB), raw frames that
break each rule, UTF-8 across fragments, a handler that sleeps while the peer pings, a full
queue, `close` with a full queue and with a silent peer, a cancelled handler and a cancelled
sender, many tasks sending at once, the ping interval and random bytes. `scripts/test.sh` runs them with the other packages.

`packages/websocket/autobahn.sh` runs the Autobahn fuzzing client in Docker against
`examples/websocket_echo`, for cases 1 to 10. It needs Docker and an internet connection for the
first pull of the image, and is not part of `scripts/test.sh`.

## Limits

- No `wss://`, no extensions and no compression.
- A connection closes only when every copy of it is dropped.
- A cancel of the task that waits in `receive` or `close` gives `Io(Cancelled)` and leaves the
  socket open. A cancel of a `send` ends the socket (see Sending).
- `pingInterval` ends the socket after one to two intervals without a pong, depending on when
  the peer went silent within the interval.
