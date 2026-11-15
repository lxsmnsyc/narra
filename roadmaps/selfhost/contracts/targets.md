---
name: targets
version: 0
status: draft
defined_by: target-table
---

# Contract: targets

## Purpose

The one table of build targets. The driver, build conditions (`std/build`, `@target`), the runtime portability nodes and the cross-target script all read it, so a target is added in one place.

## Interface

- `selfhost/src/buildtarget.nra`: `TargetRow` gains `triple: string` (the `zig cc -target` form, for example `aarch64-macos`, `x86_64-linux-musl`, `x86_64-windows-gnu`, `wasm32-wasi`) and `libc: string` (`""`, `"gnu"`, `"musl"`). `TARGETS` lists, in this order: MacOS-Arm64, MacOS-X86_64, Linux-X86_64, Linux-Arm64, Linux-X86_64-musl, Linux-Arm64-musl, Windows-X86_64, Wasi-Wasm32.
- `std/build`: `Os` gains `Wasi`; `Arch` gains `Wasm32`. `Build.libc` is not added (musl and gnu differ only in linking).
- Target names on the command line: `<Os>-<Arch>` or `<Os>-<Arch>-<libc>`, the same text `--all-targets` prints. `narrac targets` lists them with their triples.
- The C compiler for a target other than the host is `$CC` with `-target <triple>` passed first. The driver adds per-target flags from one function, `targetCompilerArguments(row): Array<string>`, which the portability nodes extend (wasm: `-mexception-handling -mllvm -wasm-enable-sjlj` and `-lsetjmp`; Windows: none yet).
- The C runtime tests targets with the C compiler's own macros (`__wasi__`, `_WIN32`, `__APPLE__`, `__linux__`), not with new defines from the driver.

## Freezing

target-table freezes this at version 1 when it lands. A change after that is a change request.
