# http

An HTTP/1.1 client and server for Narra, built on `std/net` and the event loop of the thread. It has no other dependency than `std`.

This package speaks plain HTTP. It has no TLS: run a server behind a proxy that ends TLS, and fetch `http:` addresses only (see [Limits](#limits)).

## Using the package

Add it to the `[dependencies]` of your `narra.toml` (or as a workspace member, with `http = { workspace = true }`), then import the modules you need:

```narra fragment
import * as Bodies from "http/body"
import * as Client from "http/client"
import * as Responses from "http/response"
import * as Server from "http/server"
import { Request } from "http/request"
import { Response } from "http/response"
```

The package exposes these modules: `http/headers`, `http/error`, `http/body`, `http/request`, `http/response`, `http/form`, `http/cookie`, `http/server` and `http/client`. Its parsing and framing code is private.

## A server

A server takes a listener, a set of limits and a handler. The handler is an async function that gets a `Request` and gives a `Response`.

```narra fragment
import * as Net from "std/net"
import * as Responses from "http/response"
import * as Server from "http/server"
import { Request } from "http/request"
import { Response } from "http/response"

async function handle(request: Request): Response uses io, alloc {
  if (request.url.path == "/hello") {
    return Responses.text(200, "hello")
  }
  return Responses.text(404, "not found")
}

export async function run(_args: Array<string>): int {
  const listener = Net.listen("127.0.0.1", 8080)
  if (listener is Err(_)) {
    return 1
  }
  await Server.serve(listener.value, Server.defaultOptions(), handle)
  return 0
}
```

`serve` accepts connections until its task is cancelled. Each connection runs in a task of its own and handles its requests one after the other. `serveConnection` does the same for one connection you already have, which is useful in tests.

What the server does for you:

- **Keep-alive.** HTTP/1.1 connections stay open after a response unless the request or the response says `Connection: close`. An HTTP/1.0 connection closes unless the client sends `Connection: keep-alive`.
- **Limits.** A request head over `maxHeaderBytes` gets 431, and a body over `maxBodyBytes` gets 413. A request head that does not finish within `headerTimeout` gets 408. A connection that waits longer than `idleTimeout` for its next request, for the next bytes of a body, or for a write to be accepted is closed. A whole request body must arrive within `bodyTimeout`, so a client that sends one byte at a time cannot hold a connection open; the handler then sees `TimedOut` while it reads the body.
- **Refusals.** A bad request line, a bad header, a missing or repeated `Host` header on an HTTP/1.1 request, a bad `Content-Length`, a version that is not of the form `HTTP/1.1`, or a bad chunk size gets 400. An unknown `Transfer-Encoding` gets 501. A version other than HTTP/1.0 and HTTP/1.1 gets 505. A request with both `Content-Length` and `Transfer-Encoding` gets 400. No input makes the server panic.
- **Bodies.** The body of a request is a stream that reads the connection. Content-Length and chunked bodies both work. A handler that does not read the body does not break the next request: the server reads and drops what is left. When the body fails while the handler reads it (a bad chunk, a chunked body over the limit) and the handler still answers with a success status, the server sends 400 or 413 instead.
- **Responses.** A response with bytes gets a `Content-Length`. A response with a stream is sent chunked (or ends with the connection for an HTTP/1.0 client). The server sets `Content-Length` and `Transfer-Encoding` itself and drops any that the handler gave, so the headers always match what is on the wire. A response to `HEAD`, and a 204 or 304, has no body bytes. A handler that gives a header with a line break, a status outside 200 to 599, or a 101 without `upgrade` set gets a 500.
- **Expect.** A request with `Expect: 100-continue` and a body gets `100 Continue` before the handler runs.

### Streaming a response

A body can be a stream of pieces. The server writes the head first and then each piece as it arrives, so a handler can send data over time.

```narra fragment
async function* countdown(): AsyncGenerator<Result<bytes, HttpError>> uses (io, alloc) {
  yield Ok(Bytes.fromString("three\n"))
  yield Ok(Bytes.fromString("two\n"))
  yield Ok(Bytes.fromString("one\n"))
}

async function handleStream(request: Request): Response uses io, alloc {
  let response = Responses.create(200)
  response.body = Bodies.fromStream(countdown())
  return response
}
```

### Taking over the connection

A handler that answers `101 Switching Protocols` can take the connection, for a protocol such as WebSocket. It sets `upgrade` on the response. The server writes the head of the 101 response, then calls `run` with the connection and the bytes that the client had already sent after its request. The server does nothing else on that connection.

```narra fragment
async function speak(connection: Connection, leftover: bytes) uses io, alloc {
  const sent = await Net.write(connection, Bytes.fromString("hello\n"))
  if (sent is Err(_)) {
    return
  }
}

async function handleUpgrade(request: Request): Response uses io, alloc {
  let response = Responses.create(101)
  Fields.add(mutable response.headers, "Upgrade", "chat")
  Fields.add(mutable response.headers, "Connection", "Upgrade")
  response.upgrade = Some(Upgrade { run: speak })
  return response
}
```

The connection closes when `run` has finished and no copy of the connection is left.

## A client

```narra fragment
import * as Bodies from "http/body"
import * as Client from "http/client"

const response = await Client.get("http://127.0.0.1:8080/hello")
if (response is Ok(found)) {
  const content = await Bodies.readAll(found.body, 1000000)
}
```

`get(url)` sends a `GET` with default options. `send(request, options)` sends any request. Both give `Result<Response, HttpError>`: an HTTP status such as 404 is a response, and an `Err` means that no response could be read.

- **One connection per request.** `send` opens a connection, writes the request with `Connection: close` and reads the response. There is no pooling. A request that carries `Connection: Upgrade` keeps its own `Connection` header, so a server can switch protocols. The client sets `Content-Length` or `Transfer-Encoding` from the body and drops any that the caller gave. It skips up to ten interim `1xx` responses before the answer.
- **The body is a stream.** `send` returns as soon as the head of the response has arrived. Read the body with `Bodies.readAll(body, maxBytes)` or step through `Bodies.chunks(body)`. The connection stays open until the body is read to its end or dropped.
- **Request bodies.** Build one with `Bodies.fromText`, `Bodies.fromBytes` or `Bodies.fromStream`. A stream is sent chunked.
- **Redirects.** `ClientOptions.maxRedirects` is 0 by default, so a redirect is returned as it is. With a higher number, the client follows 301, 302, 303, 307 and 308. A 303, and a 301 or 302 after a `POST`, continue as a `GET`. A redirect that would need a stream body sent again is returned as it is. `Authorization` and `Cookie` are dropped when the redirect goes to another host or port.
- **Timeout.** `ClientOptions.timeout` is the longest wait for the server in one step: connecting, writing, or reading. `None` waits as long as the server does.
- **Upgrades.** When the server answers `101`, `response.upgraded` holds the connection and the bytes that arrived after the head. The body is empty.

## Cookies

`http/cookie` reads the cookies of a request and sets cookies on a response, as RFC 6265 describes.

```narra fragment
import * as Cookies from "http/cookie"
import * as Options from "std/option"
import * as Responses from "http/response"
import * as Time from "std/time"
import { Request } from "http/request"
import { Response } from "http/response"
import { SameSite } from "http/cookie"

async function handleLogin(request: Request): Response uses io, alloc {
  const visits = Options.withDefault(Cookies.get(request, "visits"), "0")
  let response = Responses.text(200, `visits so far: ${visits}`)
  let session = Cookies.create("session", "abc123")
  session.path = Some("/")
  session.maxAge = Some(Time.seconds(3600))
  session.httpOnly = true
  session.sameSite = Some(SameSite.Lax)
  const problem = Cookies.set(mutable response, session)
  if (problem is Some(_)) {
    return Responses.text(500, "bad cookie")
  }
  return response
}
```

- **Reading.** `read(request)` gives every `Cookie { name, value }` of every `Cookie` header, in the order they were sent. `get(request, name)` gives the value of the first cookie with the name. Reading is lenient: a pair with no `=`, a name that is not a token, or a control character in the value is skipped, and a value in double quotes loses its quotes.
- **Writing.** `SetCookie` has `name`, `value`, `path`, `domain`, `maxAge` (a `Duration`), `expires` (seconds since 1970), `secure`, `httpOnly` and `sameSite` (`Some(SameSite.Strict)`, `Some(SameSite.Lax)` or `Some(SameSite.None)`). `create(name, value)` gives one with no attributes. `set(mutable response, cookie)` adds one `Set-Cookie` header and leaves the earlier ones, so several cookies can be set. `format(cookie)` gives the header value without adding it.
- **Refusals.** `set`, `format` and `validate` refuse a cookie instead of writing it: a name that is not a token, a value with a character outside the RFC 6265 set (control characters, space, `"`, `,`, `;`, `\` and bytes above 126), a path that does not start with `/` or holds a control character or `;`, a domain that is not a host name, an `expires` before 1970, `SameSite=None` without `Secure`, a `__Host-` cookie without `Secure`, `Path=/` or with a `Domain`, and a `__Secure-` cookie without `Secure`. `set` gives `Some(error)` then and adds no header, so a value from a client cannot add a header line. `None` means the header was added.
- **Lifetime.** A `maxAge` below zero is written as `Max-Age=0`. `expires` is written as an HTTP date, such as `Sun, 06 Nov 1994 08:49:37 GMT`.
- **Removing.** `remove(mutable response, name, path)` adds a cookie with an empty value that expired in 1970. The path must be the one the cookie was set with, or the client keeps it.

## Reading bodies and building responses

Narra has no methods, so these are functions of the modules, called as `Bodies.readText(request.body, maxBytes)`. Every reader takes a limit and gives `TooLarge` when the body is longer.

- `Bodies.readText(body, maxBytes)` gives the body as UTF-8 text, and `Malformed` when the bytes are not UTF-8.
- `Bodies.readJson(body, maxBytes)` gives a `std/shape` `Value`, and `Malformed` when the text is not JSON.
- `Responses.json(status, value)` is a response with the JSON text of a `Value` and `Content-Type: application/json`.
- `Responses.redirect(url, status)` is a response with a `Location` header. The status must be 301, 302, 303, 307 or 308; any other status, an empty url, and a url with a line break or a NUL give `Err(Malformed)`.
- `Responses.isOk(response)` is true for a status from 200 to 299.

```narra fragment
async function handleJson(request: Request): Response uses io, alloc {
  const value = await Bodies.readJson(request.body, 100000)
  if (value is Err(error)) {
    return Responses.text(400, Errors.describe(error))
  }
  return Responses.json(200, value.value)
}
```

## Forms

`http/form` reads the body of a form submission, with files, in the two formats that browsers and `curl` send: `application/x-www-form-urlencoded` and `multipart/form-data`.

```narra fragment
import * as Forms from "http/form"
```

`Forms.readFields(request, maxBytes)` gives every text field as `FormField { name, value }`, in order and with repeated names kept. The `Content-Type` of the request picks the format. In a multipart form every part must be text without a filename, and the fields together may hold `maxBytes` bytes at most. A file part, bytes that are not UTF-8 and any other `Content-Type` give `Malformed`.

```narra fragment
async function handleForm(request: Request): Response uses io, alloc {
  const fields = await Forms.readFields(request, 100000)
  if (fields is Err(error)) {
    return Responses.text(400, Errors.describe(error))
  }
  return Responses.text(200, `${Int.toString(Arrays.length(fields.value))} fields`)
}
```

`Forms.parts(request, options)` is for uploads. It is an async stream of `Result<FormPart, HttpError>`, one value for each part of a multipart request. A `FormPart` has `name`, `filename` (set for a file input), `contentType` (what the part says about its content) and `body`, a stream. The body reads the request as it arrives, so a large file is never held whole.

```narra fragment
async function handleUpload(request: Request): Response uses io, alloc {
  let received = 0
  for await (const piece of Forms.parts(request, Forms.defaultOptions())) {
    if (piece is Err(error)) {
      return Responses.text(400, Errors.describe(error))
    }
    if (piece.value.filename is Some(_)) {
      for await (const chunk of Bodies.chunks(piece.value.body)) {
        if (chunk is Ok(bytes)) {
          received += Bytes.length(bytes)
        }
      }
    }
  }
  return Responses.text(200, `${Int.toString(received)} bytes`)
}
```

- **Read a part before the next one.** The parts share the request body. A part that is not read is skipped when the stream goes on, and it cannot be read afterwards.
- **Limits.** `FormOptions { maxParts, maxPartHeaderBytes, maxPartBytes }` bounds the number of parts, the head of one part and the body of one part. `Forms.defaultOptions()` gives 1000 parts, 16 KiB of head and 16 MiB of body. A request that passes one gives `TooLarge("parts")`, `TooLarge("part header")` or `TooLarge("part")`.
- **No total limit.** `parts` has no limit on the size of the whole body. On a server, `maxBodyBytes` bounds it. A client that reads parts of a response must bound the body itself.
- **Errors.** A missing or bad boundary, a part head that is not valid or not finished, a part with no `name`, and a body that ends before its closing boundary give `Malformed`. The stream ends with that error.
- **The format.** The boundary comes from the `Content-Type`, bare or in quotes, 1 to 70 characters. The text before the first boundary and after the closing one is ignored. A part needs `Content-Disposition: form-data; name="..."`, with an optional `filename="..."`, and may have its own `Content-Type`. A quoted value ends at the next quote and keeps its backslashes, so a Windows path stays as sent. A boundary may be split across any two pieces of the body.
- **No other encoding.** Part headers must be UTF-8, `filename*` is not read, and `Content-Transfer-Encoding` is ignored.

## Types

- `Headers` is a list of `Header { name, value }` that keeps the order. Names match without regard to case. `get` gives the first value, `getAll` all of them, `set` replaces every field of the name with one, `add` appends and `remove` deletes.
- `Method` is `Get`, `Head`, `Post`, `Put`, `Delete`, `Patch`, `Options` or `Other(name)`.
- `Request { method, url, headers, body }` is used on both sides. On a server `url` is what the request line wrote, usually a path and a query. On a client it is absolute. `Request.create(method, url)` parses a text with `std/url`.
- `Response { status, headers, body, upgrade, upgraded }` is used on both sides. `Responses.create(status)` makes an empty one, `Responses.text(status, content)` a plain text one, and `Responses.reasonPhrase(status)` gives the standard reason.
- `Body { data, stream }` holds bytes, a stream of pieces, or both (the data goes first). A body that was read from a connection has an empty `data` and a stream. The stream type is `BodyStream`, an `AsyncGenerator` of `Result<bytes, HttpError>` with the effects `io` and `alloc`.
- `HttpError` is `Io(error)`, `Malformed(message)`, `TooLarge(what)`, `Closed` or `TimedOut`. `Errors.describe(error)` gives a line of text.

## Limits

- **No TLS.** A url with the scheme `https` gives `Malformed`. TLS needs a package that compiles C code, which is not available yet.
- **Numeric hosts only.** `std/net` does no name lookup, so the client needs an IPv4 address such as `127.0.0.1`, or an IPv6 address in brackets. A name such as `localhost` gives `Malformed`.
- **Text heads.** The request or status line and the headers must be valid UTF-8. A head that is not gets 400 on a server and `Malformed` on a client.
- **No compression, no trailers.** `Transfer-Encoding: gzip` and the like get 501. Trailers after a chunked body are read and dropped. `Content-Encoding` is left to the caller.
- **No HTTP/2.** The server answers the requests of one connection one at a time, in order, even when the client sends several at once.
- **No cap on connections or handler time.** The server serves as many connections as the system allows, and it does not stop a handler that runs for a long time. Put a limit in front of it (a proxy, or a counter in the handler), and have long handlers check their own time.
- **The head of a client response** is limited to 64 KiB.
- **Slow readers.** A write that the peer does not accept for `idleTimeout` ends the connection on a server, and for `timeout` on a client.
- **Closing with unread data.** When the server refuses a request before it has read the body (for example with 413), it closes the connection with those bytes unread. Some clients see a reset instead of the answer.

## Tests

`narrac run -sanitize packages/http` builds the test program of the package and runs it. It checks the parsers (with a fuzz run of random text), a server and a client over loopback (keep-alive, chunked bodies both ways, the 400, 408, 413, 431, 501 and 505 answers, redirects, an upgrade, timeouts), the body readers and forms (multipart bodies split at every byte, cut short, changed and random, over loopback), the cookie code (reading, writing, refusals, a fuzz run, and a cookie jar of `curl` that sets, sends back and removes cookies), and a server against `curl` when `curl` is installed, including `curl -F` and `curl -d` forms. `sh scripts/test.sh` runs it in a full run.
