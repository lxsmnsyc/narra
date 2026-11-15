# Outcome: js-interop-design

## Built

- No compiler code. Decision D123 in `DECISIONS.md`, marked "Not built yet".
- Goal debate in `roadmaps/selfhost/debate/js-interop/`: `ledger.json` (sealed after two rounds), `debate.md` (status proposed), `evidence/T1-narrac-wasm.md`, `evidence/T2-clang-wasm.md`. The debate stopped before the approval step. Q1 and Q2 are for the maintainer.
- Answer: wasm first. A Narra program, narrac first, ships to npm as one `Wasi-Wasm32` module with a JS loader and a WASI preview1 shim. The REPL checks in the page with the same module. It runs programs with clang built for wasm, loaded on the first Run press (Q1). The C library chain is parked (Q2).

## Result

Spike with two proofs, each capped at 30 minutes.

- T1: `seed/narrac.c` builds with `zig cc -target wasm32-wasi -O2 -mexception-handling -mllvm -wasm-enable-sjlj ... -lsetjmp -lm` after three patches. The module is 6.23 MB stripped and 0.99 MB brotli. `check` and `emit` match native. `check regex.nra` takes 0.07 s in wasmtime (cached), 219 ms cold in Node, and 93 to 325 ms per fresh instance in V8.
- T2: `@yowasp/clang` 22 is 72 MB plus a 28 MB sysroot, about 19.8 MB brotli. It compiles emitted C, and the output matches `testdata/regex.out`. Warm `-O0` compile and link: hello 0.47 s, regex 2.9 s. Cold: 6.2 s and 17.2 s.
- reproduce: base 9f4017d7. The commands are in the two evidence files. The scratch folders `/private/tmp/claude-501/js-interop-design/t1` and `t2` are throwaway.

## Proposed child briefs

The orchestrator writes these as node folders after the gate. Sizes follow the roadmap's S, M and L.

### Amendment to wasi-runtime (todo, batch-16)

Add two work items:

- `Env.args` on wasi reads `args_sizes_get` and `args_get`. Today it reads `/proc/self/cmdline` everywhere but macOS, so every wasi program sees no arguments.
- `Fs.writeText` to `/dev/stdout` and `/dev/stderr` writes descriptors 1 and 2 on wasi, because WASI has no `/dev`. narrac prints all its output this way.

Acceptance gains one line: a wasi build of a sample that prints its arguments, and writes through `/dev/stdout`, matches its golden under wasmtime.

### narrac-wasm: narrac built for Wasi-Wasm32

- goal: `sh scripts/wasm-build.sh` builds the compiler from the seed for `Wasi-Wasm32`, strips it with `llvm-strip` or zig's strip, and writes `dist/narrac.wasm`. cross-targets runs `check` and `emit` with it on its sample set and compares the result with native.
- size: S
- depends_on: wasi-runtime (hard, the runtime fixes), target-table (contract `targets`, the row and its flags)
- acceptance: under wasmtime, `emit` gives C byte-identical to native for `testdata/hello.nra`, `regex.nra` and `selfhost/`. The stripped module stays at most 1.2 MB brotli, and the script prints the size. The module validates with wasm exceptions turned off. The script is not part of `scripts/test.sh`, since it needs zig.

### npm-narrac: the narrac npm package

- goal: a package folder (`npm/narrac/`) holding `narrac.wasm`, a JS loader, a WASI preview1 shim over an in-memory file map (the 18 imports listed in T1), TypeScript types, and an API with `check(files)`, `format(source)` and `emit(files)`. The API returns diagnostics as data (the JSON diagnostics form) and makes a fresh instance per call. A `bin` entry runs the CLI under Node without `node:wasi`.
- size: M
- depends_on: narrac-wasm (hard)
- acceptance: a Node test calls each function on samples and matches native narrac. The shim passes the same test in a headless browser if one is on the machine, or the outcome says it was not run. `npm pack` gives one tarball under 1.5 MB. No publish: the maintainer publishes and picks the package name.

### npm-template: any Narra program as an npm package

- goal: a template folder (`npm/template/`) and a README section showing how a Narra program becomes an npm package. Build it with `--target Wasi-Wasm32` and reuse the npm-narrac loader and shim, with arguments, files and stdout as the interface. Prove it on one sample that is not narrac.
- size: S
- depends_on: npm-narrac (hard, the loader and shim)
- acceptance: the sample package installs from its `npm pack` tarball into a scratch project and prints its golden output through `npx`. A program that imports `std/process` or `std/thread` fails to build for the target with E0991.

### Outlines (not scheduled)

- npm-native (M): per-platform packages in the esbuild shape (`optionalDependencies` with `os`, `cpu`, `libc`), built by cross-targets. Starts only when a package measures the wasm speed gap.
- std-serve and js-client (M each): an esbuild-style request protocol over stdin and stdout, which keeps one instance alive across calls. Starts only when a package needs calls beyond arguments and files.

### Children of website-repl

website-repl stays a story of two children.

#### repl-check: the REPL page, check without a server

- goal: a website page with an editor. narrac.wasm runs in a Worker through the npm-narrac API. The page shows diagnostics at their positions, the `fmt` result and the emitted C as the reader types. Guide samples open in it with a link.
- size: M
- depends_on: website (hard), npm-narrac (hard)
- acceptance: the page loads the module once and checks on each edit, debounced. The time per edit is measured in Chrome, and in Firefox and Safari where they are on the machine, and recorded. No request goes to a server after the page loads. Keyboard use works, and there is no layout shift.

#### repl-run: the Run button

- goal, under Q1 (a): the first Run press loads clang built for wasm in a Worker (an npm dependency or vendored files, loaded only on that press). The Worker compiles the emitted C at `-O0` for wasm32-wasi with `-mexception-handling -mllvm -wasm-enable-sjlj -lsetjmp -lm`, runs the result with the shim, and streams stdout and the exit code to the page. A run has a time limit, and the page can stop it.
- goal, under Q1 (b): the same button sends the source to a server. The server runs narrac, `zig cc` to wasm and wasmtime with time and memory caps. This adds a node repl-run-server (L): the container image, rate limits, and a deploy that the maintainer owns.
- size: M under (a). M plus the L server node under (b)
- depends_on: repl-check (hard)
- acceptance: every guide sample that has a `.out` golden and no module left out on wasi runs in the page with matching output. The download size and the warm and cold times are recorded per browser on the machine. A program that loops forever is stopped by the limit.

## C library chain

Parked, not dropped. c-abi-design, c-layout, c-export-functions and c-abi-export go to priority `later` with the reason "D123: wasm and the esbuild form need no C library boundary". They come back if Q2 picks the oxc form, together with a new Node-API glue node (L, one native build per platform) and a wasm-exports node (M). E1610-E1639 and D122 stay reserved.

## Questions for the maintainer

1. The REPL Run button. Recommended: (a).
   - (a) The page loads a C compiler only when Run is pressed: about 20 MB once, about 0.5 s per run on a guide sample, no server.
   - (b) A server we host compiles and runs in a sandbox: about 3 s per run, and a host to pay for and keep safe.
2. A Narra library's JS API. Recommended: (a).
   - (a) The esbuild form: JS runs the Narra program (wasm, later native) and talks through arguments, files and stdout. The C chain stays parked.
   - (b) The oxc form: JS calls Narra functions in-process through Node-API or wasm exports. This unparks the four C nodes and adds two.

## Deviations

- The debate ran two rounds, not the default three, because the proofs settled the large feasibility claims in round 1.
- The goal-debate skill allows two proofs. The extra timings (per-edit instances, warm clang, exception handling turned off) were added as judge measurements in the same two scratch folders.
- `@yowasp/clang` was downloaded from npm into the T2 scratch folder (109 MB). It is not installed anywhere else.

## Stubs left

- None in the repo. The three T1 patches exist only in the scratch copy of the seed.

## Gotchas

- `wasm-opt --strip-debug --all-features` writes a module that wasmtime 50 refuses ("compact imports proposal disabled"). Use `llvm-strip`.
- The seed builds for wasm in 68 s at low load and 190 s at load 10 to 15, so the wasm build does not belong in `scripts/test.sh`.
- The unstripped module is 27 MB, because name and debug sections make up most of it.
- `regex.nra` emits 237 KB of C because it pulls in std/regex. Run time in the page grows with the emitted C, not with the sample's own lines.

## Proposed learnings

- Runtime code must not read `/proc` or `/dev` paths without a wasi branch | T1 found that `Env.args` and narrac's output both fail on wasi without an error message | selfhost/runtime | yes (cross-targets' wasi sample with arguments)

## Verification

- `python3 <goal-debate>/scripts/debate.py roadmaps/selfhost/debate/js-interop check`: ok. `status`: round 2 of 2, waiting on Q1 and Q2.
- No compiler code changed, so `sh scripts/test.sh` was not run.

## Size

- Estimated M, actual M.
