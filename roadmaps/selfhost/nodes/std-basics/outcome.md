# Outcome: std-basics

## Built

Checked first: none of it existed. `Hash.of` was in SPEC.md only (compiler/README said "does not exist yet"; only `Hashable.hash` worked), and `Ranges.stepped` was in SPEC.md only.

- `std/math` (`std/src/math.nra`): `PI`, `E`, `abs`, `absFloat`, `min`, `max`, `clamp` (generic over `Ordered`, so int and f64; `clamp` panics when `low > high`), `pow`, `powInt` (negative exponent panics), `sqrt`, `floor`, `ceil`, `round`, `trunc`, `sin`, `cos`, `tan`, `atan2`, `exp`, `log`.
- `std/time` (`std/src/time.nra`): `Instant` and `Duration` (nanoseconds), `now`, `elapsed`, `between`, `unixMillis`, `sleep`, `nanoseconds`, `milliseconds`, `seconds`, `asNanoseconds`, `asMilliseconds`. Clock reads and sleep are `io`.
- `std/random` (`std/src/random.nra`): `Random` struct (SplitMix64 on a `u64`), `seeded`, `fromSystem` (`io`, `/dev/urandom`), `nextBits`, `nextInt(mutable random, low, high)` (half-open, no modulo bias), `nextFloat`, `shuffle`. Documented in SPEC.md as not for cryptography.
- `std/hash` (`std/src/hash.nra`): `of` (calls `Hashable.hash`), `fnv1a(text): u64`. Checked against the published FNV-1a 64 values.
- `Ranges.stepped(start, end, step)`: builtin `Range.stepped` (`stdrange.nra`). `narra_range` has a third field `step`; a zero step panics; a negative step counts down; a loop over a Range value ends instead of overflowing.
- Primitives: `std/internal/math` (rows from `internalMathFuncs`, each calls the `math.h` function of its name), `std/internal/clock` (`selfhost/runtime/std_clock.c`), `std/internal/entropy` (`std_entropy.c`). All in `selfhost/src/stdinternal.nra`; registered in `stdlib.nra`, dispatched in `emitstd.nra`, runtime text in `runtimetext.nra`.
- Parser: `of` is accepted as a function name and after `.` (`parseMemberName`, `parseFuncHeader`), like `from`. Without it `Hash.of` cannot be written.
- Samples with `.out`: `testdata/std_math`, `std_time`, `std_random`, `std_hash`, `ranges_stepped`; parse golden `testdata/parse/keyword_of_name`.
- SPEC.md (module list, `Ranges.stepped`, a bullet for the four modules) and compiler/README.md updated.

## Deviations

- Files outside the owned and shared paths: none besides the generated `runtime*.nra` and `stdtext.nra` (rewritten by `--suite generated --update`).
- `Random.nextInt` takes `low, high` (half-open) instead of a `Range`: source code cannot read a Range's bounds. The orchestrator accepted this.
- Not done: sorting std source modules with `orderModules` in single-file packages. No std source module imports another source module yet (the four new ones import only builtins and `std/internal/*`).
- Not done: `scripts/linux-check.sh` (the orchestrator runs it). Linux note: `-lm` is already on the cc line in `driver.nra`; `std_clock.c` uses `clock_gettime` and `nanosleep`, which `runtime.h` exposes through `_POSIX_C_SOURCE`.

## Stubs left

- None.

## Gotchas

- `of` and `from` are keywords. Declaring or calling a function with those names needs the parser cases above; other keywords still fail.
- Editing `std/src/*.nra` needs `sh scripts/test.sh --suite generated --update` and a rebuilt compiler before the change is visible: the std text is embedded.
- Template parts must be strings: use `Show.show(x)` for bool and int in samples.
- A shift or `+%` on `u64` needs `u64` operands; use `U64.wrap(int)` to convert (`U64.from(int)` is E0940).
- Narra strings are joined with templates, not `+`.

## Proposed learnings

- A new `std/internal/<name>` with C goes in `selfhost/runtime/std_<name>.c`, one import and one case in `runtimetext.nra`, and its path is the key `runtimeFor` receives | the runtime text is included only for paths in `checker.used` | selfhost/src | no

## Verification

- Samples run with the stage-built compiler: all five print the expected values; `std_random` seed 0 gives the SplitMix64 reference value 0xE220A8397B1DCDAF first.
- `sh scripts/test.sh`: see the final report (result recorded in the commit message of the hand-back).

## Size

- Estimated M, actual M. Build cycles under machine load (5 to 10 minutes each) cost most of the time.
