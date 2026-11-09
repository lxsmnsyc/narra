# Outcome: lsp-incremental

## Built

- `selfhost/src/lspcache.nra`: `CheckCache` (on `Server.cache`), `loadPackage` and `checkLoaded`. `checkDocumentPackage` (`lspserver.nra`) calls these two in place of `Modules.loadWith`, `Body.checkPackage` and `Effects.checkEffects`.
- `selfhost/src/modules.nra`: `ParseCache`, `CachedSource`, `createParseCache` and `loadCached(path, overrides, mutable cache)`. A source with unchanged text keeps its tree and node ids. A changed source is parsed with ids from `cache.nextId`, which only grows. `loadWith` is `loadCached` with a fresh cache, so `narrac check`, `pkg --ids` and every other command number the files as before. Std sources are cached by import path.
- `selfhost/src/body.nra`: `checkModules(package, skipped)` returns `ModulesCheck { checker, unusedStart, queuedErrors }`. It declares every module, but checks field defaults and bodies only for modules not in `skipped`, and leaves the skipped modules out of the unused check. `checkPackage(package)` is `checkModules(package, Map {}).checker`. The generic struct defaults part of `checkDefaults` is now `checkQueuedDefaults`. It runs for every module, so skipped modules do not change when it runs.
- `selfhost/src/lsp.nra`: `narrac lsp [--full] [--trace]`. `--full` checks the whole package on every check. `--trace` writes one line per check to stderr: `check: whole package`, `check: modules <files>` or `check: no module changed`.
- `lsp` suite (`selfhost/tests/src/lsp.nra`): an `incremental_*` case runs with `--trace` (the trace lines are in `expected.txt` under `# stderr`), then again with `--full`, and the protocol output of both runs must be equal. New cases: `incremental_body` and `incremental_signature`.
- compiler/README.md "Language server": the incremental check and the two flags.

## How the incremental check works (for dependents)

- Changed module: its `firstId` differs from the cached one. A reparse always takes new ids.
- Signature: the module's tokens, without the bodies of its non-generic top-level functions (`makeSignature`). It is computed only for changed modules, from the old and the new `SourceModule` kept in the cache. When it differs, every module that imports the changed module, directly or not, is checked again.
- Errors are cached per module file, in two lists: the check phase (everything in `checkPackage` before the unused check) and the unused phase. The unused list is marked known, and the effect pass marked clean, only for a module that went through that phase with its current text. When a phase has to run and a skipped module has no known result for it, the check runs again with that module added. This happens once after an error is fixed.
- The merged list holds each file's errors in the order of a whole check. The published output is byte-equal to `--full`: both scripted cases check this, and five scripted edits on `selfhost/` gave equal outputs (body type error, exported signature change, body edit in `check.nra`, new unused local, effect error in a body).
- The check falls back to the whole package when:
  - this is the first check, or the package or its module list changed;
  - an error lies outside every module file;
  - a generic struct default or a generic function instance has an error (other modules' bodies queue those, so a skipped module could hide one);
  - the effect pass reports an error, which stops the pass for every module. The cache stays invalid until a whole check finds no effect error.
- Memory: when `parses.nextId` passes `4 * liveIds + 100000`, the parse cache starts over and the next check is a whole check.
- Navigation (lsp-navigation) still loads and checks the whole package for each request. `lspnavigation.nra` could call `LspCache.loadPackage(mutable server.cache, ...)` to skip the parse (about 37% of a whole check). Its checker needs every body, so the partial checker cannot serve navigation. Keeping the last whole checker on `Server` (the navigation worker's suggestion) does not fit this cache, because an incremental check never builds a whole checker. It needs its own field, filled by `checkWhole` or on demand.

## Measurement

`selfhost/` opened as a package, `src/hash.nra` open, five `didChange` messages that toggle a body edit (`offset = offset + 16` and `+ 0`). The time runs from sending `didChange` to the reply to a request sent right after it (every publish comes first). Driver: `lspbench.py` in the session scratchpad. `bench.py` ran 5 interleaved runs and 1 warmup, with base = the batch head before this node (0c1e0f8):

| | base (whole recheck) | new | change |
|---|---|---|---|
| body edit in a leaf module, median | 363.6 ms (mad 0.8) | 11.1 ms (mad 0.1) | 33x faster |
| didOpen (whole check), median | 377.2 ms | 367.1 ms | same |

Profile of the base change (macOS `sample`): load 37%, module bodies 19%, generic instances 7%, unused check 7%, effect pass 24%. In the new path most of a change goes to declaring the 113 modules again. In single runs from the comparison script (noisy machine, not bench.py), a body edit in `check.nra` took about 80 ms incremental against 560 ms with `--full`. An exported signature change in `hash.nra` checks 60 of 113 modules again.

## Acceptance

- [x] A body edit checks that module only, and a signature edit also checks its importers, directly or not. Diagnostics equal the whole recheck: `incremental_body` and `incremental_signature` with their trace lines, and the `--full` comparison in the runner.
- [x] On `selfhost/`, a body edit in a leaf module: 363.6 ms before, 11.1 ms after (bench.py), 33x.
- [x] `narrac check`, `narrac query` and the bootstrap output are unchanged. `loadWith` numbers the ids as before, and `checkPackage` runs the same passes in the same order (the queued generic defaults are checked at the same point). Suites query, errors, packages, dumps, programs and bootstrap pass.
- [x] Fast check: lsp, query, programs, errors, packages, sources, dumps and bootstrap pass (see the report).

## Gate decisions

- Stable ids from a parse cache with a growing id high-water mark. Alternatives: per-module id ranges with slack (fails when a module outgrows its slot), or renumbering the checker's side tables per module (the checker table change the brief warns about). No checker table changed. The ids are no longer gapless in the server, which only `pkg --ids` (`checkIds`) checks, and that command does not use the cache.
- Reuse errors, not checker state. A skipped module is still declared, so the checker tables are complete for signatures, and its bodies are skipped. Alternative: keep each module's checker results and merge them, which needs ids stable across checker tables.
- Signature = tokens without the bodies of non-generic top-level functions. Conservative: an edit to a private function's signature, an implement block or a generic function body also checks importers. Alternative: exported declarations only, which needs care about private types in exported signatures and about generic bodies that importers instantiate.
- Effect errors force whole checks until they are gone. Alternative: per-module effect lists. These are not independent, because one E1 (signature) error stops the walk for every module.
- `--full` and `--trace` are command-line flags, not `initializationOptions`, so the test runner can pick them per case.

## Found (outside this node)

- `reportUnusedTopLevel` (`unused.nra`) splits each module's whole source into lines for every module, even when nothing is unused. That costs 7% of a whole check (more with the private functions that look unused when bodies are skipped). The skipped modules are now left out by clearing `isSource` around the call in `checkModules`. A cheaper `reportUnusedImports` would speed up `narrac check` too.

## Touched outside owned paths

- `selfhost/src/lspserver.nra` (the `cache` field and the two calls), `selfhost/tests/src/lsp.nra` (the runner's `--full` comparison). Both are small. They are needed for the test and the wiring the brief asks for.

## New codes

- None.

## Stubs left

- None.

## Size

- Estimated M, actual M.
