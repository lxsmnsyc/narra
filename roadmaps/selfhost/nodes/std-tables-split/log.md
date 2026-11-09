# Log: std-tables-split

- Started from origin/feat/batch-1 (c1c6a02). The brief predates PR #101; the Go half is ignored.
- Already met by retire-go: the C runtime is one file per piece in `selfhost/runtime/`, and the `generated` suite rewrites and checks the generated text. Not met: one generated file per runtime source (all in one `runtimetext.nra`) and the std split.
- Baseline: built narrac from the seed, saved the C of all 1254 `testdata/**/*.nra` (`narrac emit`) and `narrac builtins` before any edit.
- Wrote a throwaway script that parses the three std files into top-level items, assigns each to a module by name, computes shared helpers by who references them, and writes `std<module>.nra` with the imports it needs. Then ran `narrac fmt`.
- Import cycle 1: the module files import `exprs` (checks), `exprs` imported `stdType` from `stdlib`, and `stdlib` imports the module files (rows). Moved `stdType` and `splitTop` to `stdtypes.nra`; `exprs.nra` imports from there (one line outside the owned paths).
- Import cycle 2: the dispatch in `stdcalls.nra` and `emitstd.nra` imports the modules, so shared helpers cannot stay in those files. They moved to `stdcheckhelp.nra` and `stdemithelp.nra`.
- `std/internal/coroutine` and `std/internal/probe` rows are in `stdinternal.nra` as two tables, because Go's order puts `fs` between them and the registry order must not change.
- Unused `pos` parameters are errors, so the per-module emit function takes `pos` only when it uses it.
- Runtime: `generated.nra` now makes one case per file of `selfhost/runtime/`; `runtimetext.nra` is written by hand (`runtimeFor`). Consumers import from the new modules (`emit.nra`, `emitmaps.nra`, `emitoptions.nra`: one import line each).
- Compared the C of every testdata program and `narrac builtins` against the baseline: identical.
