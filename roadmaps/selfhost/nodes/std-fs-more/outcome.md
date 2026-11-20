# Outcome: std-fs-more

## Built

- Five builtins, each with a registry row (`selfhost/src/std/stdfs.nra`, `stdprocess.nra`), display names and an emitter, all with the `io` and `alloc` effects:
  - `Fs.appendText(path: string, text: string): Result<int, string>`: opens for append, creates the file when missing, returns the bytes added.
  - `Fs.copyFile(from: string, to: string): Result<int, string>`: returns the bytes copied. An existing target is overwritten. A folder as source and the same file as both paths are errors. POSIX keeps the permission bits of the source.
  - `Fs.rename(from: string, to: string): Result<bool, string>`: POSIX `rename`, so an existing file target is replaced and a folder target must be empty. Windows uses `MoveFileExA` with replace-existing: a file target is replaced and a folder target is refused.
  - `Fs.removeDirAll(path: string): Result<bool, string>`: `Ok(true)` when it removed a tree, `Ok(false)` when the path was missing. It walks with `lstat` (attribute `REPARSE_POINT` on Windows), so a symbolic link or junction inside the tree is unlinked and never followed. A file or link as `path` is an error (`Not a directory`).
  - `Process.cwd(): Result<string, string>`: `getcwd`, the physical absolute path.
- Runtime: `selfhost/runtime/std_fs.c` and `std_process.c`; the generated copies `selfhost/src/runtime/runtimestdfs.nra` and `runtimestdprocess.nra`, and the std reference pages `docs/reference/std/fs.md` and `process.md`, were regenerated.
- WASI: all five are available. Built with `zig cc` for `Wasi-Wasm32` and run under `wasmtime`: append, copy, rename, read, `removeDirAll` and `cwd` (`/`) gave the expected results. The WASI tables in `buildconditions.nra` are unchanged. `fchmod` is skipped there.
- Windows: written against the Win32 calls and built with `zig cc` for `Windows-X86_64` (compiles and links). Not run: no wine or Windows host on this machine.
- Tests: `testdata/io_fs_more.nra` and `.out` (33 lines: append to a new, existing and missing-folder path, to a folder; copy new, longer target, missing, folder, itself, missing target folder; rename file, over a file, missing, folder, into a missing folder; `removeDirAll` on a link, a file, a tree holding a symlink to a folder outside it (the outside file survives), a missing path; `cwd`). It runs in the `programs` suite in debug and release with the sanitizers.
- Guide: chapter 11 drops "There is no function that appends", points to the new section and lists the four functions with a runnable sample that also calls `Process.cwd`. `compiler/README.md` records the signatures and platform behavior.

## Deviations

- `copyFile` returns the byte count and `appendText` returns the byte count, like `writeText`; `rename` and `removeDirAll` return `bool`, like the other folder functions.
- Not added to `windows_runtime.nra`: it cannot be run here, so a golden would be a guess.

## Stubs left

- None.

## Gotchas

- `selfhost/src` cannot call the new functions until the seed is regenerated. Call sites for the follow-up:
  - `selfhost/src/driver/workspace.nra:75`: `Process.run("sh", ["-c", "pwd -P"])` becomes `Process.cwd()`.
  - `selfhost/src/driver/depcache.nra:34` `INSTALL_SCRIPT` (`mv -f` of the hash file, `mv` of the work folder, `chmod -R a-w`), used at line 337: two `Fs.rename` calls. The `chmod -R` read-only step has no std function and stays (or needs its own node).
  - `selfhost/src/driver/depcache.nra:36` `REMOVE_SCRIPT` (`chmod -R u+w`, `rm -rf`), used at line 371: `Fs.removeDirAll`. The `chmod -R u+w` step is needed first because the install made the tree read-only, and `removeDirAll` does not change permissions.
  - Not covered by this node: `depcache.nra:123` and `driver.nra:1212` (`sh -c` for `GIT_TERMINAL_PROMPT` and stdin), which belong to the `Process.run` options item.
- `removeDirAll` removes a read-only file on POSIX when its folder is writable. A folder without write permission fails with the system's message.
- `rename` across file systems fails on POSIX (`Cross-device link`); Windows copies (`MOVEFILE_COPY_ALLOWED`).

## Proposed learnings

- A new std row that calls the runtime needs the C in `selfhost/runtime/` and `sh scripts/test.sh --suite generated --update` before the bootstrap | the compiler embeds the runtime text, so a stale generated file builds a compiler without the function | selfhost/runtime | yes (generated suite)

## Verification

- `sh scripts/test.sh --suite guide --suite programs --suite fmt-doc --suite sources --suite query --suite generated` (filters as noted): PASS.
- Full `sh scripts/test.sh`: `PASS, 10827 cases in 448s`.
- `pnpm --dir website build`: passes.

## Size

- Estimated M, actual M.
