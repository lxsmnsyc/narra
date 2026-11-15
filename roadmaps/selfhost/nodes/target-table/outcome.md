# Outcome: target-table

## Built

- `selfhost/src/buildtarget.nra`: `TargetRow` has `triple` and `libc`; `TARGETS` has the eight rows of contract `targets` in contract order. New exports: `describeRow(row)` (the name, `Linux-X86_64-musl` for a musl row), `findTargetRow(name): Option<TargetRow>`, `createNamedTarget(row, debug)`, `targetCompilerArguments(_row): Array<string>` (empty). `BuildTarget` has `libc: string = ""`; `describeTarget` and `describeTargets` print the libc suffix, so `check --all-targets` names `Linux-Arm64-musl` separately.
- `std/src/build.nra`: `Os.Wasi`, `Arch.Wasm32`; `selfhost/src/stdtext.nra` regenerated.
- `selfhost/src/driver.nra`: `-target name` / `--target name` for `build` (rejected for `run` and `test`), `check --target name`, `narrac targets` (`<name> <triple>` per line). `cCompilerArguments` puts `-target <triple>` and `targetCompilerArguments(row)` first when the named row is not the host's (host from `detectHostTarget`, so `NARRA_HOST_TARGET` applies). With `--target`, the build conditions use the named row.
- Codes: E1600 unknown target (lists the names), E1601 `-sanitize` with a non-host target. Both print as `E16xx: ...` on stderr, exit 2.
- Tests: `selfhost/tests/src/emitchecks.nra` cases `build-targets/cross-flag` (logging `CC`), `cross-errors`, `cross-list`; sample `testdata/build-targets/cross_hello.nra`. Goldens of the four `*.err` files in `testdata/build-targets/` now list the new rows.
- SPEC section 12 and 15, `compiler/README.md` updated.

## Proof

Not a fix node.

## Deviations

- Edited outside owned paths: `selfhost/tests/src/emitchecks.nra` (listed as shared), the four existing `testdata/build-targets/*.err` goldens (brief says "goldens updated"), `testdata/build-targets/cross_hello.nra` is owned.
- Contract draft: `TargetRow.hostOs` and `hostArch` are `Option<Os>` and `Option<Arch>` (None for Windows and Wasi). The committed seed's `std/build` has no `Os.Wasi`, so `selfhost/src` cannot name it before the seed update. After the seed update they can go back to plain `Os` and `Arch`.
- `-target` is a driver flag for `build` only; `run` and `test` reject it.

## Stubs left

- `targetCompilerArguments` returns an empty array; wasi-runtime and windows-runtime fill it.
- Windows and Wasi rows are never detected as the host (`Build.hostTarget` has no variant for them).

## Gotchas

- Generated `stdtext.nra` must be updated (`--suite generated --update`) with a seed compiler before `selfhost/src` can use a new std variant, and the source still cannot name it until the seed is regenerated.
- A musl row shares `os` and `arch` with its glibc row, so build conditions treat them alike; only `describeTarget` and the triple tell them apart.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite emit`: PASS, 25 cases.
- `sh scripts/test.sh`: PASS, 9783 cases.

## Size

- Estimated M, actual M.

## Cross-build run (acceptance criterion 1)

- `CC=<wrapper: exec zig cc "$@"> narrac build --target Linux-X86_64-musl -o h testdata/build-targets/cross_hello.nra`: static x86-64 ELF.
- `docker run --rm -i --platform linux/amd64 alpine sh -c 'cat > /h && chmod +x /h && /h' < h`: printed `cross`, exit 0, the output the sample is meant to give. (zig 0.17.1, macOS arm64; bind-mounting /private/tmp into docker failed, so the binary is passed over stdin.)
