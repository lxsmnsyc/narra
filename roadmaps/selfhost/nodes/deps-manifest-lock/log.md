# Log: deps-manifest-lock

- Extended the hand-written manifest reader with an inline-table value (`VTable`), `[package] version` and `[dependencies]`. E0407 removed from the reader and the codes table.
- New `selfhost/src/lockfile.nra`: `parseLock`, `formatLock`.
- Codes E0430-E0439 added to `diagnostics.nra`.
- Test hook: `narrac deps <dir>` prints version, dependencies and the narra.lock round trip; the manifest suite runs it for `testdata/lockfile/*`.
