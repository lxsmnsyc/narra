# Log: c-packages

- 2026-11-26: checked that no `git`/`tag` in `[c]`, no `from "<package>/<path>"` and no `c_*` lock keys existed. Started from `feat/batch-26`.
- Decision: the upstream checkout is part of the lock entry of the C package (keys `c_git`, `c_tag`, `c_commit`, `c_content`, all or none), not an entry of its own. The package is the unit that resolution selects, so the upstream follows the selected version and a lock entry stays one block. Fetched in `depresolve.nra` with `fetchVersion`, the cache and the fetcher of the dependencies.
- Decision: `[c] git` is for a package fetched as a dependency. The package being built and workspace members have no lock entry to hold the upstream, so they get E1665.
- Decision: a `from "<package>/<path>"` is parsed as a block library when the first part is a package name and a `/` follows; it needs the package in the `[dependencies]` of the binding package (E1663), a `[c]` table in the named package (E1664), the file to exist (E1660) and stay inside (E1661).
- Include folders go to packages that list the C package in `[dependencies]` (`PackageC.declared`), defines never.
- Moved tag: found when a fetch (cache gone) returns another tree than the lock (E0452). With the checkout in the cache, nothing is fetched, as for git dependencies.
- Tests: `testdata/packages/c-packages/<case>/` in the `packages` suite, `commands` variant, with `repos/` (git repositories built with fixed author and `GIT_CONFIG_GLOBAL=/dev/null`, `file://` URLs through `@REPOS@`). New steps: `delete`, `has`, `retag`, `locked`.
