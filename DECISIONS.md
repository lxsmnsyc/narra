# Design decisions

Each entry states the decision first, then the reason and the rejected alternatives. Section numbers refer to `SPEC.md`.

## D1. Memory safety through reference counting, not a borrow checker or a garbage collector

Decision: immutable values with compiler-inserted reference counting and in-place reuse (section 6).

- **Reason.** Safe code has no references, so there is nothing to annotate and no borrow rules to learn. Frees are deterministic and there are no collector pauses.
- **Rejected: borrow checker with lifetimes (Rust).** Safe and fast, but lifetime annotations and borrow errors are the main source of complexity in Rust.
- **Rejected: tracing garbage collector (OCaml, Go).** Simple to use. But it adds a runtime, pauses and memory overhead that do not fit a native target.
- **Taken in part: mutable value semantics (Hylo).** Lang adds one optional parameter convention, `mutable` (D27). Parameters default to constant values, so most signatures carry no convention.
- **Cost.** Reference count updates in hot code that shares data. A prototype must benchmark this early, on numeric and string-heavy code.

## D2. TypeScript-style syntax with Rust-style semantics

Decision: braces, `const` and `let`, arrow functions, `import`, template strings and type annotations follow TypeScript. Types and semantics follow Rust: enums with fields, `match`, `Option` and `Result`, no `null`, sized numbers (section 2).

- **Reason.** TypeScript syntax is the most widely read syntax among the target users and in model training data.
- **Rule.** Where Lang's meaning differs from TypeScript's, Lang uses different syntax or a compile error with a fix (D42). The same text never means two different things.

## D3. Explicit imports with package-rooted paths

Decision: one file is one module, names are private unless exported, and every external name comes from an `import` declaration (section 3).

- **Reason.** The import list gives the dependency graph without name resolution, which keeps build scheduling simple.
- **Reason.** Every name in a file is either declared there or listed at the top. Reading the file or a text search finds where it comes from.
- **Rejected: globals.** Any file can read or write any global, so dependencies are implicit and typos create new variables.
- **Rejected: relative paths.** They break when a file moves. Package-rooted paths plus `lang move` avoid that.

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

Decision: dependencies are git repositories pinned to tags, versions resolve with minimal version selection, and `lang.lock` stores hashes (section 15).

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

Decision: types are written as named generics: `Option<T>`, `Array<T>`, `Map<K, V>`, `Tuple<A, B>`, `Result<T, E>`, `Generator<T>`. There is no `T?`, `T | null`, `{T}` or `{K: V}` type syntax, and no `??` operator (section 4).

- **Reason.** A name says what the type is. A reader who has not seen the language before can read `Option<string>`, but not `string?` or `{string}`.
- **Reason.** Names are easy to search for. Symbols such as `?` appear in many other places.
- **Cost.** Signatures are longer.

## D24. `const` and `let`

Decision: `const` declares a binding that cannot be reassigned, and `let` declares one that can (section 2).

- **Reason.** The keyword says whether the name can change.
- **Reason.** TypeScript uses the same words. Lang goes further, since a `const` value can never change, not only its binding. Section 2 lists this as a difference from TypeScript.
- **Note.** A top-level `const` must be a compile-time constant. Inside a function, `const` accepts any expression (section 10).

## D25. TypeScript import syntax, with namespace and named imports only

Decision: `import * as Name from "path"` and `import { a, b } from "path"`. There are no default imports and no dynamic `import(...)` (section 3).

- **Reason.** These are TypeScript's own forms, so they read correctly with no new rules.
- **Reason.** `import X from "m"` means a default import in TypeScript. Lang has no default exports, so that form is a compile error instead of a different meaning.
- **Rejected: a call-style import such as `const Name = import(...)`.** It looks like a runtime call, but importing runs nothing.

## D26. Struct literals, defaults, spread updates and destructuring

Decision: structs use typed literals (`Point { x: 1.0, y: 2.0 }`), constant field defaults, `{ ...s, field: value }` for updated copies, destructuring in `const` and `match`, and `Box.new` with `.value` for boxes. Field shorthand in literals is not allowed (section 4).

- **Typed literals.** Without `Point { ... }`, every struct literal would need an annotation or a context that gives the type.
- **Constant defaults.** Optional settings need no builder functions. Constant values keep building a struct free of hidden calls.
- **Spread updates.** `{ ...s, port: 9000 }` is the TypeScript form for an updated copy. Lang limits it to one spread, first, with the same struct type, so it cannot merge unrelated values.
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
- **Cost.** Deadlocks are still possible with nested locks. `lang audit` lists nested `Shared.update` calls.

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

## D34. Entry points set in `lang.toml`

Decision: each `[[binary]]` in `lang.toml` names an entry function as `"module.function"`. Its signature is `(args: Array<string>): int` (section 15).

- **Reason.** One package can build several binaries without a special `main` name or file.
- **Reason.** A package with no `[[binary]]` is a library, which needs no extra marker.

## D35. C interop details

Decision: `@repr(<integer type>)` enums with explicit values, `@export(C, "name")` for callbacks, `CString` in `std/ffi`, and a `[link]` table in `lang.toml` (sections 9 and 15).

- **Reason.** These cover calling C libraries that take flags, callbacks and strings, which most C libraries do.

## D36. Runtime failures abort

Decision: running out of memory, stack overflow and integer division by zero abort the program, in every build (sections 4, 6 and 11).

- **Reason.** These failures cannot be handled where they happen in a useful way. Aborting matches panics (D9).

## D37. Recursion is found by the compiler, and self tail calls are loops

Decision: there is no `recursive` keyword. The compiler finds recursion from each module's call graph. A self-call in tail position compiles to a loop in every function (section 2).

- **Reason.** The compiler already builds the call graph. Imports cannot form cycles (D4), so recursion never spans modules except through function values.
- **Reason.** A keyword would repeat what the compiler knows. Accidental recursion is rare in Lang, since names are not shadowed and imported functions are called with a module prefix.
- **Reason.** The tail-call guarantee covers every function, not only marked ones.
- **Cost.** Recursion is not visible in a declaration. `lang query recursion MODULE` lists recursive functions.
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

Decision: code that is valid TypeScript but would mean something different in Lang is a compile error. The diagnostic includes an exact fix (sections 2 and 14).

- **Reason.** People and models will write TypeScript habits. A compile error with a fix is cheap. A silently different meaning causes bugs.
- **Examples.** `===` becomes `==`. `null` becomes `None`. `xs.push(v)` becomes `Array.push(mutable xs, v)`. `import X from` becomes `import * as X from`. `'abc'` becomes `"abc"`.

## D43. No semicolons

Decision: a line break ends a statement when the line could end there. A line that ends with an operator, `,`, `(`, `[` or `{` continues. A `;` is a compile error (section 2).

- **Reason.** The rule depends only on the last token of the line, as in Go, so there is one reading.
- **Reason.** The known problems with semicolons in TypeScript come from making them optional with guessing rules. Lang has no optional form.
- **Rejected: required semicolons.** They add a token to every line without removing any ambiguity the line rule leaves.

## D44. Tuples instead of multiple return values

Decision: a function that returns several values returns a `Tuple` or a struct. The caller destructures with `const [a, b] = f()` (section 4).

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

Decision: `lang check --report-copies`, `LANG_TRACE_COPIES=1` and `@unique` let variables show where an update copies a shared value (section 6).

- **Reason.** Copy-on-update is correct but hidden. A copy of a large array in a hot loop is the main source of unexpected slowness in this memory model.
- **Reason.** `@unique` turns the fast path into a checked promise in debug builds, at no cost in release builds.
- **Rejected: making every copy explicit, as Rust's `.clone()`.** Every shared value would need copy calls, which goes against the value model.

## D50. Constant parameters and reads go through pointers

Decision: large constant parameters are passed by pointer, and reads of fields and elements go through pointers without copying (section 11).

- **Reason.** Value semantics do not require physical copies. Parameters are constant and access is exclusive, so a pointer is indistinguishable from a copy.
- **Reason.** It removes the cost of passing large structs and of reading from containers, with no syntax.
