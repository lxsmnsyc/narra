# Outcome: std-source

## Answer

Yes. Std modules can ship as Narra source bundled with the compiler. `std/option` now is `std/src/option.nra` (44 lines, seven generic functions with `<effect E>`), loaded by both compilers. Both emit byte-identical C for every test program, and the bootstrap fixed point holds. The port needed no new language feature and no primitive form: `std/option` uses only the prelude `Option`/`Result` and `match`. Modules that need C (strings, files, maps) need a primitive form. This spike names that form but does not build it: builtin modules under `std/internal/`, importable only from `std/...` modules.

Recommended Go option: (b), with one small test addition (a stage 2 run test over `std/src` users). The cost of (a) is low for this loader but grows with every checker feature m10 adds.

## Built

- `std/src/option.nra`: `map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr` as generic Narra functions. Effects: `map` and `okOr` declare `uses alloc` (plus `E` for `map`), the others declare `uses E` or nothing.
- Embedding: `go:embed` cannot reach `../../../std`, so `TestStdSources` (`compiler/internal/narra/std_sources_test.go`) generates `compiler/internal/narra/std_sources.go` (`stdSources` map) and `selfhost/src/stdtext.nra` (`stdSourceFor(path)`) from `std/src/*.nra`, and fails when either is stale (`NARRA_UPDATE=1` rewrites both). Builds stay deterministic: the text is compiled into each binary, and nothing is looked up at run time.
- Go: `Module.Source` and `registerSourceModule(ns, path)` (`builtins.go`). The module stays in `stdModules` with no `Funcs`, so `moduleList`, the builtin dump, the alias checks and the `Option.map needs an import` hint keep working. `stdSourceModules` (`modules.go`) parses the std source modules a package imports, transitively, in first-import order, and puts them before the package's modules. `Check(f)` does the same. The file shown in diagnostics is `std/src/option.nra`.
- Selfhost: `loadStdModules`/`addStdImports` (`modules.nra`) do the same. Node ids come after the package's files, so `narrac pkg --ids` still sees one gap-free sequence. A `StdFunc` row with an empty key (`builtins.nra`, `stdlib.nra`) registers a source module in its place in the module order.
- Removed: the seven Go Check/Emit functions (`std_option.go` -183 lines, the helpers shared with `std/result` stay), seven `STD_FUNCS` rows, the Option dispatch and checks in `stdcalls.nra` (-100), and the Option emitters in `emitoptions.nra` (-115).
- Single-file packages now hold the std modules too. The user's module is the last one (`Package.ownModule`, and `modules[length - 1]` in `emitter.nra`).
- New sample `compiler/testdata/std_source_effects.nra`: a function with no `uses` calls `orElse` with a pure callback, `isSome` and `withDefault` and stays pure. A capturing `andThen` callback that allocates makes the caller `uses alloc`.

## Measurements

Release builds on the same machine. Base is origin/main 48c5d9d. The machine was shared with other agents' test runs (load average 14 to 19), so the times are noisy. Rerun them on a quiet machine before relying on them.

| | base (builtins) | branch (source) |
|---|---|---|
| stage 2 `emit selfhost` (`-O2`), first run + 3 warm, two rounds | 0.48 0.52 0.46 0.94 / 0.71 0.84 0.94 0.91 s | 1.53 2.66 0.59 0.76 / 0.56 0.38 0.47 0.75 s |
| best warm run | 0.46 s | 0.38 s |
| `go test -run TestSelfhostBootstrap` | 81 s, 183 s | 213 s (overlapped a base run), 154 s, 86 s |
| stage 2 binary, `-O2` | 2,071,408 bytes | 2,043,368 bytes (-28 KB) |
| stage 1 C of `selfhost` | 103,940 lines, 4,643,363 bytes | 102,355 lines, 4,565,213 bytes |
| C of `testdata/std_option.nra` | 1,325 lines | 1,393 lines (+68, +5%) |
| binary of `std_option.nra`, `-O2` | 51,656 bytes | 51,784 bytes (+128) |

- The bootstrap test times range over 2x for the same code (81 to 183 s for base). On a quiet machine the bootstrap outcome measured 21 s. Neither number shows a cost of the port.
- Under this load, the emit times show no difference beyond the noise. The compiler does less work per Option call, but checks seven generic bodies when a program imports std/option. selfhost itself does not import it.
- Stage 2 still reaches the fixed point: stage 2 `emit selfhost` equals stage 1's output byte for byte.

- Allocation: a capturing arrow passed directly to `Options.map` still gets a stack environment (`K_..._run_2 k32 = {.h = {-1, NULL}, ...}`), as with the builtin. D69 holds, with no heap environment.
- The C is longer because each instance (`map<int,int>`, `map<int,string>`, ...) is a C function instead of inline code at the call site. The C compiler can inline these at `-O2`.
- Lines, hand-written code only (not the generated text files): Go -186 +75 in the compiler, +110 test. Selfhost -226 +72. Std source +44, shared by both compilers.

## Primitive table

73 rows of `STD_FUNCS` on main. "Source" means the function can be written in Narra over the primitives in the left column.

| module | primitive (C or intrinsic) | source |
|---|---|---|
| prelude | print, assert, panic | |
| std/int | toString | abs, parse, toHex |
| std/float | fromInt, sqrt, toString, truncate | |
| std/string | length, byteAt, slice, concat, charAt, nextChar, fromChar, builder, append, appendChar, build | sliceFrom, startsWith, endsWith, contains, indexOf, split, join, trim, repeat, replace, charCount, chars |
| std/char | code, fromCode | |
| std/array | length, push, filled, values, entries (for-of lowering, element memory) | |
| std/map | entries, keys, remove, size (runtime hash table, `m[k]` syntax) | |
| std/box | new | |
| std/range | between, upTo (for-of lowering) | |
| std/env, std/fs, std/process | all 7 (syscalls) | |
| std/option | | all 7 (done) |
| std/result | | map, mapErr, andThen, withDefault, ok, err |
| std/test | eq, ne (call position in the message), panics, panicsWith (fork), golden (files) | |

Total: 45 primitive, 28 source-able. The string functions are source-able but would loop over `byteAt` where the C uses `memcmp`/`memmem`. Port them only with a benchmark (bench-harness).

## Primitive form

Not needed for `std/option`. For modules that do need it, the proposed form is **internal builtin modules**: the primitives stay `registerModule` entries (Go) and `STD_FUNCS` rows (selfhost) under paths `std/internal/<name>` (for example `std/internal/bytes`). The rules:

- Only a module whose path starts with `std/` may import `std/internal/...`. Any other import gets an error, for example "std/internal/bytes is internal to the standard library".
- They are not listed in `moduleList` hints.
- They take the existing Builtin fields (Check, Emit, Effects, Runtime), so nothing new goes into the emitter. `TestSelfhostBuiltins` compares them like any builtin.

An `@builtin` declaration form was not chosen. It would need parser, AST, clone, checker and emitter cases in both compilers, while internal modules reuse the registry as it is.

## Error messages that changed

Generic instantiation errors replace the Option-specific checks. The two `testdata/errors/` tests now expect the new text, and `typesdump/stdcalls/err_options.dump` was regenerated:

| call | before | after |
|---|---|---|
| `Options.andThen(o, (x) => x + 1)` | Option.andThen needs a function that returns Option, found one that returns int | cannot tell type parameter U of andThen here; add a type annotation |
| `Options.isSome(3)` | Option.isSome needs an Option<T>, found int | cannot tell type parameter T of isSome here; add a type annotation (+ expected bool, found void) |
| `Options.map(1, f)` | Option.map needs an Option<T>, found int | cannot tell type parameter T of map here |
| `Options.map(some)` | Option.map takes 2 arguments, found 1 | this call passes 1 arguments, the function takes 2 (+ cannot tell U) |
| `Options.map(some, 5)` / `(v) => { return }` | needs a function as argument 2 / needs a function that returns a value | cannot tell type parameter U of map here |
| `Options.map(mutable some, f)` | (none: this was hidden by the first error) | parameter "option" is not mutable; remove `mutable` |
| `Options.withDefault(3, 3)` | Option.withDefault needs an Option<T>, found int | expected Option<int>, found int |
| `Options.okOr(3, 4)` | Option.okOr needs an Option<T>, found int | cannot tell type parameter T of okOr here |
| effect error naming the callee | uses `alloc` (through `half`) | uses `alloc` (through `andThen`) |

This is the main cost of the port: when an argument does not fit a generic parameter (`int` against `Option<T>`), generic inference reports "cannot tell type parameter" instead of the mismatch. A later node should make `genericCall` report "expected Option<T>, found int" when unification fails on a concrete argument. That fix helps every source std module and every user generic, so it should come before more std modules move to source. Following the brief, no special cases were added. The coverage test (`selfhost_stdcalls_test.go`) now asks for "needs a function that returns Result" instead of the two Option strings, and `err_options.nra` gained a `Results.map` call with a void callback, so "needs a function that returns a value" is still covered.

## Go options

- **(a) Go loads std source too (what this branch does).** Cost so far: about 75 lines in Go next to about 72 in Narra. Each later loader or registry change is made twice. A std source module also needs every language feature it uses to exist in Go. For m10 that is the blocker: std-iterator needs generators, std-json needs `derives Shape`, and both are planned for the self-hosted compiler only. Under (a), each would have to be built in Go first, which the selfhost-spike gate ruled out ("port soon"). So (a) works for modules written in today's language (option, result, the string helpers, math) and not for the m10 modules that need new features.
- **(b) Go keeps only what `selfhost/src` uses; new std is selfhost-only.** Default, and recommended. What breaks:
  - `TestPrograms`, `TestPanics`, `TestCompileErrors`, the runner tests and the typesdump/declsdump goldens run the Go compiler. A sample that imports a selfhost-only std module fails under Go.
  - `TestSelfhostEmit` compares Go and selfhost C on every `testdata/*.nra`, so it would fail on those samples.
  - Instead: put such samples under a new folder (for example `compiler/testdata/stage2/`) that the Go tests skip. Add a `TestStage2Programs` that reuses the cached stage 2 binary from `TestSelfhostBootstrap` (built once per run, about 21 s, already paid), emits each sample with it, compiles with the sanitizer flags and compares `.out` and the leak check. Extra time per sample: one stage 2 emit (well under 1 s for a small file under sanitizers) plus one `cc` run (about 1 s), about the same as `TestPrograms` per file. Error tests need the same split (`testdata/stage2/errors/`).
  - `std/option` itself can stay in both compilers as long as Go keeps the loader built here. The loader has no feature coupling.
- **(c) Estimate only: a checked-in C seed replaces the Go compiler for bootstrapping.** The seed is stage 1's C for `selfhost`: today 102,355 lines and 4.57 MB (736 KB gzip -9). The update procedure: build stage 2 from the current seed, emit `selfhost` with it, check the fixed point (stage 2 output equals stage 3 output), commit the new seed in the same PR as the language change. A change that uses a new feature in `selfhost/src` needs two steps: first teach the compiler the feature (seed N builds it), then use it (seed N+1). Risks:
  - Nobody can review a 4.5 MB generated diff. Trust rests on the fixed-point check in CI, not on review, and a bug that reproduces itself (a trusting-trust miscompile) is caught only by tests.
  - Platform C: the seed must build on clang and gcc on macOS and Linux. Today's C is C11 with `-w`. `int64_t`/`size_t` differences and `unistd.h` keep it POSIX-only.
  - Git history grows by about 0.7 MB compressed per refresh. Refresh only at milestones, or keep the seed in a release asset.
  - Size: M as a node (seed file, `bootstrap.sh` from the seed, a CI fixed-point check, removal of Go from the bootstrap path). It removes the double cost of (a) and the test split of (b) for good, so it is worth doing once m10 needs features that Go does not have.

## Plan for the m10 nodes

- **std-option-result / std/result**: same pattern as this port. Six functions, no primitives. Do it under (a) now, or with the first selfhost-only module. Fix the generic mismatch message first.
- **std-json**: `Json.Value` enum, the encoder and the parser as source. Primitives: none beyond today's `std/string` builder, `byteAt`, `Int.parse` and `Float.toString`. A float parser needs `std/internal/float` (`strtod`). `derives Shape` is compiler work, not std. Selfhost-only under (b).
- **std-basics**: `std/math` min, max, clamp and integer pow as source. sqrt, the float math.h functions, `std/time`, the clock and sleep as `std/internal/` primitives (C). `std/random`: the generator (xorshift or PCG) as source, the system seed as a primitive. `std/hash`: `Hash.of` stays an intrinsic (trait dispatch). `Ranges.stepped` is a primitive next to `between`/`upTo` (for-of lowering).
- **std-iterator**: everything as source over `Generator<T>` once generators exist. No primitives. Selfhost-only by construction.
- **std-arena**: `Arena<T>` and `Handle<T>` as source over `Array<Option<T>>` plus generations. No primitives, unless benchmarks ask for an unchecked index.
- **std-unicode**: the tables as generated top-level `const` arrays in source, lookups as source. No primitives. Watch the size of the generated source and the check time.
- **std-spec-functions**: `Int.divmod`, `Int.floorDiv`, `Int.compare`, `Arrays.map`, `Arrays.slice`, `Arrays.copy`, `Str.find` and `Str.copy` as source. `Arrays.withCapacity` is a primitive (`std/array`). `Float.PI` needs exported top-level constants from a std source module, which work today through source modules (constants in `newSourceModule.Exports`). That was not tested in this spike.

## Deviations

- Files outside owned and shared paths: `compiler/internal/narra/std_sources.go` (generated), `compiler/internal/narra/std_sources_test.go` (generator and check), `selfhost/src/stdtext.nra` (generated), `compiler/internal/narra/diagnostic_json.go` and `selfhost/src/emitter.nra` (one line each: the single-file module is now the last one), `compiler/internal/narra/selfhost_stdcalls_test.go` (coverage strings), `compiler/testdata/errors/std_option_andthen_type.nra`, `compiler/testdata/errors/std_option_not_option.nra` (new message text), and `compiler/testdata/typesdump/{stdcalls/options,stdcalls/err_options,effects/basics}.dump` plus `stdcalls/err_options.nra` (regenerated, one added call).
- `selfhost/src/emitexprs.nra` still routes `std/option` builtin calls to `emitOptionCall`. That code can no longer be reached and was left alone because the file is not owned. Remove it with the next change there.
- `std/` has no `narra.toml`: the name `std` is reserved by the manifest parser, and the folder is not built as a package.

## Stubs left

- None in the port. The `std/internal/` primitive form is a proposal, not code.

## Gotchas

- A std source module is checked as a package module. Its generic bodies show up in the types dump (`body module="std/option" ...`) of every program that imports it, so dumps grow.
- `x is Some(v)` outside an if condition is an error: write `Some(_)`.
- Std modules go before the package's modules in `Package.Modules`. Code that took `Modules[0]` as the user's file must use `ownModule()`.
- Std source must be valid in both compilers. A parse error panics at load time ("std source std/option: ...").

## Proposed learnings

- Write a std module as Narra source in `std/src/` when it needs no C; register it with `registerSourceModule` and a key-less `STD_FUNCS` row, then run `NARRA_UPDATE=1 go test -run TestStdSources` | both compilers embed generated copies | compiler/internal/narra, selfhost/src | yes (TestStdSources)

## Verification

- `cd compiler && go vet ./... && go test -timeout 60m ./...`: pass (724 s under load). The default 10 min timeout ran out once under the same load; the run passed with a longer timeout. TestSelfhostBootstrap, TestSelfhostEmit, TestSelfhostBuiltins and TestStdSources pass.
- Stage 1 (built from the branch) and Go emit byte-identical C for `std_option.nra`, `options.nra` and `std_source_effects.nra`. Program output equals the `.out` files.

## Size

- Estimated M, actual M.
