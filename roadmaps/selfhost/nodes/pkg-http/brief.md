---
id: pkg-http
title: "First package: http (HTTP/1.1 client and server)"
status: todo
priority: could
type: feature
story:
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

## Acceptance

- [ ] The root `narra.toml` lists `packages/http` as a member and `scripts/test.sh` runs its tests.
- [ ] A server sample answers `curl` with keep-alive and a chunked response; the client fetches from it.
- [ ] Malformed requests (bad request line, oversized headers, bad chunk size) get 400 or 431 and never panic. A fuzz-style test feeds random bytes to the parser.
- [ ] An upgrade handler receives the connection after a `101` response.
- [ ] `sh scripts/test.sh` passes.

## Sizing

L: parsing and framing (medium), server loop (medium), client (low), the first workspace package layout (medium, new pattern).
