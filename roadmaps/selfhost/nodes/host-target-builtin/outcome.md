# Outcome: host-target-builtin

## Built

- `std/build` gains `Target { os: Os, arch: Arch }` and `hostTarget(): Option<Target>` (`std/src/build.nra`). It is `None` for a system or machine outside the four supported targets.
- New internal module `std/internal/host` with `Host.os()` and `Host.arch()` (ints; 0 unknown). C in `selfhost/runtime/std_host.c` reads `__APPLE__`, `__linux__`, `__aarch64__`/`__arm64__`, `__x86_64__`/`__amd64__`. Registry rows and `emitHostCall` live in `selfhost/src/stdbuild.nra`.
- Dispatch only in shared files: `emitstd.nra` (`Host.` prefix and runtime module), `runtimetext.nra` (import and case), `SPEC.md` (one bullet in the build conditions section). Generated: `runtimestdhost.nra`, `stdtext.nra`.
- Sample `testdata/build_host_target.nra` (+ `.out`) runs `uname -s -m` and checks `Build.hostTarget()` against it; prints "host target matches uname" on any host, so one golden covers macOS and Linux.
- The compiler still runs `uname` (`buildtarget.nra` untouched) until seed-switch.

## Deviations

- Public name `Build.hostTarget` returns `Option<Target>` rather than a bare string: the brief left the shape open.
- `sh scripts/linux-check.sh` was not run (the batch runs it at the end); the sample is host-independent by construction.

## Stubs left

- seed-switch replaces `detectHostTarget`'s `uname` call with `Build.hostTarget()` after the seed update.

## Gotchas

- An `import` in a std source module must come before any declaration (the module loader panics otherwise).
- The seed rule applies: `selfhost/src` does not use `Build.hostTarget` yet.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: PASS (254s) on the final tree.
- `narrac run testdata/build_host_target.nra` on macOS arm64: "host target matches uname".

## Size

- Estimated M, actual M.

## Review notes

- `selfhost/runtime/std_host.c` and `selfhost/src/runtimestdhost.nra` are new files outside the brief's owned paths (the brief listed `std_build.c`).
- An x32 ABI build (`__x86_64__` with `__ILP32__`) reports Linux X86_64. Out of scope.

Landed on feat/batch-9 as a741d03e.
