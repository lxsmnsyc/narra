# Outcome: c-abi-design

## Built

- No compiler code. Decision D122 is in `DECISIONS.md` before D123, marked "Not built yet". It answers questions 1 to 9 of the brief.
- Two capped proofs, with files in `/private/tmp/claude-501/c-abi-design/` (throwaway): `guard.c` and `cost.c` for the panic guard, `reactor.c` and `run.mjs` for the wasm reactor.

## Result

- Panic guard, real `runtime.h`, 1000 calls with every odd call panicking: the process stays alive, 500 messages come back, the next call is correct. 1000 allocations stay live, 2 per panic (the frames that were unwound are not released).
- Guard cost on macOS, 10 million calls: `setjmp` 315 ns, `_setjmp` 6.8 ns. The runtime must use `_setjmp` and `_longjmp`, or the guard costs more than the whole 0.18 µs addon call of T3.
- Wasm reactor with a guarded export, `narra_alloc` and `narra_free`: no imports, runs under Node 26 and under `wasmtime run -W exceptions=y`, and the instance works after a caught panic.
- Two build findings for wasm, both on zig 0.17.1: wasmtime 50 refuses the default legacy exception model (`-mllvm -wasm-use-legacy-eh=false` fixes it), and a program that uses `setjmp` needs the `__c_longjmp` tag defined in C or the link fails.

## Answers in short

1. Strings and bytes are copied in and out. Out buffers are `malloc` copies freed with `narra_<library>_free`. No other counted type crosses in the first build. `Opaque<T>` handles are designed and left to a later node.
2. `@repr(C)` for structs and `@repr(<integer>)` for enums. `@layout` goes out of SPEC. c-layout stays parked.
3. `@export(C, "name")` on a top-level function. Two entries: the plain `name` (a panic ends the process) and `narra_try_name` (status 0, 1 or 2, out parameters). A `Result` export has only the checked entry.
4. No init. One recursive lock per library output. The guard uses `_setjmp`. A caught panic leaks what its frames held.
5. Flags only: `--emit shared-lib`, `static-lib`, `node-addon`, `wasm-js`. File names in a table in D122.
6. Codes E1620-E1623 (export function), E1630-E1632 (library output), E1640 (node-addon, `Ptr` in an export).
7. 64-bit integers are `bigint`, other numbers `number`. `Err` and panic throw `Error` with `code`.
8. State is per process and serialized by the lock. No Node instance data.
9. Reactor exports and the allocate, write, call, free protocol are in D122 item 9.

## Proposed brief changes (the orchestrator applies them)

### c-export-functions

- Goal: replace "a wrapper that sets up runtime state on first call" with "two entries per export (plain and `narra_try_`), as D122 item 3 says". No runtime state is set up on a call: D122 item 4.
- Context: add the runtime work it owns: `NARRA_SETJMP` and `NARRA_LONGJMP` macros (`_setjmp` and `_longjmp`) for every build, the panic location saved with the message, the UTF-8 check on `string` parameters, the library lock behind `NARRA_LIBRARY`, `narra_<library>_alloc`, `_free` and `_shutdown`.
- Acceptance additions:
  - The checked entry returns 2 and the message for a panic, 1 and the message for an `Err`, and a bad UTF-8 argument gives status 1, in the exits suite.
  - A million calls of an export with `string` and `bytes` leave the live count at 0 under `NARRA_CHECK_LEAKS` after `shutdown`.
  - A bench line for the checked entry on macOS shows the guard under 15 ns.
- Codes: E1620 wrong place or arguments, E1621 type not allowed, E1622 type parameter, effect parameter, `async`, generator or `mutable` parameter, E1623 duplicate name or clash with an extern member. The name check reuses E1001.
- Remove "(for example `qsort` with an exported comparator)" only if no C harness exists by then. Otherwise keep it: an executable with `@export` functions takes no lock.

### c-abi-export

- Goal: `--emit shared-lib` and `static-lib` are flags only. Drop "(and a `[[library]]` table ...)". The name must be a C identifier (E1630). File names are in D122 item 5.
- The header: `narra_buffer`, status macros, one prototype per entry, the three support functions, `NARRA_API`, C++ guard.
- Acceptance additions:
  - `static-lib` with `--target Wasi-Wasm32` is E1631. A program with no export is E1632.
  - The reactor test needs the wasm flags of D122 item 5. wasmtime runs it with `-W exceptions=y`.
  - `nm` shows only the entries, the support functions and nothing else.
- Codes: E1630-E1632.

### node-addon

- Types are those of D122 item 7: `bigint` for `int`, `i64` and `u64`, `Uint8Array` for `bytes`, and E1640 for `Ptr<T>`. The `.d.ts` follows it.
- The glue calls `narra_try_name` as a `static` function in the same C file. It does not take the lock itself: the entry does.
- Acceptance additions: a test loads the `.node` in two `worker_threads` and calls an export in both. A panic gives `Error` with `code` `ERR_NARRA_PANIC`, and a later call works.

### wasm-js-glue

- The glue follows D122 item 9: `_initialize`, a 64-byte scratch area, `narra_<library>_alloc` and `_free`, and 64-bit integers as `bigint`.
- Acceptance additions: the module is built with `-wasm-use-legacy-eh=false` and runs in Node 26 and under wasmtime 50. The outcome names the browser versions that run it, or says no browser was tried.

### wasi-runtime and the contract `targets`

- `targetCompilerArguments` for `Wasi-Wasm32` needs `-mllvm -wasm-use-legacy-eh=false` besides the flags in the contract, or wasmtime 50 refuses any module that links `setjmp`. The contract names only `-mexception-handling -mllvm -wasm-enable-sjlj` and `-lsetjmp`. The orchestrator decides whether this is a change request on `targets`.
- The emitted C of a wasm build defines the `__c_longjmp` tag (see D122) while zig 0.17.1 does not link it.

### c-layout

- No change. It stays parked. D122 item 2 fixes the spelling (`@repr`) for when it is built.

### SPEC and docs

- SPEC section 9: delete the `@layout(C)` text, write the rules of D122 when c-export-functions lands. This is part of that node.

## Open questions

- None that change what gets built. Choices the maintainer may overrule when reading D122: `bigint` for 64-bit integers, the flat pointer and length shape, one lock per library instead of per-thread heaps, and a leak (not cleanup) after a caught panic.
