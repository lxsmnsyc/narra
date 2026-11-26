---
id: pkg-http-fetch-parity
title: "http: readers without a required limit, and sending forms"
status: active
priority: could
type: feature
story: batch-27
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-http-fetch-parity/**
  - packages/http/**
shared:
  - packages/websocket/**
  - examples/**
reserves: []
provides: []
depends_on: []
---

# http: readers without a required limit, and sending forms

## Goal

Reading and sending bodies works like the Fetch API: `readText(body)` and `Forms.readFields(request)` need no byte limit, and a program builds a form and sends it as a body, with the `Content-Type` and boundary set for it.

## Context

- Maintainer, 2026-11-26. When `Body` became opaque and Fetch-like (`packages/http/src/bodystate.nra`, `http/body` as the public alias), the readers kept a required `maxBytes` argument, and `http/form` only parses forms.
- Read `packages/http/README.md` and the sources under `packages/http/src` first, then `roadmaps/selfhost/nodes/pkg-http/outcome.md` and `pkg-http-forms/outcome.md`.

## Design (maintainer approved direction, 2026-11-26)

Limits belong to the body, not to every read call:

- An incoming body carries the limit of the side that read it: the server's `ServerOptions.maxBodyBytes` (already enforced while reading), and a new `ClientOptions.maxBodyBytes` for responses (default 64 MiB). Going over it gives `TooLarge("body")` from any reader.
- A body the program built itself (`fromText`, `fromBytes`, `fromStream`, a form) has no limit.
- `Bodies.limit(body, maxBytes): Body` gives a body with a tighter limit, for one route or one call.
- The readers lose the argument: `readAll(body)`, `readText(body)`, `readJson(body)`, `Forms.readFields(request)`. `Forms.parts(request, options)` keeps `FormOptions` for part count, part head and part size, which are not byte limits of the body.

A body knows its content type, like Fetch, so callers do not set it:

- A body may carry a content type: `fromForm` gives `multipart/form-data; boundary=...`, `fromQuery` gives `application/x-www-form-urlencoded`, `fromText` gives `text/plain; charset=utf-8`; `fromBytes`, `fromStream` and `empty` give none.
- The writer (server and client) adds `Content-Type` from the body when the message has none; a `Content-Type` the caller set wins. `Responses.text` and `json` keep the type they set.

Sending forms:

- `http/form` gets a builder: `FormData` (opaque), `Forms.create()`, `Forms.append(mutable form, name, value)` for a text field, `Forms.appendFile(mutable form, name, filename, contentType, body)` for a file whose body is any `Body`. Order is kept, names may repeat.
- `Bodies.fromForm(form)` gives a multipart body: a random boundary that does not occur in the text parts, each part's `Content-Disposition` (and `Content-Type` for files) written safely (quotes and line breaks in names and filenames escaped or refused), streamed so a large file is not held whole. Its length is known when every part is text or held bytes, so it is sent with `Content-Length`, else chunked. It is replayable only when every part is.
- `Bodies.fromQuery(params: Array<QueryParam>)` gives a urlencoded body, using `std/url` `queryToString`.

## Acceptance

- [ ] The readers and `Forms.readFields` take no limit. An incoming body over the server's or client's `maxBodyBytes`, or over a `Bodies.limit`, gives `TooLarge("body")`; a body the program built has no limit.
- [ ] A form built with `Forms.create`, `append` and `appendFile` and sent with `Client.send` is read back by the package's own server with `Forms.readFields` and `Forms.parts`: fields, repeated names, a file's bytes, filename and content type all match. The same form posted to a server is also read by `curl`-free raw checks of the wire format (boundary, part heads, final delimiter).
- [ ] `fromQuery` round-trips through `Forms.readFields`.
- [ ] `Content-Type` comes from the body when the caller set none, and a caller's type wins.
- [ ] Names and filenames with quotes or line breaks cannot break the multipart framing (checked).
- [ ] The README documents the readers, `Bodies.limit`, content types and sending forms. Every caller in `packages/` and `examples/` is updated.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: the limit moves into the body (low), the content type on the body (low), the multipart writer (medium), docs and callers (low).
