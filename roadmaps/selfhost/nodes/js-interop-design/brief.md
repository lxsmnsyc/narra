---
id: js-interop-design
title: "Spike: Narra in JS and TS packages, and the website REPL"
status: active
priority: could
type: spike
story: batch-16
early: false
outline: false
milestone: p5-distribution
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/js-interop-design/**
  - roadmaps/selfhost/debate/js-interop/**
shared:
  - DECISIONS.md
reserves: [decision D123]
provides: []
depends_on: []
---

# Spike: Narra in JS and TS packages, and the website REPL

## Goal

Decide how code written in Narra reaches the JS and TS ecosystem, the way Go and Rust do today, and how the website REPL compiles and runs Narra. Run it as a goal debate (the goal-debate skill, output in `roadmaps/selfhost/debate/js-interop/`), write the answer as decision D123, and propose the child briefs. Maintainer, 2026-11-14: "how would we utilize Narra for JS/TS packages ... esbuild and oxc are such examples. We will have this problem when we decide to do the REPL for the website."

## Prior art to weigh

- esbuild (Go): a native binary per platform in npm packages (`optionalDependencies` with `os`, `cpu`, `libc`), run as a child process; `esbuild-wasm` for browsers.
- oxc and swc (Rust): Node-API addons through napi-rs, plus wasm builds (`@oxc-parser/wasm`, `@swc/wasm-web`) for browsers.
- Playgrounds of compilers that emit C or native code: how the Go, Rust, Zig and Nim playgrounds compile (server or in-browser).

## Questions

1. **npm form.** A native binary per platform, a Node-API addon, a wasm module, or more than one. What each costs in this compiler: binaries need cross-targets and release-binaries only; Node-API needs the C library boundary (nodes c-abi-design to c-abi-export, held for this decision); wasm needs wasi-runtime and a JS loader.
2. **REPL compile and run path.** narrac emits C, so pick one, with measured sizes and times:
   - narrac in wasm for check, fmt and query only (diagnostics, no run).
   - narrac in wasm plus a C compiler in wasm (clang, or tcc with a wasm backend) compiling the emitted C in the page.
   - Compile and run on a server (a Vercel function), with the cost of running untrusted code.
   - A wasm backend or an interpreter in the compiler.
3. **Browser runtime.** What the runtime needs in a browser that wasm32-wasi under wasmtime does not give (a WASI shim, threads, async, files, stdout).
4. **First user.** Which package or page proves the choice: the REPL, or a small npm package that calls Narra code.

## Proofs

Small capped proofs settle feasibility (30 minutes and 100k tokens each): build `seed/narrac.c` for wasm32-wasi with zig and run `narrac check` on a sample under wasmtime; measure the wasm size; try one C-in-wasm compiler on an emitted sample. This machine has zig, wasmtime, docker, Node 26 and llvm@22.

## Acceptance

- [ ] An approved `debate.md` in the debate folder, with the costs weighed on device, time and tokens.
- [ ] D123 written in the style of D100 (Decision, Reason, Cost, Rejected), marked "Not built yet".
- [ ] Proposed child briefs in `outcome.md`, including whether the held C library chain (c-abi-design, c-layout, c-export-functions, c-abi-export) is kept or dropped, and the children of website-repl.
- [ ] The maintainer reads D123 before any child starts (gate).

## Sizing

M: a debate and a decision with small proofs, no shipped code.
