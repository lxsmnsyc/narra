---
id: linux-check
title: "Full check on Linux in Docker"
status: todo
priority: should
type: chore
story:
early: false
outline: false
milestone: m13-distribution
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify: scripts/linux-check.sh
exclusive: [docker]
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - scripts/linux-check.sh
  - scripts/linux/**
  - roadmaps/selfhost/nodes/linux-check/**
shared:
  - compiler/internal/narra/runtime.h
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---

# Full check on Linux in Docker

## Goal

The fast check has only run on macOS. Add a script that runs `go vet ./... && go test -timeout 45m ./...` inside a Linux container, and fix what fails. First user: std-io-more's stdin loop (`fgetc` instead of `getline`, PR #91), never compiled on Linux.

## Context

- Docker runs through colima on the maintainer's machine (`colima start` first). The repo has no CI; this script is the Linux check.
- The tests need Go, a C compiler with AddressSanitizer and LeakSanitizer (clang or gcc), and whatever `run_test.go` invokes. LeakSanitizer works on Linux, so the leak check may find leaks macOS misses.

## Acceptance

- [ ] `scripts/linux-check.sh` builds a small image (Dockerfile under `scripts/linux/`), mounts the repo read-only with a writable build cache, and runs the fast check. It exits non-zero on failure and prints how to start colima when Docker is not running.
- [ ] One full run passes on Linux, or the outcome lists each failure and the fix made (portability fixes in `runtime.h` and the runtime are in scope; anything larger becomes an inbox entry).
- [ ] `compiler/README.md` names the script next to the fast check.

## Sizing

S, unless Linux shows real portability bugs.
