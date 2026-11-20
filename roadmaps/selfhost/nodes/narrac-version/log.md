# Log

- Mechanism: `selfhost/src/driver/version.nra` holds `VERSION = "dev"`. The seed compiles it, so no runtime or seed change is needed.
- `scripts/release-build.sh` copies `selfhost/` to a temp folder, stamps the version line, and builds the release binary from the copy for every row. The host row now builds a separate binary with the native cc (before, it packed the bootstrap binary).
- `narrac --version` prints the version as the script receives it, without the leading `v` (same as the archive names).
- `narrac --help` and `-h` print the driver usage (already existed) on stdout, exit 0.
- Release archive check: `NARRA_EXPECT_VERSION` makes `--check-archive` compare `--version`; the workflow sets it on the macOS, container and Windows smoke steps.
