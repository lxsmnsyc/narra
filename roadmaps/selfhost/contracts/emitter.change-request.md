# Change request: emitter v1 to v2

- From node: selfhost-emit-core
- Status: open

## Problem

1. The contract has the C-name functions in `cnames.nra` take an `Emitter` (`typeId(emitter, type)`, `cType`, `payloadName`, `functionName`), and it lists `cnames.nra` below `emitstate.nra`'s helpers. But `Emitter` is defined in `emitstate.nra`, and `emitstate.nra` calls `cType` and `opFuncName` (`retainCode`, `releaseCode`, `temp`). Narra modules cannot import in a cycle, so one of the two cannot hold.
2. `implName(typeName, traitModule, traitName, function)` and `functionName(emitter, func: ...)` use `function` and `func` as parameter names; `function` is a keyword.
3. The contract says nothing about how to find work for a stub. The stubs need to know whether an input has a Map, a function type and so on, before `selfhost-emit-program` records the facts.

## Proposed change

1. `cnames.nra` functions that took `emitter: Emitter` take `checker: Checker` instead: `typeId(checker, type)`, `cType(checker, type)`, `payloadName(checker, enumType)`, `functionName(checker, program)`, `checkNameCollisions(checker)`. The typeId cache moves into the Checker (a `typeNames` map filled by selfhost-emit-program), or is dropped, since names are pure. `Emitter.typeNames` stays unused until then.
2. Rename the parameters: `implName(typeName, traitModule, traitName, member)`, `functionName(checker, program: ProgramFunc)`.
3. Add to `emitstate.nra`: `hasWork(emitter, node): bool` (scans the checker's body tables for the feature of a node) and one probe per node that has no section function: `reportMatchWork` (emitmatch), `reportStdWork` (emitstd), `reportOptionsWork` (emitoptions). Each node deletes its probe and `hasWork` arm when its emitter lands.
4. The section stubs are `emitHeapTypedefs`, `emitTypeDefinitions`, `emitHeapDefinitions`, `emitBoxPrototypes`, `emitRcFunctions` (decls, all reference counting prototypes and functions in Go's order), `emitMapTypedefs`, `emitMapStructs`, `emitMapPrototypes`, `emitMapFunctions` (maps), `emitFuncTypeDefinitions`, `emitClosureFunctions` (closures), `emitTraitFunctions` (traits), `emitConstPrototypes`, `emitConstBodies`, `emitFunction`, `emitTestBody` (bodies), `emitTestMain` (options). They take `mutable emitter` and return the C text as a `string`.

## Affected nodes

selfhost-emit-program (cnames, hash), and every selfhost-emit-* node that calls `cType` or `typeId`.

## Workaround in place

The code already follows the proposed change; `selfhost/src/cnames.nra` is a stub with the new signatures.

## Decision

