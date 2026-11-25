# The Narra guide

This guide teaches Narra to someone who has programmed before, in any language. Read the chapters in order: each one builds a small program step by step and uses what the earlier chapters taught.

A test compiles and runs every program in the guide. Each one works with the compiler as it is today and prints exactly the output shown under it.

## Chapters

1. [Getting started](01-getting-started.md): install the compiler, run a first program, read an error, split a program into a package.
2. [Values and types](02-values-and-types.md): `const` and `let`, numbers and division, strings and templates, `bool` and `if`, loops, arrays, maps, sized numbers, math, random numbers and hashing, tuples.
3. [Functions and closures](03-functions-and-closures.md): functions, `requires`, default arguments, recursion, functions as values, arrow functions, generics, closures and `escaping`.
4. [Structs and enums](04-structs-and-enums.md): structs, copies and spread updates, enums and `match`, recursive data and `Box`, generic types, type aliases, `typeof`, distinct and opaque types, re-exports.
5. [Pattern matching](05-pattern-matching.md): `match` with guards and literals, array patterns, destructuring, `is` and `Option`, block expressions, `if` values and `? :`, named arguments.
6. [Option and Result](06-option-and-result.md): `std/option`, `std/result`, errors without exceptions.
7. [Traits](07-traits.md): declarations, `implement`, constraints, `derives`, equality, ordering and hashing, `Dynamic` values, saving values as JSON with `Shape`.
8. [Modules and packages](08-modules-and-packages.md): imports, aliases, exports, the standard modules, packages from git, workspaces.
9. [Memory and `mutable`](09-memory-and-mutable.md): reference counting, copies, writes to shared values, `mutable` parameters and their rules, `Cell`, leaks and hidden copies.
10. [Effects and testing](10-effects-and-testing.md): `uses`, effect checking, `std/test`, `narrac test` and golden files.
11. [Files and the system](11-files-and-the-system.md): reading and writing files, paths, folders, standard input and output, the command line and the environment, other programs and their options, bytes with hex and base64, time.
12. [Generators](12-generators.md): `function*`, `yield`, the effects of a generator, resuming by hand, coroutines.
13. [The standard library](13-standard-library.md): the modules larger programs are built on: `std/iterator` for lazy and eager operations over generators, `std/set` for distinct values, `std/arena` for graphs of values, `std/args` for command lines, `std/toml` for configuration files, `std/terminal` for colors and terminal detection, `std/regex` for patterns, `std/unicode` for text as people see it, `std/url` for addresses, `std/semver` for versions, `std/digest` for hashes and `std/log` for log lines.
14. [Async](14-async.md): async functions, `await`, waiting for all or the first, cancelling, files and the network in async code, async generators, shutting down on a signal.
15. [Threads and shared state](15-threads-and-shared-state.md): threads and scopes, `Parallel.map`, channels, `Shared` and `Atomic`, `Sendable`, threads or async tasks.
16. [Unsafe code and C](16-unsafe-code-and-c.md): calling C, the `unsafe` effect and `@trusted`, C types and strings, pointers, linking, `narrac audit`.
17. [Proofs and compile-time constants](17-proofs-and-compile-time.md): `requires` between parameters, `ensures`, fixed-size arrays, refinement types, constant type parameters, `@const` functions, tagged templates.
18. [Tools](18-tools.md): `narrac emit` and `narrac targets`, `narrac check` and its JSON output, `narrac explain`, `narrac apply-fix`, `narrac fmt` and its options, `narrac query`, `narrac test`, and `narrac lsp` for editors.
19. [Building for other systems](19-building-for-other-systems.md): `std/build` conditions, `@target`, `--target` and `--all-targets`, WebAssembly, Windows, debug and release builds, `std/simd`.
20. [Narra for TypeScript programmers](20-narra-for-typescript-programmers.md): optional reading that compares Narra with TypeScript, topic by topic.

## Reading the samples

A program in the guide is a whole file that exports `run`, followed by what it prints:

```narra
export function run(args: Array<string>): int {
  print("hello")
  return 0
}
```

```text
hello
```

A program that does not compile starts with a comment that holds the message the compiler prints. The line and column in it count the comment as line 1:

```narra error
// error: 3:10: expected int, found string
export function run(args: Array<string>): int {
  return "hello"
}
```

Save a sample to a file and run it with `narrac run file.nra` to try your own changes.

## Planned

The language is meant to have these features, and the compiler does not have them yet. The guide does not teach them.

- Functions that C programs can call, and `@layout` for structs shared with C ([chapter 16](16-unsafe-code-and-c.md#not-built-yet)).
