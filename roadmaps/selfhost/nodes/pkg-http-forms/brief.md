---
id: pkg-http-forms
title: "http body readers and forms: text, json, urlencoded and multipart"
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
size: M
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-http-forms/**
  - packages/http/src/form.nra
  - packages/http/src/multipart.nra
  - packages/http/src/checkform.nra
shared:
  - packages/http/src/body.nra
  - packages/http/src/response.nra
  - packages/http/narra.toml
  - packages/http/src/checks.nra
  - packages/http/README.md
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: reads the Body and Headers types of packages/http
---

# http body readers and forms: text, json, urlencoded and multipart

## Goal

A handler or client reads a body as text, JSON or a form (file uploads included) in one call, and builds JSON and redirect responses in one call, as the Fetch API's `text()`, `json()`, `formData()`, `Response.json()` and `Response.redirect()` do.

## Body readers and response helpers

Narra has no method calls, so these are functions of the `http/body` and `http/response` modules, called as `Bodies.readText(request.body, maxBytes)`:

- `http/body`: async `readText(body, maxBytes)` (UTF-8, `Malformed` otherwise) and async `readJson(body, maxBytes)` (through `std/json`, `Malformed` on a parse error).
- `http/response`: `json(status, value)` (sets `Content-Type: application/json`), `redirect(url, status)` (301, 302, 303, 307 or 308, sets `Location`; any other status is an error), `isOk(response)` (status 200 to 299).

## Context

- Maintainer, 2026-11-26: the http package has no FormData parser. `std/url` `parseQuery` already decodes `application/x-www-form-urlencoded` text; nothing reads `multipart/form-data`.
- Read `roadmaps/selfhost/nodes/pkg-http/outcome.md` for the Body, BodyStream and Headers types and the package's test binary (`narrac run -sanitize packages/http`, checks registered in `src/checks.nra`).
- New public module `http/form` in `[exports]`:
  - `FormField { name, value }` and `FormPart { name, filename: Option<string>, contentType: Option<string>, body: Body }`.
  - async `readFields(request, maxBytes)`: urlencoded or multipart (text parts only) into `Array<FormField>`, chosen by `Content-Type`.
  - `parts(request, options)`: an async stream of `FormPart` for multipart, each part's body a stream, so a large upload is never held whole.
  - `FormOptions { maxParts, maxPartHeaderBytes, maxPartBytes }` with defaults.
  - The worker picks the exact names and records them in the outcome.
- Multipart per RFC 7578 and RFC 2046: boundary from `Content-Type` (quoted or bare, 1 to 70 characters), preamble and epilogue skipped, `Content-Disposition: form-data; name=...; filename=...` with quoted values, a part's own `Content-Type`. A boundary split across two reads must be found.
- Errors are `HttpError.Malformed` (missing boundary, bad part head, no closing boundary) or `TooLarge`.

## Acceptance

- [ ] `readText` and `readJson` read a body the client sent; invalid UTF-8 and invalid JSON give `Malformed`; `json`, `redirect` and `isOk` are checked over the wire.
- [ ] A form posted by `curl -F name=value -F file=@<path>` is read back: the field value and the file's bytes and filename match. The check skips cleanly without curl.
- [ ] A urlencoded body (`curl -d`) gives its fields, repeated names kept.
- [ ] Boundaries split at every byte position across reads are found (a check feeds the body one byte per chunk).
- [ ] Malformed multipart (no boundary, unclosed part, part head over the limit, too many parts) gives `Malformed` or `TooLarge` and never panics; a fuzz check feeds random bytes.
- [ ] The README documents `http/form`.
- [ ] `sh scripts/test.sh` passes.
