# Design decisions

Each entry states the decision first, then the reason and the rejected alternatives. Section numbers refer to `SPEC.md`.

## D1. Memory safety through reference counting, not a borrow checker or a garbage collector

Decision: immutable values with compiler-inserted reference counting and in-place reuse (section 6).

- **Reason.** Safe code has no references, so there is nothing to annotate and no borrow rules to learn. Frees are deterministic and there are no collector pauses.
- **Rejected: borrow checker with lifetimes (Rust).** Safe and fast, but lifetime annotations and borrow errors are the main source of complexity in Rust.
- **Rejected: tracing garbage collector (OCaml, Go).** Simple to use, but adds a runtime, pauses and memory overhead that do not fit a C-level target.
- **Rejected: mutable value semantics with `inout` parameters (Hylo).** Safe without lifetimes, but adds parameter conventions to every signature. Reference counting with reuse gets the same in-place updates without them.
- **Cost.** Reference count updates in hot code that shares data. A prototype must benchmark this early, on numeric and string-heavy code.

## D2. ReasonML semantics and ReScript syntax

Decision: expression-oriented code, variants, exhaustive `switch`, Hindley-Milner inference and labeled arguments (sections 2 and 4). Currying, partial application and the pipe are left out (D12).

- **Reason.** Immutable-by-default semantics fit reference counting (D1), because immutable values cannot form cycles.
- **Reason.** ReScript syntax is close to TypeScript, so both people and models can read it from prior knowledge.

## D3. Explicit imports with package-rooted paths

Decision: one file is one module, names are private unless exported, and every external name comes from an import (section 3).

- **Reason.** The import list gives the dependency graph without name resolution, which keeps build scheduling simple.
- **Reason.** Every name in a file is either declared there or listed at the top. Reading the file or a text search finds where it comes from.
- **Rejected: OCaml-style global module names.** Any file can reference any module by name without an import, so dependencies are implicit.
- **Rejected: relative paths.** They break when a file moves. Package-rooted paths plus `lang move` avoid that.

## D4. No import cycles, no `export *`, no default exports, no import side effects

Decision: the restrictions listed in section 3.

- **No cycles.** Modules compile in dependency order and in parallel. No module ever sees a partly initialized dependency.
- **No `export *`.** Wildcard re-exports slow builds and hide where names come from. This is a known problem with barrel files in TypeScript.
- **No default exports.** A default export gets a different name in each importing file, so text search misses uses.
- **No import side effects.** Initialization order can never matter, and startup runs no code before `main`.

## D5. No typeclasses or traits

Decision: behavior is passed explicitly as functions or modules. Equality and hashing are derived structurally by the compiler (section 4).

- **Reason.** Trait implementations are visible program-wide, which needs coherence and orphan rules and makes "which implementation runs here" hard to answer by reading code.
- **Cost.** More explicit arguments, such as `Array.sortBy(xs, Int.compare)` instead of `xs.sort()`.

## D6. No functors at first

Decision: modules cannot be parameterized by other modules in this draft.

- **Reason.** Functors are the most complex part of the OCaml module system. Generics plus explicit function arguments cover most uses.
- **Revisit when.** Real code shows repeated patterns that generics cannot express.

## D7. Polymorphic code compiled once, with opt-in specialization

Decision: generic functions take boxed values and compile to one copy. `@specialize` requests per-type copies (section 10).

- **Reason.** Per-type copies, as in C++ templates and Rust generics, are the main cause of large binaries and slow builds in those languages.
- **Cost.** Boxing in generic code. `@specialize` exists for hot paths.

## D8. A fixed set of effects in signatures

Decision: `alloc`, `io` and `unsafe`, declared on exported functions and inferred elsewhere (section 7).

- **Reason.** A reader, human or agent, sees what a function may do from its signature alone.
- **Rejected: a general effect system (Koka).** User-defined effects and handlers add a large feature to learn. A fixed set covers the questions readers ask most.

## D9. Results instead of exceptions, and panics that abort

Decision: `result` plus `try` for recoverable errors. Panics cannot be caught (section 5).

- **Reason.** Every way a function can fail is in its return type.
- **Reason.** Uncatchable panics need no unwinding tables, which keeps binaries small and C output simple.

## D10. C as the compile target

Decision: emit one C file per module (section 10).

- **Reason.** Every platform with a C compiler is a target from the start, and debug builds can use a fast C compiler.
- **Revisit when.** C compile time dominates debug builds. A native debug backend, like Go's and Zig's, is the next step.

## D11. Tooling made for agents

Decision: JSON diagnostics with exact fix edits, stable error codes, and a query CLI (section 13).

- **Reason.** An agent calls commands and reads their output. Exact edits remove a step where the agent rewrites a fix in its own words and can get it wrong.
- **Reason.** Explicit signatures (section 4), explicit imports (D3) and declared effects (D8) work together. An agent understands a function from its file and its imports alone.

## D12. No currying, partial application, pipes or operator overloading

Decision: a call passes every argument, there is no pipe operator, and operators are built in with one type family each (section 2).

- **Reason.** Each call has one reading. With currying, `f(a)` might be a result or a function waiting for more arguments. A missing argument then shows up as a type error somewhere else.
- **Reason.** Without currying, a function compiles to one C function with a fixed parameter list. No intermediate closures are allocated for partly applied calls.
- **Reason.** Without pipes, every call is written as `f(x, y)`. Code has one call form to read, search and generate.
- **Reason.** Without overloading, an operator's meaning follows from the operator alone. `+.` is always float addition, so the reader never needs the operand types first.
- **Cost.** Data transformations read inside out or need named intermediate values. Fixing arguments needs an explicit lambda.
- **Rejected: ReScript's `->` pipe.** It saves nesting, but adds a second way to write every call.

## D13. Uniform form with a tag bit, and heap boxes for 64-bit values in generic code

Decision: generic code receives one machine word per value. Types of 32 bits or less are stored inline with a tag bit. 64-bit numbers are heap boxed (section 10).

- **Reason.** `i64` stays a full 64-bit integer that matches C's `int64_t`.
- **Reason.** Non-generic code never boxes. Boxing happens only where a 64-bit value enters generic code, and `@specialize` removes it in hot paths.
- **Rejected: 63-bit `int` (OCaml).** Fast generic code, but `int` no longer matches a C type, which complicates C interop.
- **Rejected: a type descriptor passed with every generic call.** No boxing. But every generic call takes an extra argument, and generic code must read sizes at runtime.

## D14. Effect variables for higher-order functions

Decision: signatures can write `'e` after `uses`, bound by a function parameter (section 7).

- **Reason.** `map` with a pure function has only `alloc`, and `map` with an I/O function has `alloc, io`. A fixed effect list cannot state both.
- **Reason.** Inference writes the variable for private code. Only exported higher-order functions show it.
- **Rejected: higher-order functions accept only pure functions.** Too strict. Mapping over files to read them is common.

## D15. Strings indexed by checked byte offsets

Decision: string positions are `byteIndex` values, and slicing panics on a position inside a character (section 4).

- **Reason.** Byte offsets make slicing take constant time. Positions returned by searches are always valid.
- **Reason.** A distinct `byteIndex` type stops a character count from being used as a position.
- **Rejected: character offsets.** `s[i]` would take time proportional to `i`, which hides slow loops.
- **Rejected: iterators only.** Safe, but simple parsers need positions.

## D16. Git dependencies with minimal version selection

Decision: dependencies are git repositories pinned to tags, versions resolve with minimal version selection, and `lang.lock` stores hashes (section 14).

- **Reason.** Resolution needs no solver and gives the same result on every machine.
- **Reason.** No registry to run at the start.
- **Revisit when.** The number of packages makes discovery or name squatting a problem.

## D17. Async deferred

Decision: the first version has threads and channels only (section 15).

- **Reason.** async/await splits libraries into sync and async versions. Green threads need a scheduler, which goes against a small runtime.
- **Direction.** An `async` effect fits the existing effect system and keeps one version of each library.
