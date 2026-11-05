# Outcome: generators-check

## Built

The front half of generators, in the Go compiler and in `selfhost/` together. Nothing of it existed on main before.

- Syntax. `yield` is a keyword (`lexer.go`, `selfhost/src/lexer.nra`). `function*` sets `FuncDecl.Generator` (`generator` in `selfhost/src/ast.nra`). `yield value` and `yield* other` parse to `YieldStmt{Pos, Value, Delegate}` (`Stmt.YieldStmt(pos, value, delegate)`). A `Generator` type expression reads a `uses` list after its type arguments (`TypeExpr.Uses`, `effects` in selfhost). AST dump: `func ... generator`, `yield L:C delegate` with the value under it (`astdump.go` doc comment). Clone, astdump, typesdump, idcheck and every statement walker of both compilers have the new cases.
- Type. Go `GeneratorType{Elem, Effects, Binder, Origin}` (`types.go`); selfhost `Type.TGenerator(element, effects, binder, origin)` (`types.nra`), where binder and origin are `Checker.functions` indexes or -1. Effects are sorted, deduplicated effect names: the same `[]string` form as a function type's `TypeExpr.Uses`, effect variables included. Binder is the function whose `<effect E>` the list names. Origin is set only on the result type of a private `function*`: its effects are inferred, and the effect pass fills `FuncDecl.GenEffects` (Go) or `Checker.generatorEffects[generatorKey]` (selfhost). `GeneratorType.effects()` / `generatorEffectList` read the right one. Printed as `Generator<T> uses a, b` (decls and types dumps too). `Generator<T>` has no `Equatable`, `Hashable` or `Show`, and is not a map key ("a Generator is not hashable").
- Declarations: `resolveSignature` (Go `check.go`, selfhost `resolve.nra`) is used for declarations and generic instances. It checks rule 1 (E0195), rule 4 (E0196, E0615), the `uses` list written after a generator's return type (E0195), and records the emitter stub.
- Statements (`check_stmt.go` `yieldStmt`, selfhost `stmts.nra` `checkYield`): rule 2 placement (E0291: outside a generator, inside an arrow function, inside a block expression; a `yield` in a conditional or any expression is the parser error E0046), `yield` value type (E0119), `yield*` operand (E0291 second format), rule 3 (E0292). `for ... of` over a `Generator<T>` gives `T` (Go `forKind` `forGenerator`); E0232's text now lists "a Generator".
- Coercions (`expectType` -> `expectGenerator`, Go `check_types.go`, selfhost `exprs.nra`): a generator value may have fewer effects than the expected type, not more (E0119 "expected Generator<int>, found Generator<int> uses io"). An expected type whose Binder is another function than the one being checked is a call parameter whose variables bind, so it is accepted. When either side has an Origin or a Bound the check is queued (`generatorUses`) and runs after the effect pass (`checkGeneratorUses`), in queue order, with the file recorded at the use.
- Effects (`check_effects.go`, `selfhost/src/effects.nra`): a call of a `function*` charges only `alloc` (`funcEffects`). `for ... of` over a generator and `yield* g` charge the generator type's effects, via the type string (`(through \`Generator<int> uses io\`)`). An exported `function*` body is checked against its return type's `uses` (E0800). A private one's body set, minus its own effect variables, becomes its type's effects. A generator-typed parameter binds an effect variable at the call (`userCall`, helper `boundEffects`; selfhost `findBoundEffects`). A call whose callee's result type is a generator naming the callee's effect variables gets its own result type (`bindCallResult`: Go `GeneratorType.Bound`, selfhost a TGenerator origin of `-2 - k` into `Checker.boundGenerators`), and the effect pass fills it with the effects the arguments bind (`bindResultEffects` in selfhost), only at the call that made it (`boundEffects.call`, selfhost `Checker.boundCallIds`): a generic `id(x)` returns the same type and must not refill it. So `E` never leaves the callee: iterating `relay(noisy())` charges `io`, never the name `E` (found in self-review). `checkSignature` also scans the result type for unknown effects. E0802 now says "function or generator type of a parameter".
- Emitter stub: the checker stores the first `function*` as `Program.Unemitted` (selfhost `Checker.unemitted`). `CompileBinary` (`driver.go`) and `narrac emit` (`selfhost/src/emitter.nra`) stop with "generators are not compiled yet" (E0022, reworded from the old parser error). Each is one marked line (`// node generators removes this line` / `... check`).
- Tests: `compiler/testdata/astdump/generators.nra` + `.ast`; `compiler/testdata/typesdump/generators/basics.nra`, `shapes.nra`, `bound.nra` + `.dump` (golden test `TestTypesDumpGenerators` in the new `generators_test.go`); 37 error tests `compiler/testdata/errors/generator_*.nra`; `TestGeneratorEmitStub` (same file) checks the emitter stub on `typesdump/generators/basics.nra`. TestSelfhostParser, TestSelfhostBodies, TestSelfhostImports and TestSelfhostLexer compare narrac with Go on all of them.
- Docs: SPEC.md section 7 (generator types carry effects; generator parameters bind effect variables) and section 8 (rules 1 to 8). DECISIONS.md D78, and a line under D14. `compiler/README.md`: what is checked but not compiled, and the differences below.

Codes used: E0046 (3 parser formats), E0195 (3), E0196, E0291 (2), E0292, E0615. Decision D78.

## Deviations

- Files outside the brief's paths, each a small change: `compiler/internal/narra/driver.go` (stub line), `cnames.go` (`typeID` case for `GeneratorType`, so `Array<Generator<T>>` keys differ), `decldump.go` (`dumpTypeString` case), `std_map.go` (map key problem), `astdump_test.go` (`"yield"` in `astKinds`), `selfhost_effects_test.go` (coverage string follows the E0802 rewording), new `generators_test.go` (golden test for `typesdump/generators`), `compiler/testdata/errors/effects_var_unbound.nra` and `for_of_sources_int.nra` and `typesdump/effects/err_signature.dump` (reworded messages), `compiler/README.md`.
- The emit-stub test is `TestGeneratorEmitStub`, not a `testdata/errors/generator_*.nra` file: TestErrorCodes requires every file there to give a checker diagnostic, and a program with a `function*` checks clean.
- E0022 is reworded from "generators are not supported by this compiler yet" (parser) to "generators are not compiled yet" (emitter stub), keeping its code, instead of a new code.
- Rule 2's "inside a ternary" is a parse error (E0046 "`yield` is a statement, so it cannot be part of an expression"), because `yield` is a statement and never parses as an expression.
- Cost recorded in D78 and README: a non-generator function returning a generator type writes its own `uses` after the type's list; a private function's `Generator<T>` parameter accepts only pure generators (it is not an implicit effect variable, as a private function-type parameter is).

## Stubs left

- The emitter stub (both compilers), removed by node `generators`: delete the marked line in `driver.go`, the marked `if` in `selfhost/src/emitter.nra` and `TestGeneratorEmitStub`; `Program.Unemitted` / `Checker.unemitted` and E0022 can then go too.
- `GeneratorType.CName()` returns the placeholder `narra_generator`; selfhost `cnames.nra` gives TGenerator the default `x`. Node `generators` names the frame types.
- `Coroutine<In, Out>` and `Co.resume` (node coroutines): `Coroutine` as a generator return type is E0195 today.

## Gotchas

Facts the emitter (node `generators`) needs:

- Which statements contain a `yield`: none are marked. Walk the body: a `*YieldStmt` (selfhost `Stmt.YieldStmt`) can only appear in a `function*` body's statement tree (blocks, `if`, `while`, `for`, labeled loops, `match` statement arm blocks), never inside an expression, a block expression or an arrow function: the checker rejects every other place. So "contains a yield" is a plain statement walk that does not enter expressions.
- Element type: `fd.ResultT.(*GeneratorType).Elem` (selfhost: `checker.functions[i].result` is `TGenerator(element, ...)`, an instance's `FuncInstance.result`). A function* with a bad return type never reaches the emitter (E0195).
- for-of source kind: Go sets `ForStmt.Kind = forGenerator`; selfhost has no kind field, the iterable's node type is `TGenerator(...)` (`emitstmts.nra` dispatches on the iterable type, as for `TRange`).
- `yield*` operand and for-of over a generator iterate a copy (SPEC section 8). `return` in a generator never has a value.
- Rule 8 (self calls are not recursion, no tail loop): the checker does nothing; the emitter's recursion and tail-call detection must skip calls of generator functions.
- Generator effects for node `typed-calls-and-effects`: the representation is the `[]string` effect list on the type plus `Binder` (whose variables they are), `Origin` (inferred from a private body) and `Bound` (filled per call). `effectState.generatorEffects(t)` is the set a value has when it runs. A function type could carry the same three fields.
- `sameType` compares generator effects exactly (and Binder, Origin); only `expectType` allows a subset. Comparisons that do not go through `expectType` (none known today besides `==`, which is rejected) see different types.
- Selfhost types are values compared with `==`, so `TGenerator` equality is structural over the same four fields.
- `c.sigFunc` (Go) / `checker.functionIndex` (selfhost) is the function whose signature or body is checked; it is saved and restored around instance resolution and body checks. Impl bodies use their own FuncDecl in Go and -1 in selfhost; neither can bind an effect variable.
- Bound call results in generic instance bodies are never filled (the effect pass walks only the open bodies), so they print without effects there. Both compilers agree.
- Every new compile error must match narrac byte for byte: TestSelfhostBodies runs `narrac types` on every `testdata/**/*.nra`, error tests included.

## Proposed learnings

- A new `Type` variant needs a case in every exhaustive match and in the `_` defaults that list the non-trait types (`computeImplements`, `mapKeyProblem` in `selfhost/src/traits.nra`) | the default arm silently treated `TGenerator` as a primitive that implements Equatable | selfhost/src | yes (a generator `==` error test caught it)

## Verification

- `cd compiler && go vet ./... && go test -timeout 40m ./...`: pass, TestSelfhostBootstrap included. The default 10 minute timeout ran out once on this machine while other worktrees ran their suites.
- `narra types|parse|decls` against `narrac types|parse|decls` on all 37 error tests, the three typesdump cases and the AST case: 0 differences of 123.

## Size

- Estimated L, actual L.
