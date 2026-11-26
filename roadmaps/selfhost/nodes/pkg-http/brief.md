---
id: pkg-http
title: "First package: http (HTTP/1.1 client and server)"
status: done
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
spent: 464000
owns:
  - roadmaps/selfhost/nodes/pkg-http/**
  - packages/http/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: std-url
    kind: hard
    why: requests carry URLs and query strings; the package builds on the std/url API
---

# First package: http (HTTP/1.1 client and server)

## Goal

`packages/http` is the first first-party package: a plain HTTP/1.1 client and server over `std/net` and the async loop. It sets the pattern later packages copy: a workspace member under `packages/`, tagged `http/v<version>`, tested by `scripts/test.sh`.

## Context

- Source: `GOAL.md` Decisions, 2026-11-21: first package is `http` after `std/url`, then `csv`. First-party packages live in this repo under `packages/` as workspace members (SPEC section 14).
- No TLS: it waits for a `tls` package, which needs `package-c-sources`. Servers sit behind a TLS proxy until then.
- Server: request parsing with limits (header size, body size), keep-alive, chunked bodies both ways, a handler per request, and a way to hand the raw connection to a protocol upgrade (`pkg-websocket` needs it). Client: `get`/`request` with redirects off by default.
- Write the package's own docs page under `packages/http/README.md`.

## API (maintainer approved, 2026-11-26)

One `Request` and one `Response` type serve both sides, as in the Fetch API. `Body` is the same in both directions: an incoming message gets a stream that reads the connection, an outgoing one holds bytes or a stream. Names below are the approved shape; small signature details (effects, helper names) are the worker's call, recorded in the outcome.

Modules, through `[exports]` (parser and framing modules stay private):

- `http/headers`: `Header { name, value }`, `Headers { entries }` (order kept, names matched case-insensitively); `get`, `getAll`, `set` (replaces), `add` (appends).
- `http/error`: `enum HttpError { Io(error: AsyncError), Malformed(message), TooLarge(what), Closed, TimedOut }`.
- `http/body`: `Body { data: bytes, stream: Option<AsyncGenerator<Result<bytes, HttpError>>> }`; `empty`, `fromText`, `fromBytes`, `fromStream` (sent chunked), async `readAll(body, maxBytes)`, `chunks(body)`.
- `http/request`: `enum Method { Get, Head, Post, Put, Delete, Patch, Options, Other(name) }`, `Request { method, url: Url, headers, body }` (url relative on the server, absolute on the client), `create(method, url: string): Result<Request, HttpError>`.
- `http/response`: `Response { status, headers, body, upgrade: Option<Upgrade>, upgraded: Option<Upgraded> }`; `Upgrade { run: (connection, leftover: bytes) => Async<void> uses io, alloc }` (server: runs after the 101 head is written); `Upgraded { connection, leftover }` (client: filled by `send` on a 101); `create(status)`, `text(status, content)`, `reasonPhrase(status)`.
- `http/server`: `ServerOptions { maxHeaderBytes (16 KiB), maxBodyBytes (8 MiB), headerTimeout (10 s), idleTimeout (60 s) }`, `defaultOptions()`, async `serve(listener, options, handler: (request: Request) => Async<Response> uses io, alloc)` (cancel the task to stop), async `serveConnection(connection, options, handler)`. One detached task per connection.
- `http/client`: `ClientOptions { maxRedirects (0), timeout }`, async `send(request, options): Result<Response, HttpError>`, async `get(url)`. One connection per request; pooling is later and must not change this API.

Server answers: 400 bad request line, header syntax or chunk size; 431 headers over the limit; 413 body over the limit; 501 unknown `Transfer-Encoding`; 505 a version other than 1.0 or 1.1. Keep-alive is the HTTP/1.1 default, `Connection: close` ends it; HTTP/1.0 closes unless it sends `keep-alive`.

Known limits, documented in the README: `std/net` has no name lookup, so the client takes numeric hosts only (a name gives `Malformed`); `https:` is an error until the `tls` package. SPEC says a closure in an enum variant field cannot be called after narrowing; check early that an `AsyncGenerator` inside `Option` (in `Body.stream`) can be stepped after `is Some`, and if not, change the `Body` shape and record it.

## Acceptance

- [ ] The root `narra.toml` lists `packages/http` as a member and `scripts/test.sh` runs its tests.
- [ ] A server sample answers `curl` with keep-alive and a chunked response; the client fetches from it.
- [ ] Malformed requests (bad request line, oversized headers, bad chunk size) get 400 or 431 and never panic. A fuzz-style test feeds random bytes to the parser.
- [ ] An upgrade handler receives the connection after a `101` response.
- [ ] `sh scripts/test.sh` passes.

## Sizing

L: parsing and framing (medium), server loop (medium), client (low), the first workspace package layout (medium, new pattern).
