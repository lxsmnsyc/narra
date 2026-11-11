# build-conditions log

- 2026-11-11: branch feat/build-conditions from feat/batch-7 (7f0e0534). Nothing of the node existed (SPEC sample on spec-skip, no std/build).
- std/build is Narra source (`std/src/build.nra`): `Os`, `Arch` and `os`/`arch`/`debug` with stand-in values. Registered like std/time (stdbuild.nra row, stdlib.nra registerAll, stdtext regenerated, `Build` alias in fixes.nra).
- Host detection needs io, and checkPackage is `uses alloc`, so the driver chooses the target (`chooseTarget`: `uname -s -m` only when the package imports std/build) and passes it to `Body.checkPackageFor` / `checkModules(..., target)`.
- The stand-in values are replaced in the std/build AST before the module is declared (`applyBuildTarget`), keeping node ids, so checker refs and the emitter see the target's variant with no special case.
- Folding: `recordBuildCondition` after the condition check in `checkIf`; `Checker.foldedConditions[conditionId]`; `emitIf` writes the kept branch in a `{ }` block. `is` with a fieldless variant folds too.
- Effects: effects.nra walks every branch already, so dropped-branch effects count (probe: Env.get in a Windows branch of an exported function without `uses io` is an error).
- E0989: host with no TARGETS row, reported at each `import ... from "std/build"`. Tested with a fake `uname` first on PATH (emit suite).
- Test infra: programs suite reads `<name>.release.out` for the release run and `<name>.<darwin|linux>.out` when `<name>.out` is missing.
- Probes: closures in both branches, generator with a folded if, generic function: run clean under -sanitize.
- Fresh review fixes: `/usr/bin/uname` instead of PATH; host read once per process (LSP caches it in CheckCache.hostTarget); `NARRA_HOST_TARGET=<os>-<arch>` override, used by the emit test instead of a fake uname; "could not determine the host: <reason>" wording; E0989 once per module; SPEC notes how the host is found and that release C comes from `build -release -emit-c`.
