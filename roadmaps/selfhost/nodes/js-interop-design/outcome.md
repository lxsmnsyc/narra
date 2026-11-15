# Outcome: js-interop-design

## Built

- No compiler code. Decision D123 is in `DECISIONS.md`, marked "Not built yet". It was revised after the maintainer's answers of 2026-11-15.
- Goal debate in `roadmaps/selfhost/debate/js-interop/`: `ledger.json` (two sealed rounds, with the answers and the user rulings), `debate.md`, and `evidence/T1-narrac-wasm.md`, `T2-clang-wasm.md` and `T3-node-addon.md`.
- Answer, after the second answers: Narra sits on the layer of Go and Rust. The compiler builds programs, C libraries, Node addons (`--emit node-addon`) and wasm modules with JS glue it writes itself (`--emit wasm-js`). npm packaging is a script of the package that ships, not a compiler feature. narrac is the first npm package, as a command. The REPL is parked.

## Result

Spike with three proofs, each capped at 30 minutes.

- T1: `seed/narrac.c` builds with `zig cc -target wasm32-wasi -O2 -mexception-handling -mllvm -wasm-enable-sjlj ... -lsetjmp -lm` after three patches. The module is 6.23 MB stripped and 0.99 MB brotli. `check` and `emit` match native. A fresh instance per call costs 93 to 325 ms in V8.
- T2: clang 22 in wasm compiles emitted C. Warm runs take 0.47 s on hello and 2.9 s on regex, and the download is about 19.8 MB brotli. This is kept for a later REPL.
- T3, run after the answers: emitted C wrapped in about 30 lines of hand-written Node-API glue builds into a 51 KB `.node`. `require` loads it, and a call costs 0.18 µs. No init call was needed.
- reproduce: base 9f4017d7. The commands are in the evidence files. The scratch folders under `/private/tmp/claude-501/js-interop-design/` are throwaway.

## Proposed child briefs

Revised after the maintainer's second answers (2026-11-15): narrac writes the wasm glue itself, and npm packaging stays outside the compiler. The `[npm]` table, `narrac package`, the shared `@narra/loader` package, node-addon-wasm through emnapi, and npm-addon-sample are gone. The orchestrator writes these as node folders. narrac-wasm and npm-narrac ship the compiler. The C chain (3 to 5) starts in parallel, since c-abi-design is a gate spike.

### Amendment to wasi-runtime (todo, batch-16)

Add two work items:

- `Env.args` on wasi reads `args_sizes_get` and `args_get`. Today it reads `/proc/self/cmdline` everywhere but macOS, so every wasi program sees no arguments.
- `Fs.writeText` to `/dev/stdout` and `/dev/stderr` writes descriptors 1 and 2 on wasi, because WASI has no `/dev`. narrac prints all its output this way.

Acceptance gains one line: a wasi build of a sample that prints its arguments, and writes through `/dev/stdout`, matches its golden under wasmtime.

### 1. narrac-wasm: narrac built for Wasi-Wasm32 (first)

- goal: `sh scripts/wasm-build.sh` builds the compiler from the seed for `Wasi-Wasm32`, strips it with `llvm-strip`, and writes `dist/narrac.wasm`. cross-targets runs `check` and `emit` with it on its sample set and compares the result with native.
- size: S
- depends_on: wasi-runtime (hard: the runtime fixes); target-table (contract, via `targets`)
- acceptance:
  - Under wasmtime, `emit` gives C byte-identical to native for `testdata/hello.nra`, `regex.nra` and `selfhost/`.
  - The stripped module stays at most 1.2 MB brotli, and the script prints the size.
  - The module validates with wasm exceptions turned off.
  - The script is not part of `scripts/test.sh`, since it needs zig.

### 2. npm-narrac: narrac on npm, command form

- goal: `npm/` in this repository holds narrac's npm packages in the esbuild layout (D123 item 6): `narrac-<os>-<arch>[-<libc>]` per row cross-targets builds, `narrac-wasm`, and a main `narrac` package with a `bin` entry, a small loader that picks the platform package or falls back to wasm, and the WASI preview1 calls narrac imports (no `node:wasi`). `scripts/npm-package.sh` fills them from the cross builds and `dist/narrac.wasm`. A JS API (`check(files)`, `format(source)`, `emit(files)`, diagnostics as data) with TypeScript types runs the command.
- size: M
- depends_on: narrac-wasm (hard); cross-targets (behavior, via `targets`)
- acceptance:
  - The tarballs from `npm pack` install into a scratch project. `npx narrac check` and each API function match native narrac, once with the host package and once with only the wasm package.
  - Nothing in `selfhost/src` or `narra.toml` changes.
  - No publish: the maintainer publishes and confirms the package names.

### 3. c-abi-design (held, comes back): brief amendments

The brief still fits. Replace the `depends_on` reason with "D123 builds C libraries, Node addons and wasm glue on this boundary (maintainer, 2026-11-15)" and add these questions:

- **7. Node-API.** Which types the generated glue converts in the first build (numbers, `bool`, `string`, `bytes`) and how. What a panic and an `Err` become: a thrown JS `Error` with the message, with the panic stopped at the boundary by the runtime's `setjmp` guard, not ending the process. Whether a later build adds arrays, `Option` (as `undefined`) and plain structs as JS objects.
- **8. Node threads.** An addon may load in several `worker_threads`. Say whether runtime state is per Node environment (Node-API instance data) or per thread (D111 heaps), and what an export called on a worker does.
- **9. wasm reactor.** The exports built for `Wasi-Wasm32` as a reactor module with no `main`, and the memory calls the JS glue of wasm-js-glue uses to pass strings and bytes.

Change question 5 to name the `.node` output (`--emit node-addon`) next to the shared and static libraries.

### 4. c-export-functions (held, comes back): brief amendments

The brief fits. Add to acceptance: "The export wrapper is the function the glue of node-addon and wasm-js-glue calls. A panic inside an export called through it is stopped at the boundary as D122 says, tested in the exits suite." Remove the context line saying c-layout lands with it. c-layout is no longer on this path, so the signature check covers only the types D122 allows without layouts.

### 5. c-abi-export (held, comes back): brief amendments

The brief fits. Two changes:

- Change "A wasm reactor module, if D122 keeps it" to a required output: `--emit shared-lib --target Wasi-Wasm32` writes a reactor module whose only exports are the `@export(C)` functions plus the allocator calls D122 names. Add an acceptance line that runs one export under wasmtime.
- Add `depends_on: wasi-runtime (behavior, via targets)` for the reactor row.

### c-layout (held): stays parked

Values cross Node-API as `napi_value` and the wasm glue as scalars, strings and bytes, so C struct layouts are on neither path. It keeps its brief at priority later, for C users. Nothing depends on it.

### 6. node-addon: Node-API glue and `.node` output

- goal: `narrac build --emit node-addon` writes `<name>.node` for the host or a `--target` row, as D123 item 3 says, with one generated Node-API wrapper per `@export(C)` function, `napi_register_module_v1`, and `<name>.d.ts`. Every export is synchronous.
- size: M
- depends_on: c-abi-export (hard: shared library output); c-export-functions (hard: exports and the boundary wrapper); c-abi-design (hard: D122 conversions, panics, threads)
- acceptance:
  - A sample library with `string`, `bytes`, number and `bool` exports loads with `require` in a test. The suite skips it with a message when `node` is missing. Each call returns the expected value, and a panic and an `Err` arrive as thrown `Error`s.
  - The `.node` file exports only `napi_register_module_v1` (`nm`).
  - The `.d.ts` file passes `tsc --noEmit` if TypeScript is on the machine. Otherwise the outcome says so.
  - The leak check is clean after a million calls in one process.

### 7. wasm-js-glue: JS wrappers for a wasm library, written by narrac

- goal: `narrac build --emit wasm-js --target Wasi-Wasm32` writes the reactor module and `<name>.js` plus `<name>.d.ts`, as D123 item 4 says. The JS module instantiates the reactor, holds the WASI preview1 calls the runtime imports, copies strings and bytes through the D122 allocator calls, and throws a JS `Error` for a panic or an `Err`. No outside npm package.
- size: M
- depends_on: c-abi-export (hard: reactor output); c-abi-design (hard: D122 conversions); wasi-runtime (hard); node-addon (behavior: same sample and same `.d.ts`)
- acceptance:
  - node-addon's sample gives the same results through `<name>.js` in Node as through the `.node` build, and in a headless browser if one is on the machine. Otherwise the outcome says so.
  - The generated JS imports nothing outside itself, and `<name>.d.ts` matches node-addon's.
  - A panic inside an export leaves the instance usable for the next call.

### Outlines (not scheduled)

- std-serve and js-client (M each): esbuild's request protocol, which keeps one command process or wasm instance alive across calls. Starts only when a command package needs it.
- node-addon-async (M): exports that return a `Promise`, after a design over D109's loop and Node's.

### website-repl

Parked: the orchestrator sets priority later and proposes no children (maintainer, 2026-11-15). D123 item 10 and the debate keep the measurements for check in the page and for Run with clang in wasm or on a server.

## C library chain

- c-abi-design, c-export-functions and c-abi-export come back as the base for C libraries, Node addons and the wasm glue, with the brief changes above. D122 and E1620-E1639 stay reserved for them.
- c-layout stays parked with E1610-E1619 reserved. Nothing on these paths depends on it.

## Maintainer answers (second round, 2026-11-15)

1. emnapi: no. narrac writes the JS glue for wasm exports (node wasm-js-glue). The maintainer also asked to keep npm out of the language's center: Narra sits on the layer of Go and Rust, so the `[npm]` table and `narrac package` were dropped from D123.
2. Synchronous exports first: left to the orchestrator. Yes: a `Promise` export needs a design over two event loops, and every proof so far is synchronous.

The gate is passed: children may start.

## Deviations

- The debate ran two rounds, not the default three, because the proofs settled the large feasibility claims in round 1.
- Three proofs, one over the goal-debate skill's limit of two. T3 was run after the maintainer's answer to Q2, to check that emitted C runs inside Node before proposing the Node-API nodes. `debate.py check` warns about it.
- The debate's first settled structure (wasm only, chain parked, REPL children) was overruled by the answers. `debate.md` and the ledger record both.
- `@yowasp/clang` was downloaded from npm into the T2 scratch folder (109 MB). It is not installed anywhere else.
- The commit subject asked for is 75 characters. The plain-writing hook blocks subjects over 72, so the first commit used a shorter subject.

## Stubs left

- None in the repo. The T1 patches and the T3 glue exist only in the scratch folders.

## Gotchas

- `wasm-opt --strip-debug --all-features` writes a module that wasmtime 50 refuses ("compact imports proposal disabled"). Use `llvm-strip`.
- The seed builds for wasm in 68 s at low load and 190 s at load 10 to 15, so the wasm build does not belong in `scripts/test.sh`.
- Emitted function names are hashed statics (`f_436f9208_greet`). The T3 glue had to `#include` the C file. c-export-functions gives them real symbols.
- The emitted file has its own `main`. A library build must leave it out (c-abi-export).

## Proposed learnings

- Runtime code must not read `/proc` or `/dev` paths without a wasi branch | T1 found that `Env.args` and narrac's output both fail on wasi without an error message | selfhost/runtime | yes (cross-targets' wasi sample with arguments)

## Verification

- `python3 <goal-debate>/scripts/debate.py roadmaps/selfhost/debate/js-interop check`: ok, with the three-proof warning.
- No compiler code changed, so `sh scripts/test.sh` was not run.

## Size

- Estimated M, actual M.
