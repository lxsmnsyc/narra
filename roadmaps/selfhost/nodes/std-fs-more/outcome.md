# Outcome: std-fs-more

## Built

- Five builtins, each with a registry row (`selfhost/src/std/stdfs.nra`, `stdprocess.nra`), display names and an emitter, all with the `io` and `alloc` effects:
  - `Fs.appendText(path: string, text: string): Result<int, string>`: opens for append, creates the file when missing, returns the bytes added.
  - `Fs.copyFile(from: string, to: string): Result<int, string>`: returns the bytes copied. An existing target is overwritten. A folder as source and the same file as both paths are errors. POSIX keeps the permission bits of the source.
  - `Fs.rename(from: string, to: string): Result<bool, string>`: POSIX `rename`, so an existing file target is replaced and a folder target must be empty. Windows uses `MoveFileExA` with replace-existing: a file target is replaced and a folder target is refused.
  - `Fs.removeDirAll(path: string): Result<bool, string>`: `Ok(true)` when it removed a tree, `Ok(false)` only when the path was missing (ENOENT; Windows `ERROR_FILE_NOT_FOUND` or `ERROR_PATH_NOT_FOUND`); any other failure is an `Err`. A symbolic link or junction inside the tree is removed as a link and never followed. A file or link as `path` is an error (`Not a directory`). It refuses, before removing anything, an empty path, a root (`/`, `//`, `\`, a drive root such as `C:\` or `C:`) and a path whose last component, after trailing separators, is `.` or `..` (`C:..` included).
  - `Process.cwd(): Result<string, string>`: `getcwd`, the physical absolute path.
- Runtime: `selfhost/runtime/std_fs.c` and `std_process.c`; the generated copies `selfhost/src/runtime/runtimestdfs.nra` and `runtimestdprocess.nra`, and the std reference pages `docs/reference/std/fs.md` and `process.md`, were regenerated.
- WASI: all five are available. Built with `zig cc` for `Wasi-Wasm32` and run under `wasmtime`: append, copy, rename, read, `removeDirAll` and `cwd` (`/`) gave the expected results. The WASI tables in `buildconditions.nra` are unchanged. `fchmod` is skipped there.
- Windows: written against the Win32 calls and built with `zig cc` for `Windows-X86_64` (compiles and links). Not run: no wine or Windows host on this machine.
- Tests: `testdata/io_fs_more.nra` and `.out` (41 lines after the review: append to a new, existing and missing-folder path, to a folder; copy new, longer target, missing, folder, itself, missing target folder; rename file, over a file, missing, folder, into a missing folder; `removeDirAll` on a link, a file, a tree holding a symlink to a folder outside it (the outside file survives), a missing path; `cwd`). It runs in the `programs` suite in debug and release with the sanitizers.
- Guide: chapter 11 drops "There is no function that appends", points to the new section and lists the four functions with a runnable sample that also calls `Process.cwd`. `compiler/README.md` records the signatures and platform behavior.

## Review fixes

- POSIX walk: `openat(O_DIRECTORY | O_NOFOLLOW)` plus a device and inode check (macOS hides `O_NOFOLLOW` under `_POSIX_C_SOURCE`), `fdopendir` on a fresh `openat(".")`, `fstatat(AT_SYMLINK_NOFOLLOW)` and `unlinkat` relative to the parent. Iterative: a folder's names are read before going deeper, the parent descriptor is closed and opened again through `..` (checked by device and inode), so it holds at most three descriptors and builds no full path. WASI keeps one descriptor per level, because WASI refuses `..` through a descriptor.
- Windows: `FindFirstFileA` lists a folder (one handle open at a time); a folder link or junction is removed with `RemoveDirectoryA`, a file link or file with `DeleteFileA`, after clearing `FILE_ATTRIBUTE_READONLY`.
- Messages: a rename error is `<from> -> <to>: reason` on both systems; a copy error names the path that failed (source on stat, open and read; target on open and write) and both paths for the same-file case and Windows `CopyFileA`.
- `copyFile` keeps `mode & 0777`; a failed copy leaves a partial target (documented).
- Checked outside the suite with a C harness under ASan in a scratch folder: a 700-level tree (deeper than PATH_MAX) holding a link to an outside folder, under `ulimit -n 32`, is removed and the outside file survives; a locked parent gives `Permission denied`, a file in the path `Not a directory`, a folder without write permission an `Err`. The guard rule was checked on strings for POSIX and for the Windows rule (`C:\`, `C:`, `C:/`, `C:..`, `\`).
- WASI after the fixes: built with `zig cc` for `Wasi-Wasm32` and run under `wasmtime --dir`: a 40-level tree, the `..` guard, a missing path (`Ok(false)`) and a file in the path (`Err`) behave as on POSIX.
- Windows after the fixes: compiles and links with `zig cc -target x86_64-windows-gnu` (`narrac build --target Windows-X86_64`, and the runtime alone with `-Wall`); not run.
- New cases in `testdata/io_fs_more.nra`: a read-only file inside a tree, the empty path and `x/..`, `x/../`, `x/.` guards (the folder survives), a folder inside a locked parent (an `Err`, not `Ok(false)`; as root, where the permission is not enforced, the removal succeeds), a 300-level tree (above the macOS default of 256 descriptors). Error lines for rename and copy changed to the new rule.
- Emitters: `emitTwoPaths(hasCount)` split into `emitCopyFile` and `emitRename`; `emitWriteTextCall` renamed `emitWriteWith`.

## Second safety review fixes

- Trailing slash (POSIX, high): `removeDirAll("lnk/")` on a link to a folder gave `Ok(true)` and removed the target, because `lstat` and `openat(O_NOFOLLOW)` follow a link when the path ends in `/`. The guard now returns the length of the path without trailing separators, and the removal uses that copy. Before (C harness on the old `std_fs.c`, scratch folder): `lnk/ -> Ok(true)`, target gone. After: `lnk`, `lnk/`, `lnk//` and the absolute `.../lnk/` give `Not a directory`, the target and its file survive. New cases in `testdata/io_fs_more.nra`: `link/` and `link//` to a folder inside the base.
- Final removal (POSIX, low): the folder above the last name is opened once; `fstatat`, the first `openat` and the final `unlinkat(AT_REMOVEDIR)` are relative to it, so the path is resolved once. A parent that may be searched but not read (mode 0300) falls back to the path from the working folder; checked in a scratch folder.
- Windows roots (high): the guard now refuses `\\server\share`, `\\server\share\`, `\\server`, `\\?\C:\`, `\\?\C:`, `\\.\C:\`, `\\?\`, `\\?\UNC\server\share` (any case of `UNC`) and `\\?\Volume{...}\`: a share needs a name after the share, a device path a name after the device. The earlier refusals stay. Checked with a host C test of the guard alone, built with `NARRA_PATH_WINDOWS=0` and `=1` under ASan and UBSan (61 strings, 55 checks for Windows and 18 for POSIX, 0 failures; kept outside the repo, there is no C unit test harness here).
- Windows trailing characters (medium): outside `\\?\` and `\\.\` paths, a last part of only dots and spaces (`C:\a\.. `, `C:\a\...`, `junction\.`) is refused, and other trailing dots and spaces are dropped (`C:\a. \` is `C:\a`), as Win32 does. Trailing separators are dropped before `GetFileAttributesA`, so `junction\` is `Not a directory`. The walk runs on the `GetFullPathNameA` path, which goes through the guard again.
- Windows long paths (low): a full path of `MAX_PATH` or more, the folder's own or an entry's, is an Err (`ERROR_FILENAME_EXCED_RANGE`), never `Ok(false)`. Documented in the guide and `compiler/README.md`. `\\?\` paths were not used for the walk: the ANSI calls stay limited to `MAX_PATH` without a long-path manifest.
- Messages (low): the walk names the entry it stopped at: `<path>/sealed/file.txt: Permission denied` (POSIX builds it from the level names, Windows from the full path below the caller's). New case in `testdata/io_fs_more.nra`.
- Docs (low): chapter 11 says the removal stops at the first entry it cannot remove and leaves the rest, crosses mount points as `rm -rf` does, keeps one folder open per level on WASI, and lists the Windows roots and the length limit. The `Fs.removeDirAll` row has a display `note`, so `docs/reference/std/fs.md` prints a `Details:` line with what it does and refuses.
- WASI: built with `narrac build --target Wasi-Wasm32` (`CC` running `zig cc`) and run under `wasmtime --dir data::/work`: `lnk`, `lnk/`, `lnk//` and `file/` give `Not a directory` and the target survives; a tree with a link inside and a 40-level tree are removed; a missing path is `Ok(false)`; `t/..`, `x/../` and `/` are refused.
- Windows: `narrac build --target Windows-X86_64` with `zig cc` compiles and links; the runtime alone builds with `-Wall -Wextra` for `x86_64-windows-gnu` without warnings. Not run.

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
- Full `sh scripts/test.sh`: `PASS, 10827 cases in 448s`; after the review fixes `PASS, 10870 cases in 292s`; after the second safety review `PASS, 10895 cases in 367s`.
- After the second safety review: `--suite programs --filter io_fs`, `--suite generated` and `--suite query` pass.
- `pnpm --dir website build`: passes.

## Size

- Estimated M, actual M.
