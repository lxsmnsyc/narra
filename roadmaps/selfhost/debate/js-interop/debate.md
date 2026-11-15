---
goal: Decide how Narra code reaches JS and TS packages, the way esbuild (Go) and oxc (Rust) do, and how the website REPL compiles and runs Narra
date: 2026-11-15
rounds: 2 of 2
status: proposed
roadmap: roadmaps/selfhost
supersedes: none
decision: DECISIONS.md D123
---

# Debate: Narra in JS and TS packages, and the website REPL

Full debate: a proponent and an opponent subagent, two rounds, with the orchestrator of node js-interop-design as judge. The debate stopped before the approval step and sent Q1 and Q2 to the maintainer. The answers came back on 2026-11-15 and overruled two parts of the settled structure. This file shows the structure after the answers. Proof T3 was run after them.

## Settled structure

1. **Two forms, picked per package** (Q2, P10). The command form is esbuild's shape: JS runs the Narra program as a separate process, or as a wasm module. The Node-API form is oxc's shape: JS loads an addon and calls Narra functions in its own process. `[npm] form` in `narra.toml` picks one.
2. **Shared by both** (P1 as overruled by Q2, P2, P8). One build per row of contract `targets`. One npm package per platform, with `os`, `cpu` and `libc`, in the main package's `optionalDependencies`. A wasm package as the fallback. One JS loader that picks the platform package or the wasm one and holds the WASI shim (P5, J14).
3. **Runtime fixes on wasi** (P2, O4). `Env.args` and output through `/dev/stdout` fail on every wasi program today. Both join wasi-runtime.
4. **Command form** (P8, P13). A `bin` entry and a JS API that runs the program, as a child process or a fresh wasm instance. A request protocol that keeps one process alive is outlined (P9, O6, O7).
5. **Node-API form** (Q2, J19). The C library chain comes back without c-layout. Generated Node-API glue per export serves the native addon, and through emnapi the wasm fallback as well. T3 measured 0.18 µs per call against 93 to 325 ms for a fresh wasm instance (J15).
6. **First package** (P6). narrac, in the command form, with its wasm fallback (T1).
7. **REPL parked** (Q1). The measurements for check in the page (J15) and Run with clang in wasm (J5, J18) stay for later.

## Budget

- device: a wasm fallback of 0.99 MB brotli for narrac (J2). Native addons of about 51 KB plus the program (J19). The build machine needs zig for the cross rows, which target-table already needs (J4, J8).
- time: shared nodes narrac-wasm S, npm-loader M, npm-package M, npm-narrac M, and two items in wasi-runtime. Node-API nodes c-abi-design M, c-export-functions M, c-abi-export M, node-addon M, node-addon-wasm M, npm-addon-sample S.
- tokens: ten nodes. The debate's first answer, wasm only, had four.

## Rejected

- Wasm only for every package (the first settled structure): a library called in a loop pays a fresh instance per call (O6, J15). Overruled by Q2.
- Node-API only: it puts a C boundary under CLIs such as narrac that need none (P7, O5).
- Precomputed output instead of Run (A2 second half): website-repl's goal is that the reader types and runs. Moot while the REPL is parked.
- A precompiled runtime header to speed Run (A3). The runtime is about 0.4 s of a warm Run, and the rest is the program's own C (J18).
- An interpreter for Run. The D105 evaluator runs an alloc-only subset with no `print`, closures or generators (J9).
- A wasm backend in the compiler. D10 picked C.

## Proofs

- T1 for J2: does the seed build for wasm32-wasi and run `check` under wasmtime? Done. Yes, after three patches: 6.23 MB stripped, 0.99 MB brotli, 0.07 s cached per check (cap 30 min, about 45k tokens). `evidence/T1-narrac-wasm.md`.
- T2 for J5: does a C compiler in wasm compile emitted Narra C? Done. Yes, with @yowasp/clang 22: 72 MB module, 0.5 to 2.9 s warm, 6 to 17 s cold (cap 30 min, about 25k tokens). `evidence/T2-clang-wasm.md`.
- T3 for J19: does emitted Narra C run inside a Node-API addon? Done, after Q2. Yes: 51 KB `.node`, 0.18 µs per call, no init call needed (cap 30 min, about 15k tokens). `evidence/T3-node-addon.md`. It is the third proof, one over the skill's limit of two, and the script warns about it.

## Prior art

- esbuild: a native binary per platform through `optionalDependencies`, plus esbuild-wasm (14.5 MB unpacked). Its JS API talks to the binary over stdin and stdout (J10, J17).
- oxc and swc: Node-API addons through napi-rs, with the same per-platform layout, plus wasm builds of 1.5 MB and 18.4 MB (J10). napi-rs runs its wasm builds through emnapi.
- The Go and Rust playgrounds and Compiler Explorer compile and run on servers in sandboxes. No major compiler playground compiles C in the page (J11).

## Open risks

- The addon's wasm fallback through emnapi is not proven. It is the first acceptance item of node-addon-wasm.
- A panic inside an addon call ends the Node process today (T3 did not test it). c-abi-design decides the rule.
- A Narra program that uses threads, sockets or processes cannot target Wasi-Wasm32 (O9). Accepted: its packages ship native only, and `narrac package` says so.
- Firefox, Safari and phones are unmeasured (O3, O8). Accepted while the REPL is parked.

## Inferred choices

- The only users are the two the maintainer named: JS and TS packages written in Narra, and the website REPL (J12).
- "Narra for JS/TS packages" means any Narra program or library, with narrac first (J17).

## Answers from the user

- Q1, the REPL Run button: park the REPL for now. Keep the measurements. No REPL children, and website-repl goes to priority later.
- Q2, the JS API form: "ideally it should be either". An author picks the esbuild form (a wasm module or a native binary run as a separate process) or the oxc form (a Node-API addon in process). Both have native packages per platform and a wasm fallback.

## Overruled

- P7 (park the C chain): rejected by Q2. The chain comes back for the Node-API form, without c-layout.
- P1 (native per-platform packages): upheld by Q2, in both forms.
- P3, P11, P12, A1 (REPL check and Run): left open by Q1, because the REPL is parked.
- P10 (cost of both Q2 branches): both branches are taken.

## Cost of this debate

- 2 of 2 rounds and three proofs, about 260k tokens.
