# Release and installer review fixes

Branch fix/batch-17-release. Nothing was pushed, tagged or run on GitHub. The workflow was checked by reading and by running the same commands locally.

## Should fix

1. Archives were never run before release. Fixed. A new `smoke` job runs every archive on a machine that can run it, and `release` now needs `[build, smoke]`. macOS arm64 runs on macos-14, macOS x86_64 on macos-14 through `arch -x86_64`, Linux rows in a Debian 12 (glibc) or Alpine (musl) container on ubuntu-24.04 or ubuntu-24.04-arm, Windows on windows-latest with PowerShell. Each unpacks the archive and runs `narrac run` on a hello program. `scripts/release-build.sh --check-archive <archive>` is that check (also used by the build script); it also fails when the archive holds more than `narrac`. Proof, local: the check passed for linux-arm64 in Debian 12 and linux-arm64-musl in Alpine (colima). Not run locally: the Windows step, the macOS Rosetta step and the amd64 Linux rows (no runner of that kind here).
2. glibc floor. Fixed. The glibc Linux rows (Linux-X86_64 and Linux-Arm64) are now always built with zig, and a wrapper rewrites the target to `<arch>-linux-gnu.2.28`. Proof: `objdump -T` of the built binaries shows the highest versions GLIBC_2.14 (x86_64) and GLIBC_2.17 (arm64), both under 2.28. The old native build on ubuntu-24.04 needed glibc 2.39 (from the review). `docs/install.md` states glibc 2.28 or newer.
3. Prerelease text in `docs/install.md`. Fixed: it now says the one-liner 404s while only prereleases exist, and names the two ways around it (fetch the script from the release, set `NARRA_VERSION`). Comments in the installers say "newest release of any kind".
4. zig pin. Fixed: `pip install ziglang==0.17.1`, the same as the local zig and the one used in review. `scripts/cross-check.sh` uses whatever zig is on PATH. zig is now also installed on the Linux native row, which needs it for the glibc build.

## Nits

- `install.ps1`: TLS 1.2 is set for Windows PowerShell 5.1; an empty or malformed checksum file gives an `install:` message. Not run (no pwsh here); checked by reading.
- `NARRA_NO_PATH` is in the options table.
- `install.sh`: under Rosetta (`sysctl -n sysctl.proc_translated` is 1) it picks the arm64 archive. Documented in `docs/install.md`.
- Size. Cause: zig cc keeps debug info and symbols in ELF output (`file` said "with debug_info, not stripped"), and the Windows binary keeps its debug info in a separate .pdb. Fix: the cc wrapper adds `-s` for every cross build. Before: Linux-Arm64-musl 39.3 MB, Linux-X86_64 37.1 MB, Windows 5.6 MB (plus an 18.5 MB .pdb). After: Linux-Arm64 4.9 MB, Linux-X86_64 5.2 MB, Windows 5.7 MB; archives 1.9 to 2.3 MB. The native macOS row has no debug info already (the driver passes no `-g` in release).
- `std/` dropped from the archives. narrac embeds std: the archive with the binary alone ran `narrac run` in Debian 12 and Alpine, and `--check-archive` checks the archive holds only `narrac`.
- `docs/guide/12-tools.md`: one paragraph on `hovers` with an error in another file.
- `SPEC.md`: the musl rows are listed as rows a compiler runs on; only `Wasi-Wasm32` is a target only.

## Notes

- The Debian 11 image no longer has working apt mirrors (404), so the smoke job uses Debian 12. The glibc floor is proven by `objdump`, not by the image.
- `ubuntu-24.04-arm` is free on public repositories only.
