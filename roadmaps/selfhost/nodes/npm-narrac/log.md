# Log: npm-narrac

- Branch feat/npm-narrac from feat/batch-16.
- Built a stage with `sh selfhost/bootstrap.sh <dir>` (stage 3 ran), then `scripts/wasm-build.sh <dir>/stage2.c`: 6,332,654 bytes. The module imports 18 WASI calls and nothing else; `wasi.js` implements those 18.
- wasi-libc strips the leading `/` and `./` from every preopen name and every path, so a `/` preopen and a `.` preopen are the same empty prefix, and an absolute path looks like a relative one (ties go to the last preopen). Solved with one preopen for the working directory and one more per top folder of each absolute argument (longest prefix wins). Setting `PWD` did not change the libc working directory.
- `narrac fmt --stdin` opens `/dev/stdin` as a file, so the shim answers that path (under the working directory, when the folder has no `dev/stdin`) with the call's stdin buffer.
- A Windows narrac does not build yet: `narrac build --target Windows-X86_64 selfhost` stops on `poll.h`. The win32 package and row were dropped; Windows uses the wasm package until windows-process-async lands.
- The row loop first used `zig cc -target` on `stage2.c`, which fails for Windows because the emitted C depends on the target. Other rows are now built with `narrac build --target` and a zig wrapper as `cross-check.sh` does.
- Load was 5 to 40 during the work.
