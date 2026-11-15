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

Full debate: a proponent and an opponent subagent, two rounds, the orchestrator of node js-interop-design as judge. Stopped before the approval step. Q1 and Q2 go to the maintainer with the recommended answers below.

## Settled structure

1. **wasm first, for narrac and for any Narra program** (P2, P8, P13, A2 first half). A Narra program reaches npm as one package that holds the program built for `Wasi-Wasm32` plus a small JS loader with a WASI preview1 shim. One file runs on every OS in Node and in the browser. narrac is the first such package: 6.23 MB stripped, 0.99 MB brotli (J2), check and emit equal to native (J3). Per-platform native packages in the esbuild shape stay an opt-in speed path, added only when a package measures the 7 times gap (P1 conceded in P13 after O1).
2. **Runtime fixes on wasi** (P2, O4). Two of the three T1 patches are bugs for every wasi program, not only narrac: `Env.args` reads `/proc/self/cmdline`, and output through `/dev/stdout` fails. They join wasi-runtime. The process block is already in its brief.
3. **The shim** (P5, J14). A JS preview1 shim for the 18 imports narrac uses (`evidence/T1-narrac-wasm.md`), over an in-memory file map. It also runs under Node, so the package does not need `node:wasi`, which is experimental (J13). Browsers need no wasm exception handling for narrac. Only programs that reach the coroutine panic guard do (J14).
4. **REPL v1, no server** (P3). The page runs narrac.wasm in a Worker: diagnostics, `fmt` and the emitted C as the reader types. A fresh instance per edit costs 93 to 325 ms in V8 (J15).
5. **REPL Run, in the page** (P11, A1, subject to Q1). On the first Run press the page loads clang built for wasm (about 19.8 MB brotli, J5), compiles the emitted C at `-O0` to wasm32-wasi, and runs it with the same shim. Warm Run: about 0.5 s for a guide-sized sample, 2.9 s for `regex.nra` (J18). No server and no untrusted code on our side.
6. **First users** (P6, O9). The REPL page, and the `narrac` npm package with a Node test. The program template is proven on one sample that is not narrac.
7. **Outlined, not scheduled** (P9, O6, O7). An esbuild-style request protocol (`std/serve` plus a JS client) waits until a package needs calls beyond arguments and files.
8. **C library chain parked** (P7, O5, P10, subject to Q2). c-abi-design, c-layout, c-export-functions and c-abi-export move to priority later. They come back only for the oxc form (in-process calls through Node-API or wasm exports).

## Budget

- device: the REPL reader downloads 0.99 MB brotli for check, and about 19.8 MB more on the first Run. The Run Worker peaks at 194 to 277 MB (J2, J5, J6). The build machine needs zig for the wasm row, already required by target-table (J4, J8).
- time: four scheduled children (narrac-wasm M, npm-narrac M, npm-template S, repl-check M) plus repl-run M after Q1, and two items added to wasi-runtime. The REPL waits for website, which waits for language-guide, textmate-grammar and install-guide (website brief).
- tokens: six M or S nodes, against ten (including the four C chain nodes and a Node-API node) for the oxc form (P10).

## Rejected

- Native per-platform packages as the first form (P1): every platform is a package and a publish step, for a speed gain no user asked for (O1, P13).
- Precomputed output instead of Run (A2 second half): website-repl's goal is that the reader types and runs.
- A precompiled runtime header to speed Run (A3): the runtime is about 0.4 s of a warm Run. The rest is the program's own C (J18).
- An interpreter for Run: the D105 evaluator runs an alloc-only subset with no `print`, closures or generators (J9). A full interpreter is a second semantics of the whole language.
- A wasm backend in the compiler: D10 picked C. A second backend is far larger than one Worker over clang.
- A Run server as the first form (P4, P12): a sandbox host, rate limits and an uptime duty for a could page (O2). It stays the answer if Q1 picks it.

## Proofs

- T1 for J2: does the seed build for wasm32-wasi and run `check` under wasmtime? Done: yes after three patches. 6.23 MB stripped, 0.99 MB brotli. 0.07 s cached per check (cap 30 min, about 45k tokens). `evidence/T1-narrac-wasm.md`.
- T2 for J5: does a C compiler in wasm compile emitted Narra C? Done: yes, @yowasp/clang 22, 72 MB module, 0.5 to 2.9 s warm, 6 to 17 s cold (cap 30 min, about 25k tokens). `evidence/T2-clang-wasm.md`.

## Prior art

- esbuild: native binary per platform through `optionalDependencies`, plus esbuild-wasm (14.5 MB unpacked). Its JS API talks to the binary over stdin and stdout (J10, J17).
- oxc and swc: Node-API addons through napi-rs, plus wasm builds of 1.5 MB and 18.4 MB (J10).
- Go and Rust playgrounds and Compiler Explorer compile and run on servers in sandboxes (gVisor, Docker). No major compiler playground compiles C in the page (J11).

## Open risks

- Firefox, Safari and phones are unmeasured for check and Run (O3, O8): measured in repl-check and repl-run, with a stated limit for each.
- A Narra program that uses threads, sockets or processes cannot target Wasi-Wasm32 (O9): accepted. Wasi-runtime makes it a compile error, and native packages are the path for such programs.
- A library called many times from JS pays a fresh instance per call (O6): accepted until a package needs it. The outlined protocol keeps one instance alive.

## Inferred choices

- The only users are the two the maintainer named: JS and TS packages written in Narra, and the website REPL (J12).
- "Narra for JS/TS packages" means any Narra program or library, with narrac first (J17).

## Answers from the user

- Q1, the Run button: (a) clang in the page, loaded on Run (recommended), or (b) a server we host. Not answered yet.
- Q2, a Narra library's JS API: (a) the esbuild form, child process or wasm over stdin and stdout, with the C chain parked (recommended), or (b) the oxc form with in-process calls, which unparks the four C chain nodes and adds a Node-API node. Not answered yet.

## Overruled

- none

## Cost of this debate

- 2 of 2 rounds, about 245k tokens.
