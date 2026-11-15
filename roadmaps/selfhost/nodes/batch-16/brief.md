---
id: batch-16
title: "Batch 16: packages, distribution and docs"
status: active
priority: could
type: story
story:
early: false
outline: false
milestone: p4-packages
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/batch-16/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-15
    kind: hard
    why: builds on the batch-15 seed
---
# Batch 16: packages, distribution and docs

## Goal

The remaining phases p4 to p6 in one batch, grouped on 2026-11-14 from the open outlines. Children land on `feat/batch-16` as direct commits after a worker check and a fresh review. At the end: seed updates, full check and Linux check, debug copy check, one fresh review of the whole branch, bench row, one pull request.

Children:

- p4 packages: deps-manifest-lock, deps-fetch, deps-resolve, package-dependencies, package-exports, workspace-design (spike, gate), selfhost-layout (last, after the seed update).
- p5 distribution: target-table, wasi-runtime, windows-runtime, windows-process-async, cross-targets, js-interop-design (spike as a goal debate, gate), release-binaries (gate, own pull request).
- p6 docs: dead-code-check, let-never-reassigned, guide-tools, install-guide, textmate-grammar, website (gate), narra-skills (gate).
- Not in the batch: workspaces (children from D101), website-repl (children from D123), and the C library chain c-abi-design, c-layout, c-export-functions, c-abi-export, held until D123 says whether Node-API is used. node-addon was dropped for js-interop-design (2026-11-14).

## Order and shared files

- Machine: 8 GB, at most four building workers at once; hold new builds while load is above 30.
- Gates first: workspace-design and js-interop-design start in the first wave so the maintainer reads D101 and D123 while the other children run.
- Most nodes touch `selfhost/src/*.nra`; feature logic goes in its own module where it can, shared files carry dispatch rows only. The second to land rebases and keeps both sides.
- `diagnostics.nra` is sorted by code and guarded by the `codes` suite: insert rows at their code position, never append.
- Codes: deps E0430-E0479, dead-code-check E1500-E1509, let-never-reassigned E1510-E1519, distribution E1600-E1604 (target-table and the runtime nodes).
- Generated files (`stdtext.nra`, `runtime*.nra`) only through `sh scripts/test.sh --suite generated --update`; after each landing, run the generated and guide suites.
- The dependency suites need `git` on PATH and in the Linux check image, with local fixtures only.
- selfhost-layout moves every file in `selfhost/src/` and changes every C name; it runs alone after every other code child has landed.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. let-never-reassigned sweeps the compiler only after the batch seed update. Runtime text edits (wasi-runtime, windows-runtime) need two seed updates at the end of the batch.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [ ] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- [ ] A fresh review of the whole story branch, before the pull request.
- [ ] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
