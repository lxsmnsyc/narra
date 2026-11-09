---
id: language-service
title: "narrac lsp: protocol, diagnostics and formatting"
status: done
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent: 120798
owns:
  - selfhost/src/lsp.nra
  - selfhost/src/lspserver*.nra
  - testdata/lsp/**
  - roadmaps/selfhost/nodes/language-service/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/modules.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtconfig.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: std-json
    kind: hard
    why: every message is parsed with Json.parse and written with Json.stringify
  - on: std-io-more
    kind: hard
    why: reads Content-Length framed messages with Io.readLine and Io.readBytes, and writes with Io.write and Io.flush
  - on: fmt
    kind: hard
    why: document formatting calls the formatter (done)
  - on: query-cli
    kind: hard
    why: diagnostics reuse the load-and-check pipeline of narrac query (done)
---

# narrac lsp: protocol, diagnostics and formatting

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

`narrac lsp`, a Language Server Protocol server in the self-hosted compiler, usable from any LSP client (editor plugins stay a non-goal). This node builds the protocol layer, document sync, diagnostics on open and change, and whole-document formatting. Hover, definition, references and symbols are node lsp-navigation; rechecking only changed modules is node lsp-incremental.

## Context

- Read `nodes/std-json/outcome.md` and `nodes/std-io-more/outcome.md` first. `selfhost/src` cannot use `derives Shape` or `Json.decode<T>` (the Go compiler builds it and has neither): read messages by matching on `Value`. Small helpers such as `field(value, name): Option<Value>` go in `lsp*.nra`.
- Pipeline to reuse: `runQuery` in `selfhost/src/query.nra` loads with `Modules.load`, checks with `Body.checkPackage` and `Effects.checkEffects`, and writes diagnostics with `writeDiagnostic` (`queryjson.nra`). Diagnostics have a position, not an extent (D59); send a zero-width range. The self-hosted checker reports code `E0000` (query-cli outcome).
- Formatting: `formatSource` in `selfhost/src/fmt.nra`, with the `[fmt]` settings of the nearest `narra.toml` (`fmtconfig.nra`).
- Dispatch: one `if` in `run` in `selfhost/src/main.nra`, as for `fmt` and `query`.

## Acceptance

- [ ] Framing: `Content-Length` header (other headers ignored), body read with `Io.readBytes`; responses written with the same framing and flushed. End of input without `exit` ends the server with code 1.
- [ ] Lifecycle: `initialize` answers capabilities (full text sync, `documentFormattingProvider`, `positionEncoding` `"utf-8"` when the client offers it, else `"utf-16"` with conversion of columns), `initialized`, `shutdown`, `exit` (code 0 after `shutdown`, 1 before). Requests before `initialize` get error -32002; unknown requests get -32601; unknown notifications are ignored; a malformed message gets -32700 and the server keeps running.
- [ ] Documents: `didOpen`, `didChange` (full), `didClose` keep in-memory text that overrides the file on disk. After each open or change the server checks the package that holds the file (nearest `narra.toml`, as `fmtconfig.nra` finds it; a file with none is checked alone) and sends `publishDiagnostics` for every module of the package, an empty list for a module that has none. Load errors are reported at their file and position, or at the top of the changed file when they have none.
- [ ] `textDocument/formatting` returns one edit that replaces the whole document, or an empty list when it is already formatted; a document with a syntax error gets `null`.
- [ ] `TestSelfhostLsp` (`compiler/internal/narra/selfhost_lsp_test.go`) runs scripted sessions from `testdata/lsp/<case>/` against the stage 1 binary and compares the framed output: lifecycle, an error appearing and disappearing on change, a two-module package, formatting, UTF-16 columns on a line with non-ASCII text, a malformed message. `NARRA_UPDATE=1` rewrites the expected output.
- [ ] SPEC section 14 and `compiler/README.md` describe `narrac lsp` and the supported requests.
- [ ] Fast check passes, including `TestSelfhostBootstrap`.

## Sizing

M. New modules only plus one dispatch line and an override hook in the loader; uncertainty medium (protocol details), tests automatic.

## Notes for the worker

- `Modules.load(path)` reads files from disk. Add an optional map of overriding sources, path to text, in the smallest way that keeps `narrac check` unchanged; lsp-incremental builds on it.
- Log to stderr only; stdout carries the protocol.

Landed on feat/batch-1 (27ceacc). Gate decided by the orchestrator: messages go through std/json and std/shape; didClose rechecks from disk and clears the closed document; documents outside src/ get no diagnostics; every diagnostic has severity 1. Spent counts only the worker resumed after the device crash.
