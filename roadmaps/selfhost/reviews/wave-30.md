# Wave 30 report: batch-22, one form per std API

Batch-21 is open as PR #126. Batch-22 is stacked on it: the local branch `feat/batch-22`, in the worktree `../narra-batch-22`. It answers the maintainer's requests of 2026-11-20 before the `v0.1.0` tag.

## Landed

Each child landed as one commit or a few after a worker check; compiler, std and runtime changes also got a fresh review, and their findings were fixed before landing.

- std-async-fold: `std/async/file`, `std/async/time` and `std/async/net` are gone. Files are sync only (`std/fs`), `Time.sleep` is async, sockets are async in `std/net` (D125: async when the operating system can wait in the background, plain otherwise, one form per function).
- async-thread-functions: `Thread.startAsync(scope, work)` runs an async function on a thread's own loop; `Thread.start` takes plain functions only (E1192 names `startAsync`). A first version rewrote `Thread.start` in the checker by argument kind; the orchestrator replaced it with the explicit function (no hidden implementations).
- std-gap-survey: 33 missing pieces ranked; 8 before the tag, the rest in the inbox with the Crystal comparison.
- std-array-more: `Arrays.sort`, `sortBy` (stable), `pop`, `truncate`, `contains`, `indexOf`, `reverse`.
- std-text-more: `Float.parse`, `Float.toFixed`, ASCII `std/char` classes, `Str.trimStart`, `trimEnd`, `padStart`, `padEnd`, `lastIndexOf`.
- std-set: `std/set`, an opaque `Set<T>`.
- std-fs-more: `Fs.appendText`, `copyFile`, `rename`, `removeDirAll`, `Process.cwd`. Two safety reviews: `removeDirAll` refuses empty, root, drive, share and device-root paths and paths ending in `.` or `..`; it never follows a symbolic link, also with a trailing slash; it walks relative to folder descriptors (no path-length or descriptor limit on POSIX).
- entry-args-unused: the compiler already exempted `run`'s parameters; the warnings on the site came from the twoslash plugin's sample package having no binary entry. Fixed in `website/plugins/twoslash.ts`.
- snippet-warnings: Narra has no warnings; the extra diagnostics were on error samples. A `narra error` fence now names every diagnostic it prints, and the guide suite and the site build fail on any it does not (contract guide-samples v3).
- nested-void-match: a `match` with no value inside a match arm compiled to `void rN;` in C.
- bench-zig-odin: Zig and Odin versions of the seven benchmarks; Zig measured (load 1.97 to 2.65), Odin written but unmeasured (not installed).

## Orchestrator rulings

- One form per std API (D125); `Time.sleep` is async, so thread functions may be async through `Thread.startAsync`.
- `Parallel.map` and `Shared.update` callbacks stay plain.
- ASCII character classes keep plain names (`isLetter`); a Unicode version can come later in `std/unicode`. Padding width counts bytes, like `Str.length`.
- `removeDirAll` returns `Ok(false)` for a missing path, unlike `removeDir`.

## Checks

- `sh scripts/test.sh`: PASS, 10925 cases.
- `sh selfhost/bootstrap.sh`: fixed point, seed current (regenerated in `348f0a06`).
- `pnpm --dir website build`: passes; 0 highlighter problems; no unannounced sample diagnostics.

## Still to do

- [ ] `sh scripts/linux-check.sh`.
- [ ] Push `feat/batch-22` and open its pull request, stacked on PR #126.
- [ ] After both merge: replace the compiler's hand-written sorts, sets, character checks and `sh -c` calls with the new std functions (they need the regenerated seed), run `odin check` and the Odin column, then first-release (`v0.1.0`).
