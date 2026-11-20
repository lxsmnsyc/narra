# Outcome: compiler-std-files

## What changed

- `driver/workspace.nra` `makeAbsolute`: `Process.cwd()` replaces `sh -c "pwd -P"`.
- `driver/depcache.nra`:
  - `makeAbsolute`: `Process.cwd()` replaces `runProgram("pwd", [])`.
  - `rm -rf <work>/.git` is `Fs.removeDirAll`.
  - `INSTALL_SCRIPT` is gone. The install checks `Fs.exists(path)`, then `Fs.rename(hashFile, path + ".hash")` and `Fs.rename(work, path)`. Any failure of a rename goes through the new `failInstall`: when `path` is a folder (another fetch won the race) it reads the recorded hashes, otherwise it returns E0445 with the system message.
  - `REMOVE_SCRIPT` is gone. The temporary folder is removed with `Fs.removeDirAll`.

## Calls kept and why

- `GIT_SCRIPT`: `Process.run` cannot set environment variables (inbox item already open).
- `chmod -R a-w <path>` after the install, through `Process.run("chmod", ...)`: Fs has no permission function. A failed chmod is ignored, as before (`{ chmod ...; true; }`).
- `mktemp -d`: no std function for a unique temporary folder.

## Behavior notes

- `Process.cwd` is `getcwd(3)`, which returns the physical path (symbolic links resolved) on macOS and Linux, so it matches `pwd -P`.
- The old `chmod -R u+w` before removing the temporary folder is dropped. The tree under it is never read-only: the chmod runs on the final folder after the move, and `removeDirAll` removes read-only files (git objects) when their folder is writable.
- `rename` replaces an existing empty folder on POSIX, but the earlier `Fs.exists` check keeps the old guard, and a non-empty target fails with ENOTEMPTY and takes the race path.
- A rename of the hash file followed by a failed folder rename leaves `<path>.hash` behind, as the shell script did.
- No new case: the existing `deps` cases cover fetch, install and race paths that can run offline.

## Acceptance

- [x] The calls use std functions, except those kept above. Evidence: `selfhost/src/driver/workspace.nra`, `selfhost/src/driver/depcache.nra`.
- [x] `sh selfhost/bootstrap.sh`: "ok: stage 1 and stage 2 emit identical C".
- [x] `sh scripts/test.sh`: PASS, 10925 cases. `sh scripts/linux-check.sh`: PASS (10925 cases in the container).

## Inbox

- `Fs` has no permission function (`chmod`), so the read-only cache step still shells out. A `Fs.setReadOnly` or `Fs.chmod` would remove the last `sh`-free blocker here.
- `mktemp -d` could become `Fs.createTempDir`.
