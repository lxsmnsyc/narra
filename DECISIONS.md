# Design decisions

Each entry states the decision first, then the reason and the rejected alternatives. Section numbers refer to `SPEC.md`.

## D1. Memory safety through reference counting, not a borrow checker or a garbage collector

Decision: immutable values with compiler-inserted reference counting and in-place reuse (section 6).

- **Reason.** Safe code has no references, so there is nothing to annotate and no borrow rules to learn. Frees are deterministic and there are no collector pauses.
- **Rejected: borrow checker with lifetimes (Rust).** Safe and fast, but lifetime annotations and borrow errors are the main source of complexity in Rust.
- **Rejected: tracing garbage collector (Lua, OCaml, Go).** Simple to use. But it adds a runtime, pauses and memory overhead that do not fit a native target.
- **Rejected: mutable value semantics with `inout` parameters (Hylo).** Safe without lifetimes, but adds parameter conventions to every signature. Reference counting with reuse gets the same in-place updates without them.
- **Cost.** Reference count updates in hot code that shares data. A prototype must benchmark this early, on numeric and string-heavy code.

## D2. Lua syntax with static types

Decision: Lua 5.4 syntax with type annotations, `enum`, `struct` and `match` added, and metatables removed (section 2). This replaces the ReasonML syntax of the first draft, kept at the git tag `reasonml-draft`.

- **Reason.** Lua has a small grammar that people and models already know.
- **Reason.** Lua's `require` already keeps imported names local to a file, which matches the module model (D3).
- **Rejected: keeping metatables.** They allow operator overloading and dynamic field lookup, so the meaning of `a + b` or `t.x` would depend on runtime state.
- **Prior art.** Teal compiles typed Lua to Lua. Luau adds gradual types and its own VM. Pallene compiles a typed Lua subset to C.

## D3. Explicit imports with package-rooted paths

Decision: one file is one module, names are private unless exported, and every external name comes from a `require` declaration (section 3).

- **Reason.** The import list gives the dependency graph without name resolution, which keeps build scheduling simple.
- **Reason.** Every name in a file is either declared there or listed at the top. Reading the file or a text search finds where it comes from.
- **Rejected: Lua globals.** Any file can read or write any global, so dependencies are implicit and typos create new variables.
- **Rejected: relative paths.** They break when a file moves. Package-rooted paths plus `lang move` avoid that.

## D4. No import cycles, no re-export statement, no default exports, no import side effects

Decision: the restrictions listed in section 3.

- **No cycles.** Modules compile in dependency order and in parallel. No module ever sees a partly initialized dependency.
- **No re-export statement.** Exporting is done only with an `export` declaration, so every exported name appears as a declaration in its file. Wildcard re-exports slow builds and hide where names come from, which is a known problem with barrel files in TypeScript. A named re-export form was also dropped, because `require` plus `export const` or `export type` already does the same job.
- **No default exports.** A default export gets a different name in each importing file, so text search misses uses.
- **No import side effects.** Initialization order can never matter, and startup runs no code before `main`.

## D5. No typeclasses or traits

Decision: behavior is passed explicitly as functions. Equality and hashing are derived structurally by the compiler (section 4).

- **Reason.** Trait implementations are visible program-wide, which needs coherence and orphan rules and makes "which implementation runs here" hard to answer by reading code.
- **Cost.** More explicit arguments, such as `Array.sortBy(xs, Int.compare)` instead of `xs.sort()`.

## D6. No functors or module parameters at first

Decision: modules cannot be parameterized by other modules in this draft.

- **Reason.** Parameterized modules are a large feature. Generics plus explicit function arguments cover most uses.
- **Revisit when.** Real code shows repeated patterns that generics cannot express.

## D7. Polymorphic code compiled once, with opt-in specialization

Decision: generic functions take values in a uniform form and compile to one copy. `@specialize` requests per-type copies (section 11).

- **Reason.** Per-type copies, as in C++ templates and Rust generics, are the main cause of large binaries and slow builds in those languages.
- **Cost.** Boxing in generic code. `@specialize` exists for hot paths.

## D8. A fixed set of effects in signatures

Decision: `alloc`, `io`, `yield` and `unsafe`, declared on exported functions and inferred elsewhere (section 7).

- **Reason.** A reader, human or agent, sees what a function may do from its signature alone.
- **Rejected: a general effect system (Koka).** User-defined effects and handlers add a large feature to learn. A fixed set covers the questions readers ask most.

## D9. Errors as return values, checked by the caller

Decision: functions that can fail return `Result<T, E>`. The caller checks it with `if ... is Err(e)` or `match`. There is no `try`, no `catch` and no early-return operator. Panics cannot be caught (section 5).

- **Reason.** Every way a function can fail is in its return type, and every place an error is handled is a visible branch.
- **Reason.** One way to handle errors. A propagation operator would be a second, shorter way.
- **Reason.** Uncatchable panics need no unwinding tables, which keeps binaries small and C output simple.
- **Cost.** Each fallible call takes three or four lines when the error is passed up, as in Go.
- **Rejected: exceptions and `pcall`.** The failure paths are invisible in signatures.
- **Rejected: Go-style `value, err` pairs.** Both values can be set or both empty at once. `Result` makes exactly one of them present.

## D10. C as the compile target

Decision: emit one C file per module (section 11). The goal is native binaries, the same target as C, Rust and Go.

- **Reason.** Every platform with a C compiler is a target from the start, and debug builds can use a fast C compiler.
- **Revisit when.** C compile time dominates debug builds. A native debug backend, like Go's and Zig's, is the next step.
- **Rejected for now: a bytecode VM.** Useful for scripting and hot reload, but slower than native code. It stays a possible later target (section 16).

## D11. Tooling made for agents

Decision: JSON diagnostics with exact fix edits, stable error codes, and a query CLI (section 14).

- **Reason.** An agent calls commands and reads their output. Exact edits remove a step where the agent rewrites a fix in its own words and can get it wrong.
- **Reason.** Explicit signatures (D22), explicit imports (D3) and declared effects (D8) work together. An agent understands a function from its file and its imports alone.

## D12. One call form and built-in operators

Decision: a call passes every argument. There is no currying, partial application, pipe, method call syntax, varargs or labeled arguments. Operators are built in (section 2).

- **Reason.** Each call has one reading and one form, `f(x, y)`. Code has one call form to read, search and generate.
- **Reason.** A function compiles to one C function with a fixed parameter list. No closures are allocated for partly applied calls.
- **Rejected: Lua's `obj:method()`.** It is a second way to pass the first argument, the same as a pipe.
- **Rejected: varargs.** Their types and counts are checked only at runtime. An array argument has one type.
- **Cost.** Data transformations read inside out or need named intermediate values. Fixing arguments needs an anonymous function.

## D13. Uniform form with a tag bit, and heap boxes for 64-bit values and structs in generic code

Decision: generic code receives one machine word per value. Types of 32 bits or less are stored inline with a tag bit. 64-bit numbers and structs are heap boxed (section 11).

- **Reason.** `i64` stays a full 64-bit integer that matches C's `int64_t`.
- **Reason.** Non-generic code never boxes. Boxing happens only where a value enters generic code, and `@specialize` removes it in hot paths.
- **Rejected: 63-bit `int` (OCaml).** Fast generic code, but `int` no longer matches a C type, which complicates C interop.
- **Rejected: a type descriptor passed with every generic call.** No boxing. But every generic call takes an extra argument, and generic code must read sizes at runtime.

## D14. Effect variables for higher-order functions

Decision: signatures declare effect variables with `effect E` in the type parameter list and use them after `uses`. Each one is bound by a function parameter (section 7).

- **Reason.** `map` with a pure function has only `alloc`, and `map` with an I/O function has `alloc, io`. A fixed effect list cannot state both.
- **Reason.** Inference writes the variable for private code. Only exported higher-order functions show it.
- **Rejected: higher-order functions accept only pure functions.** Too strict. Mapping over files to read them is common.

## D15. Strings indexed by checked byte offsets

Decision: string positions are `ByteIndex` values starting at 0, and slicing panics on a position inside a character (section 4).

- **Reason.** Byte offsets make slicing take constant time. Positions returned by searches are always valid.
- **Reason.** A distinct `ByteIndex` type stops a character count from being used as a position.
- **Rejected: character offsets.** Indexing by character would take time proportional to the index, which hides slow loops.
- **Rejected: iterators only.** Safe, but simple parsers need positions.

## D16. Git dependencies with minimal version selection

Decision: dependencies are git repositories pinned to tags, versions resolve with minimal version selection, and `lang.lock` stores hashes (section 15).

- **Reason.** Resolution needs no solver and gives the same result on every machine.
- **Reason.** No registry to run at the start.
- **Revisit when.** The number of packages makes discovery or name squatting a problem.

## D17. Async I/O deferred

Decision: the first version has threads, channels and coroutines. An async I/O event loop is deferred (section 16).

- **Reason.** async/await splits libraries into sync and async versions. Green threads need a scheduler, which goes against a small runtime.
- **Direction.** An event loop in `std/async` that resumes coroutines when I/O is ready. The `yield` effect already marks which functions can suspend.

## D18. Indexes start at 0

Decision: arrays and string positions start at 0, unlike Lua (sections 4 and 6).

- **Reason.** C, the byte offsets of D15 and most other languages start at 0. Code that calls C or ports algorithms needs no index shifting.
- **Cost.** Lua's numeric `for` is inclusive, so loops over an array write `for i = 0, #xs - 1 do`. `Array.entries` avoids the subtraction.

## D19. Lua-style coroutines, compiled to state machines

Decision: coroutines keep Lua's create, resume and yield model, typed as `Coroutine<In, Out>`. Functions that suspend have the `yield` effect. The compiler turns each coroutine into a state machine (section 8).

- **Reason.** No separate stack is allocated per coroutine, and no assembly or platform-specific context switching is needed in the runtime.
- **Reason.** The `yield` effect tells the compiler exactly which functions need the state-machine form. Code that never yields compiles as plain C functions.
- **Rejected: stackful coroutines (Lua's own runtime, Go).** They allow yielding from any call depth. But each coroutine needs its own stack, and switching stacks needs platform-specific code.
- **Cost.** A generic function called with a `yield` function gets a separate specialized copy.

## D20. Structs stored inline

Decision: `struct` types are stored inline, like C structs. `Box<T>` puts a value on the heap (sections 4 and 6).

- **Reason.** Small data such as points, ranges and colors needs no allocation and no reference count.
- **Reason.** `@repr(C)` structs match C layouts directly for interop.
- **Cost.** Copying a large struct copies all its fields. `Box<T>` exists for large or recursive data.

## D21. No truthiness and no `nil`

Decision: conditions must be `bool`. A missing value is `None` of type `Option<T>`, and there is no `nil` (section 4).

- **Reason.** In Lua, `0` and `""` are true and only `nil` and `false` are false. A typed condition removes that source of bugs.
- **Reason.** A plain `T` always holds a value, so the compiler checks every place a missing value is possible.

## D22. Every function signature is written

Decision: function parameters and return types are always written. Local variable types are inferred (section 4).

- **Reason.** Each function can be type-checked from its signature without reading other bodies. This keeps checking fast and local.
- **Reason.** A reader sees a function's contract without running inference in their head.
- **Exception.** Anonymous functions passed as arguments take their types from the parameter they are passed to.

## D23. Type syntax uses names, not symbols

Decision: types are written as named generics: `Option<T>`, `Array<T>`, `Map<K, V>`, `Tuple<A, B>`, `Result<T, E>`, `Generator<T>`. There is no `T?`, `{T}` or `{K: V}` type syntax, and no `??` operator (section 4).

- **Reason.** A name says what the type is. A reader who has not seen the language before can read `Option<string>`, but not `string?` or `{string}`.
- **Reason.** Names are easy to search for. Symbols such as `?` appear in many other places.
- **Cost.** Signatures are longer.
- **Rejected: Lua-style table types (`{T}`, `{K: V}`) from Teal.** One symbol form for two different types.

## D24. `const` and `let` instead of `local`

Decision: `const` declares a binding that cannot be reassigned, and `let` declares one that can. `local` is not a keyword (section 2).

- **Reason.** The keyword says whether the name can change. Lua's `local` says only where the name is visible, and every binding in Lang is local anyway, since there are no globals.
- **Reason.** `const` and `let` mean the same in TypeScript, so people and models read them correctly.
- **Note.** A top-level `const` must be a compile-time constant. Inside a function, `const` accepts any expression (section 10).
- **Rejected: `local` and `local mut`.** Two words for the common mutable case, and `mut` is not a Lua or TypeScript word.
