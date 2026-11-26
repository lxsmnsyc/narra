---
id: std-net-lookup
title: "std/net: host name lookup, and the http client connects by name"
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
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent: 284000
owns:
  - roadmaps/selfhost/nodes/std-net-lookup/**
  - testdata/net-lookup*
shared:
  - std/src/net.nra
  - selfhost/runtime/std_async_io.c
  - selfhost/runtime/runtime.h
  - selfhost/src/runtime/*.nra
  - selfhost/src/std/stdtext.nra
  - selfhost/src/std/*.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
  - packages/http/src/client.nra
  - packages/http/README.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: the http client switches from numeric hosts to lookup
---

# std/net: host name lookup, and the http client connects by name

## Goal

`std/net` resolves a host name to addresses without blocking the event loop, `connect` accepts a name, and the http client fetches `http://example.com/`.

## Context

- Maintainer, 2026-11-26. Today `listen` and `connect` take numeric IPv4 or IPv6 addresses only, and the http client answers a host name with `Malformed`.
- Add async `lookup(host): Result<Array<string>, AsyncError>` (numeric addresses, IPv6 and IPv4 in the system's order), and let `connect` take a name, trying each address in turn. A numeric host skips the lookup.
- Resolve through the system resolver (`getaddrinfo`), so `/etc/hosts` and the system DNS settings apply. `getaddrinfo` blocks, so it runs off the loop thread (a helper thread that wakes the loop when done), and a cancel ends the wait at once. Writing DNS over UDP is out of scope.
- `Wasi-Wasm32` has no resolver: `lookup` of a name gives `Failed` there.
- Runtime C lives in `selfhost/runtime/`; `selfhost/src/runtime/*.nra` and `stdtext.nra` are generated (`sh scripts/test.sh --suite generated --update`). New runtime text needs two seed updates (see `GOAL.md`); the orchestrator runs them, the worker says what is needed.
- In `packages/http/src/client.nra`, drop the numeric-host check and the README limit that names it.

## Acceptance

- [ ] `lookup("localhost")` gives a loopback address; a name that does not exist (`nonexistent.invalid`) gives `Failed`; a numeric host is returned as is with no lookup.
- [ ] `connect("localhost", port)` reaches a local listener; the http client fetches `http://localhost:<port>/` from a local server.
- [ ] A cancelled lookup ends at once, and other tasks keep running while a lookup waits.
- [ ] SPEC and the std reference describe `lookup` and the name support in `connect`.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
