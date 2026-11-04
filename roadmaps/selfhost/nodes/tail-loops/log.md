# Log: tail-loops

- Nothing existed (no `tail` in the emitters, no `tail_loops*` sample).
- Design: a function with a self tail call owns its parameters. After the signature it retains each non-mutable heap parameter and registers it as a scope local, then writes `narra_tail:;`. `return f(...)` evaluates every new argument into an owned temporary, runs `unwindAll` (which releases the old parameters), assigns the parameters and writes `goto narra_tail;`.
- Not converted (stay plain calls): any function with a function-type parameter (a closure built for the call may live on the stack of the finishing iteration), and a call that does not pass a `mutable` parameter on as the same parameter. Tail position means `return f(...)` directly; `return 1 + f(...)` is not.
- Closures: the Go emitter tells a closure body from its function by the identity of the outer scope frame (`tailScope`), so `emit_closure.go` is untouched. Narra has no identity, so the selfhost side keeps `Emitter.tailProgram` and `closureFunc` clears and restores it.
- bench/: only `fib` recurses and it is not a tail call; no effect.
