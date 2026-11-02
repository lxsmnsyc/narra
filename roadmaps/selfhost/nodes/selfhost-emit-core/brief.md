---
id: selfhost-emit-core
title: Self-hosted emitter: contract, state, skeleton, driver and harness
status: active
priority: must
type: feature
story:
early: false
outline: false
milestone: m8-selfhost
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit|TestSelfhostRuntimeText|TestSelfhostPackageIsClean'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - selfhost/src/emitstate.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitter.nra
  - selfhost/src/runtimetext.nra
  - compiler/internal/narra/selfhost_emit_test.go
  - compiler/internal/narra/selfhost_runtime_test.go
  - roadmaps/selfhost/nodes/selfhost-emit-core/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/check.nra              # fields only; merge before selfhost-emit-program
  - selfhost/src/hash.nra               # stub; selfhost-emit-program owns it next
  - selfhost/src/cnames.nra             # stub; selfhost-emit-program
  - selfhost/src/emitdecls.nra          # stub; selfhost-emit-decls
  - selfhost/src/emitconsts.nra         # stub; selfhost-emit-bodies
  - selfhost/src/emitexprs.nra          # stub; selfhost-emit-bodies
  - selfhost/src/emitstmts.nra          # stub; selfhost-emit-bodies
  - selfhost/src/emitmatch.nra          # stub; selfhost-emit-match
  - selfhost/src/emitclosures.nra       # stub; selfhost-emit-closures
  - selfhost/src/emittraits.nra         # stub; selfhost-emit-traits
  - selfhost/src/emitmaps.nra           # stub; selfhost-emit-maps
  - selfhost/src/emitstd.nra            # stub; selfhost-emit-std
  - selfhost/src/emitoptions.nra        # stub; selfhost-emit-options
  - compiler/testdata/emit/skips/**     # creates one file per node; each node owns its own file next
reserves: []
provides: [emitter]
depends_on:
  - on: selfhost-check-effects
    kind: hard
    why: emits from the complete checked package (bodies, instances, effects); done
  - on: selfhost-names-exported
    kind: hard
    why: reads the AST and checker under their final field names; done
---

# Self-hosted emitter: contract, state, skeleton, driver and harness

## Goal

Freeze `contracts/emitter.md` and build what every emitter node plugs into: the Emitter state and helpers, `emitProgram` with Go's section order, stub modules for every feature, `narrac emit`, the embedded runtime C, and `TestSelfhostEmit` with per-node skip files. After this node the preamble section matches Go on every input, and nine nodes can work on their own files.

## Context

- Go to port: `emit.go` (330 lines; `Emit` 46-160, helpers 162-294, `localRef`), `retain`/`release`/`needsRC` (`emit_types.go` 35-85, `types.go` 250), `cString`/`literal` (`emit_expr.go` 37-60), `driver.go` (`CompileBinary` messages), `ioRuntime` (`std_fs.go` 14).
- Contract draft: `roadmaps/selfhost/contracts/emitter.md`. Refine it while building, then freeze it.
- Runtime C: a Go generator in `selfhost_runtime_test.go` writes `runtimetext.nra` from Go's own values. These are `runtimeH`, `Module.Runtime` of every registered std module (`runtimeFor(path)`), `runtimeMapH`, `envRuntime` and `testRunnerC`. Use plain string literals with escapes (all ASCII). The test fails when the file is stale; `NARRA_UPDATE=1` rewrites it.
- Read the outcomes of selfhost-check-bodies (Walker pattern, BodyInfo layout), selfhost-check-effects ("For selfhost-emit") and selfhost-names-exported (field names).

## Rules

- Stubs report `unsupported:` only when they have work for the input. Exported functions declare their effects. `narra build selfhost` must stay clean.
- Add the contract's checker fields to `check.nra` with their defaults; do not fill them (that is selfhost-emit-program). Until `programRecorded` is set, emit.nra writes the program-facts marker at the start of the Types and Functions sections.
- Write the initial skip files from a run of narrac over all inputs. The harness has no auto-update mode for skips.

## Acceptance

- [ ] `contracts/emitter.md` matches the code, with every dispatch signature in it; the outcome asks the orchestrator to freeze it.
- [ ] `narrac emit [--tests] [--bin name] <path>` exits 0, 1 or 3 as the contract says. Go's no-entry and unknown-binary messages are ported.
- [ ] `TestSelfhostEmit` runs every input and enforces U == S. The preamble section is byte-equal to Go on every input, including std module runtimes in `checker.used` order and std/test only in test builds. It logs compared and skipped counts per node.
- [ ] `TestSelfhostRuntimeText` passes; earlier selfhost harnesses keep their counts; `narra emit` output unchanged.
- [ ] Skip files exist for each of the nine nodes and list exactly the inputs that report that node.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M (upper end): about 650 Narra lines of real code plus stubs, and about 300 Go lines of harness and generator; the generated runtime text is not hand-written. If it grows past about 900 Narra lines, move the generator to selfhost-emit-decls.

## Notes for the worker

- `Emitter.checker` is read-only. Do not pass it `mutable` anywhere: a write copies every checker table.
- `tempCount` and `literalCount` are program-wide in Go; never reset them per function.
- Use string builders everywhere. The selfhost package emits 2.5 MB of C.
- `go test` caches ignore `selfhost/`; use `-count=1`.
