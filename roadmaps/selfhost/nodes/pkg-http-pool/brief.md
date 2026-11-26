---
id: pkg-http-pool
title: "http client connection pool"
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
  - roadmaps/selfhost/nodes/pkg-http-pool/**
shared: []
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: changes how packages/http/src/client.nra opens connections
---

# http client connection pool

## Goal

The http client reuses keep-alive connections per host and port, as Fetch does, with a cap on idle connections and an idle timeout. `send` and `get` keep their signatures; the pool lives in the client options or a client value.

Maintainer, 2026-11-26: worth doing, timing left to the orchestrator. Refine when a program (a benchmark, the game server port, or a crawler) shows the cost of one connection per request.
