# Design decisions

Each entry states the decision first, then the reason and the rejected alternatives. Section numbers refer to `SPEC.md`.

## D1. Memory safety through reference counting, not a borrow checker or a garbage collector

Decision: immutable values with compiler-inserted reference counting and in-place reuse (section 6).

- **Reason.** Safe code has no references, so there is nothing to annotate and no borrow rules to learn. Frees are deterministic and there are no collector pauses.
- **Rejected: borrow checker with lifetimes (Rust).** Safe and fast, but lifetime annotations and borrow errors are the main source of complexity in Rust.
- **Rejected: tracing garbage collector (OCaml, Go).** Simple to use. But it adds a runtime, pauses and memory overhead that do not fit a native target.
- **Taken in part: mutable value semantics (Hylo).** Narra adds one optional parameter convention, `mutable` (D27). Parameters default to constant values, so most signatures carry no convention.
- **Cost.** Reference count updates in hot code that shares data. A prototype must benchmark this early, on numeric and string-heavy code.

## D2. TypeScript-style syntax with Rust-style semantics

Decision: braces, `const` and `let`, arrow functions, `import`, template strings and type annotations follow TypeScript. Types and semantics follow Rust: enums with fields, `match`, `Option` and `Result`, no `null`, sized numbers (section 2).

- **Reason.** TypeScript syntax is the most widely read syntax among the target users and in model training data.
- **Rule.** Where Narra's meaning differs from TypeScript's, Narra uses different syntax or a compile error with a fix (D42). The same text never means two different things.

## D3. Explicit imports with package-rooted paths

Decision: one file is one module, names are private unless exported, and every external name comes from an `import` declaration (section 3).

- **Reason.** The import list gives the dependency graph without name resolution, which keeps build scheduling simple.
- **Reason.** Every name in a file is either declared there or listed at the top. Reading the file or a text search finds where it comes from.
- **Rejected: globals.** Any file can read or write any global, so dependencies are implicit and typos create new variables.
- **Rejected: relative paths.** They break when a file moves. Package-rooted paths plus `narra move` avoid that.

## D4. No import cycles, no re-export statement, no default exports, no import side effects

Decision: the restrictions listed in section 3.

- **No cycles.** Modules compile in dependency order and in parallel. No module ever sees a partly initialized dependency.
- **No re-export statement.** Exporting is done only with an `export` declaration, so every exported name appears as a declaration in its file. Wildcard re-exports slow builds and hide where names come from, which is a known problem with barrel files in TypeScript. A named re-export form was also dropped, because `import` plus `export const` or `export type` already does the same job.
- **No default exports.** A default export gets a different name in each importing file, so text search misses uses.
- **No import side effects.** Initialization order can never matter, and startup runs no code before the entry function.

## D5. Traits implemented only where the type is declared

Decision: traits exist, but a type's trait implementations must be in the module that declares the type and listed with `implements` on its declaration. Trait functions are called through the trait name, such as `Show.show(x)` (section 4).

- **Reason.** Formatting, serialization, ordering and mixed-type collections all need a shared interface. Without traits, each needed its own workaround.
- **Reason.** With implementations only next to the type, there is exactly one place to look and no coherence or orphan rules are needed.
- **Reason.** Calling through the trait name keeps one call form (D12) and makes every use searchable.
- **Reason.** Constrained generic code receives the implementation table as a hidden argument, so generic functions are still compiled once (D7).
- **Cost.** No retroactive implementations. Adding a trait to another module's type requires a wrapper type.
- **Rejected: Rust-style traits with implementations anywhere.** They need global coherence rules, and finding the implementation that runs needs tooling.
- **Rejected: no user traits, only compiler-derived abilities.** Simpler, but users could not define `Ordered`-like contracts of their own.

## D6. No functors or module parameters at first

Decision: modules cannot be parameterized by other modules in the first version.

- **Reason.** Parameterized modules are a large feature. Generics plus explicit function arguments cover most uses.
- **Revisit when.** Real code shows repeated patterns that generics cannot express.

## D7. Generic code with layout tables, specialized in release builds

Decision: unspecialized generic code receives a layout table per type parameter. Debug builds compile each generic function once. Release builds specialize within a package. `@specialize`, `@inline` and `generics = "compact"` adjust this (section 11).

- **Reason.** Debug builds stay fast to compile, since each generic function is compiled once.
- **Reason.** Release builds get direct code for the generic calls inside a package, where most hot code is.
- **Reason.** Calls across packages use layout tables by default. Binary size and build time do not grow with every use, as they do with C++ templates and Rust generics.
- **Cost.** Unspecialized generic code copies values through table functions instead of direct moves.

## D8. A fixed set of effects in signatures

Decision: `alloc`, `io` and `unsafe`, declared on exported functions and inferred elsewhere (section 7). Suspension is marked by `async` and `function*` instead of an effect.

- **Reason.** A reader, human or agent, sees what a function may do from its signature alone.
- **Rejected: a general effect system (Koka).** User-defined effects and handlers add a large feature to learn. A fixed set covers the questions readers ask most.

## D9. Errors as return values, checked by the caller

Decision: functions that can fail return `Result<T, E>`. The caller checks it with `if ... is Err(e)` or `match`. There is no `try`, no `catch` and no early-return operator. Panics cannot be caught (section 5).

- **Reason.** Every way a function can fail is in its return type, and every place an error is handled is a visible branch.
- **Reason.** One way to handle errors. A propagation operator would be a second, shorter way.
- **Reason.** Uncatchable panics need no unwinding tables, which keeps binaries small and C output simple.
- **Cost.** Each fallible call takes three or four lines when the error is passed up, as in Go.
- **Rejected: exceptions (`throw`, `try`, `catch`).** The failure paths are invisible in signatures.
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

Decision: a call passes every argument. There is no currying, partial application, pipe, method call, varargs or optional parameters. Labeled arguments were ruled out too; D87 later added named arguments, which the checker puts back in parameter order. Operators are built in, and there are no operators for joining strings, length or powers (section 2).

- **Reason.** Each call has one reading and one form, `f(x, y)`. Code has one call form to read, search and generate.
- **Reason.** A function compiles to one C function with a fixed parameter list. No closures are allocated for partly applied calls.
- **Rejected: method calls such as `xs.push(v)`.** They are a second way to pass the first argument. They also make the called function depend on the value's type.
- **Rejected: varargs and optional parameters.** A struct with defaults covers optional settings with one rule.
- **Rejected: `+` for strings.** An operator's meaning would depend on its operand types. Template strings join text instead.
- **Rejected: `#` or `.length`.** Named functions such as `Array.length` say which type they work on.
- **Cost.** Data transformations read inside out or need named intermediate values. Fixing arguments needs an arrow function.

## D13. Layout tables instead of boxing in generic code

Decision: generic code handles values of a type parameter through a table with size, alignment, copy, drop and move functions. Values are never boxed for generic code, and containers store raw values (section 11).

- **Reason.** `Array<int>` holds raw integers in all code. A boxed design would allocate one heap object per element in generic containers.
- **Reason.** `i64` stays a full 64-bit integer that matches C's `int64_t`.
- **Prior art.** Swift's value witness tables work the same way.
- **Rejected: a uniform one-word form with boxing for large values.** It made generic containers of numbers and structs allocate per element.
- **Rejected: 63-bit `int` (OCaml).** Fast generic code, but `int` no longer matches a C type.

## D14. Effect variables for higher-order functions

Decision: signatures declare effect variables with `effect E` in the type parameter list and use them after `uses`. Each one is bound by a function parameter (section 7).

- **Reason.** `map` with a pure function has only `alloc`, and `map` with an I/O function has `alloc, io`. A fixed effect list cannot state both.
- **Reason.** Inference writes the variable for private code. Only exported higher-order functions show it.
- **Rejected: higher-order functions accept only pure functions.** Too strict. Mapping over files to read them is common.
- **Extended by D78.** A parameter of generator type binds an effect variable too.

## D15. Strings indexed by checked byte offsets

Decision: string positions are `ByteIndex` values starting at 0, and slicing panics on a position inside a character (section 4).

- **Reason.** Byte offsets make slicing take constant time. Positions returned by searches are always valid.
- **Reason.** A distinct `ByteIndex` type stops a character count from being used as a position.
- **Rejected: character offsets.** Indexing by character would take time proportional to the index, which hides slow loops.
- **Rejected: iterators only.** Safe, but simple parsers need positions.

## D16. Git dependencies with minimal version selection

Decision: dependencies are git repositories pinned to tags, versions resolve with minimal version selection, and `narra.lock` stores hashes (section 15).

- **Reason.** Resolution needs no solver and gives the same result on every machine.
- **Reason.** No registry to run at the start.
- **Revisit when.** The number of packages makes discovery or name squatting a problem.

## D17. `async` and `await` on an explicit event loop

Decision: async functions use TypeScript's `async function` and `await`. Every call to an async function needs `await`. Tasks run on an event loop from `std/async`, started with `Async.run` and `Async.spawn` (section 8).

- **Reason.** `await` marks every point where a task can suspend, at the call itself. A reader does not need to look up the called function.
- **Reason.** A call without `await` is an error, so there are no forgotten background tasks. Concurrent work is started on purpose with `Async.spawn` or `Async.all`.
- **Reason.** The runtime stays small. The event loop is library code, and there is no work-stealing scheduler.
- **Cost.** Function coloring. An async function can only be called from another async function or through the event loop. Higher-order functions such as `Array.map` do not take async callbacks, so `std/async` has its own versions.
- **Rejected: a `yield` effect for suspension.** It showed suspension only in signatures, not at calls, and it did not match the TypeScript-style syntax.
- **Rejected: green threads with a runtime scheduler (Go).** No coloring. But the scheduler and growable stacks go against a small runtime and a C target.
- **Rejected: TypeScript's eager start and `Promise`.** An unawaited call that runs in the background is a common source of lost errors.

## D18. Indexes start at 0

Decision: arrays and string positions start at 0 (sections 4 and 6).

- **Reason.** C, the byte offsets of D15 and most other languages start at 0. Code that calls C or ports algorithms needs no index shifting.
- **Reason.** Counting loops use `Range.upTo(n)`, which stops before `n`, so loops over an array need no `- 1`.

## D19. Generators with `function*` and `yield`, compiled to state machines

Decision: generators use TypeScript's `function*`, `yield` and `yield*`. `yield` is valid only directly inside a generator body. The compiler turns each generator into a state machine (section 8).

- **Reason.** The syntax is familiar from TypeScript.
- **Reason.** Because `yield` cannot cross a normal call, only generator bodies need the state-machine form. Every other function compiles as a plain C function, and generic code needs no special copies for generators.
- **Reason.** No separate stack is allocated per generator, and no platform-specific context switching is needed in the runtime.
- **Rejected: stackful coroutines (Go).** They allow yielding from any call depth. But each coroutine needs its own stack, and switching stacks needs platform-specific code.

## D20. Structs stored inline

Decision: `struct` types are stored inline, like C structs. `Box<T>` puts a value on the heap (sections 4 and 6).

- **Reason.** Small data such as points, ranges and colors needs no allocation and no reference count.
- **Reason.** `@repr(C)` structs match C layouts directly for interop.
- **Cost.** Copying a large struct copies all its fields. `Box<T>` exists for large or recursive data.

## D21. No truthiness and no `null`

Decision: conditions must be `bool`. A missing value is `None` of type `Option<T>`. There is no `null` or `undefined` (section 4).

- **Reason.** TypeScript treats values such as `0`, `""` and `null` as false in conditions. A typed condition removes that source of bugs.
- **Reason.** A plain `T` always holds a value, so the compiler checks every place a missing value is possible.

## D22. Every function signature is written

Decision: function parameters and return types are always written. Local variable types are inferred (section 4).

- **Reason.** Each function can be type-checked from its signature without reading other bodies. This keeps checking fast and local.
- **Reason.** A reader sees a function's contract without running inference in their head.
- **Exception.** Arrow functions passed as arguments take their types from the parameter they are passed to.

## D23. Type syntax uses names, not symbols

Decision: types are written as named generics: `Option<T>`, `Array<T>`, `Map<K, V>`, `Result<T, E>`, `Generator<T>`. Tuples are the exception: their type is `#[A, B]`, like their literal (D53). There is no `T?`, `T | null`, `{T}` or `{K: V}` type syntax, and no `??` operator (section 4).

- **Reason.** A name says what the type is. A reader who has not seen the language before can read `Option<string>`, but not `string?` or `{string}`.
- **Reason.** Names are easy to search for. Symbols such as `?` appear in many other places.
- **Cost.** Signatures are longer.
- **`T[]` was considered and rejected (2026-10-28).** In a function type it is ambiguous, because `() => int[]` can mean either a function returning an array or an array of functions. It would also make `Array` the only collection type written with a symbol, unlike `FixedArray<T, N>` and `Map<K, V>`; tuples (D53) are not a collection. TypeScript accepts `Array<T>` as well, so the named form costs TypeScript users nothing. `T[]` gets a diagnostic with a fix instead (section 2).

## D24. `const` and `let`

Decision: `const` declares a binding that cannot be reassigned, and `let` declares one that can (section 2).

- **Reason.** The keyword says whether the name can change.
- **Reason.** TypeScript uses the same words. Narra goes further, since a `const` value can never change, not only its binding. Section 2 lists this as a difference from TypeScript.
- **Note.** A top-level `const` must be a compile-time constant. Inside a function, `const` accepts any expression (section 10).

## D25. TypeScript import syntax, with namespace and named imports only

Decision: `import * as Name from "path"` and `import { a, b } from "path"`. There are no default imports and no dynamic `import(...)` (section 3).

- **Reason.** These are TypeScript's own forms, so they read correctly with no new rules.
- **Reason.** `import X from "m"` means a default import in TypeScript. Narra has no default exports, so that form is a compile error instead of a different meaning.
- **Rejected: a call-style import such as `const Name = import(...)`.** It looks like a runtime call, but importing runs nothing.
- **Note.** D72 adds renaming to named imports: `import { map as mapArray } from "std/array"`.

## D26. Struct literals, defaults, spread updates and destructuring

Decision: structs use typed literals (`Point { x: 1.0, y: 2.0 }`), constant field defaults, `{ ...s, field: value }` for updated copies, destructuring in `const` and `match`, and `Box.new` with `.value` for boxes. Field shorthand in literals is not allowed (section 4).

- **Typed literals.** Without `Point { ... }`, every struct literal would need an annotation or a context that gives the type.
- **Constant defaults.** Optional settings need no builder functions. Constant values keep building a struct free of hidden calls.
- **Spread updates.** `{ ...s, port: 9000 }` is the TypeScript form for an updated copy. Narra limits it to one spread, first, with the same struct type, so it cannot merge unrelated values.
- **Destructuring.** `const { x, y } = p` and `{ x: 0.0, y } =>` read several fields in one step, using TypeScript's pattern syntax.
- **No literal shorthand.** `{ name: name }` is always written in full. `{ name }` in a literal would be a second way to write a field. In a pattern, `{ name }` is the only way to bind a field to its own name, so it stays.
- **Rejected: positional struct literals such as `Point(1.0, 2.0)`.** They depend on field order, and that form already means an enum constructor.

## D27. `mutable` parameters, marked at the declaration and the call

Decision: a function changes a caller's variable only through a parameter declared `mutable`. The call writes `mutable` before the argument. Access is exclusive for the call, and the parameter cannot escape (section 6).

- **Reason.** Every change to a caller's data is visible in the signature and at the call. A reader of either file knows the value changes.
- **Reason.** In-place updates are part of the language, not an optimization. The compiler passes a pointer, with no copies.
- **Reason.** Exclusive access and no escaping keep it memory safe with one local check per call. The reference ends when the call returns, so no lifetimes are needed.
- **Reason.** It gives generators and collections one explicit form: `Co.resume(mutable g, x)` and `Array.push(mutable xs, v)`.
- **Rejected: guaranteeing in-place updates for `x = f(x)`.** The fast path would depend on a pattern that is not visible. `y = f(x)` looks almost the same and copies.
- **Rejected keywords.** `inout` describes the mechanism instead of the meaning. `mut` is an abbreviation. `ref` suggests references that could be stored.
- **Amended by D94.** Passing a `mutable` argument that another live name shares is a compile error (E0978), not a copy. Write `copy(x)` where the sharing starts.
- **Cost.** Two ways to produce changed data: a `mutable` parameter or a returned value. The rule is `mutable` to change the caller's data, and a return value to build new data.

## D28. Scoped threads

Decision: threads start inside `Thread.scope`, which waits for all of them before it returns. A `Scope` cannot escape (section 8).

- **Reason.** No thread outlives the scope that started it, so no thread is leaked and no result is forgotten.
- **Reason.** Every value is immutable or copied, so any value can be passed to a thread. No `Send` or `Sync` traits are needed.
- **Rejected: detached `Thread.spawn`.** It allows threads that run past the code that started them, with results nobody reads.

## D29. A few identity types for shared state

Decision: `Sender<T>`, `Receiver<T>`, `Shared<T>` and `Atomic<T>` are the only safe types with identity. `Shared<T>` is used through `Shared.update` and `Shared.read` callbacks (section 8).

- **Reason.** A lock must be one object that every copy refers to. Plain values cannot express that, so these types are named and few, and easy to find with a search.
- **Reason.** Callbacks release the lock when they return, so a lock cannot be left held.
- **Reason.** Callbacks cannot be `async`, so a task cannot suspend while holding a lock.
- **Cost.** Deadlocks are still possible with nested locks. `narra audit` lists nested `Shared.update` calls.

## D30. Literals, equality constraints and floats

Decision: strings use `"..."`, template strings use backticks, and characters use `'...'`. Literals must fit their type. `Equatable` and `Hashable` are standard traits implemented automatically when possible. `==` on floats follows IEEE 754, and floats cannot be hashed (sections 2 and 4).

- **Reason.** One quote style for plain strings, and `char` needed a literal form. `'abc'` is a compile error with a fix, because TypeScript reads it as a string.
- **Reason.** Generic code that uses `==` or hashing must know its type argument supports them. The constraints `<T: Equatable>` and `<T: Hashable>` say this.
- **Reason.** Float keys break maps when `NaN` is involved, because `NaN` is not equal to itself.

## D31. Iterators are generators

Decision: `for (const x of iter)` takes a `Generator<T>`. Collections return generators (section 2).

- **Reason.** No new protocol or trait is needed. Generators already define suspend and resume.

## D32. Maps keep insertion order

Decision: map iteration follows insertion order (section 4).

- **Reason.** Output, tests and builds are the same on every run.

## D33. Slices share memory

Decision: `Str.slice` and `Array.slice` return views into the original buffer. `copy` functions make a separate copy (section 4).

- **Reason.** Parsers slice often, and a copy per slice is slow.
- **Cost.** A small slice keeps a large buffer alive until the slice is dropped.

## D34. Entry points set in `narra.toml`

Decision: each `[[binary]]` in `narra.toml` names an entry function as `"module.function"`. Its signature is `(args: Array<string>): int` (section 15).

- **Reason.** One package can build several binaries without a special `main` name or file.
- **Reason.** A package with no `[[binary]]` is a library, which needs no extra marker.

## D35. C interop details

Decision: `@repr(<integer type>)` enums with explicit values, `@export(C, "name")` for callbacks, `CString` in `std/ffi`, and a `[link]` table in `narra.toml` (sections 9 and 15).

- **Reason.** These cover calling C libraries that take flags, callbacks and strings, which most C libraries do.

## D36. Runtime failures abort

Decision: running out of memory, stack overflow and integer division by zero abort the program, in every build (sections 4, 6 and 11).

- **Reason.** These failures cannot be handled where they happen in a useful way. Aborting matches panics (D9).

## D37. Recursion is found by the compiler, and self tail calls are loops

Decision: there is no `recursive` keyword. The compiler finds recursion from each module's call graph. A self-call in tail position compiles to a loop in every function (section 2).

- **Reason.** The compiler already builds the call graph. Imports cannot form cycles (D4), so recursion never spans modules except through function values.
- **Reason.** A keyword would repeat what the compiler knows. Accidental recursion is rare in Narra, since names are not shadowed and imported functions are called with a module prefix.
- **Reason.** The tail-call guarantee covers every function, not only marked ones.
- **Cost.** Recursion is not visible in a declaration. `narra query recursion MODULE` lists recursive functions.
- **Rejected: a `recursive` keyword.** It made recursion visible, but added a keyword for information the compiler already has.
- **Rejected: `@tailcall` attribute.** It marked only the optimization, not the recursion.

## D38. No shadowing, and unused names are errors

Decision: a name cannot be declared again in the same function while an earlier binding is in scope. Unused locals, parameters, imports and private functions are compile errors unless they start with `_` (section 2).

- **Reason.** Each name in a function means one thing, so a reader or an agent never has to work out which binding is meant.
- **Reason.** Unused code is removed as soon as it appears instead of accumulating.

## D39. Opaque structs

Decision: `export opaque struct` exports a struct type without its fields. Only the declaring module can build, read or change the fields (section 4).

- **Reason.** A library can enforce rules on its values, such as `start <= finish`. Without it, any module could build an invalid value.
- **Reason.** The standard library needs it for types such as `Handle<T>`.
- **Rejected: per-field visibility.** It adds a keyword to every field, and a struct with some hidden fields still cannot be built from outside.

## D40. Distinct types

Decision: `distinct type A = B` declares a new type with the representation of `B`, converted only through generated `from` and `to` functions (section 4).

- **Reason.** IDs, units and indexes with the same representation cannot be mixed up by accident.
- **Reason.** No runtime cost, since the C type is the same.
- **Reason.** With `opaque`, only the declaring module can create values, which is how `ByteIndex` is defined.

## D41. Build conditions in code instead of per-platform files

Decision: platform differences use `if` on compile-time constants from `std/build` and `@target` on declarations. Every branch is type-checked on every build (section 12).

- **Reason.** Shared code is written once. Small platform differences stay inside one function.
- **Reason.** Checking every branch on every build stops code for other platforms from breaking unnoticed, which is the main problem with C's `#if`.
- **Reason.** A reader or an agent sees every platform's behavior in one place.
- **Rejected: per-platform file suffixes (Go).** Small differences need a whole separate file, and shared logic gets copied between files.
- **Rejected: a preprocessor.** Code inside an inactive branch is never checked.

## D42. TypeScript forms with a different meaning are errors with fixes

Decision: code that is valid TypeScript but would mean something different in Narra is a compile error. The diagnostic includes an exact fix (sections 2 and 14).

- **Reason.** People and models will write TypeScript habits. A compile error with a fix is cheap. A silently different meaning causes bugs.
- **Examples.** `===` becomes `==`. `null` becomes `None`. `xs.push(v)` becomes `Array.push(mutable xs, v)`. `import X from` becomes `import * as X from`. `'abc'` becomes `"abc"`.

## D43. No semicolons

Decision: a line break ends a statement when the line could end there. A line that ends with an operator, `,`, `(`, `[` or `{` continues. A `;` is a compile error (section 2).

- **Reason.** The rule depends only on the last token of the line, as in Go, so there is one reading.
- **Reason.** The known problems with semicolons in TypeScript come from making them optional with guessing rules. Narra has no optional form.
- **Rejected: required semicolons.** They add a token to every line without removing any ambiguity the line rule leaves.

## D44. Tuples instead of multiple return values

Decision: a function that returns several values returns a tuple or a struct. The caller destructures with `const #[a, b] = f()` (section 4).

- **Reason.** One concept covers both returning several values and grouping values.
- **Reason.** A tuple can be stored, passed and returned like any value.
- **Rejected: multiple return values (Go).** A second concept for the same job, and the values cannot be stored as one value.

## D45. Counting loops through `Range`

Decision: there is no C-style `for (init; cond; step)`. Counting uses `for (const i of Range.upTo(n))` (section 2). Amended by D76: the C-style loop exists; `++` and compound assignment are still open there.

- **Reason.** One loop form, `for ... of`, covers arrays, maps, ranges and generators.
- **Reason.** `Range.upTo(n)` stops before `n`, which fits 0-based indexes.
- **Rejected: `++` and compound assignment.** They are a second way to write `x = x + 1`.

## D46. Function parameters cannot escape by default

Decision: a function-type parameter can only be called during the call. `escaping` allows it to be stored, returned, captured or sent to a thread. `@inline` copies a function's body into callers (section 2).

- **Reason.** A non-escaping closure's environment lives on the caller's stack, so passing an arrow function to `Array.map` allocates nothing.
- **Reason.** The cost of an escaping closure, one heap allocation, is visible in the signature.
- **Reason.** `@inline` on small higher-order functions removes the indirect call, so `Array.map` with an arrow function compiles like a loop.
- **Prior art.** Swift makes closure parameters non-escaping by default and requires `@escaping`.
- **Rejected: escape analysis without a marker.** It would give the same speed. But a change inside one function could turn a free closure into an allocation for every caller, with no visible sign.

## D47. Fixed-size arrays and vector types

Decision: `FixedArray<T, N>` is stored inline, with `const N: int` generic parameters. `std/simd` provides vector types compiled to vector instructions (section 4).

- **Reason.** Small arrays such as a 4x4 matrix or a 16-byte key need no heap allocation and can live inside structs.
- **Reason.** Numeric code needs vector instructions without writing C.

## D48. Regions for bulk allocation

Decision: `Region.run(f)` allocates every heap value created during `f` from one block and frees the block when `f` returns. Values leave only by copying (section 6).

- **Reason.** Work with a clear end, such as one request or one compile, frees thousands of values in one step. No reference counts are updated.
- **Reason.** Copying on the way out keeps the rule that no value can point into freed memory.
- **Cost.** The result is copied out. Large results should be kept small, such as a summary instead of the whole tree.

## D49. Copies on shared values are made visible

Decision: `narra check --report-copies`, `NARRA_TRACE_COPIES=1` and `@unique` let variables show where an update copies a shared value (section 6).

- **Reason.** Copy-on-update is correct but hidden. A copy of a large array in a hot loop is the main source of unexpected slowness in this memory model.
- **Reason.** `@unique` turns the fast path into a checked promise in debug builds, at no cost in release builds.
- **Rejected: making every copy explicit, as Rust's `.clone()`.** Every shared value would need copy calls, which goes against the value model.

Superseded by D94: copies are written as `copy(x)` and a debug build reports the copies on write that no `copy(...)` explains. `--report-copies`, `NARRA_TRACE_COPIES` and `@unique` were never built and are removed.

## D50. Constant parameters and reads go through pointers

Decision: large constant parameters are passed by pointer, and reads of fields and elements go through pointers without copying (section 11).

- **Reason.** Value semantics do not require physical copies. Parameters are constant and access is exclusive, so a pointer is indistinguishable from a copy.
- **Reason.** It removes the cost of passing large structs and of reading from containers, with no syntax.

Amendment (node const-params-pointer, 2026-11-10). Built for parameters; reads of fields and elements still copy:

- **Limit.** A constant parameter whose C value is over 16 bytes is passed as `const T*`: structs, inline enums (D99) and ranges. The C ABI already passes such a struct through memory, as a pointer to a copy the caller makes on arm64 and as a copy on the stack on x86-64, so the copy was all the by-value form added. The size is the C layout (field alignment included), computed once per struct and enum in `finish` (`isPassedByPointer`, `selfhost/src/check.nra`).
- **One convention per type.** The choice depends only on the parameter's type, so direct calls, generic instances, trait implementations, function types, thunks for named functions and arrow functions agree. The generated `retain_*` and `release_*` functions of a large type take a pointer too.
- **Arguments.** A place (a variable, a field path, a parameter) is passed by address. Any other value (a call result, a literal, a constant) is first stored in a temporary of the caller.
- **Where the value is copied.** The callee cannot change the value: the caller keeps it for the call, E0319 rejects passing a variable as `mutable` and using it in another argument, and E0614 rejects a closure argument that assigns a variable another argument uses. Three places keep the parameter past the call or change it, and copy it: a generator or coroutine frame copies it when created; a closure environment copies a captured parameter, and the closure reads it through a pointer to that copy; a function with self tail calls stores each new value in a variable of its own (`<name>_held`) after all new arguments are evaluated, and points the parameter at it.
- **Measured.** `narrac emit selfhost`: 3.16 G to 2.64 G instructions retired (-16.6%) on the batch-4 branch after option-inline and perceus-reuse.

## D51. Parameters reached through a function value are escaping

Decision: a function type has no `escaping` marker, so the parameters of a function value are treated as escaping. An argument passed through a function value must be an escaping value, and a function-type parameter of an arrow function may be stored. This makes D46 stricter.

- **Reason.** The caller of a function value cannot see which function it calls, so it cannot know whether that function keeps an argument. Treating every such parameter as escaping keeps a stack environment from outliving its call.
- **Cost.** A non-escaping parameter cannot be passed through a function value, and an arrow function passed that way gets a heap environment.
- **Rejected: `escaping` inside function types.** It would add a marker to every function type that takes a function, for a rare case.

## D52. A parameter whose type is a type parameter is escaping

Decision: only a parameter written with a function type, such as `f: (x: T) => U`, can be non-escaping. A parameter written as a type parameter, such as `v: T` in `id<T>(v: T): T`, is escaping, even when `T` is a function type at a call. This makes D46 stricter.

- **Reason.** A generic body may store or return a `T`. Whether the parameter escapes must not depend on the type argument, or the same body would be valid for some calls and not others.
- **Cost.** A non-escaping parameter cannot be passed to a generic `T` parameter.

## D53. Tuple literals and types use `#[...]`

Decision: `#[1, "a"]` builds a tuple, `const #[a, b] = t` destructures one, and `#[0, name]` matches one in a pattern. The type is written the same way: `#[int, string]`, nested as in `Array<#[int, string]>`. A one-element `#[T]` is an error, as type and as literal. `[a, b]` is always an array (sections 2 and 4).

- **Reason.** It is the syntax of the JavaScript Records & Tuples proposal, so it reads as a tuple to TypeScript users.
- **Reason.** `#[` starts nothing else, so a tuple never depends on the expected type. It works in generic calls such as `swap(#[1, "a"])`, where no expected type exists.
- **Reason.** Tuples are the one type written with a symbol rather than a name (D23). The type mirrors the literal and the pattern, so `#[int, string]`, `#[1, "a"]` and `#[n, s]` read as one form.
- **Rejected: `Tuple<A, B>` as the type.** It would be the only type whose values are not written with its name, and it keeps `Tuple` out of user code as a name.
- **Rejected: `(a, b)`.** The parser would need lookahead to tell it from the parameters of an arrow function.
- **Rejected: `Tuple(a, b)`.** It is verbose, and it makes `Tuple` a reserved value name.
- **Rejected: `[a, b]` with a tuple as the expected type.** The same text would build an array or a tuple depending on context.
- **Note.** D44 now destructures with `const #[a, b] = f()`, and D23 lists tuples as its one exception. `Tuple` is not a prelude name.
- **Note.** An element is read with `pair[0]`, as in the JavaScript proposal and in TypeScript, so the parser needs no new syntax. Rejected: `pair.0` (Rust, Swift), which would also make the lexer split `pair.0.1`.

## D54. C names are hashed, built in one file

Decision: every C name comes from `compiler/internal/narra/cnames.go`. A declaration is `<kind>_<module hash>_<name>`. An interned type (generic instance, tuple, array, function type, box, enum payload) is `<kind>_<head>_<hash>`, with the hash over `head|arity|arg ids`. The hash is xxHash32, seed 0, written by hand. A process-wide table panics with an internal error if two different things get one name. `Emit` writes a comment table that maps each hashed name to its type. `__` and leading or trailing `_` are allowed in user names.

- **Reason.** Hand-built names caused three collision bugs in wave 2 (`__` in names, prelude names, tuple arity). One builder and a checked table remove the class.
- **Reason.** xxHash32 was 1.7 to 2.7 times faster than FNV-1a on short names, and mixes better. The self-hosted compiler ports it with `*%` and rotates.
- **Rejected: FNV-1a.** In the Go standard library, but slower and weaker.
- **Rejected: `hash/maphash`.** The seed is random per process, so names would change between runs.
- **Cost.** Generated C reads less well. The comment table and the readable head in each name make up for it.

## D55. Fielded enum variants live on the heap, and `Box<T>` exists

Decision: an enum with at least one fielded variant is `{ tag, payload pointer }` (since D99, only when it reaches itself or its payload is over 32 bytes). The payload is counted and shared by copies, and never mutated. Freeing uses an iterative worklist, so deep structures do not overflow the C stack. `Box<T>` is a counted heap cell with a read-only `.value`, for recursive structs (`Box.new(x)`). A user type named `Box` hides the builtin.

- **Reason.** SPEC section 6 asks for this, and recursive enums (an AST) need it.
- **Cost.** Every `Some(x)`, `Ok(x)` and `Err(e)` allocates (since D99, only for heap enums). The user accepted this. It made the `interp` benchmark about 2x slower (0.39 s against 0.24 s) through a retain and release of each loop item; since node `perceus`, for-of items and `is`/`match` bindings borrow, and `interp` runs in 0.16 s.
- **Rejected: inline payload with a pointer only for recursive fields.** Layout would depend on the type's shape, and the self-hosted compiler would need two code paths. Superseded by D99: an enum that does not reach itself and has a payload of up to 32 bytes is stored inline; this decision still holds for the rest.
- **Limit.** User functions still recurse on the C stack; about 20000 deep overflows under ASan. Not solved here.

## D56. Polymorphic recursion is an error

Decision: a generic type whose declaration reaches itself with a larger argument (`G<Array<T>>` inside `G<T>`) is a compile error at the declaration. Recursion with the same arguments is allowed.

- **Reason.** Each instance would need a larger one without end, so monomorphization never stops.
- **Rejected: an instantiation depth limit.** The error would point at a use, not the cause.

## D57. Fallible I/O returns `Result<T, string>`

Decision: `std/fs`, `std/process` and `std/env` return `Result<T, string>` or `Option<string>`. The language has no unit value, so `writeText` returns `Result<int, string>` with the byte count.

- **Reason.** The language has no exceptions, so failure is a value the caller must check.
- **Cost.** `writeText` carries a number nobody needs, until a unit type exists.

## D58. Effects: exported functions declare, private ones are inferred

Decision: an exported function's body effects must be inside its declared `uses`, and callers trust the declared set. Private functions and arrow functions are inferred by a fixed point over all bodies, so recursion works. The entry function has every effect. `alloc` comes from array literals, template strings with parts, string `+` and enum constructors with fields, besides builtins that declare it. An effect variable (`<effect E>`) takes the effects of the arguments passed to the parameters whose function type names it.

- **Reason.** Signatures are the module boundary, so they must state effects. Inside a module, writing `uses` on every helper is noise.
- **Reason.** Making private higher-order functions strict rejected valid code (found in review of #14).
- **Cost.** A call through a function value charges the effects of its type, so a function value is only as precise as the type it is stored in (D86). A top-level constant that holds a function with effects must write them in its type.

## D59. Diagnostic codes come from the message format

Decision: each code names one message format in `diagnostic_codes.go`, and `codeFor` finds the code from the format string at the call site. Codes are grouped in ranges of 100 by area. A test fails when a format has no code. Codes are never renumbered or reused.

- **Reason.** The alternative, a code argument at about 230 error sites, would touch every hot file.
- **Cost.** Two sites with the same format share a code, and a non-literal format gets `E0000`.
- **Note.** Diagnostics have a position, not an extent; JSON `range.end` equals `range.start`. Fix edits have real ranges.

## D60. String positions are `int` byte offsets, and slices copy

Decision: `Str.slice`, `charAt`, `nextChar` and `indexOf` use `int` byte offsets, and panic on a position inside a character. A slice is a copy of the bytes. `StringBuilder` is counted and copied on write when shared. Both differ from the spec and are listed in `compiler/README.md`.

- **Reason.** A view needs a parent pointer in `lstr`, which touches `runtime.h` and every release path. `ByteIndex` as a distinct type can come later, now that distinct types exist.
- **Cost.** Slicing allocates.

## D61. Top-level constants: scalars inline, others built on first use

Decision: a scalar or fieldless-constructor constant is emitted at each use. Any other constant gets storage, a ready flag and a getter that builds it on first use and returns it borrowed; `narra_consts_release` frees it before the leak report. No code runs at startup. A constant may not call a function.

- **Reason.** SPEC section 10 asks for no initialization code. Static C data for structs, arrays and maps would need a constant evaluator and a static layout for counted values.
- **Cost.** The first-use build is not thread safe; threads must revisit it. Errors such as `1 / 0` panic at first use, not at compile time.

## D62. Distinct types convert by name

Decision: `distinct type A = B` makes a new type with `B`'s C type. `A.from(b)` wraps and an unwrap named after the base (`A.toInt`, `A.toString`) unwraps; both compile to the value itself. `==`, `!=` and map keys work when the base supports them. Arithmetic and mixing with the base are type errors. An `opaque` distinct type can only be wrapped in its own module.

- **Reason.** Naming the unwrap after the base reads like the existing `Int.toString` style and needs no generic `unwrap`.
- **Cost.** No `implements` or `derives` on distinct types yet (built, D108), and no generic distinct types.

## D63. The self-hosted compiler uses integer node ids and is ported soon

Decision: the self-hosted parser gives every expression, pattern and local an integer id, and the checker keeps its results in side tables indexed by id. Types are a recursive `enum Type`; structs and enums are an index into a declaration table. The port starts now: features of milestones m4-m7, except opaque-distinct, wait until after self-hosting.

- **Reason.** The Go checker writes into AST nodes and compares pointers. Narra has value semantics and immutable shared enum payloads, so a line-by-line port is impossible there. Ids work today with no new library.
- **Reason.** Each feature built in Go first must also be ported, so building them later costs less.
- **Rejected: `std/arena` first.** Same design with checked handles, but one more M node before the port.
- **Cost.** About 10-15% more code in check and emit, and every pass threads a context struct.

## D64. Maps keep insertion order

Decision: `Map<K, V>` iterates in insertion order. Removal leaves a dead entry; inserting the key again moves it to the end. A key type with a hand-written `implement Equatable` must also implement `Hashable`, and the reverse. Float keys, functions and maps are not keys. `m[k]` is `Option<V>`.

- **Reason.** Deterministic order keeps the compiler's output stable without sorting.
- **Cost.** Maps do not implement the standard traits yet, so a map cannot be compared or used as a key.

## D65. `std/test` catches panics in a child process

Decision: `Test.panics(f)` and `Test.panicsWith(f, message)` run `f` in a child process made with `fork`. The child sets a pipe as the panic hook (`narra_panic_fd` in `runtime.h`), so a panic writes its message there and calls `_exit(101)`; a closure that returns calls `_exit(0)`. The parent reads the pipe until it closes, then waits for the child and checks the exit status. The child has a timeout (`alarm`, 10 seconds or `NARRA_TEST_TIMEOUT`); a closure that runs past it fails the test. Only `narra test` builds contain the std/test runtime and the test blocks. The functions of `std/test` can be called only inside `test` blocks (error E0907); the import itself is allowed anywhere, since test blocks are not part of a non-test build.

- **Reason.** Measured on macOS (clang, `-fsanitize=address,undefined`, `NARRA_CHECK_LEAKS=1`): `testdata/std_test.nra` forks 11 times (once inside a child) and the failure cases in `std_test.go` fork 3 times, all clean in debug and release builds. The child's heap is a copy, so its allocations never reach the parent's leak count, and the child skips the leak report through `_exit`. No change to how values are freed was needed.
- **Rejected: `setjmp`/`longjmp`.** It would skip every release between the panic and the jump, so the leak check fails after each caught panic unless panics unwind, and a panic can leave a value half built. Not tried, because fork passed.
- **Rejected: an import error in non-test builds.** A file with tests and a `run` entry must still build as a program.
- **Leaks on failure.** A failed assertion frees what it built except the final message, which `narra_panic` prints; the process then exits 101 and the leak report, which runs only at the normal end of `main`, does not run.
- **Cost.** POSIX only (`fork`, `alarm`); Windows needs another design (a re-run of the test binary per closure). Each call costs one fork. Changes `f` makes to mutable state are lost, which is also what makes the capture safe. A crash in the child (signal, sanitizer report) fails the test instead of counting as a panic. When a nested `Test.panics` runs past the timeout, the inner and the outer timeouts race. An empty `message` matches any panic, and the match stops at a NUL byte.

## D66. The test runner forks per test and golden files live in `testdata/`

Decision: the main of a test build runs each `test` block in a child process (`fork`). The child sets `narra_panic_fd` and `narra_panic_pos`, so a panic writes `<position>\n<message>` to a pipe and exits 101; the parent prints `ok - <name>` or `FAIL - <name>: <message> (<position>)`, goes on, and ends with `N passed, M failed` and exit code 1 when a test failed. The child frees its constants and runs the leak check after its test, so a leak fails that test alone. `--filter <text>`, `--time` and `--golden-dir <dir>` are arguments of the test binary; `narra test` passes them on (the golden directory is computed by the Go driver, so the generated C holds no path and stays byte-identical across machines). Each test has an `alarm` limit (10 seconds or `NARRA_TEST_TIMEOUT`); the parent reports `timed out after N s`. A filter with no match prints `no tests match "<text>"` and exits 1; a file with no tests exits 0. `Test.golden(name, actual)` compares with `<dir>/<name>.golden` (`<package>/testdata`, or the file's directory for a single file; without the flag, `testdata/` relative to the working directory) and rewrites the file when `NARRA_UPDATE_GOLDEN=1`.

- **Reason.** D65 already showed fork works under the sanitizers and the leak check. Per-test processes give isolation and let the run continue, with no change to how panics unwind.
- **Filter by argv.** An environment variable would leak into the nested processes of the tests and is invisible in a command line; argv also lets the binary run alone.
- **Timing is opt-in.** The output is compared in tests, and times vary.
- **`io` in test blocks.** The effect pass already ignores the effects of a `test` body (`Test.panics` declares `io` and is called freely), so `Test.golden` declares `io` and needs no `uses`. Helper functions called from a test still declare their own effects.
- **Golden location.** `testdata/` keeps the files out of `src/`. A single file has no package, so the golden file sits next to it.
- **Cost.** POSIX only. A `Test.panics` closure inside a test gets `min(limit, left - 1 s)`, at least 1 s, where `left` is read from the pending alarm, so the closure's timeout message comes before the test's; a test limit of 1 s therefore races (use 2 s or more). Crashes print the signal name (`strsignal`). The runner has a test-only hook, `NARRA_TEST_FORCE_LEAK`, to test the leak report; it exists in every test binary, and setting it makes every test fail with a false leak. `--golden-dir` is a public option of the test binary, documented in SPEC section 13; `narra test` sets it. The failed-assertion message is no longer on stderr as `panic at ...`; it is on standard output in the `FAIL` line. The exit code of a failed run changed from 101 to 1. Generated C of every test build changed.

## D67. Prelude types do not bring their functions

Decision: the only global functions are `print`, `assert` and `panic`. The functions of `Array`, `Map`, `Range` and `Box` are the std modules `std/array`, `std/map`, `std/range` and `std/box`, registered with `registerModule` like `std/string`. The types and constructors stay global because syntax produces them (`[]`, map literals, `is`, `match`, `derives`). An import alias may equal a type name, so `import * as Array from "std/array"` makes `Array.push` the module function while `Array<int>` stays the type. A call such as `Array.push` without the import is error E0908 with a fix that adds the import line. A user type named `Box` or `Map` still hides the prelude type, and then `Box.new` is the user's, unless the module also writes `import * as Box from "std/box"`: then the module wins for `Box.new`.

- **Reason.** User decision 2026-10-30: do not expose global functions. One rule for every function: it comes from a module.
- **Rejected: a prelude that re-exports the functions as globals.** It keeps the old behavior for no gain.
- **Cost.** Every program that uses these functions gains one to four import lines. Positions in goldens shift.
- **Note.** D72 replaces the alias rule: an import alias may no longer equal a type name, and the std modules take plural aliases (`import * as Arrays from "std/array"`). The case where `import * as Box` wins over a user type `Box` is gone.

## D68. Type arguments on map and generic struct literals

Decision: a map literal or a generic struct literal may carry type arguments after the name: `Map<string, int> {}`, `Map<string, int> { "a": 1 }`, `Pair<int, string> { a: 1, b: "x" }`. The type arguments fix the type of the literal, so it no longer depends on the expected type. When both exist they must be the same type (E0237). Arity, the map key rules and each key, value and field are checked against the written types with the same code that checks a type annotation. `Map<K, V> { ...spread }` does not exist, and `Pair<A, B> { ...s, a: 1 }` requires `s` to have the written type. Arrays get no form: `let xs: Array<int> = []` stays.

The parser decides by lookahead. After a capitalized name, a `<` starts type arguments only when the tokens up to the matching `>` are names, `.`, `,`, `#[`, `]` and nested `<...>` (`>>` closes two levels) and the next token is `{`. Otherwise `<` is the less-than operator. Function types are not allowed inside literal type arguments, because `(` is not in the set; they can still be passed through an annotation.

- **Reason.** User decision 2026-10-30: an empty map needs no variable annotation and can be passed straight to a generic function. D23 keeps names over symbols, so there is no `{K: V}` type and no `#{...}` literal.
- **Rejected: a bare `Map {}` with inference from later uses.** The checker checks one expression at a time and has no place to hold an unresolved type.
- **Rejected: `Map::<K, V> {}` or `Map.of<K, V>()`.** The first adds a token the language does not have. The second is a function call, and the literal form already exists.
- **Cost.** Two parsers carry the lookahead. A comparison such as `A < B > { ... }` between capitalized names is read as a literal, which no valid program writes.

## D69. `std/option` and `std/result` helpers

Decision: two std modules, `std/option` (`map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr`) and `std/result` (`map`, `mapErr`, `andThen`, `withDefault`, `ok`, `err`), registered like `std/array` (D67): the alias equals the type name and the call needs the import. Callbacks are non-escaping function parameters whose effects flow to the call, like `Array.map` with `<effect E>`. A capturing arrow function passed directly gets a stack environment, so a call allocates only the new `Some`, `Ok` or `Err` payload and nothing for the closure. There is no `unwrap` or `expect`.

- **Reason.** User decision 2026-10-30. `SPEC.md` already documented `Option.withDefault`. The helpers never return early from the caller, so section 5 stays true: an error path that stops is an `is` or `match`.
- **Implementation.** Builtins with a new field `Builtin.Callbacks` (argument indexes); the effect pass adds the effects of those arguments to the call. Five new diagnostics, E0909 to E0913.
- **Rejected: `unwrap` and `expect`.** They hide a panic where the spec wants a visible branch.
- **Cost.** `Option.map` and the other helpers that build a value allocate its payload (`alloc` effect). A callback passed as a stored local (`const f = ...`) is a heap closure, created once, not per call.
- **Note.** Under D72 the alias is plural: `import * as Options from "std/option"`, `import * as Results from "std/result"`.

## D70. The self-hosted checker keeps body results per checked body

Decision: the self-hosted body checker keeps the type of each expression, what each name or call resolves to, and the locals in tables of one checked body (`BodyInfo` in `selfhost/src/check.nra`), indexed by `id - first`, where the body's node ids lie in `[first, first + length)`. A function, test, constant value, field default and each instance of a generic function is one body. Tables that do not depend on the instance, such as `paramTypes`, stay package-wide and are indexed by id.

- **Reason.** Go checks a generic function once per instance on a cloned AST, so the same node gets a different type in each instance. Narra does not clone: an instance reuses the same syntax and ids, so the results must be keyed by body as well as by id.
- **Reason.** A table sized by the package's id count for every body would make checking quadratic in the number of bodies. The id range of one body is about its own size.
- **Rejected: one package-wide table with a `Map<int, Type>` per instance.** It needs the same body key, and every lookup hashes.
- **Cost.** Each body walks its nodes once to find its id range, and the passes after the checker (the emitter) must look up the body of the code they emit.

## D71. Enum constructors are always qualified

Decision: a user enum's constructors are written with the enum's name, in expressions and in patterns: `SchoolPerson.Teacher`, `SchoolPerson.Student("x")`, `SchoolPerson.Teacher =>`. Importing the enum is enough. A bare constructor is a compile error with a fix that qualifies it. `Some`, `None`, `Ok` and `Err` stay bare as prelude names. Constructors of different enums never clash, and two constructors with the same name in one enum are an error (section 3). Debate: `roadmaps/debates/qualified-enum-constructors/debate.md`.

- **Reason.** A bare constructor can be hidden by a local with the same name, which then compiles with a different meaning when the types match. D2 forbids that.
- **Reason.** TypeScript writes enum members as `Color.Red`, and Rust writes `Enum::Variant`. Both readers expect the qualified form.
- **Reason.** It removes the ambiguity rule the spec never finished. Two enums may share a constructor name, and a struct `Circle` can sit next to `Shape.Circle(c: Circle)`.
- **Rejected: bare names in patterns, resolved from the matched type.** User decision 2026-10-31: one rule for every position is worth the longer match arms.
- **Rejected: a Swift-style `.Teacher` shorthand where the type is known.** It has no TypeScript precedent. User decision 2026-10-31: never.
- **Rejected: keeping bare constructors with a written ambiguity rule.** It keeps the shadowing hazard.
- **Cost.** About 1,100 sites in `selfhost/` and `compiler/testdata/` change, through the compiler's own fix. Names that were prefixed to avoid clashes (`Type.TStruct`, `Token.TokIdent`) stutter until they are renamed.

## D72. Import aliases do not equal type names, and named imports can rename

Decision: an import alias may not equal a type name visible in the module (prelude, declared or imported), and two imports may not share an alias. The std modules of the capitalized prelude types take plural aliases by convention: `Arrays`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`; their paths do not change. Named imports accept `{ a as b }`. Importing a constructor by name is an error whose fix imports its enum (section 3). Debate: `roadmaps/debates/qualified-enum-constructors/debate.md`.

- **Reason.** Under D71, `Shape.Circle` is a constructor. An alias `Shape` would give the same text a second meaning. TypeScript also rejects an import that conflicts with a local declaration.
- **Reason.** Plural is one rule the fix can suggest for every module, as in Java's `Arrays` and Guava's `Maps`. It is only a convention: any alias that follows the rules works.
- **Reason.** `map` exists in `std/array`, `std/option` and `std/result`, so named imports need renaming to be usable. `as` means the same thing as in TypeScript.
- **Rejected: lowercase aliases (`array.push`).** They break the capitalized style of `Str` and `Int` and clash with locals named `array` or `map`.
- **Rejected: renaming the module paths to `std/arrays` and the rest.** The alias is enough.
- **Cost.** 107 imports and about 700 calls in `selfhost/` and `compiler/testdata/` change, through the compiler's own fix. Amends D25 and D67.

## D73. Sum types keep the `enum` keyword

Decision: sum types are declared with `enum`, not with ML-style `type T = | A | B(x: int)`. Structs keep `struct`, not `type T = { ... }`. `Red = 1` in an enum without `@repr` gets an error whose fix adds `@repr(i32)` (section 9). Debate: `roadmaps/debates/sum-type-syntax/debate.md`.

- **Reason.** In TypeScript, `type Shape = Circle | Square` is a union of existing types. The ML form would make the same text declare new constructors, which D2 forbids.
- **Reason.** Every TypeScript enum that Narra accepts keeps its meaning: fieldless members, read as `Color.Red` under D71. Payloads are new syntax with no TypeScript meaning. Rust and Swift use `enum` the same way.
- **Reason.** `type` already declares aliases. Making it declare a nominal type depending on the first token after `=` gives it two meanings.
- **Rejected: `type Point = { x: f64 }` for structs.** In TypeScript that is a structural alias, while Narra structs are nominal. No case in the checker or emitter would go away.
- **Cost.** None beyond the `Red = 1` diagnostic.

## D74. No local function declarations

Decision: a function cannot be declared inside another function. `function` in a function body is a compile error (E0930) that says to declare the function at module level as a private function and pass the enclosing values as arguments. The error offers no automatic fix. Arrow functions bound to `const` stay the way to write a helper that captures (section 2). Debate: `roadmaps/selfhost/debate/local-functions/debate.md`.

- **Reason.** In TypeScript a nested function captures the enclosing locals. A non-capturing version reads a module-level name where TypeScript reads the local that shadows it, so the same text would print 2 in Narra and 11 in TypeScript. D2 forbids that.
- **Reason.** Module-level private functions already recurse and call each other. The self-hosted compiler writes all 648 of its helpers that way.
- **Reason.** An error can later become local functions without changing any program that compiles. Shipping non-capturing local functions and adding capture later would change working programs.
- **Rejected: non-capturing local functions, hoisted within their block (PR #66).** Renaming at parse time left six defects, and fixing them needs scope analysis in both compilers. The design is recorded in `roadmaps/selfhost/nodes/local-functions/outcome.md`.
- **Rejected: capturing local functions.** Nothing uses them, and they would add capture analysis, an ordering rule for hoisting and environments to both compilers.
- **Rejected: a fix that moves the function to module level.** It changes the meaning when the body reads a shadowed name, and it breaks helpers that use the enclosing type parameters.
- **Cost.** TypeScript code with nested functions must be rewritten by hand.

## D75. Block expressions, and `if` and `match` as values

Decision: a block has two forms. A block in statement position is a block statement, as in TypeScript. A block in value position (after `=`, as an argument, as an operand, as an `if` branch or `match` arm used as a value) is a block expression. A block expression produces its value with `return v`. `return`, `break` and `continue` stop at the nearest block expression, as at a function body: a `return` anywhere inside it, also inside nested block statements, loops and `match` arms, ends the block expression, and a `break` or `continue` to a loop outside it is a compile error. `if (c) { ... } else { ... }` and `match` with block arms produce values when they are in value position. Decided by the maintainer on 2026-11-04.

- **Reason.** The two forms are told apart by position, so each `return` has one target: the nearest enclosing function body or block expression.
- **Reason.** Narra never returns implicitly. A block's value comes from `return`, not from its last expression as in Rust or ML.
- **Reason.** Narra has no object literals. Struct literals are named (`P { x: 1 }`), or context-typed when the first tokens after `{` are `}`, `...` or `name:` (`const p: P = { x: 1 }`; field shorthand `{ x }` is already an error). Maps write `Map<K, V> { ... }`. Every other `{` in value position starts a block expression.
- **Rejected: the last expression of a block as its value.** That is an implicit return, which Narra does not have.
- **Rejected: a keyword for block expressions (`do { ... }`).** Position already tells the forms apart.
- **Rejected: `break` and `continue` crossing a block expression.** The block would then have no value on that path. One boundary rule for all three jumps is simpler.
- **Cost.** A block expression cannot return from the enclosing function, as in JavaScript's `(() => { ... })()`. TypeScript's `const o = { a }` is an object literal there and a block expression with no `return` here, which is a compile error. Every path of a block expression must end in `return`, and every returned value must have the same type.
- **Ternary.** The conditional expression `c ? a : b` is part of this decision (maintainer, 2026-11-04). It is an `if` in value position with one expression per branch and the TypeScript precedence: looser than `||`, right-associative.

## D76. C-style `for` loops and compound assignment

Decision: Narra has C-style `for (init; cond; step) { ... }`, including `for (;;)`, next to `for ... of`. It also has compound assignment (`+=`, `-=`, `*=`, `/=`, `%=` and the bitwise and shift forms) and `++`/`--`. All are statements with the same meaning as in TypeScript. Amends D45, whose rejections of the C-style loop and of `++` and compound assignment no longer hold. Decided by the maintainer on 2026-11-04.

- **Reason.** Loops whose step is not a fixed range (two indexes moving together, a step computed in the body, scanning until a condition) read more directly as `for (init; cond; step)` than as `while` with the step at the end of the body.
- **Reason.** TypeScript code and models write these forms constantly. Accepting them with TypeScript's meaning is allowed by D2, and spares a rewrite.
- **Kept from D45.** `for (const i of Ranges.upTo(n))` stays the usual way to count over indexes.
- **Cost.** Two loop forms where D45 had one. The details (whether `init` must be `let`, whether `++` has a value, `continue` jumping to the step, labels, the semicolon rule inside the parentheses) are settled by nodes c-style-for and compound-assign.
- **Built: the C-style loop.** `init` is `let` or empty (`const` is E0960). `condition` is a `bool` or empty (E0959). `step` is an assignment, a call or empty (E0961). `continue` runs the step. The `;` is a token only inside parentheses, so `;` elsewhere is still E0003. The name from `init` ends with the loop.

Amendment (node compound-assign, 2026-11-09). Rules for compound assignment and `++`/`--`:

- **Meaning.** `x op= e` means `x = x op e` with the target place evaluated once, so `items[next()] += 1` and `items[next()]++` call `next` once. As for `=`, the value is evaluated before the target place. The operators are `+= -= *= /= %= &= |= ^= <<= >>=` and the wrapping forms `+%= -%= *%=`. Changed by D92: the wrapping forms `+%= -%= *%=` do not exist; `+= -= *=` wrap.
- **Statements only.** They have no value and cannot be nested in an expression (E0964). Only the postfix spelling `x++` and `x--` exists; `++x` and `--x` are an error with a hint (E0964), so `--x` no longer parses as a double negation.
- **Targets.** Every target `=` accepts: a `let`, a field or element of a changeable value (D81), the contents of a `mutable` parameter. A `const` binding and a whole `mutable` parameter are rejected as for `=`.
- **Types.** The value has the type of the target (E0962). `++` and `--` take an integer of any width (E0963). Overflow, division by zero and shift checks are those of the binary operator.
- **Strings.** `s += t` joins strings and releases the old value; `+=` is the only compound form on a `string`, and `a + b` stays an error.
- **Map index.** `counts[key] += 1` panics when the key is missing, since `counts[key]` is an `Option` and the old value is needed. Insert the key first with `Maps.set`.

## D77. Non-escaping closures capture `let` and `mutable` by reference

Decision: an arrow function passed straight to a non-escaping function-type parameter may read and assign the `let` variables and `mutable` parameters of the enclosing function, by reference. Every `let` and `mutable` it uses is captured by reference, also when it only reads it. `const` bindings and non-mutable parameters keep the by-value capture. Every arrow function between the use and the declaration must be of this kind; any other closure (stored, returned, passed to an `escaping` parameter or through a function value) still cannot capture them. A variable that such a closure assigns is borrowed for the call: it cannot be passed as `mutable`, used in another argument, or captured by another closure argument. Decided by the maintainer on 2026-11-05.

- **Reason.** Same semantics as TypeScript for the common callback shapes.
- **Reason.** One simple rule for which captures are by reference: `let` and `mutable`.
- **Reason.** Cheaper than copying: the environment is on the caller's stack (D46), and a by-reference capture needs no retain or release of heap values.
- **Reason for the borrow rule.** Constant parameters are passed by pointer (D50), so `each(items, (x) => { items = [] })` would free the array the callee is reading.
- **Why it is sound.** A non-escaping closure can only be called during the call it is passed to, so a pointer to the caller's variable cannot dangle. A `let` still cannot be stored, so no cycles form.
- **What non-escaping means.** The callback runs on the calling thread during the call. A future std function that runs a callback on another thread (for example `Parallel.map`) must take an `escaping` parameter.
- **Prior art.** Swift's non-escaping closures capture `inout` parameters.
- **Rejected: by-value capture of a `let` that a closure only reads.** It would look the same only because of the borrow rule, so it buys nothing.
- **Codes.** E0612 and E0613 replace E0603 and E0604, whose rows stay unused. E0614 is the borrow rule.

## D78. Generator types carry effects, charged where the body runs

Decision: a generator type carries the effects of its body, written as on a function type: `Generator<T> uses io`. `Generator<T>` alone is pure. Creating a generator charges only `alloc`, for its frame. The body's effects are charged where it runs: a `for ... of` over the generator, a `yield*` of it, and `Co.resume`. An exported `function*` writes them on its return type; a private one gets them inferred into its type. A value with more effects than the expected generator type is a compile error; one with fewer is allowed. A parameter of generator type binds an effect variable, as a function-type parameter does (D14): `export function sum<effect E>(numbers: Generator<int> uses E): int uses E`. When a result type names the callee's effect variables (`relay<effect E>(g: Generator<int> uses E): Generator<int> uses E, alloc`), each call's result has the effects its arguments bind, so the variable's name never leaves the callee. `Generator<T>` is a value type with no `==`, hashing or `Show`. `yield` and `yield*` are statements, valid only directly in a `function*` body (section 8). Decided by the maintainer on 2026-11-05.

- **Reason.** Calling a generator function runs none of its body, so the call does no I/O. The I/O happens at each resume, which may be in another function than the one that made the generator.
- **Reason.** The effect is visible in the type that reaches the code that runs it, so a function that iterates a generator states what the iteration does.
- **Reason.** A statement-only `yield` keeps the state machine simple: no expression is ever suspended halfway.
- **Rejected: charging the body's effects at creation.** A pure function could then run I/O through a generator value it was given, with no effect in its signature.
- **Cost.** A function that returns a generator it did not declare with `function*` cannot write `uses` for itself after the result type, since the list belongs to the type; it writes a second list (`(): Generator<T> uses io uses alloc`). A private function's parameter of type `Generator<T>` accepts only pure generators; it does not act as an implicit effect variable, as a private function-type parameter does.

## D79. Coroutines are a separate type from generators

Decision: `Coroutine<In, Out>` is its own type, declared with `function*` and a `Coroutine<In, Out>` return type, apart from `Generator<T>`. Inside a coroutine, `const input = yield output` binds the value passed to `Co.resume(mutable g, input)`; the input of the first resume is dropped. `Co.resume(mutable g)` resumes a `Generator<T>` and `Co.resume(mutable g, input)` a `Coroutine<In, Out>`; both return `Step<Out>` from `std/coroutine`, whose constructors are `Step.Yielded(value)`, `Step.Done` and `Step.Panicked(message)`. Requested by the maintainer on 2026-11-07; the name `Step` was decided on 2026-11-05.

- **Reason.** `for ... of` has no input to send. TypeScript's `Generator<T, TReturn, TNext>` lets a loop send `undefined`; Narra has no `undefined` (D21), so a loop over a type that needs an input would have nothing to pass.
- **Reason.** Only coroutines need `yield` in a `const` initializer. Keeping the value-carrying `yield` to coroutines leaves plain generators with the statement-only `yield` of D78, so their state machines and their C are unchanged.
- **Reason.** The common case, a lazy sequence, keeps one type parameter: `Generator<T>`.
- **Rejected: one generator type with an input parameter, as in TypeScript.** `Generator<T, TNext>` would put an input type on every generator and make `for ... of` and `yield*` check it. The cost of the choice is two types, and `Co.resume` takes both (one argument for a generator, two for a coroutine).
- **Statement form.** `const x = yield v` is a statement with a binding, not an expression, so no expression is ever suspended halfway (D78). The binding is a `const` or `let` with one name and no annotation; its type is `In`.
- **Panics.** `Panicked` is only reported in test builds, where `Co.resume` catches a panic of the coroutine with `setjmp`. D65 rejected `setjmp` for `Test.panics`, which had a fork; here the coroutine runs in the caller's process and its state must survive, so `setjmp` is the one option. The coroutine is finished afterwards and what it held is not freed; the live-allocation count is reset to its value before the resume, so the leak check of a test stays quiet. Outside test builds a panic aborts, as everywhere.
- **Cost.** Every program carries `setjmp.h`, the panic jump pointer, the 256-byte panic text buffer and `narra_co_next`, in non-test builds too, and every generator frame grows by one field (`in`). The design stays; a build flag could drop them later. Two types, and `Co.resume` is one name for two functions chosen by argument count (the compiler maps it to `resumeGenerator` or `resumeCoroutine` of the source module `std/coroutine`; `import { resume }` by name is an error with a hint to use `import * as Co`). A resume allocates the `Step` payload, so `Co.resume` charges `alloc` besides the effects of the generator type. Memory held by a panicked coroutine leaks until the test process ends. Diagnostics name `Co.resume` for a missing `mutable` and for effects; a generator passed with an input (or a coroutine without one) still reports the mapped function's parameter type, `expected Coroutine<int, int> uses E, found Generator<int>`, because the choice of function is by argument count and the argument is not inspected first.

## D80. Cycles through identity types leak and are caught by the leak check

Decision: Narra has no cycle collector and no type rule against cycles. A reference cycle through identity objects (`Shared<T>`, `Sender<T>`, `Receiver<T>`, `Task<T>` and the planned `Cell<T>`) is a leak: memory safety holds, but the objects in the cycle are never freed and the resources they hold never close. Plain values cannot form a cycle, so every cycle passes through an identity object, and the debug leak check (`NARRA_CHECK_LEAKS`, section 13) reports it at exit. SPEC section 6 "Cycles" names the two patterns that leak and how to break them. Decided by the orchestrator on 2026-11-09, under the gate the maintainer delegated for batch 1.

- **Why plain values cannot form a cycle.** To store a reference to a value inside that value, the program holds the reference, so the value's count is at least 2 during the write, and the write copies first (section 6, In-place updates). A `let` is never stored by reference, and only a non-escaping closure captures one (D77). A `const` cannot name itself in its own initializer. The spike checked this with the self-hosted compiler: an array field that receives its own struct, a `Box` stored in its own payload and a closure stored in the struct it captures all run clean under the leak check, and the self-referencing and escaping forms are compile errors.
- **Reason.** Zero cost for programs that use no identity types, and no pause or extra header for those that do. Frees and resource closes stay deterministic, as section 6 promises.
- **Reason.** Identity objects are created by hand and are visible in the source, so a leak always has a named cause, and a test that creates one fails.
- **Prior art.** Swift (ARC) and Rust (`Rc`, `Arc`) also leak reference cycles and treat them as safe; both add weak references to break them. Nim's ORC adds a trial-deletion collector on top of reference counting, run only on objects whose type can form a cycle. CPython runs a generational cycle collector over every container object.
- **Rejected: a type rule that bans identity types inside identity payloads.** It breaks request and reply channels (`Sender<#[Request, Sender<Reply>]>`) and stops a `Cell` from holding a closure, the main reason `Cell` exists.
- **Rejected for now: a trial-deletion collector for identity objects (Nim ORC, CPython).** Each identity decrement that leaves a count above zero would record a candidate, and objects shared across threads would need a concurrent collector. It can be added later without changing any program, because it would only free memory that this rule leaks.
- **Rejected: weak references (`Weak<T>`).** They add a second handle kind for back edges, which `std/arena` handles already cover.
- **Cost.** A cycle in a long-running program grows its memory, and release builds report nothing. The leak check reports a count, not the objects, so finding the cause is by hand.
- **Cost.** A cycle that holds the last `Sender` of a channel keeps the channel open, so a receiver that reads until the channel closes waits forever. In a test this shows as a timeout, not as a leak report.
- **Limit.** The check runs at exit. In test builds, memory allocated inside a coroutine that panics is taken off the count (D79), so a cycle built there is not reported.

## D81. `mutable` changes the contents of a value, never the binding

Decision: `mutable` reaches the contents of a value, its fields and elements, and never the variable that holds it. `const` fixes the name only, as in TypeScript, and `let` adds reassignment of the name. Amends D27 (the argument no longer has to be a `let`, and the callee can no longer replace the whole value) and the deep `const` of section 2.

**Decided by the maintainer on 2026-11-07:** rules 1 to 5, including the type list of rule 1 and read-only plain parameters in rule 5.

1. **Contents types.** A `mutable` parameter's type must be a type whose contents change in place: a struct (an `opaque struct` too), `Array<T>`, `FixedArray<T, N>`, `Map<K, V>`, `Box<T>`, `StringBuilder`, `Generator<T>` or `Coroutine<In, Out>`. Any other type is a compile error at the parameter: numbers, `bool`, `char`, `string`, `bytes`, distinct types, enums (`Option` and `Result` included), tuples, function types, the prelude `Range`, and the identity types (`Sender`, `Receiver`, `Shared`, `Atomic`).
2. **Inside the callee.** Through a `mutable` parameter the callee may assign fields and elements at any depth, call the std functions that change contents (`Arrays.push`, `Maps.set`, `Str.append`, `Co.resume`) and pass the value or a part of it on as `mutable`. Assigning the parameter itself (`player = other`, `count += 1`) is a compile error.
3. **Arguments.** The argument may be a local `const`, a local `let` or a `mutable` parameter that holds a contents type, or a field or element path of one. A literal, a call result, a plain parameter and a top-level constant are compile errors. Pattern, `for ... of` and destructured bindings are the open question below.
4. **`const`.** A local `const` fixes the name. Its fields and elements can be assigned directly (`hero.hp = 1`, `items[0] = 2`, `b.value = v`), and it can be passed as `mutable`, so `mutable` adds no power that direct code lacks. A local `let` adds reassignment of the name. `@unique` may be written on either.
5. **Copies stay copies.** `const copy = hero` is a copy (copy on write, as before); changing `hero` afterwards does not change `copy`. A plain parameter is still a read-only copy: its fields cannot be assigned and it cannot be passed as `mutable`. Amended by D94: `const copy = hero` followed by a write to `hero` while `copy` is still used is a compile error (E0978), because the write would copy silently. The copy is written `copy(hero)`.

**Proposed by the spike and decided by the maintainer on 2026-11-07:** rules 6 to 10.

6. **Top-level constants are fixed in depth.** A top-level `const` cannot be changed at any depth or passed as `mutable` (E0290, as before). There are no globals, and a changeable top-level value would be one.
7. **Enums and tuples.** A struct inside an enum payload or a tuple is not reachable as a place: enum fields are fixed (E0303) and tuples have no element assignment. A struct field that holds an enum or a tuple can be replaced whole (`state.current = Some(next)`). To change a struct held in an `Option`, read it with `is`, build the changed copy, and store it back.
8. **Generics.** A parameter written `mutable value: T` with a bare type parameter `T` is a compile error, whatever its constraint. A generic instance of a contents type is fine (`mutable items: Array<T>`, `mutable g: Generator<T>`). In a trait, `mutable value: Self` is allowed, and each `implement` of that trait must be for a contents type (checked at the `implement` block). In a generic function bounded by such a trait, a local `const` or `let` of type `T` may be passed as `mutable` to that trait's `mutable Self` functions, and only to them: the bound guarantees a contents type. A `Dynamic<Trait>` cannot be passed to a `mutable Self` function (compile error at the call), because changing the boxed value through a shared box has no rule yet.
9. **Closures (amends D77).** A non-escaping closure that assigns a field or element of a captured local `const`, or passes it as `mutable`, directly or through a nested closure, captures that `const` by reference. Every closure between the change and the declaration captures it by reference, as for a `let`, and the `const` is borrowed for the call (E0614). A `const` that no closure in the chain changes is still captured by value. A closure that is not passed straight to a non-escaping parameter cannot change a captured value at all. Reason for the chain rule: a stack environment stores a by-value capture as an unretained bitwise copy, so an inner closure that pushes to the outer closure's copy could reallocate the array and leave the caller's variable pointing at freed memory.
10. **Exclusive access.** E0319, E0929 and E0609 apply to every binding passed as `mutable`, `const` included. `x = f(mutable x)` stays E0929.

**Bindings that copy part of another value are read-only (decided by the maintainer on 2026-11-07).** `match` and `is` pattern bindings (`x is Some(v)`), `for (const p of players)` bindings and destructured bindings (`const { a, b } = s`, `const #[x, y] = t`) cannot have fields or elements assigned and cannot be passed as `mutable`. On main they are owned, retained copies (`emit_match.go`), so a write would change only the copy. `for (const p of players) { p.hp = 0 }` is valid TypeScript that changes the array, so it is a compile error whose hint writes through the source (`players[index].hp = 0`); `const copy = p` first still allows a changed copy.

- **Reason.** It keeps the promise of SPEC sections 2 and 14 that valid TypeScript with another meaning is a compile error, for exactly the bindings where a TypeScript reader expects the original to change.
- **Rejected: these bindings behave like `const`** (a write changes the copy). One rule for every keyword binding, but `for (const p of players) { p.hp = 0 }` would compile and silently leave `players` unchanged.
- **Cost.** "Any binding may be passed as `mutable`" narrows to `const` and `let` declarations with a plain name.
- **Later.** A `mutable` marker on these bindings (`for (mutable p of players)`, `x is Some(mutable y)`), meaning a reference to the part rather than a copy, can be added on top of this rule without breaking programs. Node mutable-markers explores where it fits.

**Left to the node that builds `Shared`.** SPEC section 8 shows `Shared.update(hits, (mutable count) => { count = count + 1 })`, an `int` `mutable` arrow parameter that rule 1, rule 2 and E0028 reject. `Shared` is not built; that node decides the form, for example a callback that returns the new value.

- **Reason.** One meaning per keyword: `const` and `let` are about the name, `mutable` is about the contents. TypeScript programmers already read `const` this way.
- **Reason.** Requiring a `let` only to pass a value as `mutable` gave `let` two jobs. In `selfhost/src`, at least 399 of the 1016 `let` locals are never reassigned by name; they are `let` only so their contents can change.
- **Reason.** A `mutable` parameter of type `int` was an output parameter that replaced the caller's variable. Returning the value says the same thing in the signature. The corpus has three (outcome of node mutable-reference).
- **Reason for rule 8.** No trait in Narra guarantees a contents type, and no code in the four trees has a bare generic `mutable` parameter. A compiler-known marker trait can be added when a program needs one.
- **Rejected: keep `const` deeply immutable and require `let`.** The rule of D27. It keeps `const` meaning "never changes" but makes `let` the only way to change contents, so `let` marks two different things.
- **Rejected: a third binding kind `mutable name = value`.** A local reference binding. It adds a third keyword where two suffice, and it raises the question of what such a binding aliases. It goes to the inbox with `x = f(mutable x)`, to weigh again once this model ships.
- **Prior art.** TypeScript `const` fixes the name and not the object. Swift `inout` and Hylo `inout` pass exclusive access to a value for one call. Hylo's `inout` reaches the whole value; Narra's `mutable` reaches only its contents.
- **Cost.** `const` no longer means deeply immutable: a reader must look for field writes and `mutable` arguments to know a `const` struct does not change.
- **Cost.** `const b = a` followed by `a.hp = 1` now compiles and leaves `b` unchanged, where TypeScript would change both. It has no diagnostic: finding it needs a check of every copy across the function. SPEC sections 2 and 14 now name it as the one exception to "valid TypeScript with another meaning is a compile error". The case was already open for `let`.
- **Cost.** Output parameters of scalar types are gone; a function returns the value, or wraps the counter in a struct. The checker splits "the name can be reassigned" from "the contents can change".

## D82. Sized integers follow Go and Rust: context-typed literals, no implicit conversion

Decision: `i8`, `i16`, `i32`, `u8`, `u16`, `u32` and `u64` are separate types; `i64` is another name for `int`. SPEC section 4 "Numbers" states the rules. Conversions are functions named after the target type, one module per type, built by node sized-conversions.

**Chosen by the orchestrator on 2026-11-09 (the maintainer delegated this batch's gate decisions):**

1. **Literals.** A literal takes the sized integer type its context expects, else `int`, and must fit (E0294). A negative literal under an unsigned type is E0294, not E0296; E0296 is for `-x` on an unsigned value that is not a literal. `-128` is a valid `i8`: a literal directly under unary minus is checked with its sign.
2. **Context.** An operator's wanted type types the left operand, so `const m: u8 = 1 << 3` works. A literal on the left of a binary operator takes the type of the right operand (`1 + d` with `d: u8`).
3. **No mixing.** Two integer types that differ, one of them sized, are E0295 for arithmetic, bitwise operators, shifts and comparison, including `==`. The message names `<Target>.from`, `.tryFrom` and `.wrap` of the left type. Bitwise operators and shifts need operands of one integer type. Shift counts have the type of the value.
4. **Overflow.** Checked at the type's own width with `__builtin_*_overflow` on that C type, panicking in debug builds and wrapping in release builds. `MIN / -1` follows `int`. Division and remainder by zero panic in every build. Shifts by the width or more, or by a negative count, panic. Changed by D92: arithmetic wraps in every build and nothing panics on overflow.
5. **Runtime.** One macro in `runtime.h` defines the add, sub, mul, neg, div, rem, shift, wrap, `eq`, `cmp`, `hash` and `show` functions of each type, suffixed with the type (`narra_add_u8`). `hash` goes through the `int` hash, so equal values hash alike.
6. **Printing.** Template string parts stay strings (E0201). A sized value prints through `Show.show(x)`; `U8.toString` comes with sized-conversions.

- **Reason.** Go and Rust: mixed-type arithmetic hides where values change size, and the lossy cases (`wrap`) are visible at the call.
- **Rejected: implicit widening as in C.** Hides size changes.
- **Rejected: `as` casts.** A keyword with one form for lossless and lossy conversions.
- **Rejected: `Int.toI32(x)` names in `std/int`.** More than 60 names for pairs of types.
- **Rejected: literal suffixes (`200u8`).** Context typing covers them; they can be added later without breaking programs.
- **Rejected: a negative literal under an unsigned type reports E0296.** The literal is the mistake, and it names the value that does not fit.
- **Cost.** A literal on the left takes its type from the right operand, so the right operand is checked first and an error in it appears before an error in the left one.

## D83. The dynamic tree is `Value` in `std/shape`; `std/json` has `parse` and `stringify`

Decision: `Value` (`Null`, `Bool`, `Int`, `Float`, `Text`, `List`, `Record`) is declared in `std/shape`. `std/json` has `parse(text): Result<Value, string>` and `stringify(value): string`, the names of TypeScript's `JSON`. `Record` is a `Map<string, Value>`, so it keeps insertion order (D64); a duplicate key takes the last value in the first place. A number without fraction or exponent that fits `int` is `Int`, any other number is `Float`. Errors read `<line>:<column>: <reason>`; nesting deeper than 512 is an error.

- **Reason.** The `Shape` trait (node shape-trait) converts types to and from `Value` without knowing the format. A later TOML or binary format reuses it, and `std/shape` never imports `std/json`, so there is no import cycle.
- **Reason.** Narra has no single number type, and a 64-bit integer id must survive a round trip, which an `f64` cannot hold, so `Int` and `Float` are separate.
- **Alternative rejected.** `Json.Value` in `std/json`: it ties `Shape` to JSON.
- **Alternative rejected.** Reading JSON straight into typed values without a tree: faster, but needs `derives Shape` first, and JSON-RPC needs the tree for messages whose type depends on a field.
- **Cost.** One primitive, `FloatText.parse` in `std/internal/float`, for `strtod`. `stringify` recurses on the depth of the `Value` it is given; only `parse` bounds depth.

## D84. `Shape` converts to and from `Value`; the target type is an explicit type argument

Decision: `trait Shape { function toValue(value: Self): Value uses alloc; function fromValue(value: Value): Result<Self, string> uses alloc }` is declared in `std/shape`. `std/json` has `encode<T: Shape>(value: T): string` and `decode<T: Shape>(text: string): Result<T, string>`, called as `Json.decode<User>(text)`. The compiler provides `Shape` for `bool`, `int`, `f64`, `char`, `string`, `Array<T>`, `Map<string, V>`, `Option<T>` and tuples of two to four items; `Value` and user types use `implement` blocks. A decode error is `<path>: <reason>` with the path from the root (`.users[2].age: expected int, found Text`). Decoding is strict on types (`int` takes only `Int`; `f64` takes `Int` and `Float`) and ignores extra data. Chosen by the orchestrator on 2026-11-09.

- **Reason.** An explicit type argument is TypeScript's form for a generic call, so no type-witness value is needed.
- **Reason.** Going through `Value` keeps `Shape` independent of JSON, and every implementation is ordinary Narra code with no runtime type layout.
- **Mechanism.** The prelude implementations are private generic functions in `std/shape.nra`. The checker points a `Shape` trait call on a prelude type at the matching function and its instance, so no C is written for them and `derives Shape` (node shape-derive) can emit Narra-level conversions the same way.
- **Alternative rejected.** `Json.decode(text, User)`, a type name passed as a value: it needs a new type-witness type.
- **Alternative rejected.** A constant description walked by a generic runtime: it needs layout tables (parked node layout-tables) and C code that builds reference-counted values from raw memory.
- **Alternative rejected.** `implement Shape for Array<T>` blocks written in `std/shape`: implement blocks for generic and prelude types are not supported (E0513, E0514), and lifting that is a larger change.
- **Alternative rejected.** Generating C for each type in the emitter: more code than the Narra functions it replaces.
- **Derived encoding (for shape-derive).** A `None` field is left out, and an enum value is `{"type": "Circle", ...fields}`, TypeScript's discriminated union. Rust serde's `{"Circle": {...}}` is rejected because TypeScript code cannot narrow it with a `switch`.
- **Cost.** One intermediate `Value` tree per encode or decode. Tuples above four items and sized integers have no `Shape`. A nested `Option<Option<T>>` decodes `Null` as the outer `None`. The path of an error is text; a user `fromValue` that returns a message starting with `.` or `[` is read as a path by the containers around it.

## D85. A `mutable` marker only on `for ... of` over an array

Decision: `for (mutable item of items)` binds `item` to each element of the array `items` in place, so `item.hp = 0` changes `items`. `for (const #[index, mutable item] of Arrays.entries(items))` does the same and also gives the index. `Arrays.values(items)` may be written for `items`. This is the only position the marker gets for now. Pattern bindings, destructuring and a local reference binding stay as D81 left them. Amends D81, whose bindings that copy part of a value stay read-only everywhere else. Chosen by the orchestrator on 2026-11-09 (the maintainer delegated this batch's gate decisions).

Rules of the loop form:

1. **The source is a local place.** `items` must be a local `const` or `let`, a `mutable` parameter, or a field or element path of one (`game.players`, `teams[t].players`), of type `Array<T>`. A call result (`for (mutable p of makePlayers())`), a plain parameter, a top-level constant, a `Map`, a `Range` and a generator are compile errors. These are the arguments D81 rule 3 allows for `mutable`.
2. **The element has contents.** `T` must be a contents type (D81 rule 1): a struct, `Array`, `FixedArray`, `Map`, `Box`, `StringBuilder`, `Generator` or `Coroutine`. `for (mutable n of numbers)` with `numbers: Array<int>` is a compile error whose hint is `numbers[index] = ...` in an `Arrays.entries` loop. Enum and tuple elements are errors too (D81 rule 7).
3. **The binding acts as a `mutable` parameter.** Its fields and elements can be assigned at any depth, it can be passed as `mutable`, and a non-escaping closure can capture it by reference (D77). Assigning it whole (`item = other`) is a compile error, as D81 rule 2. It cannot escape: `const copy = item` copies, as for a `mutable` parameter.
4. **The source is borrowed for the whole body.** Inside the body the source and every path that overlaps it cannot be read, assigned, passed or captured: for `for (mutable p of game.players)`, `game`, `game.players` and `game.players[0]` are errors, `game.round` is fine. This keeps `Arrays.push(mutable players, ...)` and `players = []` from moving or freeing the storage `p` points at, and keeps a read from seeing a half-changed array through a second name. `break`, `continue`, labels and `return` work as usual. A nested `for (mutable q of p.items)` borrows `p.items` and is fine.
5. **Uniqueness first.** Before the first iteration the source array is made unique, copying it when its count is above 1, as for a `mutable` argument or `items[i] = v` (D50, section 6). The loop then takes no reference to the array and the element binding takes no reference to the element.
6. **Not in `function*` or `async` bodies in the first build.** The pointer would be a frame field kept across `yield`. That is sound (the body cannot reach the source while suspended either) but needs its own emitter path; the build node reports a compile error there until a program needs it.

Positions weighed, with what happens to each:

- **Chosen: `for (mutable item of items)` and `#[index, mutable item]` in `Arrays.entries`.** `for (const p of players) { p.hp = 0 }` is the most common TypeScript loop that changes data, and D81 turned it into a compile error whose fix is an index loop. The marker makes the fix one word, with the TypeScript meaning. The element is reached through an index into storage the body cannot touch, so the reference cannot dangle.
- **Deferred: pattern bindings, `if (x is Some(mutable y))` and `Shape.Circle(mutable c) =>` in `match`, nested patterns and `while (cur is Some(mutable node))`.** D81 rule 7 wins over the marker for now. A payload is a counted heap value that copies share and that is never written (D55); a write through `y` would need a new copy-on-write step for every fielded enum, a place as subject (`x`, not `f()`), and a borrow of `x` over the arm or `if` body. selfhost/src has 0 places that read a struct out of an `Option` or enum, change it and store it back (8 text matches, all list walks that reassign the subject or a `Map` entry, `count.py -v`). The current rewrite stays: read with `is`, build the changed copy, store it back. A program that needs it reopens this item; the rules above (place subject, borrow, uniqueness) carry over, plus the payload step.
- **Rejected: destructured bindings, `const { mutable position } = player` and `const #[mutable a, b] = pair`.** Each is a local reference binding with another spelling, so it falls with the next item. Tuple elements are not places (D81 rule 7), so `#[mutable a, b]` has nothing to point at outside the `Arrays.entries` loop, where the tuple is built by the loop and the marker reaches the array element.
- **Deferred: a local reference binding `mutable frame = emitter.frames[index]`, the inbox item from the first draft of mutable-reference.** D81 rejected a third binding kind, and the counts do not overturn it. selfhost/src has 11 functions that write through one element path 3 times or more (49 writes), and most of them call `f(mutable checker, ...)` between the writes, which a reference into `checker` would block until its scope ends; the user would wrap the writes in a block. The form that already exists does the same job: a function with a `mutable` parameter, called once with the path, `markHinted(mutable checker.bodies[body].closures[closure], ...)`. Weigh again when a hot loop needs it.
- **Rejected: `x = f(mutable x)`.** It stays E0929 (D81 rule 10). With D81 a `mutable` parameter changes contents and the result is a separate value; assigning the result to the same name in one statement would replace the value the call just changed, or the call would change a value about to be dropped. Either reading surprises. Write `f(mutable x)` and return nothing, or `x = g(x)` with a plain parameter.
- **Out of scope: `mutable` arrow parameters** (`(mutable p) => ...`, E0028) and `Maps.entries` values. Arrow parameters are parameters, section 6. A map value in place needs a pointer into a hash table that the body could rehash; it waits for a program that needs it, with the same borrow rule.

- **Reason.** The marker reaches the same contents a `mutable` parameter does, so the loop form adds no new kind of reference: the element pointer is what `mutable items[index]` passes today, taken once per iteration.
- **Reason for borrowing the whole source.** Rust allows no use of the vector during `iter_mut` either. A rule per overlapping path (reads allowed, writes not) would need the checker to prove a read does not see the element being changed; one rule is simpler to explain and to check, and the 2 selfhost loops it would replace touch the array only through the index.
- **Prior art.** Rust: `for p in players.iter_mut()` and `ref mut` in patterns (`if let Some(ref mut y) = x`), with the borrow checker forbidding other uses of the source. C#: `foreach (ref var p in span)` over a `Span<T>` and `ref` locals (`ref var frame = ref frames[i]`). Swift: `inout` parameters only, no reference bindings, loops write `players[i].hp = 0` over `indices`. Hylo: `inout` bindings (`inout x = &y.z`) and `for inout p in &players`, with exclusivity checked statically. Narra takes the loop form (Rust, C#, Hylo) and leaves out pattern and local references (Swift's line).
- **Cost.** About 400 lines in selfhost/src: parser (the marker in a `for` head and in an `Arrays.entries` tuple pattern), AST and its dump and `fmt` (one flag on `ForStmt`), checker (place source, element type, whole-body borrow, assignment and escape rules on the binding), emitter (a pointer local in `emitArrayFor`, uniqueness before the loop, no retain of the array or the element). New diagnostics: the source is not a place, the element type has no contents, the source is used in the body, the binding is assigned whole (or reuse of that D81 code), and the generator or async body case.
- **Cost.** A loop that also needs to read another element of the source (`players[0]`) must be an index loop.
- **Counts.** `roadmaps/selfhost/nodes/mutable-markers/count.py`. In selfhost/src, 2 of 119 index loops over an array write an element of that same array (`check.nra:621`, `resolve.nra:532`); 0 `for ... of` loops write their source by index; 23 `mutable path[...]` arguments; 0 enum payload write-backs. std/src has none of these. The value of the form is the TypeScript reading, not the corpus.

## D86. Function types carry their effects

Decision: a function type is `(params) => result uses effects`, and the effects belong to the type, as D78 did for generator types. Type equality and the type key compare them; the C name of a function type ignores them, so the emitted C does not change. A function value fits a slot (a field, element, argument, return, struct literal, annotated `let`) when it does only the effects the slot lists. A function with fewer effects may go where more are allowed; one with more is a type error: `expected (int) => int, found (int) => int uses io`. A call through any function value charges the effects of its type. An effect variable in a function type names the effect variables of the function whose signature it is in (the binder, as for generators); the effect pass replaces each by the effects bound at the call.

- **Reason.** The effect pass used to follow the written type of a callee (PR #98). Stores, type parameters, and bindings without a written type (`is Some(f)`, match arms, `for ... of`, destructuring, arrow parameters) lost the effects, so a later call counted as pure.
- **How it checks.** A function value written as an arrow or a name has effects only the effect pass knows, so its type lists none. `expectType` records each function value stored where a function type with the same signature is expected; the effect pass checks what the value does (an arrow's set, a named function's, a local's tracked set, or the effects of its type) against the slot, after the fixed point. A nested position (`Array<F>` against `Array<F uses io>`) needs equal types.
- **Private functions.** A function-type parameter without `uses` on a private function stays an implicit effect variable: the call has the effects of the argument, and the store is not checked. Assigning to a local adds the effects to the local (it is tracked) instead of being checked against its type.
- **Inference.** A type argument inferred from a function value takes the value's type, which lists no effects, so `Holder { value: loud }` with an effectful `loud` is rejected. Write the type: `Holder<(value: int) => int uses io>`.
- **Constants.** A top-level constant is used from other functions and modules through its type alone, so its value counts as a store. Without a written type the constant takes the type of its value: `const step = loud` and `const steps = [quiet, loud]` are rejected when `loud` does `io`. Write the type: `const step: (value: int) => int uses io = loud`, `const steps: Array<(value: int) => int uses io> = [quiet, loud]`. An array literal's element type is the written type or the first element's, never a union of the elements' effects: a union needs the effects of private functions while types are checked, and the effect pass finds them only afterwards. The stores inside a constant's value (elements, fields, spreads) are checked like those in a function.
- **Prior art.** Koka and Effekt put effects on function types and infer them; D78 for generators in this repo.
- **Rejected: follow the written type in the effect pass.** It cannot see stores or type parameters.
- **Rejected: a unique type per arrow or named function with its effects resolved later.** Two closures in one array would have different types, and a generic function would be instantiated once per closure.
- **Cost.** Slots typed without `uses` that hold effectful closures now need `uses` (4 fields in selfhost, 2 test programs). Instances of one generic with function types that differ only in effects share a C name.

## D87. Named arguments, in any order, with every argument passed

Decision: a call can name its arguments, `move(x: 1, y: 2)`, spelled like struct literal fields. Positional arguments come first, then named ones in any order. A positional argument after a named one is an error, and each parameter is passed exactly once, by position or by name. `mutable` follows the colon: `buf: mutable xs`. Names work only when the callee is a declaration with parameter names; a call through a function value or a std builtin registered without names takes positions only, with an error that says so. Arguments are evaluated in the order they are written. Amends D12, which ruled out labeled arguments; the rest of D12 stays, and parameter defaults are a separate decision. Debate: `roadmaps/selfhost/debate/named-arguments/debate.md`.

- **Reason.** Calls with many parameters or bare `true`/`false` arguments do not show what each argument means: 363 of 1477 functions in selfhost/src and std take 4 or more parameters, and about 30 callees take a trailing bool literal.
- **Reason.** The checker puts named arguments back in declaration order, so each function still compiles to one C function with a fixed parameter list (D12).
- **Reason.** `name: value` is the struct literal form. In TypeScript `f(a: 1)` is a syntax error, so the text has no second meaning (D2). `f(a = 1)` would read as an assignment to TypeScript readers.
- **Reason.** Kotlin, the model for this feature, evaluates arguments in written order. A reordered argument with side effects is evaluated into a temporary first.
- **Rejected: `name = value`** (Kotlin's spelling). It is an assignment in TypeScript.
- **Rejected: labels only, in declaration order.** It documents a call but does not allow reordering.
- **Rejected: editor hints, required labels on bool literals, or enum parameters instead.** None gives named arguments in any order. Replacing flag parameters with enums is still fine on its own.
- **Prior art.** Kotlin `foo(a = 1)`, C# `foo(a: 1)`, Swift argument labels, Python keyword arguments. TypeScript, Rust and Go have none.
- **Cost.** Parameter names become part of a function's interface for callers that use them: renaming a parameter breaks those calls.
- **Cost.** About 650k tokens over three nodes, after the fix for argument evaluation order.

## D88. No async blocks or generator blocks for now

Decision: Narra does not add `async { ... }` blocks or inline generator expressions (`function* (): Generator<T> { ... }`) until a program needs one. An inline async body is an async arrow, and an inline generator is a top-level `function*` that takes the values it needs as arguments (D74). Debate: `roadmaps/selfhost/debate/suspending-blocks/debate.md`, which records both designs for a rerun.

- **Reason.** No code in selfhost/src or std declares a generator or an async arrow.
- **Reason.** `async { ... }` would mean exactly `async () => { ... }`, a second spelling, and it cannot ship before async functions are built.
- **Reason.** A generator expression needs capture and lifting work in both compilers, estimated at 450k tokens or more.
- **Prior art.** Rust has async blocks and unstable `gen` blocks; Kotlin builds generators with `sequence { }`; TypeScript has neither block and uses anonymous `function*` expressions and async IIFEs.

## D89. The Go compiler is retired; a checked-in C seed bootstraps Narra

Decision: the repository has no Go compiler. `seed/narrac.c`, the C that the self-hosted compiler emits for itself, builds the compiler with `cc` alone (`selfhost/bootstrap.sh`), and `scripts/seed-update.sh` regenerates it. The tests run on that compiler through a test runner written in Narra (`selfhost/tests`, `scripts/test.sh`). Go stays only as a benchmark language in `bench/`. The compiler source may only use features that the current seed compiles, so a new feature lands first without itself, then the seed is regenerated.

- **Reason.** Each language feature would otherwise be built twice, in Go and in Narra.
- **Reason.** Stage 1 and stage 2 already emit the same C, so Go is not needed to start the chain.
- **Alternative rejected.** Keep Go as the bootstrap: needs a Go toolchain on every machine and a second implementation in step with the first.
- **Alternative rejected.** Download a release binary to bootstrap: needs a network and a trusted host for every build. GitHub Releases come later, for distribution only.
- **Cost.** The seed is a 6 MB generated file; a merge that changes `selfhost/` on two branches regenerates it instead of merging it.
- **Cost.** C is the only backend, so the seed is the same on every platform (macOS arm64 and Linux x86_64 are checked).

## D90. `Cell<T>`: a source module over a box, shared by every copy of the handle

Decision: `std/cell` is a Narra source module (`std/src/cell.nra`) that exports `opaque struct Cell<T>` and four functions: `Cells.create(value)`, `Cells.get(cell)`, `Cells.set(cell, value)` and `Cells.update(cell, change)`, where `change` is `(value: T) => T uses E`. A `Cell` holds one `Box<T>`; copying a `Cell` (a `const` captured by an escaping closure, a struct field, an argument) copies the handle, so every copy reads and writes the same slot. `Cells.set` and `Cells.update` write the box in place through one primitive, `std/internal/cell` `store`, which skips the copy that a write to a shared `Box` makes, and drops the old value at the call (`Cells.set(cell, None)` frees a closure stored in the cell at once, as D80 needs). Escaping closures share changing state by capturing a `const` `Cell`; E0612 (an escaping closure that captures a `let`) now ends with the hint to keep the state in a `Cell` from `std/cell`. Chosen by the orchestrator's worker on 2026-11-09, under the gate the maintainer delegated for batch 2.

- **Import.** `import * as Cells from "std/cell"` for the functions and `import { Cell } from "std/cell"` for the type, the pattern of `std/json` and `Value` (D83). `Cell` is not a prelude type, so there is no `Cell.get` spelling; the plural alias follows D72. SPEC text that wrote `Cell.set` now writes `Cells.set`.
- **Names.** `create`, `get`, `set` and `update` match `Shared.create` and `Shared.update`; `Shared.read(shared, f)` is `Cells.get(cell)` without the lock and without a callback. `update` takes a function that returns the new value, not a `mutable` parameter, because D81 and E0028 do not allow a scalar as `mutable`. `set`, `get` and `update` have no `alloc` or `io` effect; a `Cell` write is not tracked by the effect system, as D46 only makes the allocation visible, at `create`.
- **Equality: `==` and `!=` on a `Cell` are a compile error (E0965), and a `Cell` is not `Hashable`, so it cannot be a `Map` key.** Comparing the contents would be the wrong meaning for an identity type. Identity equality can come later if a program needs it; it would need `implement Equatable` and `implement Hashable` together plus a pointer-identity primitive. Built as a check in `nominalImplements` (`traits.nra`): the struct `Cell` of `std/cell` never gets the automatic Equatable or Hashable. The hint names `Cells.get`. Decided by the coordinator after the worker's first version.
- **Threads.** A `Cell` uses plain reference counts, so it must not reach another thread. Threads are parked, so no thread function or channel exists to reject it today, and the checker says nothing. The node that builds threads must reject a `Cell` (or a type that holds one) captured by a thread function or sent through a channel, and point at `Shared`. SPEC section 8 records the rule.
- **Cycles.** None handled, as D80 decided. A closure stored in a `Cell` that captures the same `Cell` leaks, and the leak check reports it (`testdata/leaks/cell_cycle.nra`, new `leaks` suite of the test runner, which expects exit code 102 and a `leak check:` report); `Cells.set(cell, None)` before the last outside handle drops frees it (`testdata/cell_cycle_broken.nra`).
- **Checker change.** A bare `None` argument of a generic call takes its type from the earlier arguments (`Cells.set(slot, None)` with `slot: Cell<Option<string>>`). Before, it reported "cannot tell the type of None". It applies to every generic function, not only the cell ones.
- **Rejected: a builtin type `TCell` like `Shared`.** It needs a case in about 30 places of the checker, emitter and dump code, and it is one more type kind to carry for a box with a different write rule. The source module needs one primitive and no new type.
- **Rejected: `Cell` as `Box<T>` itself or a type alias.** `Box` has value semantics: a write to a shared box copies it first (section 6), and `==` compares contents. A `Cell` must share.
- **Rejected: the `Cell.set` spelling with a prelude `Cell` type.** It would put `Cell` in every module's scope (D67 keeps the prelude small) and make `Cell` a namespace type like `Box`, which needs entries in the namespace table and the missing-import fixes.
- **Rejected: `get` that borrows.** `Cells.get` returns a copy that holds one more reference, so an array in a `Cell` is copied by the next in-place change to the array made through `update`. That is O(n) per update for a large array; a program that updates a large array often should keep it in a struct passed as `mutable`.
- **Cost.** One heap allocation per `Cell`, plus one reference count per copy of the handle. `Cells.update` on a counted value (string, array, struct) holds the old value while `change` runs.
- **Cost.** E0966, reserved for this node, is not used: the other misuses (a wrong value type, reaching the field of the opaque struct) are reported by existing codes.

## D91. A call through a value binds the effect variables of the function the value names

Decision: when a function with effect variables is used as a value (`const apply = twice`, or a local constant), a call through the value binds each variable to the effects of the arguments passed to the parameters whose type names it, as a direct call does (D14, D78). `apply(loud, 1)` has the effects of `loud`, and `apply(quiet, 1)` is pure. A variable of the function being checked stays a name, because the caller of that function binds it.

- **Reason.** The effect pass dropped the variable: a local constant tracked the function's effects without its variables, and a top-level constant took the type effects, so `apply(loud, 1)` counted as pure.
- **Reason.** The type of the value (D86) already holds the parameter types and the function that binds the variables, so the call can bind them with the rule direct calls use. No new syntax and no new error.
- **Rejected: an error for a function with effect variables used as a value** (hint: wrap it in a closure). It is simpler, but it rejects `const apply = twice` and `values.map(twice)` style code that is correct, and a closure cannot write the effect variable it needs.
- **Cost.** A variable that no parameter of the value's type names is not bound at the call. Passing such a value on as an argument is unchanged.
- **Generator results.** A function type cannot list the `alloc` that creating a generator charges apart from the generator's own `uses`, so storing a generator function in a slot whose result is a generator type does not charge it. A call through any function value whose result type is a generator charges `alloc` at the call instead, so a pure function cannot create a generator frame through a value. A value that returns a generator it did not create pays an `alloc` it may not use; the type cannot tell the two apart.

## D92. No panics from arithmetic, missing returns or std functions

Decision (maintainer, 2026-11-09): a Narra program panics only through `panic`, a failed `assert`, an array index out of range, running out of memory and stack overflow. Everything else that panics today becomes a defined result or a compile error.

- **Integer overflow wraps in every build.** `+`, `-`, `*` and unary `-` wrap like `+%`, `-%` and `*%`, so those three operators and their compound forms do not exist. `MIN / -1` gives `MIN` and `MIN % -1` gives `0`. Built (`Int.addChecked`, `subChecked`, `mulChecked` and the same in `std/i8` to `std/u64`).
- **Division and remainder need a divisor proven non-zero.** Integer `/` and `%` compile only when the divisor is a non-zero constant or a name narrowed by a check (`if (count != 0) { total / count }`). Any other divisor is a compile error with a hint to add the check. Float division is unchanged (IEEE: infinity and NaN).
- **Shift counts must be proven in range.** A constant count must be at least 0 and below the bit width. A variable count must be narrowed first (`if (n >= 0 && n < 64) { x << n }`; an unsigned count needs only the upper bound). Any other count is a compile error.
- **A function with a result must return on every path.** A path ends at `return`, a call to `panic`, an `if`/`else` whose branches all end, an exhaustive `match` whose arms all end, or `while (true)` / `for (;;)` with no `break` out of it. The end of any other path is a compile error. The runtime check stays only as an internal compiler error.
- **Std functions return `Result` or `Option` instead of panicking** on bad input: `Str.slice`, `Str.split`, `Str.repeat`, `Bytes.at`, `Bytes.slice`, `Float.truncate`, `Arrays.filled`, `Ranges.stepped`, `Math.clamp`, `Math.powInt` and `Random.nextInt`. Compound assignment on a Map index target (`counts[key] += 1`) is removed; `Maps.update` names the missing-key case.
- **Array indexing still panics.** `items[i]` is bounds checked; `Arrays.get(items, i)` returns an `Option`. Proving indexes in range was weighed and rejected for now.
- **Reason.** The maintainer wants a language without hidden panics. Each rule keeps the operator syntax and adds no runtime cost: wrapping is what the hardware does, and the proofs reuse narrowing.
- **Rejected: integer `/` returning a float, or division as a function.** Too verbose (maintainer).
- **Rejected: `x / 0 = 0` (Gleam, Pony) and masking shift counts (JavaScript, Rust release).** Both give a quiet wrong answer to what is almost always a bug.
- **Rejected: a NaN for integers.** It needs a reserved value, a check in every operation, and breaks `==` on `int`.
- **Cost.** Wrapping hides overflow bugs that debug builds reported; checked functions (`Int.addChecked` and the like) return `Option` where it matters. Variable divisors and shift counts need an explicit check.

Built: wrapping arithmetic and checked functions; the missing-return rule (E0969 for functions, E0970 for arrow functions), with the emitted runtime check now an internal compiler error. Division and shift proofs are built (E0979 for a divisor, E0980 for a shift count); the std functions return `Option` (table in SPEC.md section 5) and `Maps.update(mutable map, key, initial, change)` replaces compound assignment on a Map index, which is a compile error. Std panics that remain: `Str.sliceFrom`, `Str.charAt` and `Str.nextChar` on a bad position, and array indexing.

Amendment (proven-operands, 2026-11-11):
- **What counts as a check.** The narrowing of `let` and `is` (let-narrowing) carries over: a comparison of a local integer name with a constant (`!= 0`, `> 0`, `< 0`, `>= 0`, `< 64` and the other orders, either side), joined by `&&`, `||` and `!`. A check narrows the `then` branch, the `else` branch with the opposite, the body of a `while`, the right operand of `&&` and `||`, and the rest of the block after an `if` whose `then` branch ends. Constants are literals, negated literals and top-level `const` values built from them with `+`, `-`, `*` and `<<`, and a `const` local set from one.
- **How a narrowing ends.** Like the variant narrowing: an assignment to the name, a call that passes it as `mutable`, and the start of a loop that does either. A name an arrow function captures by reference is not narrowed there.
- **Field paths are not narrowed.** `size.count` cannot be proven, because any call could change the field. The message says to copy the field into a local and check the local.
- **Cost.** Three sites in the compiler (`hash`, `alignUp`, the test runner's shard split) and one in std (`Random.nextInt`) needed an added check; the runtime has no division or shift panic.

## D93. `Async<T>` is an eager handle on one implicit event loop per thread

Decision (maintainer, 2026-11-09), amending D17 and reopening D88 for async blocks:

- **`Async<T>` is a handle to a result on its way,** like a TypeScript `Promise`. Creating one starts the work. It is an identity type (D80): copies share the same computation, and the finished value is kept until the last handle is dropped. `Task<T>` is folded into it.
- **An async function returns `Async<T>`.** `async function fetchUser(id: int): User` has the type `(id: int) => Async<User>`, and `fetchUser(1)` starts it and returns the handle. `await handle` gives the `User`; awaiting the same handle again gives the same value.
- **`async { ... }` is an expression of type `Async<T>`.** It starts at once, like a call. Generator blocks stay parked (D88).
- **An unused `Async<T>` is a compile error,** like an unchecked `Result`. `Async.detach(handle)` lets it run with nobody waiting.
- **`Async.then(handle, callback)`** runs `callback` with the value when it is ready and returns nothing. The callback is an ordinary escaping function (no `await` inside); the runtime drops it after it runs or when the handle is cancelled, so it forms no lasting cycle. Its effects are charged at the call.
- **`Async.peek(handle): Option<T>`** gives the value once finished, `None` before that or after a cancel. It never blocks.
- **`Async.cancel(handle)`** is cooperative: the I/O the task waits on next returns `Err(Cancelled)`, and the body returns a `T` as for any error. `await` always gives a `T`; a cancelled handle's `then` callbacks do not run.
- **`Async.all(handles)`** waits for all and gives the values in order; **`Async.race(handles)`** gives the first to finish. `Arrays.map(items, async (x) => ...)` gives `Array<Async<T>>`, so `Async.map` and `Async.forEach` are not needed, and async callbacks are no longer an error.
- **One implicit event loop per thread.** The entry function runs as the loop's first task and may be `async`; the program exits when it has returned and no work is left. `Async.run` and the explicit `Loop` are removed.
- **No blocking wait** (`block_on`). Inside a running loop it would hang or run other tasks in the middle of a function, and refusing at runtime would be a panic (D92).
- **No `Thenable` trait.** `await` works only on `Async<T>`, so it never runs user code (D12, no hidden implementations). A type that can be awaited offers a named function that returns an `Async<T>`.
- **Effects.** Creating an `Async<T>` charges the effects of its body, since the body starts there.
- **Reason.** An async expression in a non-async function has to produce a value at once, so the result needs a type of its own. An eager handle matches TypeScript; D17's worry about lost errors does not apply because failures are `Result` values and an unused handle is a compile error.
- **Rejected: lazy `Async<T>` (Rust futures).** It needs a separate `Task<T>` for running work and an executor to find at runtime.
- **Rejected: an explicit loop or a thread-local runtime (tokio).** Starting work where no loop runs could only fail at runtime.
- **Open.** The exit code when tasks are still running after a failed entry function, and how loops on several threads share work (threads are parked).

Not built yet. SPEC.md section 8 still describes D17 and is rewritten by the async nodes.

## D94. Shared writes are compile errors, and copies are written as `copy(x)`

Decision (maintainer, 2026-11-09, from the debate `roadmaps/selfhost/debate/value-semantics/debate.md`): Narra keeps value semantics and TypeScript-style in-place writes, and makes every copy visible in the source.

- **Shared write.** A write to an array, map, struct (including all-scalar ones), `FixedArray`, `Box`, `StringBuilder`, generator or coroutine is a compile error while another live name shares the place's root or a containing path. Sharing is tracked in both directions and through chains. Passing `mutable R` writes `R` and everything below it. A non-escaping closure that writes a captured `const` writes it where the closure is passed. Parameters, pattern and `for (const ...)` bindings, captures and values read from identity types always share their source. Assigning a whole place ends its sharing. Liveness is checked inside one function body; each error names the later use.
- **`copy(x)`** is the one way to duplicate a value, at any depth. It is a retain at the call; copy-on-write copies only the written paths. `Arrays.copy`, `Maps.copy` and spread are not copy tools.
- **`from`.** A function whose result shares a parameter writes `from <parameter>` after its result type, and the marker is part of its function type: `(grid: Grid) => Array<int> from grid`. A body that returns a shared parameter without it is an error. A function with fresh results fits a type with `from`, not the other way round. Nothing is inferred across functions (D46).
- **Debug copy check.** Debug builds report, at exit, every copy-on-write that no `copy(...)` explains. It replaces `--report-copies`, `NARRA_TRACE_COPIES` and `@unique`, which were never built; `@unique` is removed. Release builds are unchanged.
- **Unchanged:** `mutable` parameters, `for (mutable ...)`, `Cell` as the shared reference. `Cells.modify` is not added.
- **Amends** D1 ("immutable values" now reads "mutable value semantics: a value changes through one name, and no other name sees the change"), D27 (a shared `mutable` argument is an error, not a copy) and D81 rule 5. **Restores** D2 for `const b = a; a[0] = 9`.
- **Reason.** Copy-on-write hid where copies happen and let TypeScript-valid code mean something else. The rule keeps the syntax and adds no runtime cost.
- **Rejected:** immutable arrays and structs with `ref` and a mutable array type (rebinding a shared value still copies silently, so it needs the same analysis on a larger rewrite); inference of shared results across functions; shallow copy tools; `Cells.modify`.
- **Cost.** About 120 to 220 edits in the compiler and 30 to 50 `from` marks; errors depend on liveness, so a later use can turn an accepted write into an error. About 1.45M tokens over six nodes.

`copy(x)` and the debug copy check: built. `copy` is a prelude function (E0975 for a wrong argument count, `copy(mutable x)` and an argument of a type that is never changed in place, with a fix that removes the call). A local may be named `copy` and hides it, as for `panic`. The check runs in debug builds with `NARRA_CHECK_COPIES` set and exits with code 103; release builds (`-DNARRA_RELEASE`) compile none of it. `@unique`, `--report-copies` and `NARRA_TRACE_COPIES` are gone from SPEC.md. The shared-write rule is built (amendments below), so every part of this decision is built. SPEC.md section 6 describes the final rule (Copies on shared values, Shared writes, Shared results).

Amendment (node from-marker, 2026-11-10). The `from` marker is built (SPEC.md section 6, Shared results), with two answers the debate left open, settled by the maintainer:

- **Closures.** `from` names only a parameter, never a capture. An arrow function whose result shares a captured value is an error (E0976) whose fix wraps the value in `copy(...)`. An arrow function whose result shares one of its own parameters writes `from <parameter>` after its result type, like any function.
- **Generic results.** `from items` covers the whole parameter and everything reached through it, at any depth, whatever the type arguments are. A result of type `T` taken out of `items: Array<T>` shares `items`. When `T` is a scalar type the marker is still allowed and changes nothing.
- **How it is checked.** One walk per function body follows each returned value through fields, elements, patterns, loops, locals, `mutable` arguments and calls whose callee has `from`. A value whose type holds no tracked value shares nothing. E0976 reports a returned value that shares a parameter the marker does not name, and a `from` that names no parameter; E0977 reports a function with `from` stored in a slot without the same `from`, unless its result holds no tracked value. A std builtin's result counts as fresh, except that a `for ... of` over a std iterator shares what it walks.
- **Order.** `from` comes after the result type and before `uses`: `(grid: Grid) => Array<int> from grid uses alloc`.
- **Rollout.** The compiler's 185 sharing returns cannot carry `from` until the seed reads it, so `shared-results = "unchecked"` in `[package]` of `narra.toml` turned the check off for `selfhost/` while the compiler was marked. The switch was removed on 2026-11-11, once every compiler module carried its marks (node shared-write-sweep-rest); every package is checked.

Amendment (node shared-write-check, 2026-11-11). The shared-write rule is built as warning E0978, which `narrac check` and `apply-fix` report when a package has no errors; build, run and test do not run it. The fix wraps the value that started the sharing in `copy(...)`. Settled for the node:

- **Std builtins** need no `shared` column. Arrays and maps have no reading builtins (`items[i]` and `table[key]` are places), and `for ... of` over a std iterator shares the elements it walks. A std function's `mutable` argument (`Arrays.push`, `Str.append`) changes only the container, so names that share its elements are not hit; the pushed value becomes an element of the container.
- **Liveness.** One forward walk per body keeps the pairs of places that share a value. Assigning a local or a place whole ends its pairs, so a reassigned local starts fresh. A write that hits a pair warns at the next use of the other side; branches are merged, and a loop body is walked again until its state stops growing. Parameters, captured values inside an arrow function, constants and cells are always live.
- **Its own walk.** The analysis does not reuse the flow-insensitive walk of `sharedresults.nra`; it calls `findSharedArgument` for `from` results and `isCopyCall` for `copy(...)`.
- **Loops.** A `for ... of` loop's iterator holds the place it walks while the body runs, so a write that copies that place warns too. Assigning one element keeps the sharing of the other elements, because indexes are not compared. Known limit: reassigning the same index still warns (`let row = grid[0]; grid[0] = [9]; row[0] = 5` while `grid[0]` is used later); shared-write-error decides whether it stays.
- **Counts at the start of shared-write-sweep:** 198 in `selfhost/src` (34 of them writes inside a loop over the written place), 0 in `std/src`, 3 in `selfhost/tests/src`, 36 in `testdata` (the copy samples that keep a hidden copy on purpose).

Amendment (node shared-write-sweep, 2026-11-11, decided by the orchestrator). `copy(x)` also takes an enum whose payload can hold a tracked value, at any depth: `Option<Array<int>>`, `Result<Team, string>`, a recursive enum with a struct payload. An enum value is never changed in place, but a part taken out of its payload can be, so a function whose result is such an enum may need to copy a second shared parameter, and only `copy` can do that. The copy retains the payload; a write to a part taken out of the copy copies only that part. The debug copy check gives a credit to each counted value inside a heap enum payload that the copy reaches, so those copies on write are explained. An enum whose payloads hold only scalars and strings is still E0975. The E0976 and E0978 fixes offer `copy(...)` for exactly the types `copy` takes.

- **Rejected:** treating enum results as unshared in E0976 and E0978. It is unsound: `match (lookup(table)) { Some(row) => row[0] = 1 }` would copy a shared array without showing it.

Amendment (node shared-write-error, 2026-11-11). E0978 is an error. Every command that checks a package runs it (`build`, `run`, `test`, `emit`, `check`, `apply-fix` and the language server), also when the package has other errors; a write that another error already points at is not reported twice. Warnings are gone from the compiler: nothing else used them. Settled for the node:

- **Indexes.** Two different integer literal indexes are different places: `const first = rows[0]; rows[1][0] = 4` is clean (the reviewer's probe; no copy happens either). Every other index is conflated with every element: an identifier or any other expression (`rows[i]` and `rows[j]`, `rows[i]` and `rows[0]`) and a map key. The same literal index stays one place, and assigning it does not end its sharing, so `let row = grid[0]; grid[0] = [9]; row[0] = 5` still reports while `grid[0]` is used later. Ending it would need a whole-place rule for one element; for identifiers, proof the index did not change. No site in the compiler, std or the samples needs it, and the fix is one edit.
- **Function values share their captures.** A function value made from an arrow function now pairs with the values it captured, so calling it counts as a use of them: `const read = () => items[0]; let other = items; other[0] = 5; read()` is an error. Before, the pair was dropped because a function type holds no tracked value, and the debug copy check found the hidden copy.
- **The debug copy check runs on every program sample** (`NARRA_CHECK_COPIES` is set for every program the test runner starts). A `for ... of` or `yield*` over a generator it does not own resumes its own copy, as section 8 says; that copy is explained like a `copy(...)`. The credits that a copy of a heap enum gives the arrays, maps and boxes in its payload end with the copy: the copy marks the payload box it retained, once per copy, and each release takes back the credits of the copies that can no longer hold the box.
- **Known gap:** a write after the other name's last use is accepted, but the other name still holds the value until its scope ends, so the write copies at run time. `const team = teams[0]; resetTeam(mutable team)` with `teams` unused afterwards copies `team`; the debug copy check reports it. The sample `mutable_for` now takes `copy(teams[0])`.

## D95. Designs for the parked syntax: async exit, contracts, fixed arrays, build conditions, inline, C interop, dynamic traits

Decision (maintainer, 2026-11-09). Each part replaces the matching SPEC.md text when its node is built.

- **Async exit (amends D93).** The program exits when the entry function finishes; for an async entry, when its `Async<int>` resolves. Tasks still running are cancelled. `await Async.idle()` waits until no other task on the loop is running or ready, including tasks started while it waits. `Async.detach` work lives as long as the program. Async generators (`async function*`, `AsyncGenerator<T>`, `for await`) are a separate node, needed for HTTP and streaming.
- **Contracts: proven at compile time.** `requires <condition>` names parameters (`requires count >= 0`); the caller proves it from constants and narrowing. `ensures <pattern>` is a pattern the output must match, with no named subject: `ensures >= 0`, `ensures Some(>= 0) | None`, `ensures Ok(> 0) | Err(_)`, `ensures #[>= 0, _]`, `ensures Rect { width: > 0 }`, `ensures "GET" | "POST"`. Comparison patterns (`>= 0`) need an ordered type; literal and constructor patterns work for any type with `==` and literals. On a generator `ensures` holds for each yielded value; on an async function, for the awaited value. Anything unproven is a compile error. The prover is the one proven-operands builds (D92); contracts extend it. Or-patterns become a prerequisite.
- **Fixed arrays.** `FixedArray<T, N>` with constant type parameters `const N: int`, inferred from argument or expected types or written as `f<3>(x)`; `N` is also a constant value in the body; no arithmetic on `N` in types. Indexing never panics: a constant index is checked at compile time, and a variable index must be proven (`if (i < 4)`, `for (let i = 0; i < N; i++)` with `i` unassigned in the body, `Ranges.upTo(N)`). `FixedArrays.get(values, i): Option<T>` covers the rest. No literal types: Narra has no union types (D73), and contracts and refinements cover restricting a type to listed literals.
- **Build conditions.** As SPEC.md section 12: `Build.os`, `Build.arch`, `Build.debug` are constants; an `if` testing only them keeps one branch. Every branch is type-checked and only the matching one is emitted (dead-code elimination is guaranteed). `@target` declarations are type-checked on every target and emitted only where kept code reaches them (see the build notes below). One target table is shared with cross-targets.
- **Inline.** `@inline` is a guarantee, not a hint: a recursive `@inline` function is an error; it works across packages; inlining never changes behavior.
- **C interop.** C functions are imported in a block that names their library, like an import: `extern { function crc32(crc: u64, data: Ptr<u8>, length: u32): u64 } from "z"`; `from "c"` needs no extra link. Every function in the block is `uses unsafe`. `@extern("deflateInit_")` gives a member its C name; on a Narra function with a body, `@extern("compare_ints")` makes it callable from C under that name. `@layout(C)` fixes a struct's C layout and `@layout(u8)` gives enum values (renamed from `@repr`). `unsafe { ... }` is the only place pointer operations and unsafe calls are allowed; pointers are functions (`Ptr.read`, `Ptr.write`, `Ptr.offset`, `Ptr.alloc`, `Ptr.free`), not operators. `@trusted` stops the `unsafe` effect at a hand-checked wrapper; `narra audit` lists them. `CString.from(text): Option<CString>` (None when the text holds a zero byte). Code inside `unsafe` may crash: the stated exception to D92. `[link] libraries` in narra.toml is replaced by `from`. Shipped in two steps: imports, `unsafe`, `@trusted`, `Ptr`, `CString` and audit first; `@layout` and C-callable functions with c-abi-export.
- **Dynamic traits.** `Dynamic<Trait>` and `Dynamic.of(value)`; calls name the trait (`Drawable.draw(item, mutable canvas)`). Only traits whose functions take `Self` first and never return `Self`. No `==` on a `Dynamic`, one trait only, no downcast in the first version (`Dynamic.as<T>(d): Option<T>` would be sound, since it compares the stored type, but encourages type switches).
- **Reason.** Each design keeps the syntax TypeScript-like, adds no hidden calls (D12) and no panics outside `unsafe` and array indexing (D92).

Built: `requires` (node contracts, 2026-11-11). Amendments from the build:

- **Conditions are integer comparisons.** A condition compares one integer parameter with an integer constant (the constants the D92 division proofs accept), joined by `&&`, because the prover keeps facts only for integer names. A float parameter, a field path, a call, `||` or two parameters compared is E0984. Floats and fields are a later extension.
- **Calls prove each condition from the argument**: a constant, or a name with a narrowed fact, including the caller's own `requires`. An unproven argument is E0985. In the body the conditions are facts about the parameters, so `requires d != 0` lets the body divide by `d`.
- **No function values.** A function with `requires` cannot be used as a value (E0986) and `typeof` of it is E0982 (maintainer, 2026-11-11): a function type carries no condition, so a call through the value could pass an unproven argument, while the body relies on it with no runtime check left. Carrying conditions in function types is a later design. `ensures` has no such problem: dropping a known fact about a result is safe.
- **No `requires` on functions of an `implement` block** (E0984): they are called through the trait.
- **The ends of the int range are not facts.** A name with no check proves no bound, even `value >= -9223372036854775807`. A u64 constant from 2^63 up (a literal such as `18446744073709551615`, also as a divisor) counts as above every bound.

`ensures` (node contracts-ensures, 2026-11-11). Amendments from the build:

- **Comparison patterns are integer-only and live in `ensures` only.** `< c`, `<= c`, `> c`, `>= c`, `!= c` with a constant bound that fits the place's type; a non-integer place, a bound that is not a constant or does not fit, a pattern that binds a name, or `ensures` on a function that returns nothing is E0987. `!=` was added by the maintainer (2026-11-11): a return is proven by a constant other than the bound or a fact that excludes it (for `!= 0` the non-zero flag or a range without 0), never from the unbounded ends; a caller learns non-zero from `!= 0` and nothing from another `!= c`. A comparison in `match` or `is` is a parse error.
- **A struct pattern is written `{ width: > 0 }`**, as in `match`; patterns carry no type name.
- **Returns are proven where they are written**, from the shape of the value: constructors, tuples and struct literals in place, literals written the same, and integer facts. A call is not proven from the callee's `ensures` (no pattern implication); a `yield*` in a generator with `ensures` is unproven. Unproven is E0988.
- **Callers learn facts about integer names only**, as the prover stores facts per local: `const n = f()`, `let n = f()` until assigned, pattern bindings of `f()` in `if`, `match` and destructuring, the same for a `const r = f()` whose type cannot change in place (not a struct, array, map, box or generator, which a `mutable` argument could change), and `for (const n of g())` for a generator. An or-pattern in the caller's pattern learns nothing.
- **Generic functions and `implement` members** may have `ensures`; the declaration's body is proven, and an `implement` member gives callers no facts (a call through the trait carries none). An async function has no `ensures` yet, since Narra has no async functions.

Build conditions are built (batch-7, node build-conditions), except `@target` and `--all-targets` (node build-targets). The host is the target: `Build.os` and `Build.arch` come from `/usr/bin/uname -s -m`, read once per compiler process and only for a package that imports `std/build`, through one target table (`selfhost/src/buildtarget.nra`); `NARRA_HOST_TARGET=<os>-<arch>` overrides it (fix of the fresh review: a `uname` searched on `PATH` could hang or vary), and `Build.debug` is false under `-release`. The effects of every branch count toward the function's effects, the dropped branches included, so a function has the same signature on every target (orchestrator, 2026-11-11). Besides `==`, `!=`, `!`, `&&` and `||`, an `is` test of a variant without fields folds too, since it tests only the constants. A host with no row in the table is E0989 at the import of `std/build`. Batch-9 (nodes host-target-builtin and seed-switch) replaced the `uname` call: the host now comes from `Build.hostTarget()`, which reads the C compiler's predefined macros, so checking does no I/O.

`@target` and `--all-targets` are built too (batch-7, node build-targets), for Narra functions; batch-8 (node c-audit-target) added the `extern` block form. A use of a `@target` function is checked per target: it is E0991 when the target compiles the use but has no such function. A target does not compile a branch its build conditions drop, nor the body of a `@target` function it does not have, so over all targets this is "a use outside a branch that guarantees that target". `narra check --all-targets` runs the check for each row of the target table as a debug and as a release build (a `Build.debug` condition can keep a use in one only), so a use that fails only off the host is found without that machine; each use is reported once, with the targets it fails on. Malformed attributes are E0990. A `@target` function is type-checked on every target; tree shaking leaves it out of the C of a target where no kept code reaches it.

Status of D95 after batch-7: built are `requires`, `ensures`, build conditions, and `@target` with `--all-targets`. Not built yet: async exit and `Async.idle`, fixed arrays, `@inline`, C interop (including `@target` on `extern` functions) and dynamic traits.

`@inline` (batch-11, node inline, 2026-11-12). Amendments from the build:

- **E1030 and E1031.** A call chain through `@inline` functions back to the first is E1030, reported once at the first call of the chain with the chain in the message; a call to a function without `@inline` ends it. A generator or an extern block member with `@inline` is E1031. An `async` function waits for async functions (batch-12).
- **Only called parameters are copied.** An arrow argument is copied into the body when every read of its parameter is a call outside any arrow function of the body. Any other use (passed on, captured, stored) builds the closure at the call, on the caller's stack as before.
- **A value use keeps a C function.** An `@inline` function may be used as a value; that use gets one ordinary C function and every call is still copied (maintainer recommendation in the brief: an error would break every program that passes a std helper as a value once std uses `@inline`).
- **No `static inline`.** The SPEC sentence that `@inline` functions are emitted as `static inline` C functions is replaced: C's `inline` is a hint and keeps the closure call.

Constant type parameters (batch-8, node const-generics, 2026-11-11), the first part of fixed arrays. Amendments from the build:

- **The kind is `int` only.** `const N: u8` and the other sized kinds are E0992 for now; they can come later without a syntax change.
- **Functions and structs only.** Enums and type aliases take no constant parameter (E0992).
- **An argument is an integer literal (`-3` included) or a constant parameter in scope.** A local, a top-level constant or a type in a constant slot, and a literal or constant parameter in a type slot, are E0993. Arithmetic in a type argument (`Vector<N + 1>`) is E0992 in the parser. In a call, `f<3>(x)` reads as type arguments because an integer between the brackets counts like a type name.
- **The generic declaration defers proofs on `N`.** It is checked once with `N` open, so a divisor or shift count that is `N` is proven in each instance, where `N` is a constant; `f<0>` dividing by `N` is E0979 at the division.

Unsafe code (batch-8, node unsafe-ptr, 2026-11-11), C interop step 1 part 1. Amendments from the build:

- **`unsafe { }` is a statement.** `unsafe` is a keyword only before `{`, so it stays an effect name. The block is not an expression.
- **What needs a block is a call with the `unsafe` effect**: the std/ptr functions that read, write, move or free (`alloc`, `free`, `read`, `write`, `offset`) and any function with the effect, so the effect pass checks both with one rule (E0996). `Ptr.null` and `Ptr.isNull` touch no memory and are safe.
- **What a `Ptr<T>` may point at**: numbers, `bool`, `char`, `Ptr`, `FixedArray`s of such values and structs whose fields are all such values (FixedArray fields included; maintainer, 2026-11-11, after unsafe-ptr and fixed-arrays landed together). A `FixedArray` of anything else is E0998, and the message names the element. Anything that holds a reference count, an enum, an opaque struct and a type parameter are E0998, because the pointer would bypass reference counting. Generic code over pointers waits for a bound that says "no counted fields".
- **A `Ptr` is plain data**: it can be stored in structs and arrays and passed across any function, `@trusted` or not. It is not `Equatable`, `Hashable` or printable; `Ptr.isNull` tests for null.
- **`@trusted` without an `unsafe` block is an error, not a warning** (E0997), and so is `@trusted` with `uses unsafe`: Narra has no warnings outside lint, and a trusted function that trusts nothing hides a mistake.
- **`Ptr.alloc` and `Ptr.null` take `T` from the expected type**, like `Arrays.withCapacity`.
- **Generators and traits cannot hide `unsafe`** (fresh review). A `for` loop or `yield*` over a generator whose body has `unsafe` is E0996 outside a block. An implementation may use `unsafe` only when its trait function declares it (E0800), since a call through a bound sees only the trait's effects; `@trusted` is allowed on implement block members to wrap unsafe code. The same subset rule for `alloc` and `io` is not enforced: implementations of pure trait functions allocate in existing samples and the guide.

Fixed arrays (batch-8, node fixed-arrays, 2026-11-11). Amendments from the build:

- **An index is proven at both ends.** `if (i < 4)` alone proves nothing, because `i` may be negative; the check is `i >= 0 && i < 4`. A C-style for loop proves its counter: `for (let i = 0; i < N; i++)` gives `i >= 0` when nothing in the loop assigns `i` and the condition bounds `i` with `<` (or with a constant bound the step cannot pass), so the step never wraps to a negative value. `Ranges.upTo(n)` and `Ranges.between(a, b)` give their bounds from what is known about `n`, `a` and `b`. Only a plain `int` name carries a proof; `values[i + 1]` is E0994.
- **The length is never negative** (E0995), and an empty `FixedArray<T, 0>` has no valid index. One value takes at most 1 MiB (E0995): a larger local can overflow a default thread stack, and C rejects arrays past a size of its own (fresh review, 2026-11-11).
- **Literals give every element**, and only an expected `FixedArray` type makes one; a literal argument does not tell a generic call its length (write `sum<2>(...)`), and a generic body with an open `N` cannot build one.
- **`FixedArrays.get` and `FixedArrays.length`** live in `std/fixedarray`, written in Narra; `get` proves its own index per instance.

C interop blocks (batch-8, node c-interop, 2026-11-11), C interop step 1 part 2. Amendments from the build:

- **Members are signatures only.** No type parameters, no `uses` (the effect is always `unsafe`), no body, one per line; `@extern("name")` is the only attribute a member takes. A member may be `export`. An empty block is an error.
- **Types are numbers, `bool` and `Ptr<T>`** (E1002), and a member may return nothing. `mutable` parameters are rejected.
- **C prototypes use assembler labels.** The C code calls a member by its generated name and declares it `extern T name(...) __asm__("symbol")`, so a header declaring the same function with other C types (`size_t strlen(const char*)` against `strlen(text: Ptr<u8>): u64`) never conflicts. `__asm__` labels are a GNU extension that clang and gcc both accept under `-std=c11`.
- **Reserved C names** (E1001): `main`, `narra_` and `lstr_` names, string literal names, names with a module hash, and the program's generated function and type names with their helper suffixes.
- **Only reached blocks link.** A library is passed to `cc` only when reached code calls a member of its block, so a block used only in a branch a build condition drops needs no library on that target. `"c"` passes nothing and `-lm` is always passed.
- **`CString` lives in `std/cstring`**, imported as `CStrings` beside the type, since a module alias cannot share a type's name: `CStrings.from`, `CStrings.pointer`, `CStrings.toString` and `CStrings.read` (copies a C string back, `Err` on invalid UTF-8).
- **`[link]` keeps only `search`**, passed as `-L` relative to the package; any other key is E1000.

Audit and `@target` on extern blocks (batch-8, node c-audit-target, 2026-11-11). Amendments from the build:

- **`narra audit [path]` reads source only** and prints one line per `unsafe` block and `@trusted` function, sorted by file, line and column: `src/x.nra:12:3: unsafe block in function name`, `src/x.nra:4:1: @trusted function name`. A block in a test reads `in test name`, and a trait or implement member is named with its trait or implementation (`@trusted function measure in implement Measure for Cells`). It exits 0, or 1 when the package does not load. Nested `Shared.update` calls are not listed.
- **`@target` before an `extern { } from` block applies to each member**, as if each were a `@target` function: a member the target lacks is not reached, and its library is not linked there. `@target` on a member inside the block is E1000, and so is any other attribute before the block.

Dynamic traits (batch-11, node dynamic-traits, 2026-11-12). Amendments from the build:

- **The trait rule.** Every function of the trait takes `Self` as its first parameter, once and not nested, and mentions `Self` nowhere else, the result included (E1023 at the written `Dynamic<Trait>`, naming the function). `Hashable` qualifies: `Hashable.hash(item)` calls the stored type's hash, and a `Dynamic` is still never a map key, since it implements no trait (orchestrator, 2026-11-12). A trait may still have `mutable value: Self` functions; a call of one through a `Dynamic` is E1025 (D81 rule 8).
- **A `Dynamic` implements no trait.** It does not meet a bound such as `<T: Drawable>` (E0525 as before), because a bounded body may pass a local of type `T` as `mutable Self`. `Dynamic<Show>` does not implement `Show` either, so `Show.show(items)` on an `Array<Dynamic<Show>>` is rejected; loop and call `Show.show(item)`. No `==` and no map keys follow from the same rule.
- **`Dynamic` is a prelude name with one function.** `Dynamic.of(value)` takes the trait from the expected type (E1024 without one, with more than one argument, or for another function name); a user type named `Dynamic` hides it. `Dynamic.as<T>` stays out.
- **One allocation, the table in the cell.** A `Dynamic` is a pointer to a counted cell `{ rc, table, value }`; a copy is one retain and the value is never copied, so `copy(item)` copies nothing (E0975). The table of a (trait, type) pair is a static const struct: a drop function first, then one thunk per trait function that calls the user block, the instance of a generic block (generic-implement), or the compiler's `Show` function of the type.
- **Calls keep `Ref.Unresolved`**, as a call through a bound does, with the `Dynamic` in `traitSelf`; so the effect and shared-result passes treat it as a bound call, and its effects are the trait function's declared ones. `Dynamic.of` is `alloc`.

## D96. `typeof` in type positions

Decision (maintainer, 2026-11-09): `typeof name` in a type annotation stands for the declared or inferred type of a value in scope: a `const`, `let`, parameter, top-level constant or function. For a function it is the full function type, with effects (D86) and `from` (D94).

```ts
const table = Maps.create<string, Array<Entry>>()
function fill(target: typeof table): void { ... }

export function handle(request: Request): Response uses io { ... }
export type Handler = typeof handle      // (request: Request) => Response uses io
```

- Names only in the first version, not field paths or expressions. A `typeof` that refers to its own declaration is an error.
- **Rejected: runtime `typeof x === "string"`.** Every type is known at compile time; enums and `match` cover values of several shapes.

Built (type-of). Amendments from the build:

- A parameter can be named by the parameters after it in the same signature.
- A generic function, a function with effect variables and a function with a `mutable` parameter have no function type to write, so `typeof` of one is E0982, like a type or a module.
- A constant without a written type has its type from its value, which is known only once the constant is checked. `typeof` of it works in bodies and across modules, but not in a declaration of the same module (a signature, field or alias), where it is E1003 and asks for the written type. Lifting this needs the declaration passes to check constants on demand.
- E0981 and E0982 are the codes; a field path or an expression after `typeof` is the parser's ordinary unexpected-token error.

## D97. A package lists the module paths it exposes

Decision (maintainer, 2026-11-09): `narra.toml` may have an `[exports]` table that maps import paths to module files. Only listed paths can be imported from other packages; every other module is internal to its package.

```toml
[exports]
"json" = "src/index.nra"                  # import * as Json from "json"
"json/parse" = "src/parse.nra"
"json/stream" = "src/streaming/reader.nra"
```

- Without `[exports]`, every module is exposed by its path, as today.
- Inside the package, modules import each other by their real paths.
- A key must be the package name or start with it; the bare package name maps to one module.
- Importing an unlisted path is a compile error that lists the exposed paths. `narra move` updates the table.
- `std` uses the same table, so `std/internal/*` becomes an ordinary unexported path and E0426 folds into the general rule.
- **Reason.** A package chooses its public surface, and its file layout can change without breaking importers.

Not built yet. Dependencies themselves (`[dependencies]`) are not built either (E0407).

## D98. Or-patterns in `match` arms

Decision (maintainer, 2026-11-09): a `match` arm may list alternatives with `|`, after ReasonML: `Token.Plus | Token.Minus => ...`. Alternatives nest inside a pattern: `Some(1 | 2)`.

- Every alternative binds the same names with the same types: `Shape.Circle(r) | Shape.Sphere(r) => ...`.
- A guard covers the whole or-pattern. When the guard fails for one matching alternative, the next alternative that matches is tried with its own bindings; the arm is taken if any alternative matches with the guard true (Rust semantics).
- A binding of the whole value keeps its full type; there is no narrowing to the listed constructors, so no union types appear (D73).
- Exhaustiveness and unreachable-arm checks count each alternative. Unreachable alternatives are reported for constructors and bools, not for integer or string literals.
- **Only in `match`.** An `is` test and a `const`, `let` or `for ... of` destructuring stay one pattern (E0974): allowing alternatives there only without bindings would make `is` and `match` patterns differ (maintainer).
- `narrac fmt` keeps short lists on one line and puts one alternative per line, each starting with `|`, when the arm does not fit.
- Contract patterns (D95) use the same alternatives.

Built. Contract patterns (D95 `ensures`) use the same alternatives (node contracts-ensures). E0967 and E0968 are alternatives that bind different names or types, E0974 reports an unreachable alternative and alternatives in `is`.

## D99. Enums that do not reach themselves store their payload inline

Decision (maintainer accepted, 2026-11-10): an enum with a fielded variant is a tagged union held by value, `struct E { int32_t tag; union { ... } u; }`, unless it reaches itself or its payload is over 32 bytes. Those two keep the counted heap payload of D55. This replaces D55's "Rejected: inline payload" line for enums that do not reach themselves.

- **Reaching itself.** A variant field reaches the enum through structs, enums and distinct types held by value. `Array`, `Map`, `Box`, generators and function values are pointers and stop the walk. Two enums that reach each other both stay on the heap. An enum that holds a heap enum (an AST) without reaching itself is inline.
- **Size limit.** The payload is the largest variant, each field rounded up to 8 bytes. Over 32 bytes (four words) the payload stays on the heap.
- **Copies.** An inline value is copied like a struct: retain and release switch on the tag and count the fields of that variant. An inline enum with no counted fields has no reference count functions. A variant without fields (`None`) never takes a retain.
- **Monomorphized.** `Option<T>` and `Result<T, E>` get the layout of each instance. A future layout table (layout-tables) must carry the size of each instance.
- **Language.** Unchanged. Payloads stay immutable, and constructors with fields still count as `alloc` for effects (D58), so a program's effects do not depend on a layout choice.
- **Reason.** Instructions retired, base and branch compilers built from `feat/batch-4` 8de99b4, macOS arm64. `narrac emit selfhost`: 3.28 G to 3.08 G (6.0% fewer), user time 0.26 s to 0.24 s, peak memory 130 MB to 125 MB. A loop of 20 M `Option<int>` and `Result<int, string>` returns and matches: 1.77 G to 0.44 G, 0.09 s to 0.02 s. Part of the compiler gain comes from caching which structs and enums hold counted values (`Checker.countedStructs`, `countedEnums`); on e64faa2 the same compiler with every enum on the heap took 3.53 G against 3.67 G for the base and 3.45 G inline.
- **Cost.** `bench/interp` takes 1.4% more instructions and about 20% more user time (0.15 s to 0.18 s). With the payload in the value, clang turns the `Add` and `Sub` arms of its hot `match` into conditional selects, which lengthens the dependency chain on the accumulator. The other bench programs compile to the same C.
- **Rejected: no size limit.** The compiler took 3.87 G instructions on e64faa2, 5.6% more than the base: `Option<Emitter>` is 808 bytes and a `Result` of a checker struct 704, and both are copied by value. With the cache, limits of 8, 16, 48 and 64 bytes gave 3.48, 3.46, 3.51 and 3.56 G against 3.45 G at 32.
- **Rejected: a pointer only for the recursive fields.** It needs a second rule for which fields are pointers, and a recursive enum then allocates for each recursive constructor anyway.

## D100. Regions are parked, and the allocator comes first

Decision (spike regions-design, for the maintainer at the gate): do not build `Region.run` now. Node regions stays a user feature with priority could, and is parked. Most of the measured cost is in `malloc` and `free`, not in reference counts, so a size-class allocator in `narra_alloc` comes first. The design below is what regions would need when someone builds them. Measurements: `roadmaps/selfhost/nodes/regions-design/prototype/measure.sh` on an Apple M2 (8 GB, macOS 27), instructions retired from `/usr/bin/time -l`, load average 2 to 5.

Measured on `narrac emit selfhost` (the compiler emitting its own C, 183786 lines). Each variant rewrites the compiler's emitted C and emits byte-identical C:

| variant | instructions | user s | peak memory |
| --- | --- | --- | --- |
| base (`malloc`/`free`) | 9.27 G | 0.57 | 135 MB |
| size-class free lists, frees reuse blocks | 5.45 G | 0.42 | 153 MB |
| bump allocation, counts kept, `free` does nothing | 5.04 G | 0.42 | 807 to 911 MB |
| bump, counts kept, no drop walk | 4.63 G | 0.36 | 898 MB |
| bump, no retain and no release at all | 3.75 G | 0.29 | 909 MB |
| base plus a sign check on every retain and release | 9.64 G | 0.58 | 135 MB |

One run makes 19.6 M allocations (580 MB), 510 K reallocs, 60.2 M retains, 70.3 M releases and 15232 copies on write.

- **Skipping counts (question 1).** A region value would carry a negative count, as string literals and stack closure environments do. The sign check costs 3.9% more instructions on the compiler, with no change in user time. On `bench/` it costs 4.9% on `interp`, 0.3% on `strings`, and under 0.4% either way on `sieve`, and nothing on `fib`, `matmul` and `particles`. But a negative count is never 1, so every `_unique` copies, and `Arrays.push` on a region array becomes quadratic. Region values can be written in place only when the checker proves no other name shares them, which is D94 (node shared-write-check). `copy(x)` of a region value must then copy at once instead of relying on copy-on-write.
- **Routing (question 1).** A per-thread "in region" flag read by `narra_alloc` routes arrays, maps, strings, boxes, enum payloads, closure environments and `StringBuilder`. It is not enough alone. A heap value stored into a region container is never released, because the container's release is skipped. Each store site must add the value to a list that the region releases at its end. A generator or coroutine created outside and resumed inside must clear the flag for the resume, and `Co.resume` must copy a region input. `Cell`, `Shared`, `Sender`, `Receiver` and `Task` objects stay on the heap with counts.
- **Escape (question 2).** Drop the parameter: `Region.run(() => ...)`. Allocation follows the flag, so the region value has no use, and no `Scope` rules are needed. Two compile errors need codes. First, the arrow assigns or passes as `mutable` a captured `let` or `mutable` parameter (D77 captures by reference, and reading stays allowed). Second, the result type contains an identity type, `Generator`, `Coroutine` or `Async`. A callee can still write a region value into a `Cell`, `Shared` or channel created outside, and the checker cannot see that. Those writes copy at run time when the identity object is not in the current region. SPEC already says so for channels and `Shared`. `Thread.start` inside a region copies the closure environment to the heap.
- **Copying out (question 3).** The emitter generates one copy function per type that holds boxed values: `string`, `bytes`, `Array<T>`, `Map<K, V>`, `Box<T>`, enum payloads, tuples and structs with boxed fields, `StringBuilder`, and function values (one per closure environment). Only region objects are copied. Heap objects inside the result are retained. A copied object leaves a forwarding pointer in its dead header, so shared subtrees are copied once. The cost is linear in the result's region part. For `narrac fmt` it is one string of the file's size. The compiler has no per-file summary today: the checker reads whole trees.
- **Growth and runtime buffers (question 4).** `narra_realloc` cannot call `realloc` on a region block. Growth takes a new region block and copies. The old block stays until the region ends, so doubling wastes at most the final size. The choice follows the owner's mark (`a->rc < 0`), not the thread flag, since a heap array with count 1 can grow inside a region. The raw `malloc` buffers in `selfhost/runtime/std_io.c`, `std_fs.c` and `std_test.c` are freed in the same call, `std_env.c` keeps one process-wide buffer, and the drop worklist uses plain `realloc`. None of them hold Narra values, so they stay outside the region.
- **Value to the compiler (question 5).** Per-module parsing does not fit: `Modules.load` keeps every tree for the whole compile. Per-function emission does not fit: `Consts.emitFunction` writes the shared `mutable emitter` (literals, closures, counters), which the first escape error forbids. Two phases have a clear end and a string result. `formatSource` (`selfhost/src/fmt.nra:257`) per file: `narrac fmt --check selfhost/src` is 14.21 G instructions base, 9.22 G with free lists, and 8.77 G at most with a region that skips all counts, 5% below free lists. `compileToC` (`selfhost/src/driver.nra:361`) as one region: at most 3.75 G against 5.45 G with free lists (31% fewer, 0.29 s against 0.42 s user), at 6 times the peak memory (909 MB against 153 MB). The 3.75 G bound skips the 15232 copies on write, which is safe only once D94 rejects shared writes.
- **Reason.** The size-class allocator gives 41% fewer instructions and 26% less user time on the compiler, with no language change and 13% more peak memory. A region adds at most 31% more on one phase, needs D94, deep copies, a store list, two errors and run-time copies at identity objects, and keeps all memory until the region ends.
- **Cost.** SPEC section 6 "Regions" and D48 stay as written until the node is built. The SPEC example's `(region) =>` parameter is wrong under this design.
- **Rejected: keeping counts in a region and skipping only `free` (bump row).** 7.5% fewer instructions than free lists, at 5 times the peak memory.
- **Rejected: skipping counts on outer values inside a region.** An outer value can be dropped during the region through a `Cell` write, so a borrowed pointer would dangle. The sign check is per object.
- **Rejected: a region per module parse or per emitted function.** The trees live for the whole compile, and emission writes shared state.
- **Rejected: never freeing in the compiler (one process-wide region).** 6 times the peak memory on the compiler's own source, and it grows with program size on an 8 GB machine.

Size-class allocator: built (node fast-alloc). `narra_alloc` takes blocks of up to 512 bytes, header included, from 32 free lists in steps of 16 bytes, carved from 1 MiB chunks that are never returned to the system. Larger blocks use `malloc`. An 8-byte header holds the class (0 for a `malloc` block), so blocks are 8-byte aligned. `narra_realloc` keeps a small block that still fits, moves it to another class or to `malloc` otherwise, and calls `realloc` on a `malloc` block. AddressSanitizer builds (`-sanitize`) and builds that define `NARRA_PLAIN_MALLOC` keep plain `malloc` and `free`. `narra_live` counts as before. The lists are one global set; making `NARRA_HEAP_LOCAL` `_Thread_local` gives one set per thread. Measured on `narrac emit selfhost` (same machine, median of 9 alternating runs, load average 25 to 46): 9.35 G instructions before, 5.33 G after (43% fewer), user time 1.48 s to 1.12 s, peak memory 133 MB to 150 MB (13% more).

Regions: not built yet.

## D102. Contracts get `||` and relations between parameters; no `!`, no `&&` inside a pattern

Decision (spike contract-logic-design; maintainer, 2026-11-12). Facts about one name stay one interval (D92, D95). A second kind of fact, a relation between two names, is added. Counts: `roadmaps/selfhost/nodes/contract-logic-design/count.py` on `feat/batch-11` (text heuristics; `-v` lists each site). `selfhost/src` and `std/src` hold no `requires` or `ensures` line today, so the counts are of checks that a contract could replace: 70 `if` and `while` conditions read an integer parameter, and 42 of them are entry guards (among the first three statements, the block returns or panics). Measured base: `narrac emit selfhost` retires 2.41G to 2.43G instructions (two runs, batch-11 start).

**1. `||` in `requires` (built).** `requires x < 0 || x > 10`, `requires (index >= 0 || index == -1) && count != 0`, `requires a != 0 || b != 0`.

- **Syntax.** A `requires` line is clauses joined by `&&`. A clause is one comparison, or comparisons joined by `||`. Each comparison is an integer parameter against an integer constant as today, or two parameters (part 2). Parentheses group a clause. `&&` inside `||` (`a > 0 && b > 0 || c > 0`) is E0984, with a hint to split it. The condition stays a conjunction of disjunctions, so each clause is proven alone.
- **Caller.** A clause is proven when one of its comparisons is proven from what the caller knows: a constant, a name a check narrowed, or a relation (part 2). A caller that checked `if (x < 0 || x > 10)` knows only the hull of the two sides: every int except zero, which proves neither side. So it passes `x` after a check that proves one side, or a constant. An unproven clause is E0985 at the argument of its first parameter, and the message names the whole clause (`requires x < 0 || x > 10`).
- **Body.** A clause whose comparisons all name one parameter gives that parameter the hull of its sides (`joinFacts`). `requires x == 1 || x == 3` gives `1 <= x <= 3`. `requires d < 0 || d > 0` gives non-zero, so the body may divide by `d`. A clause over two parameters gives nothing.
- **Each side is built as `refineFact` builds it.** On a `u64` name a bound from 2^63 up is negative as an int, and `refineFact` (`operands.nra:350`) drops such a comparison. A side with such a bound is the unknown fact, so the hull is unknown too. Otherwise `x == 9223372036854775813 || x == 1` on a `u64` would join to `x <= 1`, and `y << x` would pass with a count of 2^63 + 5.
- **The same hull in checks.** D92 narrowing gives nothing for a true `||` (or a false `&&`) today. It gets the same rule, so a body and an `if` know the same: in `if (d < 0 || d > 0)` the `then` branch knows `d` is non-zero. Only a `||` whose comparisons all name one local against constants. Anything else still gives nothing. Amends D92.
- **Rejected: a set of intervals per name.** `x < 0 || x > 10` would live as two intervals, and a caller's `if` could prove it. Every fact read (divisors, shift counts, fixed array indexes, contract arguments, `ensures` returns) and every narrowing (about 6000 checks in `selfhost/src` and `std/src`) would work on lists instead of one `low`, `high` and `nonzero`. No check needs it: the one guard that excludes a range, `std/src/json.nra:409` (`unit < 55296 || unit > 56319`), negates to an interval that `&&` already writes.
- **Count.** 0 entry guards need `||` in a `requires`. A guard joined by `&&` negates to a `||` condition. The script finds 12, and read by hand none rejects its input. 9 pick a branch for valid input (`byte >= 48 && byte <= 57` in hex digit readers, `lexer.nra:1004`, `lspserver.nra:56`, `json.nra:278`). 3 are the checked accessors that turn an index into an `Option` (`std/src/fixedarray.nra:5`, `lspnavigation.nra:266`) or a saturating multiply (`check.nra:1330`). Built because it is small, TypeScript writes it, and SPEC section 13 lists `||` among the forms E0984 rejects (maintainer).

**2. Relations between two parameters (built).** `requires low <= high`, `requires first < end`, `requires left != right`.

- **Syntax.** A comparison of two integer parameters of the same type with `<`, `<=`, `>`, `>=`, `==` or `!=`, with no offset. A relation may stand alone or be a side of a `||` clause. An offset (`low + 1 <= high`) is E0984, since the `+` wraps; so is the same parameter on both sides.
- **The fact.** A relation between two integer locals is the set of orders still possible between them: less, equal, greater (three bits). `<` is {less}, `<=` is {less, equal}, `!=` is {less, greater}. Narrowing intersects the sets. A set proves a comparison when it is a subset of the comparison's set. It lives in `checker.localNames` beside the intervals, keyed by the pair with the smaller local first, so it is scoped and undone like them. Each local keeps a short list of the locals it has a relation with, also in `localNames`, so ending one name's narrowing (an assignment, a `mutable` call, a loop that assigns it) ends its relations without a scan.
- **Where it comes from.** A `requires` relation, at the start of the body. A D92 check between two integer locals of the same type: `if (low <= high)` in the `then` branch, its opposite in the `else` branch, the body of a `while`, the right operand of `&&` and `||`, and the rest of the block after an `if` whose `then` branch ends. A local captured by an arrow function gets none, as for intervals.
- **Bounds flow one step.** When a relation is recorded, and when a name with relations is narrowed against a constant, each related name takes the bound the relation passes on. With `low <= high`, `low` takes the high end of `high`, and `high` takes the low end of `low`. `==` passes both ends. The step does not repeat through a third name, and relations do not chain: `a < b` and `b < c` give nothing about `a` and `c`.
- **What the body can use it for.** `high - low` can still wrap (`low = -6e18, high = 6e18` gives a negative result), so the relation proves nothing about a difference. It proves three things. (a) A call to another function with `requires low <= high` or a weaker relation (`<` proves `<=` and `!=`). (b) Bounds through the step: `requires low >= 0 && low <= high && high < 64` makes `low` a valid shift count and a valid index of a `FixedArray<T, 64>`, and gives `high >= 0`. (c) `ensures` returns from those bounds (`return Some(low)` against `ensures Some(< 64) | None`).
- **Caller.** A relation requirement is proven by any of these. Two constants that meet it, ordered as the parameters' type orders them (a `u64` constant from 2^63 up is above every other). Intervals that do not overlap in the needed way (`x <= y` when the high end of `x` is at most the low end of `y`, both bounded). The same local passed twice (`f(n, n)` proves `<=`, `>=` and `==`). A relation the caller holds between the two argument names, from its own checks or its own `requires`. The call records the relation set between its argument names when it is checked, since `checkContractCalls` runs after the narrowing has ended.
- **`u64`.** A relation is about the true values, which the emitted C compares unsigned, so the set is sound on `u64` whatever the bounds. The bound step and the interval proof use only bounded ends. On a `u64` an end from 2^63 up is no bound (as `narrowFact` treats it today), so neither passes a value from 2^63 up on as a small bound.
- **Count.** 7 checks compare two integer parameters, and 2 reject their input: `idcheck.nra:54` (`end < first`, returns `Err`) and `std/src/random.nra:32` (`low >= high`, returns `None`, as D92 requires of std functions). The other 5 are shortcuts for equal arguments (`sharedwrites.nra:571`, `603`, `789`) and `Int.compare` (`std/src/ext__int.nra:4`, `7`). About 100 checks compare two names with `<`, `<=`, `>` or `>=`, and about 117 with `==` or `!=`. Each records a relation when both are integer locals.
- **Rejected: relations with an offset (`a - b <= c`, octagons).** The offset needs an overflow argument at every use under wrapping, and the storage grows with the square of the number of names.

**3. `!` in `requires` (not built).** `requires !(x < 0)` stays E0984. The message gives the rewritten condition (`write x >= 0`), computed by negating the comparison (`!(a || b)` becomes `!a && !b`, each negated).

- **Reason.** `!` over a comparison is always rewritable, and the condition text is what a caller reads in E0985 (`requires x >= 0 of pick`). One spelling keeps the declaration and the error the same. 0 of the 70 parameter checks negate a comparison. 6 lines in all put `!(` before a comparison, 3 of them in checks (`count.py -v`, the `negation` rows). None negates a single comparison of an integer name with a constant.
- **Rejected: `!` over a comparison only, as D92 checks allow.** It costs as little as the hint, but adds a second spelling for each comparison and nothing a caller can prove.

**4. `&&` inside an `ensures` pattern (not built); several lines stay.** `ensures Some(>= 0 && < 64) | None` stays a parse error. Write one line per bound: `ensures Some(>= 0) | None` and `ensures Some(< 64) | None`.

- **The facts are the same.** A caller narrows by each line in turn (`seedPatternFacts` intersects them), and each alternative joins to its hull. So the two lines give `0 <= n <= 63` exactly as `&&` would. Checked with the batch-11 compiler: a result of `digit(byte)` with the two lines is a valid shift count (`1 << count`). Each `return` is proven against every line, so each value meets both bounds.
- **The one difference.** With two alternatives that bound the same place, lines are weaker on the return side. `ensures Some(>= 0) | Some(>= 100)` with `ensures Some(< 10) | Some(< 110)` accepts `return Some(50)`, which `Some(>= 0 && < 10) | Some(>= 100 && < 110)` would reject. A plain integer result has the same gap (`ensures >= 0 | >= 100`). Callers learn `0 <= n < 110` either way.
- **Count.** 24 functions return an `Option` or `Result` of an integer. 2 have a result bounded on both ends by constants, `readHex4` (`0` to `65535`) and `readCodePoint` (`0` to `1114111`) in `std/src/json.nra`, and two lines describe both.
- **Rejected: `&&` in patterns.** A new `Pattern` variant legal only in `ensures`, with a case in every pattern switch (checker, matrix, emitter, `fmt`, `astdump`, `idcheck`, `trivia`, `querysites`, `typesdump`), for no fact a caller does not already get.
- **Rejected: a range pattern (`0..64`).** Narra has no range syntax. Ranges are `Ranges.between(0, 64)`, a call.

- **Out of scope (maintainer, 2026-11-11).** Field paths and floats in conditions and facts: bind the field to a local first.
- **Prior art.** Ada and SPARK preconditions, Dafny `requires` and Eiffel `require` take any boolean, and prove it with an SMT solver or check it at run time. Narra proves with intervals and no solver, so only forms the facts can hold are allowed. Abstract interpretation over intervals joins a disjunction to its hull, which is the body rule here. A relation as a set of orders is the sign of `b - a` kept without computing `b - a`, so it needs no overflow argument.
- **Cost.** Node contract-logic, size L: `||` clauses and relations in `requires`, the hull and relations in D92 narrowing with the bound step, the `!` message, SPEC sections 13 and 4 (the D92 checks paragraph), samples and error tests. No new error code: E0984 and E0985 change their messages. No `Pattern` change. Every check between two integer locals in `selfhost/src` now writes a fact, so `emit selfhost` may rise; the node measures it.

Built (node contract-logic, 2026-11-12): parts 1 to 3 as decided. A call reports one E0985 for the failed one-comparison clauses of an argument, joined by `&&` as before, and one per failed `||` clause; a clause whose parameters all got constants names their values: `5 does not satisfy`, or each parameter in parameter order, `x = 5 and y = 1 do not satisfy` (a new E0985 format). An offset is any comparison of two parameters with `+` or `-` on a side, `x - y >= 0` included. In D92 checks the hull covers a `||` whose sides are joined by the same operator; a side that is not a comparison of the one local with a constant gives nothing. The bound step also runs when an `ensures` fact or a fixed array loop narrows a local, since both go through `narrowFact`. `emit selfhost`: 2.470G instructions before, 2.482G after, on the same source (+0.5%).

## D103. Array patterns; no destructured parameters and no destructuring assignment

Decision (spike destructuring-design; maintainer, 2026-11-12): build array patterns, with a named rest that copies. Do not build destructured parameters or destructuring assignment. Array patterns keep TypeScript's syntax and meaning where Narra's rules allow. Bound names follow the rule for bindings that copy part of a value (D81): they are read-only copies. Counts: `roadmaps/selfhost/nodes/destructuring-design/count.py` on `feat/batch-10` (text heuristics; `-v` lists each site).

**1. Array patterns (built first, the only form built).** `[]`, `[only]`, `[a, b]`, `[first, ..._]` and `[first, ...rest]` in `match` arms and `is` tests, over `Array<T>` and `FixedArray<T, N>`.

- **Meaning.** A pattern without `...` matches an array of exactly that length. A pattern with `...` after `k` elements matches arrays of `k` elements or more. Element patterns nest and can be any pattern, including literals: `[_, "hash", text, output]`. `...` appears at most once, last.
- **Named rest.** On an `Array<T>`, `...rest` binds a new `Array<T>` that holds the remaining elements, retained. Like a slice (D60), it costs linear time in the tail and one allocation per match that binds it. `[first, ...rest] => sum(rest)` is therefore quadratic over the whole array. `..._` binds nothing and copies nothing. On a `FixedArray<T, N>` a named rest is a compile error: its type would be `FixedArray<T, N - k>`, and D95 allows no arithmetic on `N` in types. Write `..._` and index the array.
- **Bindings.** Each bound element is a retained copy, read-only (D81). The emitter tests the length once and reads the elements without a bounds check.
- **Exhaustiveness.** Lengths are constructors in the usefulness matrix of `patterns.nra`. `L` is the largest element count in any arm, with or without `...`. The cases are the exact lengths 0 to `L` and "more than `L`". A pattern with `...` after `k` elements covers every length from `k`. So `[]` and `[_, _, _, ..._]` leave lengths 1 and 2 uncovered. A missing case is reported as a pattern (`[_]`, `[_, _]`, `[_, _, _, _, ..._]`). Guarded arms do not count toward exhaustiveness. Unreachable arms are E0983 as today.
- **Irrefutable patterns.** `[..._]` and `[...rest]` match an `Array` of any length, so they are irrefutable when they appear alone, and `const [...rest] = items` is allowed (a copy, as `copy(items)` would be). Every other pattern on an `Array` is refutable: `const [a, b] = items` is E0715 (the existing "this pattern does not match every value; use match or `is`" error). No new panic: D92 keeps array indexing as the only array panic, and a hidden length panic would be a second one. On a `FixedArray<T, N>` with a known `N`, a pattern of exactly `N` elements, or `..._` after at most `N`, whose elements are irrefutable, is irrefutable and works everywhere a struct pattern does: `const`, `let`, `for ... of`, and nested in tuple and struct patterns. Any other length is a compile error at the pattern. In a generic body where `N` is a type parameter, only `[..._]` is irrefutable; any other FixedArray pattern in `const`, `let` or `for ... of` is the same length error, because one instance cannot prove it for every `N`. In `match` and `is` such patterns are length tests, checked per instance.
- **Other types.** An array pattern on a value that is not an `Array` or `FixedArray` (`string`, `bytes`, `Map`, `Range`, a generator) is a compile error.
- **TypeScript reading.** `const [a, b] = items` is valid TypeScript that gives `undefined` for a short array. In Narra it is a compile error with a fix, which keeps SPEC section 2's promise.
- **Count.** 352 tests of `Arrays.length(x)` against a constant; 53 of them read `x[0]` to `x[9]` within 6 lines, and 15 of those are the argv dispatch in `selfhost/src/main.nra` (`Arrays.length(arguments) == 4 && arguments[1] == "hash"` becomes `[_, "hash", text, output] =>`). 15 loops start at index 1 of an array (`Ranges.between(1, Arrays.length(x))`, for example `emitexprs.nra:270`, `fmt.nra:1454`, `ctors.nra:437`, `effects.nra:243`). They walk the tail after a first element handled apart, the shape of `[first, ...rest]`; some also use the index, and those stay index loops.

**2. Destructured parameters (rejected).** `function area({ width, height }: Rect)` and `(#[x, y]: #[int, int]) => x + y` are not built. The node param-destructuring is dropped.

- **Reason (maintainer).** A destructured parameter has no name, so a call with named arguments (D87) cannot pass it by name. Positional arguments must come before named ones, so a caller also cannot name any parameter before it. The form would weaken named arguments, and the maintainer keeps named arguments whole.
- **Reason.** 3 functions would shorten (`std/src/shape.nra` tuple2ToValue, tuple3ToValue, tuple4ToValue). 85 plain struct parameters are read only through 3 or more fields, but reading `rect.width` is as clear. 0 arrow functions destructure their parameter. `const { a, b } = parameter` on the first line does the same job.

**3. Destructuring assignment (rejected for now).** `#[x, y] = #[y, x]` and `{ x, y } = point` into existing names are not built. The node destructuring-assign is dropped.

- **Reason.** 4 sites would shorten, and every one needs field or element targets: 1 element swap (`std/src/random.nra:57`) and 3 runs of field copies from one value (`selfhost/src/stmts.nra:669`, `989`, `1115`, as `checker.loops = finished.loops`). 0 sites assign the parts of a tuple result to existing names. A form limited to `let` names would shorten none of them.
- **Reason.** Neither spelling is TypeScript text. TypeScript writes `[x, y] = [y, x]` with array literals, which in Narra would be an array pattern assignment, refutable on `Array`. The tuple spelling `#[x, y] = ...` is new syntax. `{ x, y } = point` at the start of a statement is a block in Narra, and TypeScript itself needs `({ x, y } = point)`.
- **Reason.** Field and element targets need rules of their own: all reads before any write, a target whose index reads a name that an earlier target assigns, and E0978 for each target against the temporaries. The three-line swap already states the order.
- **Rules if it is reopened.** The right side is evaluated whole into temporaries, then each target's place is evaluated and written left to right. Targets are a local `let`, a field or element of a changeable value, or the contents of a `mutable` parameter (the targets of `=`). A target whose path reads a name that an earlier target in the same pattern assigns is a compile error. Each write is checked by E0978 as the separate assignment would be. Only the tuple form; the struct form waits for a parenthesized statement form.

- **Rejected: a named rest that is a view.** Narra has no slice type, and a view into an `Array` would need a parent pointer, the cost D60 avoided for strings.
- **Rejected: `const [a, b] = items` on an `Array` with a length panic.** A new panic site against D92.
- **Rejected: only `..._` in the first build.** It avoids the hidden copy, but the 15 tail loops are the shape of `[first, ...rest]`, and the copy has the same cost as a slice, which D60 already accepts (maintainer).
- **Prior art.** TypeScript array destructuring and rest elements. Rust slice patterns (`[first, .., last]`, `[first, rest @ ..]` binding a slice view) and their exhaustiveness over lengths. Swift and Go have no array patterns. Kotlin destructures data classes but has no array patterns.
- **Cost.** A new `Pattern` variant in the AST, its dump, `fmt`, `idcheck` and `trivia`; length constructors in the matrix; a length test and a tail copy in the emitter; FixedArray length checks, which need the scrutinee type where irrefutability is decided. About 600 lines. Codes in the patterns range: E0722 (`...` twice or not last), E0723 (FixedArray length), E0724 (named rest on a FixedArray), E0726 (an array pattern on another type).

Built: part 1 (node array-patterns, 2026-11-12). Amendments from the build:

- E0722 is a parser error: anything after the rest, a second `...` included, fails at that token. A trailing comma after the rest is the same error, and `narrac fmt` writes none.
- An array pattern in `ensures` is E0987 (a new message row), since the prover keeps no array facts.
- A generic instance repeats its declaration's checks with `N` known. There, a FixedArray length that cannot match is not an error, and no `match` is judged again for unreachable arms or exhaustiveness: the declaration's own body already judged every match, against the open length where `N` is a parameter, where only `[..._]` or `_` covers every length. The emitter writes `false` for the test of an arm whose length cannot match in the instance.
- A FixedArray length error in a destructuring stands in for E0715, so `const [a, b] = values` over a generic `N` reports E0723 alone.

## D104. A local holds its value until its block ends

Decision (node checker-emitter-copies, 2026-11-12). Closes the first E0978 gap that D94's shared-write-error amendment left open, and keeps the second one documented. A debug-built compiler now compiles itself with `NARRA_CHECK_COPIES=1` and exits 0 (before: `array 956, map 971` hidden copies on `emit selfhost`, `array 3, map 1` on `check testdata/arrays.nra`).

1. **A write after another name's last use is an error when it copies.** A `const` or `let` local holds its value until its block ends, as the emitter releases it there. A write is E0978 while such a local shares the written place, also after the local's last use, when the write copies a heap object the local holds: an array, map, `Box`, `StringBuilder` or generator on the path from the shared place to the written place, or anything inside the place that a `mutable` argument passes. The shared value must be able to hold one: a struct of numbers never copies. An enum counts when a payload can hold one, since `Some(bag.items)` keeps the array of `bag.items` (fresh review); a `mutable` call on a struct whose shared part is an enum still copies nothing, because the payload is not a place. The error names the end of the block (`which load.items still holds until its block ends at line 21`) when no later use comes first, and offers the same `copy(...)` fix. A destructuring `const` or `let` counts too, and so does one initialized with `copy(...)`: a part taken out of a copy of an enum shares the original's part, and the copy's credits end when the copy or the original's part is released, so a later write would be a hidden copy (fresh review: `const snapshot = copy(bag.items); bag.items = []; let list = snapshot; Arrays.push(mutable list, 3)`). To change such a part, take `copy(...)` of the part. The bindings of an `is` or `match` on a value that is not a place (a call result, a constructor) hold it until the `if` block or the arm ends, and a `for (const ...)` binding over such a value until the body ends (fresh review: `if (makeList() is Some(found)) { let list = found; Arrays.push(mutable list, 3) }` copied unseen). Bindings of a place keep the earlier rule (a later use), since the place holds the value. A write that another error already points inside (between the start and the end of the written place) is not reported again.
   - **Rejected: releasing a local at its last use in the emitter** (no new errors). The emitter would need the liveness walk for every branch, loop, closure capture and generator state, and the point where a value is freed would move, which is also when a resource closes (section 6). It stays possible later: it only turns some errors back into accepted code.
   - **Cost.** 119 errors in `selfhost/src` and 2 in `selfhost/tests/src`. Most are a deep `mutable checker` or `mutable emitter` call beside a local read out of it, which the rule cannot tell apart from a write to that part. About 75 sites take the `copy(...)` fix (a retain at the call), the rest push a literal directly or move the value into its owner inside a helper, whose locals end at its return. A value taken out of a narrowed `Result` and then changed (`let names = listed.value; sortNames(mutable names)`) always copied; it now has to say so.
2. **Shared elements stay a documented gap.** `Arrays.filled(row, n)` and `[row, row]` make elements that share one value. The sharing belongs to the array value, not to a name: it travels through results, fields and parameters, and nothing is inferred across functions (D46, D94), so the rule would need a new marker in function types. The debug copy check reports it. The compiler had one such site (`filledArray([], n)` in `query.nra`), now one list per node.
3. **Found: a `mutable` parameter that takes another parameter's value.** `into.pairs = other.pairs` in a function with `mutable into` and a plain `other` leaves the caller's two arguments sharing after the call. Neither E0976 (results) nor E0978 (one body) sees it. It caused 1785 of the 1927 hidden copies (`mergeInto` in `sharedwrites.nra`, now `copy(other.pairs)`). Left as a known gap in SPEC section 6; a rule would treat a store into a `mutable` parameter like a result without `from`.
4. **Found: a part of a temporary copy of a heap enum.** `let row = takeLeftRow(copy(tree)); row.cells[0] = 9` compiles, and the copy check reports it: `copy(x)` shares nothing, and the temporary copy takes its credits back when the statement ends. An inline enum (`takeItems(copy(found))`) keeps the credit on the array itself and is explained. Left as a known gap in SPEC section 6.
5. **The copy check runs on `emit`.** `exitWithoutRelease` ends the process with `Process.exit` unless `NARRA_CHECK_LEAKS` is set, which skipped the copy report, so `NARRA_CHECK_COPIES=1 narrac emit` always exited 0. It now also returns normally when `NARRA_CHECK_COPIES` is set.
6. **CI.** The `copies` suite builds a debug compiler from `selfhost/` and runs it on `selfhost/` with the copy check; exit 103 fails it.

## D105. `@const` functions run in an evaluator inside the compiler

Decision (spike const-eval-design; approved by the maintainer on 2026-11-12, with every recommended answer): a top-level `const` may call a function marked `@const`. The compiler runs the call with a tree-walking evaluator over the checked AST and writes the result into the C. This amends D61's "a constant may not call a function": a constant may call an `@const` function, and nothing else. The call is written at the constant, so D12 still holds: no call is hidden. Measurements: a throwaway prototype in the spike's scratch folder (`roadmaps/selfhost/nodes/const-eval-design/log.md` has the table), Apple M2, macOS 27, `-release` builds.

1. **Evaluator, not compiled C.** The evaluator lives in `selfhost/src/consteval.nra` and walks the checked AST with the checker's side tables. Values are Narra values in the compiler (`int`, `string`, `Array`, `Map` and so on), so value semantics, wrapping arithmetic (D92) and map insertion order come free. Each std row the evaluator supports calls the same std function inside the compiler, which links the same runtime as the program. A test runs every `@const` sample twice, once at compile time and once as an ordinary call at run time, and compares the printed results.
2. **Marked, not inferred.** Only a function written with `@const` may be called from a top-level constant. Its body is checked where it is declared, against the rules of item 3, even when no constant calls it. Outside a constant it is an ordinary function. An exported function can change its body without changing its signature, so inference would let a private edit break a constant in another module; with the mark, the edit fails in the module that made it.
3. **What an `@const` body may do.** Its effects are at most `alloc` (D58: an exported one declares at most `uses alloc`). It may call other `@const` functions, constructors and the std functions of `std/string`, `std/array`, `std/map`, `std/option`, `std/result`, `std/char`, `std/range`, `std/int`, `std/float`, `std/math`, `std/bytes`, `std/fixedarray` and the sized number modules, as long as the call's effects stay inside `alloc`. `Math.sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow` are left out (see item 9). So are the functions that take a function, since an `@const` body cannot make or call a function value: `Options.map`, `andThen` and `orElse`, `Results.map`, `andThen` and `mapErr`, and `Arrays.map`. A call to any of them is E1052. The std source functions of the allowed modules run as Narra bodies. The first build (node const-functions) supports `std/string`, `std/array`, `std/map`, `std/int`, `std/float` and `std/range`: 45 registry rows. The other modules on the list follow in a later node. It may read top-level constants and use loops, `match`, `is`, patterns, `mutable` parameters and locals. It may not: call a function that is neither `@const` nor an allowed std function, use an arrow function or call through a function value, be a generator or call one, use `unsafe`, `extern` functions, `Ptr`, `Cell`, `Shared`, channels, threads or async, or have a type parameter with a trait constraint. Trait implementations: only the built-in and derived ones run (`==` and map keys on numbers, strings and derived types, templates of built-in types, the `Ordered` instances that `Math.min`, `max` and `clamp` use on numbers and strings). A user `implement` block reached from an `@const` body is E1052. Its parameter and result types hold only numbers, `bool`, `char`, `string`, `bytes`, `Array`, `Map`, `FixedArray`, `Option`, `Result`, tuples, and structs, enums and distinct types made of these.
4. **`alloc` is allowed.** Nothing outside the compiler observes an allocation at compile time, and tables, the main use, are arrays and maps.
5. **Limits.** Per top-level constant: 10 million evaluation steps (one per expression node, statement and loop turn), 64 MiB of values (counted at the evaluator's allocation points: 8 bytes per number, element, field and entry, string and bytes lengths rounded up to 8, 16 per array, map and box), and 1000 nested calls. A self-call in tail position runs as a loop, as section 11 compiles it, so it does not count toward the depth. Each limit is a compile error at the constant that names the limit and shows the call chain. The evaluator charges the memory budget before each std call that allocates, from its arguments (`Arrays.filled(value, count)`, `Arrays.withCapacity`, `Str.repeat`, `Str.join`, `Str.replace`, string concatenation and templates), so the call never runs when it would pass the budget. `Arrays.filled(0, 100000000)` would take 800 MB in the compiler; it is E1055 instead. Measured: a 256-entry CRC-32 table takes 26,370 steps (0.26% of the budget) and about 2 KiB of values; the prototype walker gives the same table as the compiled function and runs at 44 instructions per step, about 80 times slower than native code. With a margin of 10 for a full evaluator, 10 million steps take under 1 second. A 65,536-entry table with 16 bit turns per entry takes 13 million steps and 512 KiB, so it fails the step budget: that size is what `lazy` values or a generated source file are for.
6. **Panics become compile errors.** `panic`, a failed `assert`, an index out of range and the remaining std panics (`Str.sliceFrom`, `Str.charAt`, `Str.nextChar` on a bad position) stop the evaluation. The std functions that panic call `narra_panic` in the compiler, which would end the compiler, so the evaluator checks their preconditions (a position inside the string and on a character boundary) before it calls them, and reports the panic itself. The error is at the constant, with the message and the call chain from the constant down to the failing call. Division and shifts cannot fail: the checker already proves them (D92).
7. **Writing results into C.** A small writer turns the value into C, without building AST (a constant's body has no node ids for new nodes). A scalar result (number, `bool`, `char`, a constructor without fields) is a static constant: it is written at each use, as D61 does for scalars. Any other result gets the D61 getter, which builds the value from literals on first use. A float that is not finite is written as `NAN`, `INFINITY` or `-INFINITY` from `<math.h>`, since the printed form (`nan`, `inf`) is not C. Allowed code can make one (`Float.sqrt(-1.0)`, `1.0 / 0.0`). An array of scalars is written as a `static const` C array, and the getter copies it into a new counted array with one `memcpy`. The counted array itself is still built at first use, so D61's "no static C data for counted values" holds. Reason: the getter's one store per element is fast at run time (a 65,536-element getter adds 7.8 M instructions to a debug-built program's start), but clang takes 25 s at `-O2` on a 4096-element getter, 126 s on 16,384, and 153 s on 65,536 in a debug build (the `-O2` build was stopped after 4 minutes), against 0.17 s for a 65,536-element `static const` array with `memcpy`. Fully static counted values (a negative count, like string literals) are worth a node only when a profile shows first-use builds of non-scalar tables.
8. **Where calls are allowed.** Only as the value of a top-level `const`, or inside the constant parts of one (`const pair = #[square(2), square(3)]`), with constant arguments. Not in types (D95 allows no arithmetic on `N` in types), not in struct field defaults (SPEC section 4 asks for a compile-time constant there; the first build keeps defaults as they are), not in build conditions, and not needed in `lazy` values, which run at first force anyway. An `@const` body may read top-level constants. A constant that reaches itself through such a read is a compile error, like E0191.
9. **Determinism.** A result depends only on the source and the target. Maps keep insertion order, and `int` and sized numbers wrap the same way in the compiler and the program. Floats are IEEE doubles on both supported targets, and the compiler prints them with the same runtime function (`lstr_from_f64`). One gap was found: clang contracts `a * b + c` into a fused multiply-add on arm64 by default, even at `-O0`, so a program can compute a different last bit than the evaluator. The emitted C starts with `#pragma STDC FP_CONTRACT OFF` (checked: clang then emits `fmul` and `fadd`; gcc in `-std=c11` mode does not contract). The C library's `sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow` are not correctly rounded, and the compiler's host library can differ from the target's when building for another platform, so `@const` bodies may not call them. Arithmetic, `Math.sqrt`, `floor`, `ceil`, `round`, `trunc` and the conversions are exact in IEEE and stay allowed. The same constant gives byte-identical C on every run, as section 12 asks.
10. **Tooling.** `narrac check`, `build`, `emit` and the language service evaluate every constant that calls an `@const` function, since a failed evaluation is a compile error. Results are kept per constant for one run. Caching across runs waits for interface files (section 12); an exported `@const` function's body will then go into the interface file.

- **Reason.** SPEC sections 10 and 12 already promise `@const` functions. Tables computed from a rule (CRC, lookup tables, keyword maps) are short to write as code and long to write as literals, and computing them in the compiler costs nothing at program start.
- **Cost.** A second semantics for the language subset in item 3, kept equal to the emitted C by the run-twice test. About 1500 to 2000 lines in `consteval.nra` and the writer, plus one case per supported std row. Compile time grows with the work a constant asks for, up to the budget. Every float program loses FMA contraction, which can make float-heavy code slightly slower and changes the last bit of some results.
- **Rejected: compiling and running C.** `narrac check` and the language service would need `cc` and a link per evaluation (0.04 s for an empty C file, 0.42 s for a program with a 256-entry table), it runs host code when building for another target, and it makes the compiler depend on running what it builds.
- **Rejected: inferring purity from D58 effects.** See item 2.
- **Rejected: arrow functions and function values in the first build.** The evaluator would need closure values and captured places; no table sample needs them. They can come later without changing the rules above.
- **Rejected: forbidding floats in `@const` instead of the pragma.** Float tables built from arithmetic and `Math.sqrt` (gamma curves, fixed-point scales) are a common use, and the pragma also makes run-time float results the same at every optimization level.

Built in batch-11 (node const-functions) for the std modules string, array, map, int, float and range; the rest of the item 3 list is node const-functions-std. Limits as built, per constant: 10 million steps (E1054), where a std string, array or map call also costs one step per 8 bytes or elements it scans; 64 MiB of values (E1055), where `copy(...)` and storing a value read from a place charge the outer level again (`16 + 8 * length` for an array) for the copy its first write can make; 1000 nested calls (E1056), with a self-call in tail position run as a loop; a nesting depth of 5000 calls, blocks, loops and expressions under way (E1056), which keeps the evaluator near half of an 8 MiB stack on Linux, where gcc frames are larger (10000 overflowed there at about 9100 levels, found by the end-of-batch Linux check); and a result value depth of 1000 (E1056). An f32 literal is refused (E1052) when its f64 read is inexact and lands halfway between two f32 values. `check`, `build` and `emit` compute constants after the shared-write check; the language service still computes them before it. Arrays of strings are written one store per element.

## D106. Constraints and implement blocks on generic types

Decision (node generic-implement, 2026-11-12). SPEC section "Traits" lets any type implement a trait, and "Generics" lets a type parameter require traits; the compiler built both for functions and non-generic types only (E0021 and E0514, now removed).

1. **Constraints on struct and enum type parameters.** `struct Labeled<T: Show>` and `enum Tree<T: Ordered>` keep their bounds in the AST (`StructDecl.bounds`, `EnumDecl.bounds`) and in `Checker.bounds` for the type's parameters, as a function's do. Field types, implement blocks and derives see them. Each written use of the type with arguments (a type, a struct literal, an enum constructor, which also covers inferred instances) is checked once every module is declared: a failure is E1020. A type alias takes no constraints; `type P<T: Show> = ...` is the parser's ordinary `expected ">"` error.
2. **Bounds are not implied.** A type argument that is a type parameter needs the constraint declared on its own parameter: `function first<T>(pair: Pair<T>)` with `struct Pair<T: Show>` is E1020 with the hint `declare it as <T: Show>`. This matches the explicit style of the rest of the language and keeps `typeImplements` on a type parameter a plain lookup.
3. **Head syntax.** `implement Show for Pair<T>` writes the declaration's own parameter names, in order. A concrete type (`Pair<int>`), a missing list, the wrong count or another name is E1021. It keeps one name per parameter across the declaration and its blocks and rules out specialization.
   - Rejected: Rust's `impl<T>` binder (new syntax). Fresh names mapped by position (two names for one parameter).
4. **Where bounds come from.** Only the type declaration. An implement block adds no bounds of its own, so there are no conditional implementations (`Show` for `Pair<T>` only when `T: Show`). A type that wants `Show` for every instance declares `struct Pair<T: Show>`. Conditional implementations can come later on top of this.
5. **How it is built.** The block is recorded on the generic type, so every instance implements the trait. Each member is also a generic function with the type's parameters and bounds, under a hidden name, so its body is checked once with the parameters open and once per instance, and emitted per instance. A trait call on an instance, direct or through a bound, calls the member's instance. Every concrete instance of the type gets the members' instances, so `==`, map keys, derived `Ordered` and `Show` of an enclosing type call the user's functions. `findImplementation(checker, type, traitIndex, name)` in `selfhost/src/implementations.nra` returns the implementing `ProgramFunc` (a block function or a member instance); the Dynamic tables (D95) build on it.

Built (generic-implement).

## D108. Traits on distinct types

Decision (node distinct-traits, 2026-11-12). A distinct type lists traits like a struct or an enum. A listed trait needs an `implement` block in the declaring module or a `derives` entry; a derived trait uses the base type's implementation, because the C type is the same.

1. **Where the lists go.** After the base: `distinct type UserId = int implements Label derives Show`. It reads left to right like a struct head, the base stays next to `=` as in a type alias, and `implements` is a keyword, so a generic base such as `Array<int>` ends without ambiguity.
   - Rejected: `distinct type UserId implements Label = int`, which splits the name from its base.
2. **`==` and map keys stay.** SPEC gives every type without a function inside an automatic `Equatable`, and D62 kept `==` and map keys for distinct types. SPEC now says that arithmetic, bit operators and `<` do not carry over, which is all the old line meant.
   - Rejected: removing `==` and map keys from distinct types, which makes ids such as `UserId` hard to use.
3. **What a derived `Show` prints.** The base's text: `Show.show(UserId.from(7))` is `7`, and a distinct `string` is quoted. A distinct type has no fields, so there is no `UserId(7)` form to show.
   - Rejected: `UserId(7)`, which reads like an enum constructor that does not exist.
4. **`derives Shape` on an opaque distinct type** is allowed, as on an opaque struct: the declaring module chose it. `Shape.fromValue` then builds values outside the module.
5. **A derive needs the base to implement the trait.** Otherwise E1012 at the entry. The derivable traits stay Equatable, Hashable, Ordered, Show and Shape (E0503 for others).
6. **An own `implement` block wins.** A direct call and a generic instance `<T: Show>` with `T = UserId` run the user's function, not the base's: the emitter strips a distinct type to its base only when the type has no block of its own for the operation. An `implement Equatable` or `implement Hashable` block for a distinct type is E1014: `==` and map keys compare and hash the base value (D62), so the block would never run. (On a struct, `==` and map keys do call the user block.)
7. **How `derives Shape` is built.** The loader writes the two functions (`deriveshape.nra`) through the base's `Shape`; the unwrap call inside them uses the hidden name `__unwrap`, since the usual name (`toInt`, `toString`) depends on the resolved base. The checker accepts `__unwrap` only inside the generated `__shapeToValue_*` functions; anywhere else it is E0322 like any unknown function.

Built (distinct-traits).

## D109. How async is built: task frames, one loop per thread, codes E1100-E1199

Status: Approved by the maintainer, 2026-11-13. Spike async-design, 2026-11-13. It settles how the compiler and runtime build D93 and the async exit of D95, and it replaces the D17 text of SPEC section 8. Nothing here changes what D93 lets a program do, except where an item says "amends".

Amended 2026-11-13 (maintainer): idle rule.

Amended 2026-11-13 (maintainer): cancel passes down the await chain.

1. **Body transform: the generator state machine, with a task header.** An async function, an `async` arrow and an `async { ... }` block compile to a frame struct and a resume function, as a `function*` does (section 11). Every local, temporary and loop counter is a frame field, and each `await` that has to wait is a state.
   - **Shared with generators** (`emitgenerators.nra` and `emitstate.nra`): the frame-field path of `declareVariable`, `recordLive` (what is live at a state), `emitStateSwitch(emitter, false)` (the drop function), `unwindAll` before the end state, and the `inGenerator` flag that keeps body values out of C locals. They become exported, and the flag is renamed `inFrame` so both bodies set it.
   - **New** (`emitasync.nra`): a frame `A...` whose first field is the task header `narra_task` and whose second is `value`, the result slot of type `T`; a resume function `a...(narra_task* t)` that returns when the task waits or finishes; `await` (below); `return value`, which stores the value, unwinds, sets the finished state and calls `narra_task_finish`; and the creation function, which allocates the frame, stores and retains the parameters (or the captures of an arrow or block), then runs the body until it first waits, as a TypeScript async function does. There is no clone function: copies of a handle share the frame, so a task frame is never copied.
   - **`await handle`.** When the handle has finished, the value is read at once with no suspension. Otherwise the task registers on the handle (`narra_task_wait`), records its state and returns to the loop; it resumes when the handle finishes and then reads the value. The task takes its own reference to the value, so awaiting the same handle twice gives the same value twice.
   - **`await` may stand in any expression position of an async body** (maintainer, 2026-11-13): `g(await f())`, `if (await ready())`, `await a + await b`, a template, a struct literal, a block expression. The emitter keeps TypeScript's left-to-right order by spilling: in an expression that holds an `await`, each operand evaluated before it (the receiver and earlier arguments of a call, the left operand of a binary operator, earlier elements and fields of a literal) is stored in a frame temporary instead of a C scratch local, so it survives the suspension and the drop function releases it if the task is dropped there. `recordLive` lists those temporaries at the state. `&&`, `||`, conditional expressions (`c ? a : b`) and `if` values with an `await` in a branch that may not run become branches over a frame temporary, so the `await` runs only when its branch does. A `mutable` argument's place is taken as a pointer after the last `await` among the call's arguments, from index values spilled in order, since the pointer cannot be held across a suspension. Expressions without `await` emit as before, so generator and non-async code is unchanged.
   - **Rejected: a separate stack per task (stackful coroutines).** It needs a stack allocator and assembly or `ucontext`, which is deprecated on macOS, and it gives each task a fixed stack size.
   - **Rejected: the generator header and resume signature** (`bool (narra_gen*, void*)`). A task needs the loop link, the result slot and the waiter list, which a generator lacks; one more header type costs less than optional fields on every generator.

2. **Handle layout: the task frame is the handle.** `Async<T>` compiles to `narra_task*` for every `T`. The header holds a reference count, the state (running, waiting, done), a cancelled flag, the resume and drop functions, the tasks waiting on it, the `then` callbacks and a link for the ready queue. The value slot follows the header in the frame.
   - **Who holds a reference.** Each copy of the handle, each task in its waiter list, and the loop until the task finishes. A task nobody holds (after `Async.detach`) still runs to its end, and a waiting task cannot be freed while it is in a list.
   - **Finishing.** The body releases its locals when it returns, as a generator does, so a finished frame holds only the header and the value. `narra_task_finish` moves the waiters to the ready queue, runs the `then` callbacks unless the task was cancelled, drops the callbacks and drops the loop's reference. Releasing the last handle of a finished task frees the value and the frame.
   - **Dropping an unfinished frame** (only at exit, item 6) releases what is live at its state through the drop function and frees the frame.
   - Handles use plain reference counts and no copy-check hooks: nothing writes through a handle.
   - **Rejected: a handle object apart from the frame.** A second allocation per call would let a finished task free its frame early, but the body has already released every local by then; only the frame's own bytes stay.

3. **Where it lives.** `Async<T>` is a builtin type, `Type.TAsync(value: Type)`, beside `Type.TGenerator`. `Async` is a builtin type name like `Generator` and `Dynamic` (`resolve.nra`): it needs no import, and a type the module declares named `Async` hides it. Builtin names are not scope types, so `import * as Async from "std/async"` is not E0421. `std/async` is a source module (`std/src/async.nra`) over the primitives of a builtin module `std/internal/async`, whose rows are in `stdasync.nra`, registered from `stdlib.nra`.
   - **Runtime sources,** one per part, each generated into its own module by the existing `std_NAME.c` rule of the generated suite: `selfhost/runtime/std_async.c` (tasks and handles, `runtimestdasync.nra`, created by async-core), `std_async_loop.c` (the loop, `runtimestdasyncloop.nra`, created by async-core as a ready queue, then owned by event-loop) and `std_async_io.c` (sockets and files, `runtimestdasyncio.nra`, std-async-io). In `runtimetext.nra`, `runtimeFor("std/internal/async")` returns the task text and the loop text together, and `runtimeFor("std/internal/asyncio")` returns the I/O text. `runtimecore.nra` and `runtime.h` get nothing.
   - **Inclusion.** The checker adds `std/internal/async` to the used modules when it checks an async body, an `await` or a type that holds `Async<T>`. The emitter notes the helper when it emits a task frame, an `await` or an async entry. A program without async emits the same C as before.
   - **Reason for a builtin type.** The type comes from syntax, not from a std call: the checker gives an async function the result `Async<T>`, checks `await` operands and the unused-handle rule, and the emitter needs the C type of the result slot at each `await`. A source struct like `Cell` (D90) would need all of these special cases and also a typed primitive for the slot. Cost: a case in the switches that name `Type.TGenerator`, about 20 files.
   - **Amends the batch-12 brief:** the names `async.c`, `async_loop.c`, `async_io.c` and `runtimeasync.nra` become the `std_` names above, since the generated suite makes no case for the others.

4. **Where `await` is allowed.** In the body of an async function, an `async` arrow or an `async { ... }` block, in any expression position, including the block statements and block expressions nested in that body (item 1). Anywhere else it is E1100, and the message names the place: outside an async body, in a `function*` body, and in an arrow function that is not `async` (also inside an async body, since the arrow is a body of its own). `await` on a value that is not an `Async<T>` is E1102.
   - `async` and `await` become keywords. No source in `selfhost/`, `std/` or `testdata/` uses either as a name. `async` must be followed by `function`, an arrow function or `{` (E1108). `async function*` is E1107: async generators are their own node (D95). A nested `async function` stays E0930.
   - An async function's declared result is `T` (D93): `async function load(id: int): User` has the type `(id: int) => Async<User>`, and its `return` gives a `User`. A missing return is E0969 as for any function. `yield` in an async body is E0291.
   - `Async` with a number of type arguments other than one is E1109.
   - An `async { ... }` block follows the rules of a block expression (section 2): its value comes from `return`, which ends the block and not the enclosing function.

5. **Unused handle.** An `Async<T>` counts as used when it is awaited, assigned to a named `const` or `let` (the unused-local rule of section 2 then covers the name), passed as an argument (`Async.detach`, `Async.then` and `Async.cancel` included), returned, or stored in a field, an element, a tuple or a constructor. The rest is E1103: an expression statement whose value is an `Async<T>` (a call without `await`, an `async { ... }` alone), and `const _ =` or `let _ =` of one. The hint names `await` and `Async.detach`. Unlike an unchecked `Result` (E0229), `const _ =` does not discard a handle: the task keeps running, and D93 makes letting it run an explicit call.
   - **Rejected: a warning.** D93 makes it an error.

6. **Entry function.** The entry may be `export async function run(args: Array<string>): int`; this amends section 15, which says the entry is not `async`. `main` sets up the loop, calls the entry (which runs until it first waits) and runs the loop until the entry's handle is done; the exit code is its value. A plain entry runs the same way, and the `int` it returns is the code.
   - **Exit (D95).** Once the entry has finished, every task that has not is cancelled (item 8). The loop wakes each waiting task with `Err(AsyncError.Cancelled)` and runs ready tasks until none is ready. A cancelled task's next suspending call returns `Err(AsyncError.Cancelled)` at once, so a body that checks its results ends. Then the loop drops its reference to each task still unfinished (one that waits on a handle that cannot finish), which frees the frame through the drop function. The entry's code stands whatever the other tasks do (the open point of D93).
   - **No work left before the entry finishes.** When nothing is ready, no timer is set and no I/O is registered, but the entry's handle is not done, every task waits on another task. The program then stops with the message `async: every task waits on another task` and exit code 101, as a panic does. This needs a cycle of awaits, such as a task that awaits its own handle through a `Cell`. It is one more runtime stop beside the list of D92; the other choice is a hang.
   - The entry still has every effect.

7. **Loop and I/O.** One loop per thread: a `_Thread_local` struct in `std_async_loop.c`, so that threads (parked) each get one later. Each turn runs the ready queue in order until it is empty, then waits in the poller until the nearest timer deadline: `kqueue` under `__APPLE__`, `epoll` under `__linux__`, and a C `#error` on any other target. Timers are a binary min-heap ordered by deadline, then by creation order, so equal deadlines fire in the order they were set. There is no wasm32-wasi or Windows target today; when cross-targets adds one, async there is a compile error until it has a poller.
   - **`Async.idle()`** finishes when the loop has nothing ready and no timer or I/O wait pending, counting tasks started while it waits. Several idle callers finish one at a time, oldest first, so work started after an older one finishes holds the later ones. A task stuck in a cycle of awaits does not hold it, since it can never finish. A cancel wakes a task waiting in its own idle, and idle in a cancelled task finishes at once; it is not otherwise a cancel point.
   - **Sleep lives in `std/async/time`:** `sleep(duration: Duration): Result<Duration, AsyncError>`, with `Duration` from `std/time`, the name and argument of `Time.sleep`. Ok holds the time actually slept. This replaces the D17 `Async.sleep(milliseconds)`, so `std/async` keeps only the handle functions and `idle`.
   - **Blocking std calls in async code** (`std/fs`, `Time.sleep`, `Process.run`) are allowed and block the whole loop. They are not errors: a body may need them, and the effect system does not tell blocking `io` from suspending `io`.
   - **Files in the first build** (`std/async/file`): the call does the blocking read or write, then gives the other ready tasks one turn before it returns. kqueue and epoll do not wait on regular files. A worker thread can come later without changing the functions.

8. **Cancel.** `Async.cancel(handle)` sets the cancelled flag and drops the `then` callbacks; on a finished handle it does nothing. If the task waits on a timer or on I/O, the loop removes the wait and resumes it with `Err(AsyncError.Cancelled)`. If it is ready or running, its next suspending std call returns `Err(AsyncError.Cancelled)` without doing the work. A cancel passes down the await chain (amended 2026-11-13): when the cancelled task waits on another task's handle and is the only task waiting on it, that task is cancelled too, and so on down the chain, so the innermost wait (timer, I/O or `Async.idle`) returns `Err(AsyncError.Cancelled)`. A handle that other tasks also await is not cancelled. Detached tasks, and tasks the cancelled task started but does not await, keep running. A task cancelled while it runs passes the cancel when it next suspends on an `await`. `await` on a cancelled handle waits for the body to return and gives its `T`; `Async.peek` gives `None` from the cancel on.
   - **Error type.** `std/async` exports `enum AsyncError { Cancelled, Failed(message: string) }`. Every suspending function of `std/async/net`, `std/async/file` and `std/async/time` returns `Result<T, AsyncError>`, so the unchecked-result rule (E0229) makes each body look at a cancel. `Failed` holds the operating system's message, as the `string` errors of `std/fs` do.
   - **Sockets are counted handles.** `Listener` and `Connection` are opaque structs of `std/async/net` over a builtin `Handle` type of `std/internal/asyncio`, whose last release closes the descriptor (section 6). Functions take a handle by value, not as `mutable`: `Net.read(connection, 4096)`. `host` is a numeric address (no name lookup in the first build). `Net.listen(host, 0)` with `Net.localPort(listener)` lets a test take a free port.

9. **Effects and copies.**
   - **Effects** are charged where the handle is made (D93). A call of an async function charges its effects plus `alloc` for the frame, and an `async { ... }` block or a call of an `async` arrow charges the body's effects plus `alloc`. `await` charges nothing. The type of an `async` arrow is `(x: int) => Async<T> uses E`, where `E` is the body's effects. `Async<T>` itself carries no effects, unlike `Generator<T>`, because its body has already started.
   - **Parameters** are retained into the frame, and a parameter passed by pointer is copied into it (`f->name = *name`), as for generators. A `mutable` parameter on an async function or arrow is E1104: the task outlives the call, so it cannot keep a pointer into the caller.
   - **Captures.** An `async` arrow or block captures like an escaping closure wherever it is written, also as the argument of a non-escaping parameter (`Arrays.map(items, async (item) => ...)`): the frame gets each captured value at creation. Capturing a `let` or a `mutable` parameter is E1110, with the hint to keep the state in a `Cell` (D90).
   - **Shared writes (D94, D104).** The shared-write check treats a handle as holding every argument and capture of its task, as a result with `from` holds the parameter it names. While the local that holds the handle is live, a write to a place passed to the task is E0978, with the `copy(...)` fix at the argument. A direct `await f(items)` holds nothing after its statement, since the task has returned. After the handle's last use (passed to `Async.detach`, stored or returned), a task still running holds the value with no local to show it, and a later write copies; the debug copy check (`NARRA_CHECK_COPIES`) reports it. This is a documented gap of the same kind as D104 item 2.
   - `for (mutable item of items)` in an async body is E1105 (D85 rule 6).
   - `==`, `!=` and map keys on an `Async<T>` are E1106: a handle is an identity type (D80), and like `Cell` (D90) it gets no automatic Equatable or Hashable. A `Cell` or an `Async` must not reach another thread; the threads node rejects both.

10. **Debug checks.** At exit `main` drains and drops the unfinished tasks (item 6), then closes the poller descriptor, before `narra_consts_release`, the copy report and `narra_report_leaks`. A task cancelled at exit is therefore not a leak: AddressSanitizer and the leak check see every frame freed. A cycle through handles can still leak (a task whose frame holds a handle to a task that awaits it, or a handle in a `Cell` that its own task holds), and the leak check reports it with exit code 102, as D80 says. Under `NARRA_CHECK_LEAKS` an open socket counts as a live allocation until its handle is released. Async samples await or idle their tasks; one sample checks that tasks cancelled at exit leak nothing.

11. **Codes.** E1100-E1199 is the concurrency range (header of `diagnostics.nra`). Each node appends its rows inside its own block:
   - **async-core, E1100-E1129:** E1100 `await` outside an async body; E1101 left free (it was the statement-position limit the maintainer dropped; a code is never reused for another meaning); E1102 `await` on a value that is not an `Async<T>`; E1103 an unused `Async<T>`; E1104 a `mutable` parameter on an async function or arrow; E1105 `for (mutable ...)` in an async body; E1106 `==`, `!=` or a map key on an `Async<T>`; E1107 `async function*`; E1108 `async` not followed by `function`, an arrow function or `{`; E1109 the wrong number of type arguments for `Async`; E1110 an `async` arrow or block that captures a `let` or a `mutable` parameter.
   - **event-loop, E1130-E1149:** none expected; `Async.idle` and `std/async/time` are ordinary calls.
   - **async-api, E1150-E1169:** none expected; a misuse of `then`, `detach`, `peek`, `cancel`, `all` or `race` is an ordinary type error.
   - **std-async-io, E1170-E1189:** none expected; a closed handle is a runtime `Err`.
   - **E1190-E1199** stay free for threads and channels (a `Cell` or an `Async` sent to a thread).

- **The `std/async` functions (async-api).** `then(handle, escaping callback)` charges the callback's effects at the call; `detach(handle)` releases the handle; `peek(handle): Option<T>`; `cancel(handle)`; `all(handles: Array<Async<T>>): Async<Array<T>>` gives the values in order; `race(handles: Array<Async<T>>): Async<Option<T>>` gives the first value and `None` for an empty array, which could otherwise never finish. Neither `all` nor `race` cancels the handles it was given.
- **Reason.** Each part reuses something the compiler already has: the generator frame, the runtime text per module, the builtin type names, the unused-local rule and the shared-write check. Programs without async are not touched.
- **Rejected: the loop in std source.** The ready queue, the poller and the timer heap call kqueue and epoll; in Narra they would need `unsafe` code and `Ptr` in every async program.
- **Cost.** About 2500 lines over the four nodes (estimate): parser, checker and emitter cases for one type and three syntax forms, about 600 lines of C runtime, and four std modules. Spilling the operands before a nested `await` adds frame fields to async frames. A file call blocks the loop while it runs. A cycle of awaits stops the program with exit code 101.
