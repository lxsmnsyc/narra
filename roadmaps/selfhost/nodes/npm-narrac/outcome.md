# Outcome: npm-narrac

## Built

- `npm/narrac`: the main package. `bin/narrac.js` runs the compiler with the real streams. `lib/index.js` is the JS API (`check`, `format`, `emit`, `run`, `describeRunner`, `NarracError`) with `lib/index.d.ts`. `lib/run.js` picks the runner: the native package of the host row, else `narrac-wasm`. `NARRAC_WASM=1` forces wasm. `lib/platform.js` holds the row table. `lib/wasi.js` holds the 18 WASI preview1 calls `narrac.wasm` imports, on `node:fs` (no `node:wasi`, no outside package). All CommonJS, Node 18 or later.
- `npm/narrac-<os>-<arch>[-<libc>]`: `package.json` only (`os`, `cpu`, `libc` fields) for darwin-arm64, darwin-x64, linux-arm64, linux-x64 and the two musl rows. The binary is added by the script and is gitignored. `npm/narrac-wasm`: `package.json` and `index.js`; the module is added by the script.
- `scripts/npm-package.sh [--no-check] [version]`: bootstrap (or `NARRA_STAGE`), `scripts/wasm-build.sh <stage>/stage2.c`, host package from the stage binary, extra rows named in `NARRA_NPM_ROWS` built with `narrac build --target` and a zig wrapper, version set in copies under `$NARRA_NPM_OUT/package/` (default `dist/npm`), `npm pack`, then two scratch installs and `npm/test/compare.js`. Nothing is published.
- `npm/test/compare.js`: 23 cases per project against a native narrac: 14 commands through `npx narrac` and 9 API calls. Not shipped.
- `README.md` gets one paragraph on `npm/`.

## Decisions

- Wasm host target. A wasm narrac cannot detect its host (E0989 on the Wasi row), so the loader sets `NARRA_HOST_TARGET` for the wasm instance from `process.platform`, `process.arch` and the libc, with a caller's value winning:

  | platform | arch | libc | `NARRA_HOST_TARGET` |
  | --- | --- | --- | --- |
  | darwin | arm64 | | `MacOS-Arm64` |
  | darwin | x64 | | `MacOS-X86_64` |
  | linux | arm64 | glibc | `Linux-Arm64` |
  | linux | x64 | glibc | `Linux-X86_64` |
  | linux | arm64 | musl | `Linux-Arm64-musl` |
  | linux | x64 | musl | `Linux-X86_64-musl` |

  An unmapped platform sets nothing. The API needs no host for `check`, `format` or `emit` of a program without `std/build`, so a raise at load time would break calls that work. A program that imports `std/build` then stops with narrac's own error (`could not determine the host`), which names the problem. A fallback to `Linux-X86_64` would emit C for the wrong host without a word, so it was rejected. Checked: the same call without the variable fails, with it the C equals native.
- Linux libc comes from `process.report.getReport().header.glibcVersionRuntime`.
- Windows has no package. `narrac build --target Windows-X86_64 selfhost` fails on `poll.h` today (windows-process-async). Windows and every unmapped platform use the wasm package; paths there are POSIX only (the shim has no drive letters).
- File access in the shim. wasi-libc strips the leading `/` and `./` from preopen names and paths, so a `/` preopen and a `.` preopen are one empty prefix. The shim gives the working directory the empty prefix (relative paths, `..` too) and adds one preopen for the top folder of each absolute argument; the longest prefix wins. `/dev/stdin` (read by `fmt --stdin`) is answered with the call's stdin. Paths the compiler reads through a relative `..` work; an absolute path that reaches a folder other than the top folder of an argument (for example an absolute import in another top folder) does not.
- The API calls one process per file (`check` and `emit` take one path each) and is synchronous. A fresh wasm instance per call (compile once per process) is the T1 cost.
- Other rows are built with `narrac build --target`, not `zig cc` on `stage2.c`, because the emitted C depends on the target.
- `main` pins every optional dependency to the version, including rows not built in this run, so a publish needs every row first (`NARRA_NPM_ROWS`). The package names are the maintainer's to confirm.
- Not changed: `selfhost/src`, `narra.toml` (`git diff --stat` shows only `README.md`, `npm/`, `scripts/npm-package.sh` and this node's folder).

## Acceptance

- Tarballs install and match native narrac, once with the host package (`narrac` and `narrac-darwin-arm64`) and once with `narrac` and `narrac-wasm` only: `sh scripts/npm-package.sh` printed `23 passed, 0 failed` for the host project (runner native) and `23 passed, 0 failed` for the wasm project (runner wasm). The cases compare status, stdout and stderr of `check`, `check --json`, `check --all-targets`, a missing file, a relative path, `targets`, `explain`, `fmt --stdin` (good and broken), `emit` (hello, a `std/build` program, a bad file) and no arguments, and the API results with the native diagnostics, format and C. A second run with `NARRA_NPM_ROWS=Linux-Arm64-musl` and version 1.2.3 also passed both projects and packed `narrac-linux-arm64-musl` (36 MB binary, not run here).
- Tarball contents: `narrac` holds `bin`, `lib` and `package.json`; the platform packages hold `bin/narrac` and `package.json`.
- Nothing in `selfhost/src` or `narra.toml` changes.
- No publish. `npm publish` was never run.

## Tests

- `sh scripts/test.sh`: `test.sh: PASS, 9915 cases in 217s` (no timing case failed, so no rerun).
- Formatted with oxfmt (defaults; the repo has no config). oxlint was not run: the binary found on this machine would not start from this worktree.

## Gotchas

- The module imports exactly 18 calls today. A later compiler that imports more needs the call in `wasi.js`; an unknown import answers `ENOSYS` instead of failing to link.
- The native binary is not stripped (5 MB on macOS, 36 MB for linux-arm64-musl) and the package holds it as is.
- The wasm build needs a stage from `selfhost/bootstrap.sh`: the seed does not build for wasm until the batch seed update.
- No docker or colima was used.

## Size

- Estimated M, actual M.

## Review

A fresh agent (pr-review) found no must-fix and four should-fix items. Fixed: the host package name now comes from the loader's table (the script guessed the libc from `/lib/ld-musl-*`); the script clears only `package/` and `*.tgz` of its output folder; template files are copied before the binary and the module, from a fixed list, so a stale local file cannot overwrite them; `format("")` returns `""` on both runners (native narrac fails on an empty stdin file). Also fixed from the nits: only `MODULE_NOT_FOUND` means a missing package, shim bugs are no longer hidden as EIO, more errno rows, reading stdout gives EBADF, open files close on a trap, the wasm file mode is 644, the script exits on INT and TERM, function-level comments moved inside the functions. Left: the runner is looked up on every call, a symlink opened without follow is followed, writing a directory open succeeds, and `narrac-wasm` stays an optional dependency of everyone (about 1.5 MB gzip). The full script passed again after the fixes (23 and 23 cases).
