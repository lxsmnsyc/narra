# Outcome: closures

## Built

All code is in `compiler/internal/narra/`. The logic is in `check_closure.go` and `emit_closure.go`. Shared files only gain `case` lines and calls.

Syntax:
- `ArrowExpr` (`ast.go`): `Params []*Param` (with `Type` nil when the type is left out), `Result *TypeExpr`, then either `Body Expr` or `Block *Block`, plus `Closure *Closure`, which the checker fills. The parser (`arrowAhead`, `arrow` in `parser.go`) reads `(` as an arrow function when its matching `)` is followed by `=>` or `:`. A `{` after `=>` is always a block. Arrow parameters cannot be `mutable`.
- `Param.Escaping`: `escaping f: () => void` (`acceptEscaping`). `escaping` is contextual, not a keyword.
- Function types use the traits slice (`funcTypeExpr`, `TypeExpr.Func`, `FuncType`). `() => void` now resolves, through `funcResult`.
- `CallExpr.FuncValue *FuncType` is set for a call through a value. `Ident.Func` and the new `FieldExpr.Func` are set for a named function used as a value (`double`, `Alias.double`).

Checker (`check_closure.go`):
- `c.funcOf(params, result)` returns the one canonical `*FuncType` for a signature. `resolveType` and `substitute` use it. `Program.FuncTypes` lists the concrete ones, sorted by mangled name, for the emitter.
- `Closure{Type, Fn, Captures, Stack, Hint}`. `Fn` is a synthetic `FuncDecl` with the arrow's params (their `Local`s), its `ResultT`, and a `Body` block to emit. An expression body becomes `return e`, or `e` as a statement when the result is void.
- Parameter types come from the annotation, or from the expected `FuncType`, or from `Hint` (inside a generic call, `hintArrow` unifies the earlier arguments, so `map(xs, (x) => x * 2)` types `x`). The result type comes from the annotation, or the expected type, or the expression body, or the `return`s of a block body (void if none returns a value).
- A named function as a value: `funcValue`. A generic function needs an expected function type to infer its type arguments; then it uses `funcInstance`. A function with `mutable` parameters cannot be a value.
- Calls: `valueCall` handles a callee that is a local or any expression, such as `make()(x)` or `xs[0](x)`. `valueCallOf` handles `value.field(args)` when the field holds a function, through `fieldFunc`.

Capture rules:
- An arrow body has its own scope, whose parent is the enclosing scope (so the no-shadowing rule still applies). A local found outside the arrow's own scopes is a capture. Nested arrows each capture it (`Closure.Captures`, in order of first use).
- Allowed captures: `const` bindings, non-mutable parameters, `for` and pattern bindings. Errors: a `let` variable, a `mutable` parameter, or a non-escaping parameter. `break` and `continue` inside an arrow do not see outer loops.

Escaping rules as built (D46):
- `Local.NoEscape` is set for a parameter written with a function type and without `escaping` (`nonEscapingParam`). Such a parameter can only be (a) called or (b) passed directly as an argument to another parameter of that kind. Any other use of its name is an error: return, `const`, struct field, array element, capture, `==`, an escaping parameter, a generic `T` parameter, or an argument to a function value. The check is in `useLocal`, which runs on every `Ident` that names a local. `funcArg` and `valueCall` record the single `Ident` that is allowed.
- An arrow passed directly as an argument to a non-escaping parameter gets `Closure.Stack = true`. Every other arrow is escaping.
- A function type has no `escaping` marker, so parameters reached through a function value are escaping.
- `escaping` on a parameter that is not a function type is an error.

C representation (`emit_closure.go`):
- Function type `(A, B) => R` is `struct F2__A__B__R { R (*fn)(narra_env*, A, B); narra_env* env; }`. The name is `mangle(t)`, which is now `FuncType.CName()`. The typedefs and struct definitions are written after all forward typedefs and before struct definitions (`funcTypeDefs`), so structs can hold function values inline.
- `narra_env { int64_t rc; void (*drop)(narra_env*); }`, with `narra_env_retain` and `narra_env_release`. Both ignore NULL and a negative count. The runtime is emitted only when the program has function types, from `envRuntime`, not `runtime.h`.
- Arrow function: `static R k<owner>__<n>(narra_env* env, params)`. Here `<owner>` is `funcCName` of the enclosing declared function, or of instance `f_<module>__test` for a test block. `n` counts from 1 per owner, nested arrows included. Its environment is `typedef struct K<owner>__<n> { narra_env h; T c_<local CName>; ... }`. `D<owner>__<n>` releases the captures. At the start of the body, captures are copied into C locals with their original C names (borrowed), so the body code is the same as in a normal function.
- Named function as value: thunk `static R w<funcCName>(narra_env* env, P0 p0, ...)`, emitted once, with env NULL.
- The names are injective: `funcCName` is injective, and no `funcCName` ends in `__<digits>`. The prefixes `k`, `K`, `D`, `w` + `f_` are used nowhere else.
- Values: no captures gives `{fn, NULL}`, which is not owned. Stack gives a block-local `K... kN = {.h = {-1, NULL}, .c_x = x}` and `{fn, &kN.h}`, which is not owned. Captures are borrowed, and nothing is counted or freed. Heap gives `narra_alloc`, `rc = 1`, `drop = D...`, and each capture is copied and retained. The value is an owned temporary.

Reference counting:
- `needsRC(FuncType)` is true. `retain` and `release` of a function value are `narra_env_retain/release((v).env)`. So locals, struct fields, arrays and enums that hold function values count them like strings. Parameters are borrowed as usual.

Calling a function value: `valueCall` takes an owned reference to the callee (`consume`, then released at the end of the statement) and emits `t.fn(t.env, args...)`. Arguments are borrowed. A non-void result is an owned temporary, the same as a direct call. Holding the reference means an argument that reassigns the variable holding the callee cannot free its environment during the call. The checker also rejects that case: `checkCalleeExclusive` reports a variable that is passed `mutable` anywhere inside the arguments and also read by the callee expression (`f(swap(mutable f))`, `h.run(reset(mutable h))`).

Tests: `testdata/closures.nra` (generic `map`, `forEach`, string and array captures, escaping `adder`, `compose`, `greeter`, a struct field, an array of closures, a named function as a value) and `testdata/closures_nested.nra` (nested closures, a generic function as a value, captured struct and narrowed `Option`, arrows in generic bodies and in loops, block bodies with an inferred result). Package `testdata/packages/closures` (named and `F.double` values across modules, a generic `F.map` with an arrow, escaping `F.twice`, a closure in an imported struct, private functions named `inc` in both modules, and two test blocks with closures). Sample `testdata/closures_traits.nra` (a trait function with an `escaping` parameter, implemented with the same marker and called from a constrained generic). Error tests `testdata/errors/closures_*.nra` (12, including `closures_trait_escaping`, `closures_callee_mutable` and `closures_field_callee_mutable`) and `testdata/packages/errors/closure_escape`.

## Deviations

- `DECISIONS.md` gains D51 (parameters reached through a function value are escaping) and D52 (a parameter whose type is a type parameter is escaping). Both make D46 stricter. The coordinator asked for them during review.
- `check_traits.go` (traits code, not in the brief's lists): `sameSignature` now requires an `implement` function to match the trait's `escaping` on each parameter, and `signatureText` prints it. Without this, an implementation could store a parameter that the trait declares non-escaping, and the stack environment would escape (review finding, ASan stack-use-after-return).
- Edits outside the owned and shared lists, each a few lines:
  - `check.go`: `Program.FuncTypes`, the `Checker.closure` field, and a `c.finishFuncTypes()` call in `finish`.
  - `emit.go`: the `Emitter.closures` field, the `e.funcTypeDefs(&b)` and `e.closureFuncs(&b)` calls, and `Module: t.Module` on the test block's `FuncDecl`, so closures in tests carry the module.
  - `compiler/README.md`.
  - `testdata/traits_functypes.nra` and `.out`: its functions stored and returned non-escaping parameters, which is now an error. They are marked `escaping`, and `run` now calls them.
- `check_types.go` `resolveType`: the result of a function type is resolved through `funcResult`, so `void` is accepted there.
- New C name forms (`k`, `K`, `D`, `w` + `funcCName`) are not in the module-graph contract. They follow its rule through `funcCName`. The orchestrator may want to list them in the contract.
- Differences from the spec are recorded in `compiler/README.md`:
  - function types carry no `escaping`, so parameters reached through a function value are escaping;
  - a type-parameter parameter is escaping;
  - a block-bodied arrow infers its result from its `return`s.

## Stubs left

- `@inline`: node inline. A non-escaping closure is still an indirect call through `fn`.
- `async` arrow functions: node async. Effects on function types (`uses E`): node effects. `funcTypeExpr` does not parse `uses`.
- Standard functions (`Int.toString`) cannot be used as values.
- A closure that is an enum variant field and is called through a narrowed variable (`x.f()` after `is`) is not supported. Only struct fields are.

## Gotchas

- `FuncType.CName()` is no longer `void*`. Always create function types with `c.funcOf`, or the emitter will not write their struct. A `&FuncType{}` literal is only safe as a checker-internal hint.
- A non-escaping parameter is decided by syntax (`p.Type.Func`), never by the resolved type. So generic instances whose type argument is a function type keep their escaping behaviour.
- Emitting an arrow writes its C function into `closureEmit` immediately. It saves and restores `e.out`, `e.indent`, `e.frames`, `e.scopes` and `e.loopTemp`. Anything else added to the emitter's per-function state must be saved there too. `e.fn` is kept, because it names the owner.
- Stack environments are block-local C variables. They are only valid because the closure is an argument in the same statement. Do not set `Closure.Stack` anywhere else. A future tail-call-to-loop transform must not pass a stack environment through a rewritten parameter: the loop would reuse a C variable whose block has ended, or overwrite one that is still in use.
- Arrows inside generic templates are checked but never emitted. Each instance re-checks its cloned arrows (`cloneArrow`).
- threads and async must treat `Thread.start` and `Async.spawn` parameters as escaping. Builtins do not go through `funcArg`, so arrows passed to them are already heap-allocated, and non-escaping parameters are rejected.

## Proposed learnings

- Create function types only through `c.funcOf` | the emitter writes C structs only for listed types | compiler/internal/narra | yes (any program using an unlisted type fails to compile in C)
- New per-function emitter state must be saved in `closureFunc` | closures are emitted in the middle of another body | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass. All samples, including both closure samples and the package, run under ASan, UBSan and the leak check (`NARRA_CHECK_LEAKS=1`), and in release mode.
- Checked by hand: `narra emit testdata/closures.nra` shows stack environments for arguments of non-escaping parameters and heap environments for returned or stored closures. `narra test testdata/packages/closures` passes 2 tests. A `break` inside an arrow in a loop is an error. `==` on a function value is still an error.

## Size

- Estimated L, actual L.
