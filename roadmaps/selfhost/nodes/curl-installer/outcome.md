# Outcome: curl-installer

## Built

- `install.sh`: macOS and Linux, arm64 and x86_64, glibc and musl (the musl archive on a musl system, found by `ldd --version` or `/lib/ld-musl-*`). The version comes from the `releases/latest` redirect; when that finds no release, from the newest release of any kind (GitHub API). `NARRA_BASE_URL` also takes a `file://` folder, for `scripts/install-check.sh`. It ends with "open a new terminal, then run: narrac targets".
- `install.ps1`: Windows x64. Checks the SHA-256, installs `narrac.exe` to `%LOCALAPPDATA%\narra\bin` (or `NARRA_INSTALL`), adds the folder to the user PATH once (`NARRA_NO_PATH=1` skips), prints the same next step.
- `scripts/release-build.sh --target <row>`: builds one archive per row. The host row is built from the seed with the fixed-point check and the test suites. Other rows are cross builds: the bootstrapped compiler runs `narrac build -release --target <row> selfhost` with a `zig cc` wrapper in `CC`. The tests do not run on a cross row. The archive is unpacked and run where the machine can run it (host row, musl on Linux, macOS x86_64 under Rosetta). Windows archives are `.zip` with `narrac.exe`; all others `.tar.gz`. Names: `narra-<version>-<os>-<arch>[-musl]`.
- `.github/workflows/release.yml`: a matrix of the seven rows (macOS arm64 and x86_64 on macos-14; Linux x86_64 native, Linux arm64, both musl rows and Windows cross-built on ubuntu-24.04 with zig from pip). The release job checks that all 14 archive and checksum assets exist, adds `checksums.txt`, `install.sh` and `install.ps1`, and creates the release.
- `scripts/install-check.sh`: builds the archives, installs through `NARRA_BASE_URL=file://...` into a temp folder, runs `narrac targets`, and checks that a wrong checksum is refused. Linux rows run in debian (glibc) and alpine (musl) containers when docker runs. It parses `install.ps1` when pwsh exists.
- `docs/install.md`: the two one-liners first, where it installs, the check, then "Build from source". README line updated.

## Decisions

- Prerelease: the one-liner fetches `install.sh` itself from `releases/latest/download`, and that URL skips prereleases. No fallback inside the script can help, because the script is never fetched. So the first tag must have no suffix: use `v0.1.0`, not `v0.1.0-alpha.1`. `release.yml` still marks suffix tags as prereleases (and says why in a comment). `install.sh` and `install.ps1` fall back to the newest prerelease when they run from another source (for example a raw download), so `NARRA_VERSION`-less installs still work there.
- `narrac --version` does not exist in the compiler, and `selfhost/` is not an owned path. The installers and the check use `narrac targets` instead. Request for the maintainer: add `narrac --version`, then switch the installers' last line and `install-check.sh` to it.
- Cross rows are built without tests; only the host row runs them in each runner. Linux arm64 is not run natively in CI.
- zig in CI comes from `pip install ziglang` (not pinned to a version).

## Proof

Not a fix node. `sh scripts/install-check.sh` passed (exit 0) on macOS arm64 with docker through colima (arm64): MacOS-Arm64 installed natively and ran `narrac targets`; a wrong checksum was refused; Linux-Arm64 installed and ran in a debian container; Linux-Arm64-musl in an alpine container. `sh scripts/release-build.sh --target Windows-X86_64` produced a zip holding `narrac.exe` and `std/` with its `.sha256`. Not run here: the x86_64 rows, and install.ps1 (pwsh is not installed).

## Result

Not an experiment.

## Not done

- No tag, no workflow run, no release (as required). `release.yml` is checked by reading and by running the same `release-build.sh` commands locally.
- `install.ps1` was not run: pwsh is not installed here, and the Windows archive needs Windows. It is checked by reading.
- The first-release brief is untouched.
