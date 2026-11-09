# Log: lsp-incremental

- 2026-11-09: branch feat/lsp-incremental from origin/feat/batch-1; rebased on 0c1e0f8 (lsp-navigation landed).
- Baseline, `selfhost/` opened as a package, body edit in `src/hash.nra` (driver in the scratchpad, `lspbench.py`): open about 0.4 s to 1.6 s, change about 480 ms to 760 ms (machine noise is high).
- Profile of one change (macOS `sample`, 4 s): load 37% (lex and parse of all 111 modules plus std sources), module bodies 19%, generic instances 7%, unused check 7%, effect pass 24%, publish and drops about 6%.
- Plan: (1) keep parsed sources between loads; a changed source takes fresh ids from a high-water mark, so unchanged modules keep their ids and trees and no checker table changes; (2) skip the defaults and bodies of modules that did not change and whose imports kept their signatures, reuse their errors per file and phase.
- Found: `reportUnusedTopLevel` runs only when the package has no error so far, and the effect pass only when there is no error at all. A module's unused and effect lists therefore depend on other modules' errors; the cache keeps them per module with a "known" flag and reruns a module whose list is missing.
- Found: generic struct defaults and generic function instances are checked in the declaring module but queued by other modules' bodies. The incremental check falls back to the whole package when any of them reports an error.
