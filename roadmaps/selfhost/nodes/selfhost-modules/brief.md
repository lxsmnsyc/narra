---
id: selfhost-modules
title: Self-hosted package loading
status: done
spent: 512000
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestPkgDump|TestSelfhostModules|TestSelfhostParser'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - compiler/internal/narra/pkgdump.go
  - compiler/internal/narra/pkgdump_test.go
  - compiler/testdata/pkgdump/**
  - selfhost/src/manifest.nra
  - selfhost/src/modules.nra
shared:
  - compiler/cmd/narra/main.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/internal/narra/modules.go
  - compiler/internal/narra/modules_toml.go
  - selfhost/src/main.nra
  - selfhost/narra.toml
  - selfhost/src/parser.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-front
    kind: hard
    why: reads each module's imports from the parsed AST
---

# Self-hosted package loading

## Goal

Port `compiler/internal/narra/modules.go` and `modules_toml.go` (569 Go lines) to Narra: read `narra.toml`, walk `src/**`, parse each module with `narrac/parser`, order modules by imports, and resolve binaries. Compare with Go through a package dump, as the lexer and parser are compared.

## Context

- Split from selfhost-front on 2026-10-30. Read `selfhost/CLAUDE.md`, the selfhost-front outcome (parser API, node id scheme, error flow) and `contracts/module-graph.md` (v3, frozen).
- Dump on the Go side: `PkgDump(path) (string, bool)` in `pkgdump.go` and a hidden command `narra pkg <path>`, with the format defined in a doc comment at the top of the file, as `astdump.go` does. Print the package name, each `[[binary]]` (name, entry, resolved entry module and function), `[link]` libraries and search, and the modules in load order (module path, file path relative to the package, exported names in sorted order, imports in source order). Do not repeat the AST; the parser dump already covers it. An error prints `error <file>:<line>:<col>: <message>` (or the Go error text when there is no position) and exits 1, matching `FormatError`.
- `path` is a package directory or a single `.nra` file, as `Load` accepts.
- Narra side: `manifest.nra` (the TOML subset of `parseManifest`) and `modules.nra` (`load(path): Result<Package, Diag>`, module ordering, cycle errors, `binary`, `entryModule`). Wire `narrac pkg <path>` in `main.nra` with the same output.
- Number the whole package in one id sequence: pass each file's `Parsed.nextId` as the next file's `firstId`, in load order, and keep the final `nextId` on the package. selfhost-check-decls sizes its side tables from it.
- Walk `src/` in a fixed order (sort directory entries) so the dump and ids are stable. `Fs.listDir` returns Err on a file, so recurse on Ok (spike gap 15). If the Go walk order differs from sorted order, match Go and say so in the outcome.
- `TestSelfhostModules` in `selfhost_test.go`: build `selfhost/` once in debug mode (ASan, UBSan, `NARRA_CHECK_LEAKS=1`) and compare `narrac pkg` with `narra pkg` (stdout, stderr, exit code) on every package under `compiler/testdata/packages/` (including `errors/`), every single-file sample, `selfhost/` itself, and manifest edge cases in `compiler/testdata/pkgdump/` (bad TOML, `[dependencies]`, duplicate binary, bad package name, `std` as a name, import cycle, missing entry module).
- Fold in two nits from the selfhost-front review: `narrac` ignores the `Result` of `Fs.writeText("/dev/stdout", ...)` (return 1 on Err), and the comment in `selfhost/narra.toml` still says only the lexer is ported.
- Optional, if cheap: build `selfhost/` once for all the selfhost tests (`sync.Once`) instead of once per test.

## Acceptance

- [x] `narra pkg <path>` prints the dump, its format is defined in the doc comment of `pkgdump.go`, and `TestPkgDump` compares goldens in `compiler/testdata/pkgdump/`.
- [x] `TestSelfhostModules` passes: `narrac pkg` equals `narra pkg` on every input listed above.
- [x] `TestSelfhostLexer` and `TestSelfhostParser` still pass.
- [x] The outcome describes the Narra `Package` and `SourceModule` types, the id sequence across files, and the load order.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 800 Narra lines (569 Go lines at the 1.3-1.45x ratio seen so far), a small Go dump, one harness.

## Notes for the worker

- Edit `modules.go` or `modules_toml.go` only to expose what the dump needs or to fix a bug the comparison finds; record it in the outcome.
- `export` lists depend on the parser's `exported` field; imports come from `ImportDecl`.
