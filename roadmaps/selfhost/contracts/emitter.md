---
name: emitter
version: 2
status: frozen
defined_by: selfhost-emit-core
---

# Contract: emitter

## Purpose

How the Narra C emitter (selfhost/src/emit*.nra) reads the checker's results, names C, holds its state, and which module emits which part of the C. Used by every selfhost-emit-* node and by selfhost-bootstrap.

## Interface

### Checker facts (check.nra; fields added by selfhost-emit-core, filled by selfhost-emit-program)

    export enum ProgramFunc {
      PlainFunc(funcIndex: int)            // Checker.functions, non-generic, at declaration (Go check.go:279)
      InstanceFunc(instance: int)          // Checker.instanceList, at addFuncInstance (check_call.go:475)
      ImplFunc(block: int, index: int)     // Checker.implementations[block].functions[index] (check_traits.go:429)
    }
    Checker.programFunctions: Array<ProgramFunc> = []   // Go's Program.Funcs order
    Checker.programConsts: Array<int> = []              // Checker.consts, in the order each reaches CONST_DONE
    Checker.composites: Map<string, Type> = Map {}      // by typeKey: every TArray, TBox, TMap, TFunc made where Go calls arrayOf/boxOf/mapOf/funcOf
    Checker.programRecorded: bool = false               // set by selfhost-emit-program; until then emit.nra reports "unsupported: program facts (selfhost-emit-program)"
    LocalInfo.number: int = 0                           // checker.localCount right after the increment; C name v_<name>_<number>
    BodyInfo.instanceRefs: Map<int, int>                // node id of a call, Ident or FieldExpr naming a generic function -> instanceList index
    BodyInfo.traitSelf: Map<int, Type>                  // CallExpr id with Ref.TraitFunction -> Self type
    export function findBody(checker: Checker, key: string): int   // BodyInfo index or -1
      // keys: "func|<funcIndex>", "inst|<instance>", "impl|<block>|<index>", "test|<module>|<n-th test of the module>",
      //       "const|<constIndex>", "default|<struct or generic key>|<field>" (a generic struct has one default body shared by its instances)

The emitter never mutates the Checker. It holds it read-only in Emitter.checker, because a write would copy every table.

### C names (cnames.nra, hash.nra; selfhost-emit-program)

    hash.nra (imports std only; modules.nra uses it):
      export function xxh32(text: string, seed: int): int        // 0 <= result < 2^32
      export function hex8(value: int): string
      export function relativeModule(path: string): string
      export function moduleHash(path: string): string
    cnames.nra (imports types, check, hash):
      export function typeId(checker: Checker, type: Type): string      // Go typeID
      export function cType(checker: Checker, type: Type): string       // Go CName(): adds "*" for arrays, boxes, maps; "void*" for TParam
      export function declName(kind: string, module: string, name: string): string
      export function internedName(kind: string, head: string, arguments: Array<string>): string
      export function payloadName(checker: Checker, enumType: Type): string
      export function functionName(checker: Checker, program: ProgramFunc): string   // f_<hash>_<name>, instance and impl names included
      export function instanceName(module: string, name: string, typeIds: Array<string>): string
      export function implName(typeName: string, traitModule: string, traitName: string, member: string): string
      export function opFuncName(operation: string, typeIdText: string): string
      export function closureName(kind: string, owner: string, count: int): string
      export function thunkName(owner: string): string
      export function constName(module: string, name: string): string          // G_...
      export function constGetterName(module: string, name: string): string    // g_...
      export function checkNameCollisions(checker: Checker): Option<string>   // runs once after naming; Some(message) is an internal error

Names are pure functions of the Checker. cnames.nra takes a Checker, not an Emitter, because emitstate.nra defines Emitter and imports cnames.nra for cType and opFuncName. Go registers each name lazily in a process-wide table. Here one pass over all named things replaces that, which gives the same output for every valid program. A typeId cache, if needed, is a Checker field filled by selfhost-emit-program; Emitter.typeNames stays unused until then and may be removed.

### Emitter state (emitstate.nra; selfhost-emit-core)

    export struct Code { text: string, owned: bool }        // Go's (string, bool)
    export struct TempVar { name: string, tempType: Type }
    export struct ScopeFrame { locals: Array<int>, loop: bool }  // BodyInfo.locals indexes that need release
    export struct TraitNeed { operation: string, needType: Type, name: string }
    export struct Emitter {
      checker: Checker
      typeNames: Map<string, string>
      entry: int; tests: bool
      file: string                          // Program.Files[module]: the module's FIRST file, as in Go
      body: int                             // BodyInfo index being emitted
      owner: string                         // C name of the declared function or test (closure naming)
      resultType: Type
      literals: StringBuilder; literalCount: int
      out: StringBuilder; indent: int; tempCount: int   // tempCount is program-wide, never reset (as in Go)
      frames: Array<Array<TempVar>>; scopes: Array<ScopeFrame>; loopFrames: Array<int>
      traitNeeds: Array<TraitNeed>; traitSeen: Map<string, bool>                       // selfhost-emit-traits
      closureCounts: Map<string, int>; thunks: Map<string, bool>
      closureEnvs: StringBuilder; closurePrototypes: StringBuilder; closureDefinitions: StringBuilder   // selfhost-emit-closures
      unsupported: Array<string>
    }
    export trait EmitWalker {
      function emitExpression(walker: Self, mutable emitter: Emitter, expression: Expr): Code uses alloc
      function emitStatement(walker: Self, mutable emitter: Emitter, statement: Stmt) uses alloc
      function emitBlock(walker: Self, mutable emitter: Emitter, block: Block, loop: bool) uses alloc
      function emitBlockInScope(walker: Self, mutable emitter: Emitter, block: Block) uses alloc
      function emitLvaluePointer(walker: Self, mutable emitter: Emitter, expression: Expr): string uses alloc
    }
    helpers (Go names in brackets): writeLine [line], where, temp, ownedTemp, consume, beginStatement, endStatement,
      pushScope, popScope, addLocal, unwindAll, unwindLoop, localRef, literal, cString, retainCode [retain], releaseCode [release],
      needsRc, underlyingType [underlying], typeOf(emitter, id), refOf(emitter, id), localOf(emitter, index),
      switchBody(mutable emitter, body): int (returns the saved body), reportUnsupported(mutable emitter, construct, node): string
    stub probes (temporary): hasWork(emitter, node): bool scans the checker's body tables for a node's feature;
      reportMatchWork (match), reportStdWork (std), reportOptionsWork (options) report for nodes with no section function.
      Each node deletes its probe and its hasWork arm when its emitter lands.

emitter.nra (the top module, core) has `struct EmitterWalker implements EmitWalker {}`, which forwards to the bodies modules. Modules that need recursion take `<W: EmitWalker>` and pass `walker` on. This is the same pattern as the checker's Walker (D70 outcome), because Narra function types cannot take `mutable` and modules cannot import in a cycle.

### Module layout, imports bottom to top, and owners

    hash.nra, cnames.nra ................................ selfhost-emit-program
    runtimetext.nra (generated) ......................... selfhost-emit-core
    emitstate.nra ....................................... selfhost-emit-core
    emitdecls.nra ....................................... selfhost-emit-decls
    emittraits.nra ...................................... selfhost-emit-traits
    emitclosures.nra .................................... selfhost-emit-closures
    emitmaps.nra (imports emittraits) ................... selfhost-emit-maps
    emitmatch.nra (may import emittraits) ............... selfhost-emit-match
    emitstd.nra ......................................... selfhost-emit-std
    emitoptions.nra (imports emitclosures) .............. selfhost-emit-options
    emitconsts.nra, emitexprs.nra, emitstmts.nra ........ selfhost-emit-bodies
    emit.nra, emitter.nra ............................... selfhost-emit-core

selfhost-emit-core creates every file with the signatures below as stubs. Each stub calls reportUnsupported only when it has work (hasWork): a type, a map, a function type or an expression to emit. The section stubs emitHeapTypedefs, emitTypeDefinitions, emitHeapDefinitions, emitBoxPrototypes, emitRcFunctions (decls), emitMapTypedefs, emitMapStructs, emitMapPrototypes, emitMapFunctions (maps), emitFuncTypeDefinitions, emitClosureFunctions (closures), emitTraitFunctions (traits), emitConstPrototypes, emitConstBodies, emitFunction, emitTestBody (bodies) and emitTestMain (options) take `mutable emitter` and return the C text as a string.

### Dispatch points (section order = Go's Emit, emit.go:46-160)

    preamble: RUNTIME_H, then runtimeFor(path) for each checker.used module (std/test only in test builds) ....... core
    "\n// Types\n": comment table ......................................................................... core
      array typedefs ...................................................................................... core
      emitMapTypedefs(mutable emitter): string ............................................................ maps
      type typedefs ....................................................................................... core
      emitHeapTypedefs ................................................................................... decls
      emitFuncTypeDefinitions (envRuntime + F structs, only when function types exist) ................... closures
      emitTypeDefinition per type, emitArrayStructs ....................................................... decls
      emitMapStructs ...................................................................................... maps
      emitHeapDefinitions ................................................................................ decls
    "\n// Reference counting\n": emitBoxPrototypes (decls), emitMapPrototypes (maps), emitRcPrototypes, emitRcFunctions,
      emitArrayFunctions, emitBoxFunctions (decls)
    emitMapFunctions: runs HERE, before constants, because it registers trait needs; its text goes after the trait functions ........ maps
    "\n// Constants\n": emitConstPrototypes ................................................................ bodies
    "\n// Functions\n": prototypes of programFunctions .................................................... core
    emitConstBodies, emitFunction(func) per programFunctions, emitTestBody(n) per test in test builds ...... bodies
    "\n// String literals\n" + literal pool ................................................................ core (literal() helper)
    emitTraitFunctions ..................................................................................... traits
    map function text ...................................................................................... maps
    emitClosureFunctions ("\n// Closures\n") ............................................................... closures
    "\n" + bodies text; then emitTestMain (options) in test builds, else the C main (core)

    Expressions (emitexprs.nra, bodies), one arm per kind:
      Ident/FieldExpr function value -> emitFuncValue (closures); ArrowExpr -> emitArrow (closures)
      IsExpr -> emitIsCondition, MatchExpr -> emitMatchValue, StructLit with spread -> emitSpreadLiteral,
        ArrayLit of a tuple -> emitTupleLiteral (match)
      MapLit -> emitMapLiteral, IndexExpr on a TMap -> emitMapGet (maps)
      BinaryExpr == / != on struct, enum or array -> emitEquality (traits)
      CallExpr: Ref.TraitFunction -> emitTraitCall (traits); Ref.Value -> emitValueCall (closures);
        Ref.Builtin by StdFunc.module: "" (print, assert, panic), std/array, std/range, std/box -> bodies;
        std/map -> emitMapBuiltin (maps); std/option, std/result, std/test -> emitOptionCall (options);
        any other std module -> emitStdCall (std)
    Statements (emitstmts.nra, bodies): VarStmt with pattern -> emitDestructure, IfStmt on IsExpr -> emitIfIs,
      ExprStmt of MatchExpr -> emitMatchStatement (match); AssignStmt to a Map index -> emitMapSet,
      for over Map.keys / Map.entries -> emitMapFor (maps)
    Shared helpers other nodes call: enumAllocation(emitter, enumType, variant, fields) (decls);
      useEq/useHash/useCmp/useShow (traits); callFunctionValue (closures); argumentCode, constRef (bodies)

### Driver (main.nra, emitter.nra; core)

    narrac emit [--tests] [--bin name] <path>
      exit 0: the C on stdout
      exit 1: load, check or effect errors (narrac "error ..." lines; matching Go's text is selfhost-bootstrap's job),
              or Go's CompileBinary messages for a missing or unknown binary or entry function
      exit 3: C with "/* unsupported: <construct> (<node>) */" markers on stdout, and one
              "unsupported: <construct> (<node>)" line per distinct construct on stderr, in first-seen order

## Behavior

- Output is byte-identical to Go's `CompileBinary(path, bin, tests)` for every input that compiles.
- Never iterate a Map where Go iterates a slice. Composite lists are sorted by Go's key string with a byte-order merge sort.
- A static constant's value is emitted at each use inside switchBody(const body). A struct field default is emitted inside switchBody(default body).
- New per-function state goes into the Emitter, and emitClosureFunctions saves and restores it (CLAUDE.md rule for closureFunc).
- Loop over statements and declarations. Recurse only on nesting.

## Stub

selfhost-emit-core ships every module above with the frozen signatures. The stubs return `Code { text: "0 /* unsupported */", owned: false }` or a marker line, and call reportUnsupported.

## Contract tests

- `TestSelfhostEmit` (compiler/internal/narra/selfhost_emit_test.go). Inputs: testdata/*.nra (test mode for tests.nra and std_test*), testdata/panics/*.nra, testdata/packages/* except errors/, testdata/runner/mixed.nra and packages/runner (test mode). selfhost-emit adds the selfhost package.
  - U = the nodes in narrac's unsupported lines. S = the nodes whose file compiler/testdata/emit/skips/<node>.txt lists the input. U must equal S.
  - If U is empty: exit 0 and full byte equality.
  - Otherwise: split both outputs at the headers "\n// Types\n", "\n// Reference counting\n", "\n// Constants\n", "\n// Functions\n", "\n// String literals\n". Each of the first five parts must agree up to the first marker line in it. The last part is not compared, because a stub shifts the literal and temp counters.
- `TestSelfhostRuntimeText`: runtimetext.nra equals what the generator builds from Go's strings. `NARRA_UPDATE=1` rewrites it.
- `TestSelfhostXXH32`: `narrac hash <seed> <text>` against the TestXXH32 vectors.

## Changes

- v1: initial.
- v2 (2026-11-02, from selfhost-emit-core's change request): cnames.nra takes Checker instead of Emitter (import cycle); parameters `function`/`func` renamed to `member`/`program` (keyword); stub probes hasWork, reportMatchWork, reportStdWork, reportOptionsWork; section stub names and signature listed.
