# Outcome: narrac-version

## Built

- `narrac --version` prints the version; `--help` and `-h` print the usage.
- `selfhost/src/driver/version.nra` holds `VERSION = "dev"`; `main.nra` prints it.
- `scripts/release-build.sh` builds the release binary from a copy of `selfhost/` with the version stamped in, and `--check-archive` checks the printed version (`NARRA_EXPECT_VERSION`). `release.yml` sets it on every smoke step.
- `install.sh`, `install.ps1`, `docs/install.md` and `scripts/install-check.sh` use `narrac --version`.
- Test case `build-targets/version` in the `emitchecks` suite.

## Proof

Not a bug fix.

## Deviations

- `selfhost/src/driver/driver.nra` (not in the brief's paths): `usageText` is exported and lists `--version`. Two lines.
- The version has no leading `v` (`0.1.0`, not `v0.1.0`), matching archive names and what `install.sh` takes.

## Stubs left

None.

## Gotchas

- A release binary differs from the bootstrap binary only in `version.nra`, so the release host row is built once more with the native cc.

## Proposed learnings

None.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds; the built compiler prints `dev`.
- `NARRA_RELEASE_SKIP_TESTS=1 sh scripts/release-build.sh 0.0.7-check`: the archive check ran the unpacked binary and found `--version` equal to `0.0.7-check`.
- A stamped copy built by hand printed `9.9.9`.
- `sh scripts/test.sh`: see the report.

## Size

- Estimated S, actual S.
