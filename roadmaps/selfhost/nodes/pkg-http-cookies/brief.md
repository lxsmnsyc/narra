---
id: pkg-http-cookies
title: "http cookies: read Cookie, build Set-Cookie"
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
size: S
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent: 125000
owns:
  - roadmaps/selfhost/nodes/pkg-http-cookies/**
  - packages/http/src/cookie.nra
  - packages/http/src/checkcookie.nra
shared:
  - packages/http/narra.toml
  - packages/http/src/checks.nra
  - packages/http/README.md
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: uses the Headers, Request and Response types of packages/http
---

# http cookies: read Cookie, build Set-Cookie

## Goal

A server reads the cookies of a request and sets cookies on a response, per RFC 6265.

## Context

- Maintainer, 2026-11-26. Read `roadmaps/selfhost/nodes/pkg-http/outcome.md` for the package layout and test binary.
- New public module `http/cookie` in `[exports]`:
  - `Cookie { name, value }`; `read(request): Array<Cookie>` and `get(request, name): Option<string>` from every `Cookie` header.
  - `SetCookie { name, value, path, domain, maxAge: Option<Duration>, expires (unix seconds), secure, httpOnly, sameSite: enum Strict | Lax | None }`; `create(name, value)`; `set(mutable response, cookie)` adds one `Set-Cookie` header per cookie; `remove(mutable response, name, path)` sets an expired cookie.
  - The worker picks the exact names and records them in the outcome.
- Building a cookie with a name or value that is not valid per RFC 6265 (control characters, `;`, `,`, spaces, a bad name token) is an error, so no header injection is possible. `SameSite=None` without `Secure` is an error.
- Reading is lenient: a malformed pair is skipped, not fatal.

## Acceptance

- [ ] A cookie set by the server is sent back by `curl -c`/`-b` and read by the handler. The check skips cleanly without curl.
- [ ] Invalid names and values, and `SameSite=None` without `Secure`, are refused; a value with a line break never reaches the wire.
- [ ] Malformed `Cookie` headers are read without a panic; a fuzz check feeds random text.
- [ ] The README documents `http/cookie`.
- [ ] `sh scripts/test.sh` passes.
