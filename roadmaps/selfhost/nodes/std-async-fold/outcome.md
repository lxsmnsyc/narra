# Outcome: std-async-fold

## Built

- `std/net` (`std/src/net.nra`): the old `std/async/net` under its new path, names unchanged. Registry row in `selfhost/src/std/stdasyncio.nra`; the WASI left-out list in `selfhost/src/check/buildconditions.nra` names `std/net`. The guide imports it as `Net`.
- `Time.sleep` in `std/src/time.nra` is async: `export async function sleep(duration: Duration): Result<Duration, AsyncError> uses io, alloc`, the body of the old `AsyncTime.sleep`, with the same cancel and exit behavior. `std/time` now imports `std/internal/async` and `AsyncError` from `std/async`.
- Removed: `std/src/async__file.nra`, `std/src/async__time.nra`, their registry rows, the blocking primitive `Clock.sleepNanos` (row and emitter in `selfhost/src/std/stdinternal.nra`) and `narra_clock_sleep_nanos` in `selfhost/runtime/std_clock.c`. An import of an old path is an unknown module.
- Sync callers of the blocking `Time.sleep`:
  - `testdata/std_time.nra`: `run` is async, and it awaits `Time.sleep` and prints `slept: true`.
  - `testdata/windows_runtime.nra`: `run` is async, and it awaits `Time.sleep(20 ms)`.
  - `testdata/channel_full_blocks.nra`: the sleep sits in a `Thread.scope` callback, which cannot be async, so the receiver spins for 2 ms on `Time.now()` and `Time.elapsed`.
  - Guide chapter 11 "Measure time": the sample measures a loop instead of a sleep. A sentence points to chapter 14 for the async `Time.sleep`.
- `testdata/async_io_file.nra` became `testdata/async_fs.nra`: `std/fs` calls in an async entry, with a ticker task that runs only at `await Async.idle()`.
- All other testdata: `AsyncTime.sleep` became `Time.sleep`, `"std/async/time"` became `"std/time"`, and `std/async/net` became `std/net`. The module lists in 8 expected-error files and one declsdump now read `std/async, std/net`.
- `DECISIONS.md`: D125 (one form per std API, async when the OS can wait in the background) and an amended line in D109 item 7.
- `SPEC.md`: the std module list, the `std/time` line, section 8 ("One form per std function (D125)", the blocking-calls rule), the Sendable list, the WASI list and both samples.
- `compiler/README.md` (the async I/O line), `skills/narra/SKILL.md` (one paragraph on the rule), `scripts/cross-check.sh` (comment), `scripts/std/generate.nra` (module table: `std/net` row, no `std/async/*` rows, `std/time` description).
- Guide: chapter 14's sleep samples use `Time.sleep`. "Read files and talk to the network" became "Files and the network", which states the rule. Its file sample is now `std/fs` with `Async.idle`, and its net section is "Network connections: `std/net`". Chapters 8, 11, 15, 19, 20 (a new "Async code" section that compares Node.js `readFile`/`readFileSync`, `setTimeout` and `net`) and the guide README follow.
- Regenerated: `docs/reference/std` (`net.md` new, `async/*.md` gone, `time.md`, `README.md`), `selfhost/src/std/stdtext.nra`, `selfhost/src/runtime/runtimestd{async,asyncio,clock}.nra`.

## Deviations

- `Clock.sleepNanos` and its C function were removed too, since nothing uses them after the blocking sleep goes. The brief did not name them.
- The doc comment of `sleep` moved into the module header of `std/time`, since the code style rule forbids comments that describe a function.
- Chapter 20 got a short new section. The brief lists chapter 20 but the chapter had no async text before.

## Stubs left

- None.

## Gotchas

- `std/time` now pulls in `std/async`. The compiler and its tests do not import `std/time`, so the seed rule is unaffected. A future selfhost import of `std/time` would need a seed that compiles async functions.
- A plain function that must wait has no sleep now: it spins on `Time.now()` or becomes async.

## Proposed learnings

- none

## Verification

- `sh scripts/test.sh` (from the seed): `test.sh: PASS, 10820 cases in 361s`.
- Targeted: guide PASS 610 cases, spec 65, errors 907, exits 8, sources 1, generated 26; `--filter async` and `--filter time` pass.
- `sh selfhost/bootstrap.sh <scratch>`: fixed point holds.
- `pnpm --dir website build`: passes.
- plain-writing checker on the edited chapters, skill and std comments: no new warnings.

## Size

- Estimated L, actual M. The rename was mechanical, and most of the text work was in chapter 14.
