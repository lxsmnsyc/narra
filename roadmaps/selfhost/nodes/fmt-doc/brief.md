---
id: fmt-doc
title: "fmt: document IR and printer"
status: todo
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmtDoc|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - selfhost/src/fmtdoc.nra
  - compiler/internal/narra/selfhost_fmtdoc_test.go
  - compiler/testdata/fmt/doc_*
  - roadmaps/selfhost/nodes/fmt-doc/**
shared:
  - selfhost/src/main.nra
reserves: []
provides: [fmt-doc]
depends_on: []
---

# fmt: document IR and printer

## Goal

`selfhost/src/fmtdoc.nra`: the `Doc` IR and `print` of `contracts/fmt-doc.md`, a pure library with no parser or AST dependency. The story brief (`nodes/fmt/brief.md`) has the formatter's rules.

## Context

- References: Wadler, "A prettier printer"; Prettier's `src/document/printer.js` (`printDocToString`, `fits`, modes break and flat, `lineSuffix`, `breakParent` propagation).
- Selfhost tools are tested from Go: `compiler/internal/narra/selfhost_test.go` builds `narrac` once and `runNarrac` compares output. selfhost/src has no `test` blocks today, so unit tests run as a hidden subcommand.

## Acceptance

- [ ] `narrac fmt-doc <file>` reads a small text description of a document (one constructor per line or an s-expression form the worker picks, documented at the top of `fmtdoc.nra`) and prints it with given options. Golden cases `compiler/testdata/fmt/doc_<case>.txt` and `.out`, run by `TestSelfhostFmtDoc`, cover: a group that fits, one that breaks, nested groups where only the outer breaks, `HardLine` breaking parents, `BreakParent`, `IfBreak` by enclosing group and by group id, `LineSuffix` flushed before a newline, `join`, indentation with spaces and with tabs, width exactly at the limit and one over, no trailing whitespace.
- [ ] A document nested 50000 levels deep prints without a stack overflow.
- [ ] TestSelfhostBootstrap passes with the new module in the package.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty medium (known algorithm, new to the repo), coupling low (no dependencies), blast radius low (new module, one dispatch line), verifiability low (golden tests), size medium (3 to 4 files, about 500 lines), context low.

## Notes for the worker

- Iterative printer with an explicit stack of (indent, mode, doc) and an iterative `fits`; recurse only where the selfhost rules allow (selfhost/CLAUDE.md).
- `Text` must hold no newline; do not check it at runtime in the hot loop, only in the text-description reader.
- fmt-trivia runs at the same time and also adds one dispatch line to `run` in `main.nra`; rebase before review.
