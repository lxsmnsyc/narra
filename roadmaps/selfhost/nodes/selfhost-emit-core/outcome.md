# Outcome: selfhost-emit-core

## Built

- `narrac emit [--tests] [--bin name] <path>` (`selfhost/src/emitter.nra`, `runEmit`, called from `main.nra`). Exit 0: C on stdout. Exit 1: load, check or effect errors, or Go's `CompileBinary` messages for an unknown binary or a missing entry (single file and package forms). Exit 3: C with `/* unsupported: <construct> (<node>) */` markers, plus one `unsupported: <construct> (<node>)` stderr line per distinct construct in first-seen order. Exit 2: usage. `EmitterWalker implements EmitWalker` forwards to `emitexprs.nra` and `emitstmts.nra`.
- `selfhost/src/emitstate.nra`: `Code`, `TempVar`, `ScopeFrame`, `TraitNeed`, `Emitter`, trait `EmitWalker`, `newEmitter`, and the helpers `reportUnsupported`, `writeLine`, `cString`, `literal`, `where`, `typeOf`, `refOf`, `localOf`, `localRef`, `switchBody`, `underlyingType`, `needsRc`, `retainCode`, `releaseCode`, `temp`, `ownedTemp`, `beginStatement`, `endStatement`, `consume`, `pushScope`, `popScope`, `addLocal`, `unwindAll`, `unwindLoop`, `hasWork`.
- `selfhost/src/emit.nra`: `emitProgram(checker, entry, tests): EmitResult` in Go's section order. The preamble (runtime, std runtimes in `checker.used` order, std/test only in test builds) is byte-equal to Go on every input. Until `programRecorded` is set, the Types and Functions sections start with the `program facts (selfhost-emit-program)` marker.
- `selfhost/src/runtimetext.nra` (generated): `RUNTIME_H`, `RUNTIME_MAP_H`, `ENV_RUNTIME`, `TEST_RUNNER_C`, `runtimeFor(path)`. Generator and staleness check: `TestSelfhostRuntimeText` in `selfhost_runtime_test.go` (`NARRA_UPDATE=1` rewrites).
- Stub modules with signatures: `hash.nra`, `cnames.nra` (selfhost-emit-program); `emitdecls`, `emitmaps`, `emitclosures`, `emittraits`, `emitmatch`, `emitstd`, `emitoptions`, `emitconsts`, `emitexprs`, `emitstmts`. Signatures are listed in `contracts/emitter.change-request.md` item 4 and in the files.
- `check.nra`: `ProgramFunc`, `Checker.programFunctions`, `programConsts`, `composites`, `programRecorded`, `LocalInfo.number`, `BodyInfo.instanceRefs`, `traitSelf`, and `findBody` (returns -1). Defaults only; none is filled.
- Harness `selfhost_emit_test.go`: `TestSelfhostEmit` runs 64 inputs, enforces U == S per input, compares the preamble fully and the next four parts up to the first marker, and logs skipped and compared counts per node. `TestSelfhostEmitErrors` checks the no-entry and unknown-binary messages against Go.
- Skip files `compiler/testdata/emit/skips/<node>.txt` for the nine nodes (inputs per node: program 63, decls 63, bodies 63, std 52, match 34, closures 14, traits 14, options 7, maps 2), written from a run of narrac. A file starts with a `#` comment line; entries are paths relative to `testdata/`.

## Deviations

- The contract needs a v2 (`contracts/emitter.change-request.md`): `cnames.nra` functions take `Checker`, not `Emitter` (import cycle); parameter names `function` and `func` renamed; `hasWork` and three probes (`reportMatchWork`, `reportStdWork`, `reportOptionsWork`) added. Orchestrator: please freeze v2 with these.
- `BodyInfo.instanceRefs`, `traitSelf` and `Checker.composites` have no `= Map {}` default (not a constant); `startBody` and `newChecker` set them.
- Edited `check.nra` (shared, fields plus `findBody` stub) and `main.nra` (the `emit` command).

## Stubs left

- All nine nodes' stubs. Section stubs report only when `hasWork` finds the feature (decls and bodies always do). Function prototypes, the C main and the recorded-mode core code in `emit.nra` (comment table, array typedefs and structs, type typedefs, C main) are not exercised until selfhost-emit-program sets `programRecorded`; they have never run.
- `hash.nra` and `cnames.nra` return placeholders.

## Gotchas

- Narra rejects `f(mutable x, x.field)` in one call: read fields into a local first. A `Map {}` is not a constant field default. `is` binds only as the whole `if` condition. `Maps.entries` gives tuples `#[key, value]`.
- Each node's stub removal must also shrink its skip file; the harness fails both ways (reported but not listed, listed but not reported).
- Comparison of later parts is by prefix up to the first marker. If a part has no marker, it must equal Go's. A node that makes a section real must make sure the next marker still appears where Go has content.
- `go test` caches ignore `selfhost/`; use `-count=1`.

## Proposed learnings

- Pass `Checker` (not `Emitter`) to pure C-name functions | `Emitter` lives in a module that calls them | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit|TestSelfhostRuntimeText|TestSelfhostPackageIsClean'`: pass (TestSelfhostEmit 64 inputs, 0 byte-identical, none fail to compile).
- `cd compiler && go vet ./... && go test -count=1 ./...`: see the PR description for the final result.
- `narra check selfhost`: clean.

## Size

- Estimated M, actual M (about 1,000 Narra lines including stubs, 330 Go lines).
