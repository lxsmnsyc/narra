# Log: deps-resolve

- New `selfhost/src/depresolve.nra`: `resolveDependencies(roots, overrides)`, `checkAgainstLock(resolved, locked)` and the test-only command `narrac pkg --resolve <root dir> [<more root dirs>]` (dispatch row in `main.nra`). The command writes `narra.lock` in the first directory when the text differs and prints it.
- The walk is a queue of requirements. Each (name, version) is fetched and read once; the selected version is the highest one seen. Requirements of a version that loses stay in the graph, as in Go's minimal version selection.
- An override rewrites the URL and the version of a requirement before the URL check, so a name reached through two URLs is fine when it is overridden. An overridden version is fetched exactly; the versions that the manifests ask for are never fetched.
- The tag of a version: a cached member tag, a cached plain tag, else `listVersions` must list the version (E0451, E0456 for an override) and the member tag is tried before the plain tag.
- `manifest.nra`: `Manifest.overrides` and the `[overrides]` table, parsed like `[dependencies]` (same errors, same name checks).
- `depcache.nra`: `isVersionBefore` is exported.
- Codes E0450-E0458 in `diagnostics.nra`.
- Fixtures: `testdata/deps/resolve/<case>/{roots,repos,expected.txt|error.txt}`; the suite builds the git repositories from `repos/`. Cases are in `selfhost/tests/src/depsresolve.nra`, called from `deps.nra`.
