---
id: c-seed
title: Bootstrap from a checked-in C seed and retire the Go compiler
status: active
priority: must
type: story
story:
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: the stage 2 test path must exist before the Go tests can go
---

# Bootstrap from a checked-in C seed and retire the Go compiler

## Goal

After this story the repository has no Go compiler. A checked-in C file, `seed/narrac.c`, builds the self-hosted compiler with `cc` alone. The self-hosted compiler has the user commands (`build`, `run`, `check`, `test`), and the whole test suite runs on it through a test runner written in Narra. Every language feature is then built once, in Narra.

## Context

- Decided with the user on 2026-11-08:
  - The seed lives in the repository. GitHub Releases come later for distribution (node release-binaries, after this story).
  - Go goes away fully, including the Go test harness (`compiler/internal/narra/*_test.go`, 97 tests).
  - Platforms: macOS arm64 and Linux x86_64.
  - The backend stays C, not LLVM. C keeps the seed the same on every platform, and clang already optimizes through LLVM.
  - The rest of the roadmap waits until this story merges, so later nodes are built once.
- Strategy: one story branch, `feat/c-seed`. Each child lands on it as a direct commit after its targeted tests. There are no per-child pull requests and no CI. The full suite runs once at the end, then one fresh review and one pull request for the maintainer.
- The seed rule: the compiler's own source may only use features the current seed compiles. A new feature lands in the compiler first, written without itself; then the seed is regenerated; only then may the compiler source use it.

## Children

seed-parity-spike, selfhost-driver, checker-codes, checker-fixes, test-runner-core, goldens-tools, goldens-programs, seed-file, retire-go. (checker-parity and test-runner-goldens were split into these after the spike.)

## Acceptance

- [ ] `sh selfhost/bootstrap.sh` builds the compiler from `seed/narrac.c` with only `cc`, and checks the fixed point.
- [ ] The Narra test runner passes every suite that the Go tests covered, under AddressSanitizer and the leak check, on macOS and in the Linux container (`scripts/linux-check.sh`).
- [ ] No `.go` file builds or tests the compiler. `go` is not needed to build or test Narra. Go stays only as a benchmark language in `bench/`.

## After the story merges (orchestrator)

Update `GOAL.md` (verification commands, constraints that name Go), the roadmap `CLAUDE.md` project rules, and the open briefs that own Go paths (`compiler/internal/narra/...`): drop the Go half of each.
