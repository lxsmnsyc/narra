# Outcome: tail-loops

## Built

- Go: `selfTailCall`, `hasSelfTailCall`, `tailJump` in `compiler/internal/narra/emit_call.go`; `Emitter.tailScope` and the function prologue in `emit.go`; the `ReturnStmt` case in `emit_stmt.go`.
- selfhost: `selfTailArguments`, `hasSelfTailCall`, `openTailLoop`, `emitTailJump` in `selfhost/src/emitstmts.nra`; `Emitter.tailProgram` in `emitstate.nra`.
- A function with a self tail call retains its heap parameters on entry (it owns them), writes `narra_tail:;`, and `return f(...)` becomes: evaluate new arguments into owned temporaries, release everything, assign parameters, `goto narra_tail;`.
- Plain calls remain for non-tail calls, for functions with a function-type parameter, and for a `mutable` parameter that is not passed on as itself.
- `compiler/testdata/tail_loops.nra` and `.out`: 10M int tail calls, 10M calls replacing a string and an array parameter, swapped strings, a mutable array parameter, a call inside a `for` body, and a non-tail `depth`. Runs in the debug build under ASan with the leak check.
- bench/: only `fib` recurses and it is not a tail call, so no effect.

## Deviations

- Edited `selfhost/src/emitstate.nra` (new `Emitter` field), `selfhost/src/emitclosures.nra` (save and clear the field in a closure body) and `selfhost/src/emitconsts.nra` (`emitFunction` prologue). They are outside the brief's shared list but `emitFunction` and `Emitter` live there.
- Go's `emit_closure.go` is untouched: a closure body is told apart by its outer scope frame.

## Stubs left

- None.

## Gotchas

- Functions with a function-type parameter keep the recursive call.
- Tail position is `return f(...)` only; a void function's bare recursive call is not converted.
- `from` is a keyword in Narra, so samples cannot name a parameter `from`.

## Proposed learnings

- New per-function emitter state in selfhost must be saved and cleared in `closureFunc` | closures are emitted in the middle of another body | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'`: pass.
- `cd compiler && go vet ./... && go test ./...`: pass (296 s on a loaded machine), including `TestSelfhostBootstrap`.

## Size

- Estimated M, actual M.
