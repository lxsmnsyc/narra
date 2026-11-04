# Outcome: tail-loops

## Built

- Go: `selfCall`, `findTailCalls`, `tailJump`, `returnMatch` in `compiler/internal/narra/emit_call.go`; `Emitter.tailCalls` and the function prologue in `emit.go`; hooks in `emit_stmt.go` (`ReturnStmt`, `ExprStmt`) and `emit_match.go` (statement match arm).
- selfhost: new `selfhost/src/emittail.nra` (`findTailCalls`, `isTailCall`, `openTailLoop`, `emitTailJump`, `emitReturnArm`); hooks in `emitstmts.nra`, `emitmatch.nra` (`emitReturnMatch`, `returning` mode of `emitMatchArms`) and `emitconsts.nra`; `Emitter.tailProgram` and `Emitter.tailCalls` in `emitstate.nra`.
- Tail position is found before the body is emitted and recorded per call node. It covers `return f(...)`, an arm of `return match (...) { ... }` that is `f(...)`, and the last statement of a void function when it is `f(...)`, also at the end of an `if`/`else` branch or a `match` arm (block or expression) in tail position. A call inside a closure is never listed.
- A function with a listed call retains its heap parameters on entry (it owns them), writes `narra_tail:;`, and a tail call evaluates the new arguments into owned temporaries, releases everything, assigns the parameters and writes `goto narra_tail;`. A returned `match` makes each arm return its own value.
- Plain calls remain for non-tail calls, for functions with a function-type parameter, and for a `mutable` parameter that is not passed on as itself.
- `compiler/testdata/tail_loops.nra` and `.out`: 10M int tail calls, 10M calls replacing a string and an array parameter, swapped strings, a mutable array parameter, a call inside a `for` body, 1M calls through a returned `match` and through void functions (early return, `if`/`else`, `match` arms, heap parameters), and a non-tail `depth`. Runs in the debug build under ASan with the leak check.
- bench/: only `fib` recurses and it is not a tail call, so no effect.

## Deviations

- Edited `selfhost/src/emitstate.nra` (new `Emitter` fields), `emitconsts.nra` (`emitFunction` prologue) and `emitmatch.nra` (return mode and the tail-call arm), and added `selfhost/src/emittail.nra`. They are outside the brief's shared list, but `emitFunction`, `Emitter` and the match arms live there.
- Go: edited `emit_match.go` (one tail-call check in `matchStmt`), also outside the list.

## Stubs left

- None.

## Gotchas

- Functions with a function-type parameter keep the recursive call.
- A match nested inside a returned match arm keeps plain calls in its arms.
- `from` is a keyword in Narra, so samples cannot name a parameter `from`.
- The C after a `goto narra_tail;` at the end of a function (the function-end releases and panic) is still emitted; it is the same text every function gets.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass, including `TestSelfhostEmit` and `TestSelfhostBootstrap`.
- `narra run compiler/testdata/tail_loops.nra` with `NARRA_CHECK_LEAKS=1` (debug build, ASan and UBSan): pass.

## Size

- Estimated M, actual M.
