# Log: package-dependencies

- New `selfhost/src/depload.nra`: `prepareDependencies(directory, manifest, fetch)` resolves with `resolveDependencies`, compares with `checkAgainstLock`, writes `narra.lock` only when its text changes, and returns the checkout folder of each locked package. With `fetch` false (the language service) it reads the lock and the cache only (E0460, E0461). `findUndeclaredImports` (E0463) and `isDependencyPath` are there too.
- `modules.nra`: `loadPackage` takes `fetch`; the module loop became `loadPackageModules`, used for the package being built (ids first) and for each dependency (display file `<name>/src/<path>`, modules `<name>/<path>`). `load` and `loadWith` fetch; `loadCached` (the language service) does not.
- C names: `Package.dependencyNames` becomes `Checker.dependencyNames`. `hashedModule` gives the text a module hash is made of: the whole path for a dependency, the path without its package name otherwise. `cnames.nra` takes the checker where it hashed a module (`declName`, `declHead`, `instanceName`, `implName`, `cachedModuleHash`, `cachedDeclName`). Root and std names are unchanged: `narrac emit selfhost` from a compiler built before and after the change is identical.
- `narrac test` skips the test blocks of dependency modules (`emitTestMain`); the test functions are still emitted so the numbering holds.
- The language service drops diagnostics and publish targets of dependency files (`lspserver.nra`).
- Test-only command `narrac pkg --locked <dir>` (`deplocked.nra`, one row in `main.nra`) loads like the language service.
- Fixtures `testdata/deps/build/{app,error-undeclared,error-no-sources}`, cases `build-*` in `selfhost/tests/src/depsbuild.nra`, called from `deps.nra`. `depsresolve.nra` exports `buildRepositories` and `SUBSTITUTE_SCRIPT`.
- Codes E0460-E0463 in `diagnostics.nra`. SPEC section 15 and guide chapter 8 describe the build.
