# Outcome: selfhost-emit-bodies

## Built

- `selfhost/src/emitexprs.nra`: `emitExpression` (one arm per expression kind), literals, templates, name and field reads, index reads, unary and binary operators, calls (user functions, generic instances through `instanceRefs`, impl functions, enum constructors, distinct calls), the prelude (`print`, `assert`, `panic`), `Array.length/push/filled`, `Box.new`, struct and array literals with field defaults emitted inside `switchBody(default body)`, `emitLvaluePointer`, `argumentCode`, `constRef` (static constants re-emitted inside `switchBody(const body)`, others through the getter), `isStaticConstant`, `constType`, `pointerTemp`, `declaredLocal`. Dispatch arms to feature modules are one line each, with the contract's signatures.
- `selfhost/src/emitstmts.nra`: `emitStatement`, `emitBlock`, `emitBlockInScope`; var, assign (with release of the old value), expression, return, if, while, for over ranges and arrays, break and continue. `beginStatement`/`endStatement` surround the calls into `emitMapSet`, `emitMatchStatement` and `emitMapFor`, as Go does, so the feature modules must not open the frame again.
- `selfhost/src/emitconsts.nra`: `emitConstPrototypes`, `emitConstBodies` (getters and `narra_consts_release`), `emitFunction`, `emitTestBody`, `functionSignature` (exported, for the prototype section of emit.nra). A `BodyWalker` forwards to the expression and statement modules.

## Deviations

- Not run against real C: `Checker.programRecorded` is still false (selfhost-emit-program), so `programFunctions`, `programConsts`, `findBody`, C names and `LocalInfo.number` are placeholders. The code type-checks and follows Go line by line; no body has been compared byte for byte.
- While `programRecorded` is false, `emitConstPrototypes` and `emitConstBodies` report their old markers so the skip file stays valid. Once program records the facts they stop reporting by themselves.
- `constRef` keeps the stub signature (no walker). It uses a private `ConstWalker` (expressions only) defined in emitexprs.nra, because static constants only hold literals, operators and constructors.
- `findBody` keys I rely on: `func|i`, `inst|n`, `impl|b|i`, `const|i`, and `default|<module>.<name>|<field>` (module omitted for the prelude; the generic's declared name for instances), matching the body names the checker gives. Tests are found by walking `checker.bodies` for `kind == "test"`, not by key. selfhost-emit-program must make `findBody` agree.
- `emitter.file` comes from `ModScope.file` of the body's module.
- Skip file `compiler/testdata/emit/skips/selfhost-emit-bodies.txt` is unchanged (63 inputs): it can only shrink once program and decls land.

## Stubs left

- `Range.upTo/between` outside a `for` and unresolved refs report `unsupported (selfhost-emit-bodies)`.
- Everything the other emit nodes own is still their stub: closures, maps, traits, match, std, options, decls.

## Gotchas

- `test` is a keyword; Narra rejects `f(mutable emitter, g(emitter...))` even inside a template argument, so read tables into locals first.
- Acceptance "skip file deleted" needs the program and decls nodes merged first; the orchestrator should rerun this node's check after them and delete the file then.
- Each `Ref.Function` call to a generic function needs `BodyInfo.instanceRefs[callId]`; without it the call reports unsupported.

## Proposed learnings

- None.

## Verification

- `cd selfhost && narra check .`: clean.
- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit|TestSelfhostRuntimeText|TestSelfhostPackageIsClean'`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (go vet clean; narra/internal/narra ok in 234 s).

## Size

- Estimated M, actual M (about 1,000 Narra lines).
