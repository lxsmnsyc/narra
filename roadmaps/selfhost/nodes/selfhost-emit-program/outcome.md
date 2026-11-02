# Outcome: selfhost-emit-program

## Built

- `selfhost/src/hash.nra`: `xxh32`, `hex8`, `relativeModule`, `moduleHash` (imports std only). `narrac hash <seed> <decimal text>` in `main.nra` prints `hex8(xxh32)`.
- `selfhost/src/cnames.nra`: the full contract (`typeId`, `cType`, `declName`, `internedName`, `payloadName`, `functionName`, `instanceName`, `implName`, `opFuncName`, `closureName`, `thunkName`, `constName`, `constGetterName`, `checkNameCollisions`). Extra export: `functionSignature(checker, program)` is Go's `signature()` (`static <result> <name>(<params>)`, mutable parameters get `*`, parameter names are `v_<name>_<number>` read from the body's first locals). `checkNameCollisions` checks program functions, `checker.order` and composites; nothing calls it yet (selfhost-emit-bodies or bootstrap may call it after naming).
- Checker facts (`check.nra` and the sites): `programFunctions` (PlainFunc in `declare.nra declareFunc`, ImplFunc in `traits.nra`, InstanceFunc in `instances.nra`), `programConsts` (in `exprs.nra checkConstType` at CONST_DONE), `LocalInfo.number` (`check.nra declareLocal`, `match.nra createPatternLocal`), `composites` via new `internComposite(mutable checker, type)` (resolve, substitute, funcvalues, closures, literals, stdcalls, stdlib sites), `Checker.bodyKeys` plus `keyBody(mutable checker, key)` and `findBody`.
- `findBody` keys: `func|<funcIndex>` (non-generic only), `inst|<instance>`, `impl|<block>|<index>`, `test|<module>|<n-th test of the module>`, `const|<constIndex>`, `default|<module>.<name>|<field>` (module omitted in the prelude; a generic struct shares one body).
- `emit.nra`: real function prototypes (`emitPrototypes`), composite lists skip types with type parameters, the program-facts marker is gone. `programRecorded` defaults to true.
- Module hash clash check in `modules.nra loadPackage`; fixture `compiler/testdata/packages/errors/module_hash_clash/` (`maqoucq` and `mqddgyw` collide). Go and narrac report the same error (`TestSelfhostModules`).
- `TestSelfhostXXH32` in `selfhost_emit_test.go`. The skip file `selfhost-emit-program.txt` is deleted.

## Deviations

- `emit.nra` (not in the brief's lists) was edited: prototypes, type-parameter filter, removal of the facts marker, and the match/std/options probes now run unconditionally (they ran only while `programRecorded` was false, and would have gone silent once it is true; the bodies node deletes the calls when it lands).
- `selfhost_emit_test.go`: besides `TestSelfhostXXH32`, `readEmitSkips` now ignores a missing skip file (needed to delete this node's file). Other nodes deleting their files may make the same one-hunk edit; keep one copy on merge.
- `body.nra`: `checkFunctionBody` and `checkFieldDefaults` gained a key parameter (internal). `main.nra`: `hash` command.
- Effect on other nodes' stubs: none. `TestSelfhostEmit` still reports the same skip counts for the other eight nodes.

## Stubs left

- None of this node's. The other eight nodes' stubs remain. `Emitter.typeNames` stays unused (no typeId cache was needed).

## Gotchas

- `internComposite` records only composites without a type parameter, as Go drops them in finish. emit.nra still filters, harmlessly.
- The Go sites that make types without recording (e.g. `Test.panics` building `() => void`) are not recorded, as in Go.
- `ProgramFunc.ImplFunc` module is the implement block's module (`scopes[impl.scope].path`), as Go sets `fd.Module`.
- Narra rejects `f(mutable x, g(x))`; bind the inner call first.
- Short (7 byte) names find xxh32 collisions fast; sequential long names did not.

## Proposed learnings

- Record checker facts for the emitter at the Go call site | a type walk afterwards misses types made in unused signatures | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit|TestSelfhostXXH32|TestSelfhostModules|TestSelfhostBodies|TestSelfhostEffects|TestXXH32'`: pass. TestSelfhostEmit: 63 inputs, program node reports on none, comment table, typedefs and `// Functions` match Go on all.
- `narra check selfhost`: clean.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (internal/narra 208 s).

## Merge notes

- `cnames.nra` exports `functionSignature(checker, program)`; selfhost-emit-bodies (PR #48) exports `functionSignature(emitter, program)` from emitconsts.nra and also edits the prototype section of emit.nra. On merge keep one: emit.nra calls bodies' version, or bodies' version delegates to cnames. Mine reads parameter names from `LocalInfo.number` of the body found through `bodyKeys`.
- Test bodies are keyed `test|<module>|<n>` here; bodies walks `checker.bodies` instead. Both work.

## Size

- Estimated M, actual M.
