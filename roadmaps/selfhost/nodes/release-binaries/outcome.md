# Outcome: release-binaries

## Built

- `scripts/release-build.sh [--dry-run] <version>`: builds `narrac` from the seed (`selfhost/bootstrap.sh`, fixed point checked), runs `scripts/test.sh` with that compiler, packs `narra-<version>-<macos|linux>-<arm64|x86_64>.tar.gz` (the `narrac` binary and `std/`) and a `.sha256` into `dist/`, unpacks the archive and compiles a sample with the unpacked binary. It never uploads. `NARRA_DIST` sets the output folder. `NARRA_RELEASE_SKIP_TESTS=1` skips the tests.
- `.github/workflows/release.yml`: runs only on `push: tags: v*`. A matrix of `macos-14` and `ubuntu-24.04` runs the script with the tag minus `v`. One release job downloads the archives and runs `gh release create` with `--generate-notes`. Default permissions are `contents: read`; only the release job has `contents: write`.
- `install.sh` (repo root): detects OS and CPU, finds the latest release through the `releases/latest` redirect (or `NARRA_VERSION`), downloads the archive and its `.sha256` (from `NARRA_BASE_URL` when set), checks the SHA-256 before unpacking, installs `narrac` to `NARRA_INSTALL` (default `~/.local/bin`). Only macos-arm64 and linux-x86_64 have binaries; other systems get an error that points to the source build.
- `docs/install.md`: the "Prebuilt binaries" section.

## Deviations

- `install.sh` is at the repo root, not under an owned path, so the raw URL stays short (coordinator decision).
- `.gitignore` gets one line, `/dist/` (coordinator decision).
- `docs/install.md`: only the prebuilt binaries section changed.
- The Linux build in docker ran on arm64 with `NARRA_RELEASE_SKIP_TESTS=1`, not amd64: emulated amd64 did not pass stage 0 in 40 minutes, and the full suite takes 820 s natively. The tests ran in full on the macOS host.

## Stubs left

- None.

## Gotchas

- The workflow has never run. The first real run is the first tag.
- The seed is older than the source, so the bootstrap takes stage 3. This is normal and the fixed point still holds.
- A full `release-build.sh` run takes about 15 minutes on this Mac under load (820 s of tests).
- `install.sh` finds the latest version from the GitHub redirect. It fails with "there is no release yet" before the first release.

## Proposed learnings

- Any workflow that publishes must stay tag-only, with `contents: write` limited to the release job | a push or pull request trigger would run billed or unreviewed jobs | .github/workflows | no

## Verification

- macOS arm64 host: `sh scripts/release-build.sh --dry-run 0.0.0-test` exit 0; bootstrap fixed point held (stage 3), `scripts/test.sh` PASS, 9830 cases in 820 s; archive unpacked and compiled hello.nra.
- Linux: `docker run --platform linux/amd64 ubuntu:24.04` ran about 40 minutes inside stage 0 (the emulated cc build of the seed) and was killed, so the amd64 archive was not produced. The same script in `docker run --platform linux/arm64 ubuntu:24.04` (native speed, gcc, `NARRA_RELEASE_SKIP_TESTS=1`) exit 0: bootstrap fixed point held (stage 3), archive narra-0.0.0-test-linux-arm64 packed and its unpacked narrac compiled hello.nra. The linux-x86_64 job is first proven by the first tag.
- `actionlint` 1.7.12 (already installed, not installed by this node) on `release.yml`: no findings.
- `install.sh` against archives served by `python3 -m http.server` on 127.0.0.1, scratch HOME and `NARRA_INSTALL`: good checksum, exit 0, installed `narrac` ran hello.nra; corrupted `.sha256`, exit 1, "checksum ... does not match", nothing installed.
- `sh scripts/test.sh --suite guide`: PASS, 329 cases in 95s.

## Before the first tag (maintainer)

- Read `.github/workflows/release.yml`: trigger, permissions, action versions (`checkout@v4`, `upload-artifact@v4`, `download-artifact@v4`).
- Confirm the repository and the `main` branch in the raw URL and in `repository=` in `install.sh`.
- Confirm that macOS x64 and Linux arm64 stay without binaries.
- Tag `vX.Y.Z` yields `narra-X.Y.Z-...` archives. Check the first run's two build jobs and the release page, then run the curl command in a scratch HOME.

## Size

- Estimated M, actual M. The full test run on the host dominated the time.
