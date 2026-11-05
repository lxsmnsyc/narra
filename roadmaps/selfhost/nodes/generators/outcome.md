# Outcome: generators

## Built

Generators compile to C state machines in the Go compiler and in `selfhost/`, and the two emit byte-identical C for every new sample.

- Runtime (`compiler/internal/narra/runtime.h`, mirrored in `selfhost/src/runtimetext.nra` by `TestSelfhostRuntimeText`): `narra_gen { rc, state, resume, clone, drop }`, `narra_gen_retain`, `narra_gen_release` (drops through the `narra_drop` worklist), and `narra_gen_next(narra_gen** p, void* out)`. `narra_gen_next` clones a shared generator (rc > 1) before it resumes it.
- Value: every `Generator<T>` has the C type `narra_gen*` (`GeneratorType.CName()`, selfhost `cType`). `needsRC`/`needsRc`, `retain`/`release` and `refCountCode` handle it.
- Lowering: `compiler/internal/narra/emit_generator.go` and `selfhost/src/emitgenerators.nra`. Each `function*` (and each generic instance) emits a frame struct `Y<owner>`, a resume function `y<owner>`, a clone function `j<owner>`, a drop function `z<owner>` and the constructor (the declared function). The names come from `generatorCName` (Go `cnames.go`) and `generatorPart` (selfhost `cnames.nra`).
- In a generator body every C variable that can be held at a `yield` is a frame field: temporaries, Narra locals, parameters, loop counters, the match done flag, and the result variables of block expressions, match expressions and map indexes. The hook is `Emitter.cVar` / `declareUnset` / `declareLocal` / `declareCounter` (Go), and `declareVariable` / `declareUnset` / `variableRef` / `declareLocal` / `declareLocalVariable` (selfhost `emitstate.nra`). Only scratch values used inside one expression (closure environment pointers, enum payload pointers, out-parameter structs) stay C locals. Outside a generator they write the same C as before. `localRef` gives `f->name` for frame locals.
- Live values: at state 0 and at each `yield` the emitter records what `unwindAll` would release there (`liveValues` / `recordLive`). Drop releases, and clone retains, the list of the current state (`switch (f->h.state)`). Both compilers stop with an internal error when a recorded value is not a frame field (no `f->` prefix), so a new raw declaration site fails in the emitter instead of in cc.
- Statements: `yield v` writes `*(T*)out`, sets the state, returns true and places `narra_resume_N:;`. `yield* g` resumes its own copy of `g` and passes `out` through. `return` and the end of the body release what is live, set state -1 and return false. `for ... of` over a generator (Go `forGenerator`, selfhost dispatch on `TGenerator`) resumes its own copy until it finishes.
- Closures in a generator body capture frame fields by value. `closureFunc` / `emitClosureFunction` save and restore the generator state, so a closure's body uses C locals.
- Selfhost `typeId` of `TGenerator` now matches Go's `typeID` (`G_<elem>_<effects>[_O<module>.<name>][_B<k>]`), so composite names such as `Array<Generator<T>>` agree. It was `x` before.
- The emitter stub is gone: `Program.Unemitted`, the `driver.go` line, `TestGeneratorEmitStub`, `Checker.unemitted` and the `emitter.nra` check. The E0022 table entry stays (codes are never reused), reworded as retired; nothing reports it.
- Samples (ASan and the leak check, debug and release): `compiler/testdata/generators.nra` (SPEC section 8 example with `counter`, `evens`, `both` with `yield*`, and an empty generator), `generators_heap.nra` (strings and arrays held across yields, `yield` in nested `while`, in for-of over an array and a range, in `match` arms, in an `if is` branch with a `return`), `generators_values.nra` (a generic generator, generators passed to functions, iterating one variable twice, an endless generator left with `break`, generators never iterated, one in a struct field dropped half way, an array of generators, tuple elements, an arrow function capturing frame locals), `generators_held.nra` (a map index, a match expression and a block expression held across yields as loop sources, as match subjects and under `yield*`, each also dropped half way).
- Docs: SPEC.md section 11 describes the lowering; `compiler/README.md` lists generators under Supported.

## Deviations

- One C type for all generators (`narra_gen*`) instead of one C struct per element type. It needs no per-type list in either checker. Resume takes `void* out`.
- Frame fields are not zeroed when their scope ends (brief item 5). Drop and clone use the per-state live list, which gives the same behavior and reuses the unwind lists that `return` already relies on. The frame is still zeroed at allocation.
- Files outside the owned and shared lists: `compiler/internal/narra/driver.go` (stub line removed), `types.go` (`CName`, `needsRC`), `emit_match.go`, `emit_expr.go`, `std_map.go` (declarations through the hook), `generators_test.go` (`TestGeneratorEmitStub` removed). `selfhost/src/runtimetext.nra` is regenerated. All small.

## Stubs left

- None from this node. `Coroutine<In, Out>` and `Co.resume` are node coroutines.

## Gotchas

- In a generator body, any C variable that a later statement reads, or that goes into `e.frames`, must be declared through `cVar`/`declareUnset` (Go) / `declareVariable`/`declareUnset` (selfhost). The first review found three raw sites (block, match and map-index results); the `f->` check in `liveValues`/`recordLive` now catches the next one. A plain `T name = ...;` line is fine only inside one statement, because no `yield` can be inside an expression. The two compilers must agree on which names go through the hook, or the frame struct fields differ: Go `e.temp(&ptrType{...})` is selfhost `pointerTemp`/`rawTemp`, and both now go through it.
- An assignment to a frame field cannot take a brace initializer: the Array.entries loop variable uses a compound literal `(T){...}` in a generator.
- Generator bodies never go through `function`/`emitFunction`'s tail-call path, so rule 8 needs no check.
- `async` can reuse the frame, header, `suspend` and the live lists; nothing in them is generator-only except `out` and the element type.

## Proposed learnings

- A C variable declared by the emitter that outlives one statement goes through `cVar`/`declareVariable` | generator bodies keep such variables in the frame across `yield` | compiler/internal/narra, selfhost/src | yes (a generator sample using the construct across a yield, run under ASan and TestSelfhostEmit)

## Verification

- `go test -count=1 ./internal/narra -run 'TestPrograms/generators'`: pass (debug with ASan and the leak check, and release).
- `go test -count=1 ./internal/narra -run 'TestSelfhostEmit$/inputs/generators'`: pass, byte-identical C for all four samples.
- The first review's programs (`/private/tmp/claude-501/pr89rev/t/*.nra`) build and run with the leak check, except the expected checker errors (`mut.nra`; `genval*.nra`, the function-type annotation bug that is also on main).
- `cd compiler && go vet ./... && go test -count=1 -timeout 45m ./...` on cdce7e3 (after the review fixes, rebased on origin/main 83dee4e): pass, `ok narra/internal/narra 688.0s`, TestSelfhostBootstrap and TestSelfhostEmit included. Output: the session scratchpad, `fast-check-cdce7e3.txt`. An earlier run with `-timeout 20m` hit the timeout under load from other worktrees.

## Size

- Estimated L, actual L.
