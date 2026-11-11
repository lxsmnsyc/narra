# Outcome: process-exit

## Built

- `Process.exit(code: int): void uses io` in `std/process`. Rows and emitter in `selfhost/src/stdprocess.nra` (`emitProcessExit`). It emits `(fflush(NULL), _exit((int)(code)))`: pending output is flushed, live values are not released, the leak check and LeakSanitizer do not run, the OS keeps the low 8 bits of the code.
- Effect is `io`, no new effect. Result type is `void`: the language has no never type (`panic` is `void` as well). `selfhost/src/pathends.nra` `isPanicCall` also accepts `Process.exit`, so a function returning a value may end with it.
- No runtime C change (inline emit), so no `std_process.c` / `runtimestdprocess.nra` edit and no extra seed update for this node. process-signals can edit those files freely.
- New suite `exits`: `selfhost/tests/src/exits.nra`, registered in `registry.nra`; sample `testdata/exits/process_exit_code.nra` + `.out` (exit 7, output flushed through a pipe, live array and map, clean under -sanitize and the leak check).
- SPEC.md (std/process bullet after the std/fs bytes line) and `docs/guide/08-modules-and-packages.md` list the function.

## Proof

Not a fix node.

## Deviations

- Touched `selfhost/tests/src/registry.nra` and added `exits.nra` (not in the brief's paths): the programs suite requires exit 0, so a non-zero exit needs its own runner.
- Touched `selfhost/src/pathends.nra` (one condition) for the path-end rule.
- Did not edit `stdtext.nra`, `builtins.nra`, `emitstd.nra`, diagnostics (none needed); no codes used.

## Stubs left

- seed-switch replaces `skipTeardown` in `selfhost/src/emitter.nra` with `Process.exit` after the seed update.

## Gotchas

- A Process.exit call skips leak reporting by design; do not use it where leaks should show.
- The seed does not know `Process.exit` until it is regenerated.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: PASS (878 cases per shard, 0 failed).
- `--suite exits`: 1 case passed.

## Size

- Estimated M, actual S.
