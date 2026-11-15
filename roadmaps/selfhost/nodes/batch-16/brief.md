---
id: batch-16
title: "Batch 16: packages, distribution and docs"
status: done
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

- p4 packages: deps-manifest-lock, deps-fetch, deps-resolve (both amended from D101), package-dependencies, package-exports, workspace-design (spike, gate; done, D101), selfhost-layout (last, after the seed update).
- p5 distribution: target-table, wasi-runtime, windows-runtime, windows-process-async, cross-targets, js-interop-design (spike as a goal debate, gate; done, D123), narrac-wasm (after wasi-runtime), npm-narrac (after narrac-wasm and cross-targets), c-abi-design (spike, gate; done, D122 accepted 2026-11-15), release-binaries (gate; done, landed as direct commits and reviewed by a fresh agent, 2026-11-15).
- p6 docs: dead-code-check, let-never-reassigned, guide-tools, install-guide, textmate-grammar, website (gate), narra-skills (gate).
- Not in the batch, with the reason:
  - workspaces and its six children (workspace-manifest, workspace-loader, git-member-deps, workspace-commands, workspace-lsp, repo-workspace): the maintainer kept them out, and they need deps-fetch, deps-resolve and package-dependencies landed first. They start in the next batch.
  - c-export-functions, c-abi-export, node-addon, wasm-js-glue (M each): D122 is accepted (2026-11-15). They stay out of this batch: they edit the parser, emitter and driver, which many batch children already edit, and the batch is large. They start in the next batch.
  - c-layout: parked, on neither D123 path.
  - website-repl: parked by the maintainer; D123 proposes no children.
  - std-serve, js-client, node-addon-async: outlines in D123, not scheduled.

## Order and shared files

- Machine: 8 GB, at most four building workers at once; hold new builds while load is above 30.
- Gates first: workspace-design, js-interop-design and c-abi-design are done and their gates passed.
- Most nodes touch `selfhost/src/*.nra`; feature logic goes in its own module where it can, shared files carry dispatch rows only. The second to land rebases and keeps both sides.
- `diagnostics.nra` is sorted by code and guarded by the `codes` suite: insert rows at their code position, never append.
- Codes: deps E0430-E0479, dead-code-check E1500-E1509, let-never-reassigned E1510-E1519, distribution E1600-E1604 (target-table and the runtime nodes). Reserved for later batches: E1610-E1619 (c-layout, parked), E1620-E1623 (c-export-functions), E1630-E1632 (c-abi-export), E1640-E1649 (node-addon), E1650-E1659 (wasm-js-glue).
- Generated files (`stdtext.nra`, `runtime*.nra`) only through `sh scripts/test.sh --suite generated --update`; after each landing, run the generated and guide suites.
- The dependency suites need `git` on PATH and in the Linux check image, with local fixtures only.
- selfhost-layout moves every file in `selfhost/src/` and changes every C name; it runs alone after every other code child has landed.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. let-never-reassigned sweeps the compiler only after the batch seed update. Runtime text edits (wasi-runtime, windows-runtime) need two seed updates at the end of the batch.

## After the seed update

Checks that wait for the batch seed update, collected from the children:

- [x] `sh scripts/cross-check.sh` with the seed check on, for Windows-X86_64 and Wasi-Wasm32 (cross-targets, windows-runtime, wasi-runtime). Done: all 13 samples and the seed check pass on Windows-X86_64 (wine), 9 ran and 4 left out on Wasi-Wasm32, after the fixes in 415086a0 and a0767ff8.
- [x] `sh scripts/wasm-build.sh` with its default seed input (narrac-wasm). Done: 6509609 bytes, 1027727 bytes brotli.
- [x] `narrac build --target Windows-X86_64 selfhost`; if it builds, add the `narrac-win32-x64` package and row to `npm/` and `scripts/npm-package.sh` (npm-narrac dropped it on a `poll.h` failure before windows-process-async landed). Done: the build works, the exe runs `check` and `emit` under wine, the package and row are added and `npm-package.sh` packs it (5.3 MB tarball, not published).
- The narra-skills eval run, alone, with its cost recorded.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
  - Evidence: every child listed above is `status: done`; the C chain, workspaces and website-repl are out of the batch with reasons above.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
  - Evidence, after the review fixes and the last seed update: `test.sh: PASS, 10132 cases in 178s`; `linux-check: PASS in 408s`; `cross-check.sh Windows-X86_64 Wasi-Wasm32` with the seed check: `PASS`, seed emits the same C as the host on both.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
  - Evidence: after the review fixes, a second `sh scripts/seed-update.sh` printed `seed is current` and `seed unchanged`.
- [x] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
  - Evidence: the final seed built with `cc -O1 -g` and no `-DNARRA_RELEASE`; `NARRA_CHECK_COPIES=1 narrac emit selfhost` exit 0.
- [x] A fresh review of the whole story branch, before the pull request.
  - Evidence: three fresh reviewers (packages, distribution, checks and docs). Two must-fix items (a plain-tag checkout reused for a member package; CRLF checkouts on windows-latest) and the should-fix items were fixed on fix/batch-16-packages and fix/batch-16-distribution and landed here.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
  - Evidence: instructions 3.561G (main compiler) to 3.596G (batch-16 compiler) on the batch-16 source, +1.0%, inside the 2% spread.
