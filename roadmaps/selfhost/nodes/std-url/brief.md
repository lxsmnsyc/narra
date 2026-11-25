---
id: std-url
title: "std/url: URL parsing, percent encoding and query strings"
status: todo
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
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
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-url/**
  - std/src/url.nra
  - testdata/url*
shared:
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
reserves: []
provides: []
depends_on: []
---

# std/url: URL parsing, percent encoding and query strings

## Goal

A new `std/url` module parses a URL into scheme, user info, host, port, path, query and fragment, writes it back as text, and resolves a relative reference against a base. It also percent-encodes and decodes text and reads and writes `application/x-www-form-urlencoded` query strings.

## Context

- Source: `GOAL.md` Decisions, 2026-11-21: std keeps `std/url` because narrac reads git URLs and `std/net` connects to a host and port. `inbox.md` 2026-11-20 (std-gap-survey rank 17). The `http` package node (`pkg-http`) needs it.
- Go `net/url` and the WHATWG URL standard are the references. Follow RFC 3986 for parsing and resolution. Do not copy the WHATWG host rules for special schemes, which need IDNA tables.

## Acceptance

- [ ] `Url.parse(text): Result<Url, string>` and `Url.toString(url)` round-trip the RFC 3986 section 5.4 examples, and `Url.resolve(base, reference)` gives the results listed there.
- [ ] `Url.encode`/`Url.decode` and `Query.parse`/`Query.toString` handle `+`, `%20`, bad escapes (an `Err`, not a panic) and repeated keys (kept in order).
- [ ] The std reference page is generated (`scripts/std/generate.nra`).
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: a parser, a resolver and the query helpers, all in Narra.
