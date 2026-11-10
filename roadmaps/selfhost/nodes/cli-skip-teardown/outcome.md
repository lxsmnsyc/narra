# Outcome: cli-skip-teardown

## Built

- `skipTeardown(checker, package)` in `selfhost/src/emitter.nra`. Unless `NARRA_CHECK_LEAKS` is set, it parks the Checker and the Package in a `Cell` whose value holds the same Cell (a documented identity cycle), so nothing frees them and the process exit does.
- `runEmit` (`narrac emit`) calls it after `emitProgram`. `compileToC` in `selfhost/src/driver.nra` calls it too, so `narrac build` skips the release, and so do `run` and `test`, which share `compileToC`. The language server (`lsp.nra`) does not call it and keeps releasing.
- With `NARRA_CHECK_LEAKS` set (test suites, bootstrap), the function returns first and the full release runs, so the leak check still covers the compiler.
- No runtime or seed change. The std library has no process-exit function, and adding one needs a seed update, so the cycle is the way to leave the release out with the committed seed. A later node can replace it with an exit call once the seed has one.

## Proof

- base: feat/batch-5 at 61199a64. `/usr/bin/time -l <narrac> emit selfhost`, instructions retired, 7 alternating pairs, load average 18 to 22 on 8 cores. Mean 1889.8 M, range 1874.8 to 1912.8 M. `build selfhost` (with `CC=true` so only narrac counts): mean 2628.9 M, range 2611.2 to 2643.4 M.
- branch: emit mean 1848.4 M (-2.2%), range 1837.6 to 1862.2 M. Build mean 2586.3 M (-1.6%), range 2572.1 to 2596.2 M. The branch won 7 of 7 pairs on both, and the ranges do not overlap.
- Both compilers were emitted by the same seed-built compiler and built with the same `cc` flags. Per-run numbers are in `log.md`.

## Deviations

- The profile put the teardown at 6.3%; the measured gain is 2.2% (emit) and 1.6% (build). The remaining teardown is the Emitter and its tables released inside `emitProgram` (`emit.nra`), which this node does not own.
- The skip also applies to `narrac run` and `narrac test`, because they share `compileToC` with `build`. They are short-lived paths too.

## Stubs left

- None.

## Gotchas

- Do not use `NARRA_CHECK_LEAKS` unset to mean "no leaks": with it unset the Checker and the Package are leaked on purpose.
- Replace the cycle with a process-exit call once the seed has one. Then the Emitter inside `emitProgram` can be skipped too.

## Proposed learnings

- A compiler path that ends the process may leave its working memory unreleased unless the leak check runs | the OS frees it faster than reference counting | selfhost/src | no

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --suite lsp` and `--suite sources`: pass.
- `NARRA_CHECK_LEAKS=1 narrac emit selfhost`: exit 0, no leak report.
- `sh scripts/test.sh` (full), seed-built compiler: pass (8 shards, about 732 cases each).

## Size

- Estimated S, actual S.
