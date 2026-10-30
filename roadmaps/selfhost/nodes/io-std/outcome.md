# Outcome: io-std

## Built

- `std/fs`: `readText(path): Result<string, string>` (Err on I/O failure or invalid UTF-8), `writeText(path, text): Result<int, string>` (Ok = bytes written), `exists(path): bool`, `listDir(path): Result<Array<string>, string>` (sorted by bytes, no `.` or `..`; Err for the whole listing if any entry name is not valid UTF-8). In `compiler/internal/narra/std_fs.go`.
- `std/process`: `run(command: string, args: Array<string>): Result<#[int, string, string], string>`. Ok holds `#[exit code, stdout, stderr]`. The command is found through PATH, stdin is /dev/null, a signal death gives 128 + signal, and Err means it could not start (exec failure is reported through a close-on-exec pipe). Output is made valid UTF-8 with U+FFFD. In `std_process.go`.
- `std/env`: `get(name): Option<string>`, `args(): Array<string>` (same as the entry function's `args`, program name first; macOS `_NSGetArgv`, elsewhere `/proc/self/cmdline`, so no change to `emit.go`). In `std_env.go`.
- Runtime C in `compiler/internal/narra/runtime_io.h`, embedded by `std_fs.go` and split at `// @section` lines. Each module's `Module.Runtime` is the guarded `common` section plus its own, so importing several modules emits the common helpers once.
- Shared Go helpers in `std_fs.go`: `checkFixed`, `preludeEnum`, `resultOf`, `variantOf`, `ioResult` (builds a `Result` from a runtime success flag; reuse for other Result-returning std modules).
- Tests: `compiler/testdata/io_files.nra` and `.out` (write, read back, list, echo, exit code, stderr, missing command, env, args), `compiler/testdata/errors/io_unchecked.nra`, `io_wrong_argument.nra`. A README line under "Supported".

## Deviations

- `Result<void, string>` is not possible (the language has no unit value), so `writeText` returns `Result<int, string>` with the byte count.
- `compiler/testdata/errors/unknown_module.nra`: the expected text listed exactly three modules and broke when modules were added. It now expects the prefix `unknown module "std/nope" (this compiler supports std/int, std/float, std/string`. Other std nodes will hit the same line; this prefix needs no further edit.
- `compiler/README.md` got one added line (not in owned paths). The new error tests are under `compiler/testdata/errors/`, outside the `compiler/testdata/io*` glob.

## Stubs left

- None. Effects (`io`, `alloc`) are declared on the builtins; checking is node effects.

## Gotchas

- Builtin `Check` can get prelude generics through `c.generics["Result"]` and `c.instantiate`; the instance is registered for emission automatically. Call `c.arrayOf(...)` in `Check` so the array type is emitted.
- Tuple fields are `f_0`, `f_1`, ...; read them with destructuring `const #[a, b, c] = r` (`r.0` does not parse).
- Runtime helpers use plain `malloc` for scratch buffers, not `narra_alloc`, so the leak counter only sees Narra strings.
- Paths and commands containing a NUL byte return Err.

## Proposed learnings

- A std builtin that returns `Result` builds it through a runtime function with out-parameters plus `ioResult`; runtime C cannot name per-program types, since `Module.Runtime` is emitted before them | applies to compiler/internal/narra | no
- Error tests that list the std modules must match only a prefix | the module list grows with every std node | compiler/testdata/errors | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (programs run under AddressSanitizer, UBSan and the leak check in debug mode, and in release mode).

## Size

- Estimated M, actual M.
