---
id: std-gap-survey
title: "Which std modules a first user will miss"
status: active
priority: could
type: spike
story: batch-22
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: true
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-gap-survey/**
shared:
reserves: []
provides: []
depends_on: []
---
# Which std modules a first user will miss

## Goal

Maintainer, 2026-11-20: "any modules we are missing?" Compare the 55 std modules (`docs/reference/std/README.md` on this branch; after std-async-fold `std/async/*` becomes `std/net`, `Time.sleep` and `std/fs`) with what a first user of a systems language with TypeScript syntax will reach for, using the std libraries of Go, Rust, Deno/Node and Zig as the reference, and the programs in this repo (the compiler, `scripts/`, `testdata/`, the guide) as evidence of real need. Known gaps to confirm: sorting (no `Arrays.sort`), a set type, sync sockets (decided: none; net is async only).

For each missing piece: what it is, who needs it (evidence), the module it belongs in (an existing one or a new one), whether it can be written in Narra in `std/src` or needs the runtime, and a size. Rank by need. Mark the ones that should land before `v0.1.0` (a first user hits them in a first real program) and the ones that can wait. The orchestrator decides from this list.

## Acceptance

- [ ] `survey.md` in this folder with the ranked list and the before/after-release mark.
