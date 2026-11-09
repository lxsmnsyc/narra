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

- **Meaning.** `x op= e` means `x = x op e` with the target place evaluated once, so `items[next()] += 1` and `items[next()]++` call `next` once. As for `=`, the value is evaluated before the target place. The operators are `+= -= *= /= %= &= |= ^= <<= >>=` and the wrapping forms `+%= -%= *%=`.
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
5. **Copies stay copies.** `const copy = hero` is a copy (copy on write, as before); changing `hero` afterwards does not change `copy`. A plain parameter is still a read-only copy: its fields cannot be assigned and it cannot be passed as `mutable`.

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
4. **Overflow.** Checked at the type's own width with `__builtin_*_overflow` on that C type, panicking in debug builds and wrapping in release builds. `MIN / -1` follows `int`. Division and remainder by zero panic in every build. Shifts by the width or more, or by a negative count, panic.
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
