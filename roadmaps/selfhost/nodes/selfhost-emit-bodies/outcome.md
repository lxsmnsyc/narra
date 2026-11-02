# Outcome: selfhost-emit-bodies

## Built

- `selfhost/src/emitexprs.nra`: `emitExpression` (one arm per expression kind), literals, templates, name and field reads, index reads, unary and binary operators, calls (user functions, generic instances through `instanceRefs`, impl functions, enum constructors, distinct calls), the prelude (`print`, `assert`, `panic`), `Array.length/push/filled`, `Box.new`, struct and array literals with field defaults emitted inside `switchBody(default body)`, `emitLvaluePointer`, `argumentCode`, `constRef` (static constants re-emitted inside `switchBody(const body)`, others through the getter), `isStaticConstant`, `constType`, `pointerTemp`, `declaredLocal`. Dispatch arms to feature modules are one line each, with the contract's signatures.
- `selfhost/src/emitstmts.nra`: `emitStatement`, `emitBlock`, `emitBlockInScope`; var, assign (releases the old value), expression, return, if, while, for over ranges and arrays, break and continue. `beginStatement`/`endStatement` surround the calls into `emitMapSet`, `emitMatchStatement` and `emitMapFor`, as Go does, so those modules must not open the frame again.
- `selfhost/src/emitconsts.nra`: `emitConstPrototypes`, `emitConstBodies` (getters and `narra_consts_release`), `emitFunction`, `emitTestBody(emitter, number, module, moduleNumber)`. Prototypes and signatures come from `cnames.functionSignature`. A `BodyWalker` forwards to the expression and statement modules.
- The skip file `selfhost-emit-bodies.txt` is deleted (it listed 63 inputs; now no input reports this node).

## Deviations

- `emitTestBody` has two more parameters (`module`, `moduleNumber`) so the test is found by the key `test|<module>|<n>`; `emit.nra` `emitTestBodies` computes them (shared edit).
- `emitstate.nra`: `localName(local)` is now exported and used by `localRef` and by emitexprs.nra (one helper for `v_<name>_<number>`).
- `calls.nra` and `funcvalues.nra`: after `addFuncInstance`, the call records `checker.bodies[checker.body].instanceRefs[nodeId] = instance`. The contract gives this field to selfhost-emit-program, but nothing filled it, so every generic call reported "generic call without an instance".
- `compiler/testdata/emit/skips/selfhost-emit-match.txt` gains `Tuple.nra`, `patterns_tuple_shapes.nra` and `literal_type_args.nra`: with real expression dispatch, their tuple literals now reach the match stub, which reports.
- The probe calls `reportMatchWork`, `reportStdWork` and `reportOptionsWork` stay in `emit.nra`; each belongs to its own node.
- `constRef` keeps the stub signature (no walker) and uses a private expressions-only `ConstWalker`.
- `findBody` keys: `default|<module>.<name>|<field>` for field defaults (module omitted for the prelude).

## Stubs left

- `Range.upTo/between` outside a `for`, and unresolved refs, report `unsupported (selfhost-emit-bodies)`.
- Diffs from Go that remain belong to other nodes: `selfhost-emit-decls` (heap typedefs, type definitions, enum allocation, rc functions; all 60 inputs), and the stubs of match, closures, traits, std, maps and options. No input is byte-identical yet (0 of 60), so function bodies have not been compared with Go; the harness compares only the first marker-free prefix of each section.

## Gotchas

- `test` is a keyword; Narra rejects `f(mutable emitter, g(emitter...))` even inside a template argument, so read tables into locals first.
- Once decls and the feature nodes land, run `TestSelfhostEmit` again: the first full byte comparison of function bodies happens then.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (go vet clean; narra/internal/narra ok in 193 s).

## Size

- Estimated M, actual M (about 1,000 Narra lines).
