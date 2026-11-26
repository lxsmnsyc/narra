---
id: pkg-http-compress
title: "http Content-Encoding: gzip and deflate bodies"
status: todo
priority: could
type: feature
story:
early: false
outline: true
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
  - roadmaps/selfhost/nodes/pkg-http-compress/**
shared: []
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: wraps the body streams of packages/http
  - on: c-packages
    kind: hard
    why: gzip and deflate come from a compress package that binds zlib as a C package (c_zlib) fetched from upstream
---

# http Content-Encoding: gzip and deflate bodies

## Goal

The client sends `Accept-Encoding` and decodes `gzip` and `deflate` responses as a stream; the server compresses a response when the request accepts it and the handler opts in. Both stay streaming, so a large body is never held whole.

Maintainer, 2026-11-26. Needs a `compress` package (zlib bound through package C sources), which has no node yet: refining this node writes that one first.
