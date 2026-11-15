---
name: targets
version: 1
status: frozen
defined_by: target-table
---

# Contract: targets

## Purpose

The one table of build targets. The driver, build conditions (`std/build`, `@target`), the runtime portability nodes and the cross-target script all read it, so a target is added in one place.

## Interface

- `selfhost/src/driver/buildtarget.nra`: `TargetRow` gains `triple: string` (the `zig cc -target` form, for example `aarch64-macos`, `x86_64-linux-musl`, `x86_64-windows-gnu`, `wasm32-wasi`) and `libc: string` (`""` or `"musl"`). Until the batch seed update, `hostOs: Option<Os>` and `hostArch: Option<Arch>` (None for Windows and Wasi, which the committed seed's std/build cannot name); after it they go back to plain `Os` and `Arch`. `BuildTarget` gains `libc`, so `-musl` targets are their own targets in `check --all-targets`. `TARGETS` lists, in this order: MacOS-Arm64, MacOS-X86_64, Linux-X86_64, Linux-Arm64, Linux-X86_64-musl, Linux-Arm64-musl, Windows-X86_64, Wasi-Wasm32.
- `std/build`: `Os` gains `Wasi`; `Arch` gains `Wasm32`. `Build.libc` is not added (musl and gnu differ only in linking).
- Target names on the command line: `<Os>-<Arch>` or `<Os>-<Arch>-<libc>`, the same text `--all-targets` prints. `narrac targets` lists them with their triples.
- The C compiler for a target other than the host is `$CC` with `-target <triple>` passed first. The driver adds per-target flags from one function, `targetCompilerArguments(row): Array<string>`, which the portability nodes extend (wasm: `-mexception-handling -mllvm -wasm-enable-sjlj` and `-lsetjmp`; Windows: none yet).
- Commands: `narrac build --target <name>` and `narrac check --target <name>`; `run` and `test` reject `--target`. `narrac targets` prints `<name> <triple>` per line. E1600: unknown target (lists the names). E1601: `-sanitize` with a target other than the host. Both exit 2.
- Exports of `buildtarget.nra`: `describeRow(row)`, `findTargetRow(name): Option<TargetRow>`, `createTarget`, `targetCompilerArguments(row)`, `describeUnknownTarget`, `describeSanitizeError`.
- The C runtime tests targets with the C compiler's own macros (`__wasi__`, `_WIN32`, `__APPLE__`, `__linux__`), not with new defines from the driver.

## Freezing

Frozen at version 1 on 2026-11-15 when target-table landed (maintainer approved). A change after this is a change request.
