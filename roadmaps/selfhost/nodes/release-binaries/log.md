# Log: release-binaries

- Branch renamed to chore/release-binaries, reset to feat/batch-16.
- Wrote scripts/release-build.sh, install.sh (repo root), the release workflow. Added NARRA_DIST and NARRA_RELEASE_SKIP_TESTS to the script.
- actionlint 1.7.12 was already installed (I did not install it). Result: clean.
- Host run started: sh scripts/release-build.sh --dry-run 0.0.0-test.
- Host run exit 0 (tests PASS, 9830 cases). install.sh good and corrupted checksum cases pass. Linux amd64 under emulation too slow; arm64 container exit 0. Guide suite PASS 329.
