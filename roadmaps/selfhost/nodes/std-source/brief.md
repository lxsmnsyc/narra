---
id: std-source
title: Std modules as Narra source
status: todo
priority: should
type: spike
story:
early: false
outline: false
milestone: m10-std
target:
ship: pr
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
spent:
owns:
  - std/**
  - compiler/testdata/std_source*
  - roadmaps/selfhost/nodes/std-source/**
shared:
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/std_option.go
  - compiler/internal/narra/modules.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/driver.go
  - selfhost/src/builtins.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitoptions.nra
  - selfhost/src/modules.nra
  - selfhost/src/scopes.nra
  - selfhost/src/main.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: std grows on the self-hosted compiler, so nothing is built in Go and ported
---

# Std modules as Narra source

## Goal

Answer one question, with numbers: can std modules ship as Narra source files bundled with the compiler, so that only primitives (memory, syscalls, string bytes, the generic intrinsics) stay compiler builtins? Prove it by porting `std/option` from builtins to source, measure what it costs, and write the answer to `outcome.md`. The user decides at the gate how the rest of m10 is built.

## Context

Today every std function is a builtin in two places that must match:

- Go: `registerModule` / `registerNamespaceModule` in `compiler/internal/narra/std_*.go`, each entry with a Check and an Emit closure that writes C. C helpers come from embedded `runtime*.h` files (`go:embed`), split into sections per module (`ioRuntime` in `std_fs.go`).
- Narra: the hand-written table `STD_FUNCS` in `selfhost/src/stdlib.nra` (signature text plus the name of the Go Check function), checks dispatched by name in `selfhost/src/stdcalls.nra`, emit code in `selfhost/src/emit*.nra` (Option in `emitoptions.nra`), and the runtime C as generated string constants in `selfhost/src/runtimetext.nra` (`NARRA_UPDATE=1 go test -run TestSelfhostRuntimeText`).
- `TestSelfhostBuiltins` fails when the two registries differ, `TestSelfhostEmit` compares the C of both compilers on every `testdata/*.nra`, and `TestSelfhostBootstrap` checks the fixed point. So a source std module must be loaded by both compilers and produce byte-identical C.
- Imports of `std/...` are resolved by the checker from the registry (`selfhost/src/scopes.nra`, `check.go`). The module loaders (`modules.go`, `selfhost/src/modules.nra`) load only the package's own `src/` tree. The package name `std` is already reserved (`modules_toml.go`). `[dependencies]` is not supported yet.
- `std/option` (D69) is seven generic functions with callbacks and effect variables: `map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr`. User code supports `<effect E>` and non-escaping arrow arguments with stack environments (D46, D58), so the port should be expressible today.
- Sources: GOAL.md decision 2026-11-01 (m10-std), D67, D69, D72, SPEC.md sections 7 (modules) and 12 (build model).

## Steps

Time box: one session. Stop and write what you have when the box is spent; a partial answer with the blocker named is a valid result.

- [ ] Classify every entry of `STD_FUNCS` (about 75) as primitive (needs C or a compiler intrinsic: memory, syscalls, string bytes, `Array.push`, `Map.*`, `Box.new`, `Range.*`, `std/test`) or source-able (expressible in Narra over primitives). Write the table in `outcome.md`.
- [ ] Decide where std source lives (proposal: `std/src/*.nra` at the repo root, a package named `std`) and how each compiler finds it. Compare at least: embedded in the binary (Go `go:embed`; selfhost a generated string-constant file like `runtimetext.nra`) against a path found at run time (`NARRA_STD` or next to the executable). Builds must stay deterministic (SPEC section 12).
- [ ] Decide how a source module reaches a primitive without c-interop: keep the primitive as a registry builtin under an internal path (for example `std/internal/...`, not importable by user packages), or a small declaration form such as `@builtin`. Name the form and its rules.
- [ ] Port `std/option` to source in both compilers: load the std source modules on import, check them with the package, emit only functions that are used, and remove the seven Option builtins (Go Check/Emit, `STD_FUNCS` rows, `stdcalls.nra` dispatch, `emitoptions.nra` code). `compiler/testdata/std_option.nra` and `options.nra` keep their output.
- [ ] Measure before and after: `narrac emit selfhost` time (warm, three runs), `go test ./internal/narra -run TestSelfhostBootstrap` time, stage 2 binary size (`-O2`), and the generated C line count and binary size of `std_option.nra`. Count lines removed and added in both compilers.
- [ ] Decide what the Go compiler does with std source. GOAL.md non-goals keep the Go compiler only for what bootstrapping needs. Cost each option:
  - (a) Go loads std source too, so every source module works in both compilers and every test keeps comparing them. Cost: each loader and checker change is made twice for the rest of m10.
  - (b) Go keeps only the builtins that `selfhost/src` uses, and new std modules are selfhost-only. This is the default unless the spike finds it blocks testing. List what breaks: Go-run tests of user programs that use new std (`TestPrograms`, `TestCompileErrors` and the runner tests run on the Go compiler), and `TestSelfhostEmit` comparisons for those programs. Say how such programs would be tested instead (for example a test that runs stage 2 on them), and what that costs in test time.
  - (c) Estimate only, build nothing: a checked-in C seed of the selfhost compiler (stage 1 output) that replaces the Go compiler for bootstrapping, as a separate future node. Give its size (C lines and bytes of the seed, the update procedure) and its risks (review of a generated file, platform C differences, how a seed is refreshed after a language change).
- [ ] Write the answer: yes, yes with a named primitive form, or no (and what is missing; if it is c-interop, say which part), together with the recommended Go option.

## Acceptance

- [ ] `outcome.md` answers the question with one of the three results, the primitive table, the chosen location and lookup, the primitive form, and the measurements above.
- [ ] If the answer is yes: the `std/option` port passes `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBuiltins`, `TestSelfhostEmit` and `TestSelfhostBootstrap`, with every sample clean under ASan and the leak check. Error tests in `compiler/testdata/errors/` that mention Option calls still pass, or the outcome lists each changed message.
- [ ] If the answer is no: the prototype is not merged, and the outcome says what c-interop (or other) work would unblock it.
- [ ] `outcome.md` has the cost of options (a) and (b), the list of what breaks under (b) and how it is tested instead, a recommendation, and the size and risk estimate for (c).
- [ ] `outcome.md` has a short plan for each m10 node (std-json, std-basics, std-iterator, std-arena, std-unicode, std-spec-functions): which parts are source and which are primitives.

## Sizing

Uncertainty high (that is why it is a spike), coupling low, blast radius medium (module loading and the checker's import resolution in both compilers), verifiability medium (tests plus measurements, then the gate), size medium, context high (two compilers' std code). Kept at M by the time box.

## Notes for the worker

- The pull request holds `outcome.md` and the prototype. Because the node is gated, the user decides whether the prototype merges or only the outcome does. Keep the prototype in a state that can merge.
- Node ids: selfhost modules get id ranges `[firstId, endId)` in `modules.nra`, and `narrac pkg --ids` checks them. Std source modules need ranges too, and the order must be the same in both compilers.
- C names are hashed (`cnames.go`, `selfhost/src/cnames.nra`) from the module path. Std source functions must hash the same in both compilers, and a user module path must not collide with `std/...`.
- Effects: the Option builtins today take their effects from the callback (`Builtin.Callbacks`). A source port needs `<effect E>` on each function. Check that exported std functions still pass the effect rule (D58) and that a call with a pure callback stays pure.
- Allocation: D69 promises that a capturing arrow passed directly allocates no closure. Check the generated C of `std_option.nra` before and after; a heap environment where there was none is a regression to report.
- Error messages: today `checkOptionMap` and friends give Option-specific messages. Generic instantiation errors read differently. List the differences instead of adding special cases.
- `TestSelfhostBuiltins` compares the registry dumps. Removing Option rows changes both dumps; keep them equal.
- Do not change other std modules. Do not add `[dependencies]` support; std is not a dependency package.

