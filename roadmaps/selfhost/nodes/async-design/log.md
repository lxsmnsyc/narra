# async-design log

- 2026-11-13: branch spike/async-design reset to feat/batch-12 (af35fdcb). D109 absent, SPEC section 8 still describes D17. Work starts.
- Read D17, D80, D85, D88, D90, D93, D94, D95, D104; emitgenerators.nra; runtime.h (narra_gen); runtimetext.nra; the generated suite (selfhost/tests/src/generated.nra).
- Finding: the generated suite maps only `std_NAME.c` (plus four fixed names) to `runtime<name>.nra`. The brief's `async.c`, `async_loop.c`, `async_io.c` would get no generated case. D109 names them `std_async.c`, `std_async_loop.c`, `std_async_io.c` (generated `runtimestdasync.nra`, `runtimestdasyncloop.nra`, `runtimestdasyncio.nra`), so the test runner needs no change.
- Finding: runtime text is picked per module path in `runtimeFor` and kept only when a noted helper reaches it (emit.nra around line 93). Async syntax needs the runtime without an import, so the checker marks `std/internal/async` used and the emitter notes it. No hook in runtimecore.nra.
- Finding: `Generator`, `Coroutine` and `Dynamic` are builtin type names in resolve.nra, not scope types, so E0421 (alias equals a type name) does not see them. `Async` follows them, which lets `import * as Async from "std/async"` stand beside the type `Async<T>`.
- Finding: the generator emitter suspends only at statement level (`yield` is a statement, E0046), and scratch values inside one expression stay C locals. So `await` is allowed only at the top of a statement in the first build.
- Finding: no `std/net` exists, std/fs has no handles, and SPEC section 6 says sockets close when the last owner drops them. Connections need a counted OS handle type; std-async-io grows to L.
- Finding: no wasm32-wasi or Windows target exists (Build.os has Windows, but the supported targets are macOS and Linux).
- Finding: `async` and `await` are not used as names in selfhost/src, selfhost/tests, std/src or testdata, so both can become keywords.
- Wrote D109 (eleven items plus the std/async functions), the diagnostics header line, the SPEC text and the four briefs.
- The `spec` suite checks every `ts` block; `selfhost/tests/spec-skip.txt` already skips "Async functions | import * as Async from \"std/async\"". The rewritten section keeps exactly one sample with that first line, so no skip-list change is needed now.
- SPEC: grep for `Async.run`, `Async.spawn`, `Task<`, `Loop`, `Async.sleep`, `Async.join`, `Async.map`, `Async.forEach` finds nothing.
- spec suite PASS 62 cases; sources suite PASS 1 case.
- Maintainer approved D109 (2026-11-13) with one change: `await` in any expression of an async body. D109 item 1 now describes spilling earlier operands into frame temporaries; E1101 left free; SPEC row and position rule removed; async-core brief updated.
- Rebased on feat/batch-12 (94d1fb5b), exit 0. spec PASS 62 cases, sources PASS 1 case.
