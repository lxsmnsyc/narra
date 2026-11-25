---
id: pkg-websocket
title: "websocket package over http"
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
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-websocket/**
  - packages/websocket/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: pkg-http
    kind: hard
    why: the handshake runs through the http server's upgrade hand-off
  - on: std-digest
    kind: hard
    why: the handshake key is SHA-1 then base64
---

# websocket package over http

## Goal

RFC 6455 WebSocket server and client on top of `packages/http`: the upgrade handshake, framing, masking, fragmented messages, ping and pong, close codes, and limits on message size. No extensions at first (permessage-deflate waits for a `compress` package).

## Context

- Motivating user: porting a game server whose browser client talks to it over WebSocket. The Autobahn test suite is the reference for conformance.
