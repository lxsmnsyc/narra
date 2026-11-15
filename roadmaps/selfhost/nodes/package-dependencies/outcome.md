# Outcome: package-dependencies

## Built

- `selfhost/src/depload.nra`: `prepareDependencies` (resolve, check against `narra.lock`, write the lock when it changed, return the checkouts), `findUndeclaredImports`, `isDependencyPath`.
- `selfhost/src/modules.nra`: dependency packages load as modules `<name>/<path>` of one program; `selfhost/src/cnames.nra`, `hash.nra`, `check.nra`, `declare.nra`: a dependency's C names hash the package name, a colon and the path.
- `selfhost/src/deplocked.nra`: the test-only command `narrac pkg --locked <dir>`.
- Codes in `diagnostics.nra`: E0460 locked dependency not in the cache, E0461 lock has no entry for a dependency (both language service), E0462 dependency with no `.nra` file, E0463 import of a package the importer does not list.
- Suite: cases `build-app`, `build-locked`, `build-error-lock-mismatch`, `build-error-undeclared`, `build-error-no-sources` in the `deps` suite (`selfhost/tests/src/depsbuild.nra`, fixtures in `testdata/deps/build/`).
- Docs: SPEC section 15 and `docs/guide/08-modules-and-packages.md` (section "Using a package from git").

## Acceptance

- Fixture app (`app` -> `json` -> `util`, each with a module `text` that declares `render`, a trait, a constant and a test): `build-app` runs it with `-sanitize` (sanitizers and leak check) and with `-release`, prints the expected output, and a second build leaves `narra.lock` byte-identical.
- No clash: the three `text` modules and their equal function and trait names link in one program. Their C names differ because a dependency's module hash is made of `json:text`.
- Root C unchanged: `emit selfhost` by a compiler built from the base commit and by the new one are byte-identical (cmp, 12186650 bytes); the `bootstrap` suite and every golden pass.
- Errors: `build-error-lock-mismatch` (E0452 through the build), `build-error-undeclared` (E0463), `build-error-no-sources` (E0462), `build-locked` (E0461 with the lock removed, E0460 with the cache removed; the language service wrote no lock and created no cache; the locked load ran with `PATH=/nonexistent`).
- `narrac test` runs the package's tests only: json's failing test does not run (`1 passed, 0 failed`).
- Same path: `build-same-path` links a root file `src/json/text.nra` with the dependency `json` module `text` (no E0417), because a dependency module hashes `json:text` and a root path never has a colon.
- Stale lock: `build-stale-lock` removes every dependency and the next build leaves a lock with no `[[package]]`.
- Transitive lock check: `build-locked-transitive` drops `util` from a lock that keeps `json`, and the language-service load reports E0461.
- Docs: SPEC and guide chapter 8; the guide suite passes.

## Deviations

- Touched outside the owned and shared lists: `cnames.nra`, `hash.nra`, `check.nra`, `declare.nra`, `emit.nra`, `emitconsts.nra`, `emitexprs.nra`, `emitoptions.nra` (module hash text and the test list), `lspserver.nra` (drop dependency files), `main.nra` (one dispatch row and one import), `selfhost/tests/src/depsresolve.nra` (two exports) and `deps.nra`.
- Test output has no codes, so the cases match a part of each message; the `codes` suite covers the code table.
- A dependency's modules go through every check of the compiler, so its unused names and effect declarations are errors in the build that uses it.
- The language service reports E0460 and E0461 as load errors with no file.
- Not done: a dependency's package inside a repository whose root is a workspace (the manifest is read at the root of the checkout, as in deps-resolve). That needs `[workspace]` (workspace-manifest, workspace-loader).
- The hash text of a dependency module is `<name>:<path>`; the hash formula is `hashModule` in `hash.nra`, used by the loader's clash check and by `cnames.nra`.
- Not done: the files of a checkout are not hashed again on each build; the recorded tree hash is compared with the lock, as deps-fetch decided.
- `[link] search` of a dependency is ignored; only the package being built sets library paths.
- Dependency binaries are ignored.

## Stubs left

- Workspace members as dependencies and `workspace = true` (workspaces nodes).

## Gotchas

- `declName` and friends take the checker now; a new C-name helper must pass `checker.dependencyNames` to `hashedModule`.
- The known limit from deps-resolve (a version cached only under its plain tag is used even if a member tag exists) holds in `findCheckout` as well; not fixed.

## Contract

`dependencies` version 2 holds. Additions: `prepareDependencies`, `DependencyPackage` in `depload.nra`.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 10012 cases in 144s`. `emit selfhost` by the base compiler and the new one: cmp equal (12203023 bytes).
- `sh scripts/test.sh --suite deps`: `test.sh: PASS, 32 cases in 5s` (24 earlier cases and 8 `build-*`).
- `--suite generated`: `test.sh: PASS, 23 cases in 2s`. `--suite guide`: `test.sh: PASS, 331 cases in 12s`.
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C.

## Size

- Estimated L, actual L.
