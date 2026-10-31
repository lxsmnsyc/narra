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

Decision: a call passes every argument. There is no currying, partial application, pipe, method call, varargs, optional parameters or labeled arguments. Operators are built in, and there are no operators for joining strings, length or powers (section 2).

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

Decision: there is no C-style `for (init; cond; step)`. Counting uses `for (const i of Range.upTo(n))` (section 2).

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

## D50. Constant parameters and reads go through pointers

Decision: large constant parameters are passed by pointer, and reads of fields and elements go through pointers without copying (section 11).

- **Reason.** Value semantics do not require physical copies. Parameters are constant and access is exclusive, so a pointer is indistinguishable from a copy.
- **Reason.** It removes the cost of passing large structs and of reading from containers, with no syntax.

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

## D54. C names are hashed, built in one file

Decision: every C name comes from `compiler/internal/narra/cnames.go`. A declaration is `<kind>_<module hash>_<name>`. An interned type (generic instance, tuple, array, function type, box, enum payload) is `<kind>_<head>_<hash>`, with the hash over `head|arity|arg ids`. The hash is xxHash32, seed 0, written by hand. A process-wide table panics with an internal error if two different things get one name. `Emit` writes a comment table that maps each hashed name to its type. `__` and leading or trailing `_` are allowed in user names.

- **Reason.** Hand-built names caused three collision bugs in wave 2 (`__` in names, prelude names, tuple arity). One builder and a checked table remove the class.
- **Reason.** xxHash32 was 1.7 to 2.7 times faster than FNV-1a on short names, and mixes better. The self-hosted compiler ports it with `*%` and rotates.
- **Rejected: FNV-1a.** In the Go standard library, but slower and weaker.
- **Rejected: `hash/maphash`.** The seed is random per process, so names would change between runs.
- **Cost.** Generated C reads less well. The comment table and the readable head in each name make up for it.

## D55. Fielded enum variants live on the heap, and `Box<T>` exists

Decision: an enum with at least one fielded variant is `{ tag, payload pointer }`. The payload is counted and shared by copies, and never mutated. Freeing uses an iterative worklist, so deep structures do not overflow the C stack. `Box<T>` is a counted heap cell with a read-only `.value`, for recursive structs (`Box.new(x)`). A user type named `Box` hides the builtin.

- **Reason.** SPEC section 6 asks for this, and recursive enums (an AST) need it.
- **Cost.** Every `Some(x)`, `Ok(x)` and `Err(e)` allocates. The `interp` benchmark is about 2x slower until node `perceus` removes retains on borrowed values. The user accepted this.
- **Rejected: inline payload with a pointer only for recursive fields.** Layout would depend on the type's shape, and the self-hosted compiler would need two code paths.
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
- **Cost.** Function types carry no effects in the type checker yet. Calls through struct fields, call results and arrays of functions count as pure.

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
- **Cost.** No `implements` or `derives` on distinct types yet, and no generic distinct types.

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
