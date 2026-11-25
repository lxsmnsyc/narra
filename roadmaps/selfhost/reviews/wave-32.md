# Wave 32 report: batch-24, std modules for servers and release tools

Batch-24 starts from `main` at aa4e6bba. Local branch `feat/batch-24`, worktree `../narra-batch-24`. Maintainer request, 2026-11-25: "one shot the p2-std".

## Landed

- std-semver: `std/semver` with `parse`, `toString`, `compare`, `bump` and `satisfies`. `satisfies` reads a requirement as a minimum version, the only form SPEC section 14 defines; `^1.2.0` and `1.2` give `false`.
- std-hex-base64: `toHex`, `fromHex`, `toBase64`, `fromBase64`, `toBase64Url` and `fromBase64Url` in `std/bytes` (`std/src/ext__bytes.nra`). The URL alphabet is its own function pair, not a flag. Decoders reject nonzero leftover bits.
- std-url: `std/url` with `parse`, `toString`, `resolve`, `encode`, `decode`, `parseQuery` and `queryToString`. `port` is `Option<string>` so every part keeps its text. A host with an unbracketed `:` is an `Err`. Fresh review: two should-fix items, fixed.
- std-signal: `std/signal` with `wait(): Result<Signal, AsyncError>` for SIGINT and SIGTERM, through a self-pipe the async loop watches. A signal during `Process.run` reaches the child and is also recorded for `wait`. Fresh review: three should-fix items, fixed. The batch review found one more race, fixed in 8f27e698.
- std-digest: `std/digest` with SHA-256, SHA-1, HMAC, PBKDF2-HMAC-SHA-256 and `equal`; `Hasher` is opaque. 64 MiB in 0.74 s wall against 0.26 s for `shasum`. Fresh review: test gaps, fixed.
- process-run-options: `Process.run` takes `input`, `environment`, `removed`, `folder` and `inheritOutput` as defaulted parameters (a struct default is E1212). `run` is Narra code over the internal builtin `Spawn.run`. Both systems search the child's `PATH`. Fresh review: the Windows runtime did not compile. Fixed and cross-compiled with `zig cc`.
- std-log: `std/log` with levels, key-value fields, text and JSON-lines forms. A later key replaces an earlier one in both forms. Odd text keys are quoted. The first worker stopped under load. Its work was finished and reviewed separately.
- Docs: guide sections for the five modules (chapters 11, 13, 14), the guide index, and the SPEC std list (which also lacked `std/set`).

## Checks

- `sh scripts/test.sh`: PASS, 11018 cases, after the seed update.
- `sh selfhost/bootstrap.sh`: fixed point, seed current (regenerated in this batch).
- `sh scripts/linux-check.sh`: PASS, 11018 cases (before the last race fix and the seed update).
- One full macOS run before the race fix failed with an unknown case, and the signal filter failed once in five runs then. After the fix, 20 runs of the signal filter passed. The cause is not proven.

## Not verified

- The Windows code of std-signal and process-run-options never ran. Process was cross-compiled. Signal was only read.

## Process notes

- Six workers at once on 8 cores pushed the load above 90. One worker stalled for 20 minutes in the seed build. Run at most three workers that build the compiler at once.
- Every new std module edits the registry, `scripts/std/generate.nra` and 9 module-list goldens, so each landing after the first conflicted. Taking the batch side and rerunning `--suite generated --update`, `--suite errors --update` and `--suite dumps --update` resolved each one.

## Still to do

- [x] Push `feat/batch-24` and open its pull request (PR #156, merged).
- [x] After it merges: set batch-24 and its children to done in `main`'s roadmap and archive.
