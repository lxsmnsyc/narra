# Outcome: deps-manifest-lock

## Built

- `selfhost/src/manifest.nra`: `Dependency { name, git, version, line }`; `Manifest` gains `version` and `dependencies`. The hand-written reader got an inline-table value (`Value.VTable`) instead of moving onto `std/toml` (smaller change). `isVersion`, `stripComment` and `parseString` are exported. E0407 is gone from the reader and the `codes` table.
- `selfhost/src/lockfile.nra`: `LockEntry`, `parseLock`, `formatLock` (sorted by name, stable, a comment header, one `[[package]]` per entry, keys in `LockEntry` order). No I/O.
- Codes in `diagnostics.nra`: E0430 missing git or version, E0431 bad version, E0432 unknown or repeated key, E0433 name `std` or the package's own, E0434 duplicate name, E0435 not a valid inline table (also a trailing comma), E0436 invalid name, E0437 malformed lock line, E0438 lock entry unknown, missing or repeated key, E0439 duplicate lock entry.
- `narrac pkg --deps <dir>` (new, in `main.nra`): prints the version and dependencies of `narra.toml`, and when `narra.lock` exists, its parse, format and re-parse round trip. It exists so the manifest suite can test the data half; the loader ignores the dependency list.
- Tests: `testdata/lockfile/*` (17 cases: ok, each error code, round trip, escapes, empty), run by the manifest suite (`selfhost/tests/src/manifest.nra`).
- SPEC section 15: `[dependencies]`, the lock shape and the codes.

## Proof

Not a fix node.

## Deviations

- Cases live in `testdata/lockfile/` (owned), not `testdata/errors/`: manifest errors are checked by the manifest suite, not the `// error:` files. `testdata/manifest/dependencies` (the E0407 case) is deleted.
- Touched outside the owned and shared lists: `selfhost/src/main.nra` (the `deps` subcommand, one dispatch row plus `depsDump`). Without it the lock round trip has no way to run from the suite, because the test runner cannot import compiler modules.
- `testdata/pkgdump/dependencies.dump` (not owned) was regenerated: its E0407 golden now reads E0435 for `x = "1"`.
- Lock error messages carry no `narra.lock:` prefix; the file name is in the location.

## Stubs left

- Nothing reads `narra.lock` or uses `Manifest.dependencies` yet: deps-fetch, deps-resolve and package-dependencies do.

## Gotchas

- `[package] version` is now validated (`MAJOR.MINOR.PATCH`, E0431); a manifest with another form fails.
- Inline tables must sit on one line.
- The seed is not regenerated (batch rule).

## Contract

Freeze `dependencies` at version 1 with the draft interface unchanged: `Dependency`, `LockEntry`, `parseLock`, `formatLock`. Additions to record: `Manifest.version`, `isVersion(text): bool` exported from `manifest.nra`, and lock file shape (comment header, blank line, `[[package]]` tables, sorted by name). `depcache.nra` is still deps-fetch's.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: PASS, 9796 cases.

## Size

- Estimated M, actual M.
