# The Narra guide

This guide teaches Narra to someone who has programmed before, in any language. Read the chapters in order: each one builds a small program step by step and uses what the earlier chapters taught.

A test compiles and runs every program in the guide. Each one works with the compiler as it is today and prints exactly the output shown under it.

## Chapters

1. [Getting started](01-getting-started.md): build the compiler, run a first program, read an error, split a program into a package.
2. [Values and types](02-values-and-types.md): `const` and `let`, numbers, strings and templates, `bool` and `if`, arrays, maps, tuples.
3. [Functions and closures](03-functions-and-closures.md): functions, recursion, functions as values, arrow functions, generics, closures and `escaping`.
4. [Structs and enums](04-structs-and-enums.md): structs, copies and spread updates, enums and `match`, recursive data and `Box`, distinct and opaque types.
5. [Pattern matching](05-pattern-matching.md): `match` with guards and literals, destructuring, `is` and `Option`, block expressions, `if` values and `? :`.
6. [Option and Result](06-option-and-result.md): `std/option`, `std/result`, errors without exceptions.
7. [Traits](07-traits.md): declarations, `implement`, constraints, `derives`, equality and ordering.
8. [Modules and packages](08-modules-and-packages.md): imports, aliases, exports, std imports.
9. [Memory and `mutable`](09-memory-and-mutable.md): reference counting, copies, `mutable` parameters and their rules.
10. [Effects and testing](10-effects-and-testing.md): `uses`, effect checking, `std/test` and `narrac test`.
11. [Narra for TypeScript programmers](11-narra-for-typescript-programmers.md): optional reading that compares Narra with TypeScript, topic by topic.

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

- `async` and `await`, threads, channels and shared state.
- Regions, fixed-size arrays and SIMD vectors.
- `unsafe`, `extern` and C interop.
- Build conditions.
- `@inline`, `@const` functions and `lazy` values.
- C-style `for` loops and compound assignment (`+=`, `++`).
- Sized number types (`i8` to `i64`, `u8` to `u64`, `f32`). Today the number types are `int` and `f64`.
- Dependencies on other packages.
