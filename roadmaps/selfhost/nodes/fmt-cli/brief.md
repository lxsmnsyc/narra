---
id: fmt-cli
title: "fmt: config, CLI, declarations and statements"
status: done
priority:
type: feature
story: fmt
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmt|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - selfhost/src/fmt.nra
  - selfhost/src/fmtconfig.nra
  - compiler/internal/narra/selfhost_fmt_test.go
  - compiler/testdata/fmt/decl_*
  - compiler/testdata/fmt/stmt_*
  - compiler/testdata/fmt/comment_*
  - compiler/testdata/fmt/config_*/**
  - roadmaps/selfhost/nodes/fmt-cli/**
shared:
  - selfhost/src/fmtprinter.nra
  - selfhost/src/main.nra
  - compiler/README.md
  - SPEC.md
reserves: []
provides: [fmt-printer]
depends_on:
  - on: fmt-doc
    kind: contract
    via: fmt-doc
  - on: fmt-trivia
    kind: contract
    via: fmt-trivia
---

# fmt: config, CLI, declarations and statements

## Goal

The working `narrac fmt` command: the `[fmt]` config, the CLI, the driver that builds the `Printer` of `contracts/fmt-printer.md`, the printing of files, imports, declarations and statements, and the test harness with the idempotence and token checks. Expressions, types and patterns come from fmt-exprs through `fmt-printer`. The rules (options, defaults, CLI, blank lines) are in the story brief `nodes/fmt/brief.md`.

## Context

- CLI dispatch: `run` in `selfhost/src/main.nra`; `Emitter.runEmit(arguments)` is the pattern for a subcommand in its own module.
- `narra.toml` reading: `selfhost/src/manifest.nra`. Both it and the Go loader (`modules_toml.go`) already skip tables they do not know, so `[fmt]` needs no loader change. `fmtconfig.nra` reads the `[fmt]` keys itself, reusing manifest helpers where they are exported, and reports an unknown key or a wrong value type as `error narra.toml:<line>: ...`.
- Package files: `Fs.listDir` and `collect` in `selfhost/src/modules.nra`. Stdin: check what `std/fs` offers; reading `/dev/stdin` through `Fs.readText` is the current way (std-io-more adds a real stdin later).
- Tests: `compiler/internal/narra/selfhost_test.go` (`selfhostBuild`, `runNarrac`).

## Acceptance

- [ ] `narrac fmt [paths]`, `--check` and `--stdin` / `-` behave as the story brief says; a file that does not parse prints `error <path>:<line>:<col>: <message>`, is left unchanged, and the exit code is 1.
- [ ] `[fmt]` config with `printWidth`, `tabWidth`, `useTabs` and `trailingComma`, found in the nearest `narra.toml` above each file; defaults without one. Cases `compiler/testdata/fmt/config_<case>/` (a `narra.toml` and input and expected files) cover each key and a bad key.
- [ ] Golden cases `compiler/testdata/fmt/{decl,stmt,comment}_<case>.nra` and `.fmt` for imports, every declaration kind, every statement kind, blank lines (kept, collapsed to one, none at block start), and comments in every placement, run by `TestSelfhostFmt` through `--stdin` and through `--check`.
- [ ] `TestSelfhostFmt` also formats every `.nra` file under `compiler/testdata/` and `selfhost/src/` and checks: formatting twice gives the same text; the token sequence (kind and text) and comment texts are unchanged; the `narrac parse` dump with positions removed is unchanged. With the fmt-exprs stub these checks run on declarations and statements only; they cover everything once fmt-exprs lands.
- [ ] `SPEC.md` section 2 "Statements and layout" lists the `[fmt]` keys and their defaults; `compiler/README.md` documents `narrac fmt`.
- [ ] TestSelfhostBootstrap passes. Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty medium (statement layout choices follow Prettier), coupling medium (three dependencies, all through contracts), blast radius low (new modules, one dispatch line, two doc sections), verifiability low (golden, idempotence and token tests), size medium (6 files, about 800 lines including tests), context medium.

## Notes for the worker

- Merge into the story branch before fmt-exprs, which rebases on this branch and replaces the stub fields.
- Write `fmtprinter.nra` exactly as the contract says; if a change is needed, write `contracts/fmt-printer.change-request.md`.
- Do not format the repo's own files in this node; the stability checks compare the formatter with itself.
- Layout follows Prettier's TypeScript output where Narra has the same construct (`if`/`else`, `for`, `while`, `return`, declarations, imports); where Narra differs, pick the closest Prettier rule and list the choice in `outcome.md`.
