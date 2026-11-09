# Outcome: std-tables-split

## Built

- One file per std module in `selfhost/src`: `stdprelude`, `stdint`, `stdfloat`, `stdstring`, `stdarray`, `stdbox`, `stdchar`, `stdcoroutine`, `stdinternal` (internal/coroutine and internal/probe), `stdenv`, `stdfs`, `stdio`, `stdmap`, `stdoption`, `stdpath`, `stdprocess`, `stdrange`, `stdresult`, `stdtest`. Each holds its `<MODULE>_FUNCS` rows, its checks (`check<Module>Call`) and its emitters (`emit<Module>Call`).
- `stdlib.nra`: `stdRegistry` with one `registerAll` line per module, in the old order (`narrac builtins` output is identical). `stdcalls.nra`: `checkStdCall` forwards by the prefix of the check name, one `if` per module. `emitstd.nra`: `emitNamedCall` forwards by the prefix of the builtin name, one `if` per module.
- Shared helpers: `stdcheckhelp.nra`, `stdemithelp.nra`. `stdType` and `splitTop`: `stdtypes.nra`.
- One generated file per runtime source: `runtimecore`, `runtimemap`, `runtimeclosureenv`, `runtimetestrunner`, `runtimestd{env,fs,int,io,path,process,string,test}`. `runtimetext.nra` is now hand-written and only holds `runtimeFor`. The `generated` suite makes one case per file of `selfhost/runtime/` (13 cases with `stdtext`; was 2).
- `selfhost/CLAUDE.md`: where a std module and a runtime file go.

## Acceptance

- One file per module: done (above). Deviation: `stdlib`, `stdcalls` and `emitstd` hold the registration and the dispatch only. The shared helpers are in `stdcheckhelp.nra` and `stdemithelp.nra`, because imports cannot form a cycle and the dispatch files import the modules. `stdType` moved to `stdtypes.nra` for the same reason (`exprs` -> module -> `exprs`).
- Generated runtime text, one file per source: the sources were already one file each (fulfilled by retire-go: `selfhost/runtime/*` and the `generated` suite with `--update`). The generated output was one file; now one per source.
- No behavior change: `narrac emit` of all 1254 `testdata/**/*.nra` gives byte-identical C and error output before and after; `narrac builtins` is identical; the stage 1 and stage 2 C of the bootstrap are identical. No expectation file changed.
- CLAUDE.md line: done.
- `sh scripts/test.sh`: PASS (4506 sharded cases plus the bootstrap case, 0 failed, 350 s on a loaded machine).
- Go items (`TestSelfhostRuntimeText`, `TestSelfhostBootstrap`): not applicable, replaced by the `generated` and `bootstrap` suites.

## Deviations

- Edited outside the owned paths: one import line each in `exprs.nra` (stdType), `emit.nra`, `emitmaps.nra`, `emitoptions.nra`; `selfhost/tests/src/generated.nra` and one description in `selfhost/tests/src/registry.nra`.
- The seed is not regenerated. It still has the old layout; the orchestrator regenerates it (`sh scripts/seed-update.sh`) after the batch, and `generated` and `bootstrap` pass without it.

## Stubs left

- None.

## Gotchas

- A new `std_NAME.c` needs an import and an `if` in `runtimetext.nra` by hand; the suite fails with a missing file until `--update` writes its module.
- Dispatch is by name prefix: a check named `checkFooBar` goes to the module whose prefix `checkFoo` is in `checkStdCall`.
- Other workers that edit rows, checks or emitters in the old files will conflict; they must redo the edit in `std<module>.nra`.

## Proposed learnings

- Std module code lives in `std<module>.nra`; shared helpers in `stdcheckhelp.nra` / `stdemithelp.nra` | the dispatch files import the modules, so helpers cannot sit in them | selfhost | yes (cycle error)

## Size

- Estimated M, actual M.
