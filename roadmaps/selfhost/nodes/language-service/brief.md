---
id: language-service
title: Narra language service
status: todo
priority: should
type: feature
outline: true
milestone: m9-adoption
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: query-cli
    kind: hard
    why: hover, definition and references answer the same questions as narra query, from the same index
  - on: fmt
    kind: hard
    why: document formatting calls the formatter
  - on: std-json
    kind: hard
    why: LSP messages are JSON-RPC
  - on: std-io-more
    kind: hard
    why: LSP reads framed messages from stdin
---

# Narra language service

## Goal

`narra lsp`, a Language Server Protocol server written in Narra on the self-hosted compiler: diagnostics on change, hover types, go to definition, find references, document symbols and formatting. It rechecks only changed modules. Editor plugins stay out of scope (a non-goal); any LSP client can use the server.

## Notes

Outline. Likely splits into a protocol and incremental-check node and a features node.
