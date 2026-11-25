---
id: pkg-postgres
title: "postgres package: wire protocol client and pool"
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
size: L
risk: high
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-postgres/**
  - packages/postgres/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: std-digest
    kind: hard
    why: SCRAM-SHA-256 login needs HMAC and PBKDF2
---

# postgres package: wire protocol client and pool

## Goal

A PostgreSQL client written in Narra over `std/net`, not a libpq binding, so it does not wait for `package-c-sources`. Covers startup and SCRAM-SHA-256 login, the simple and extended query protocols with parameters, text and binary decoding of the common types, transactions, `COPY`, and a connection pool. Migrations come later as a small add-on.

## Context

- XL as a whole: split at refinement into at least protocol and login, typed values, and pool.
- No TLS until a `tls` package exists; local or proxied connections only.
- Tests need a PostgreSQL server; decide in refinement how `scripts/test.sh` finds one or skips (and tag the node `exclusive: [postgres]` if it uses a fixed port).
