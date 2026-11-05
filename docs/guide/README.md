# The Narra guide

A guide to Narra for a programmer who has written code before. Each chapter builds on the ones before it. You do not need to know any other particular language.

Every program in the guide is compiled and run by a test (`TestGuide`), so what you read is what the compiler does today.

## Chapters

1. [Getting started](01-getting-started.md): a package, `narra.toml`, `narra run`, `narra build`.
2. [Values and types](02-values-and-types.md): numbers, strings and templates, `bool`, `const` and `let`, arrays, maps, tuples.
3. [Functions and closures](03-functions-and-closures.md): declarations, arrow functions, function types, non-escaping parameters, tail calls.
4. [Structs and enums](04-structs-and-enums.md): declarations, literals, spread updates, `Box` and recursive enums, opaque and distinct types.
5. [Pattern matching](05-pattern-matching.md): `match`, `is`, destructuring, narrowing, block expressions and the ternary.
6. [Option and Result](06-option-and-result.md): `std/option`, `std/result`, errors without exceptions.
7. [Traits](07-traits.md): declarations, `implement`, constraints, `derives`, equality and ordering.
8. [Modules and packages](08-modules-and-packages.md): imports, aliases, exports, std imports.
9. [Memory and `mutable`](09-memory-and-mutable.md): reference counting, copies, `mutable` parameters and their rules.
10. [Effects and testing](10-effects-and-testing.md): `uses`, effect checking, `std/test` and `narra test`.
11. [Narra for TypeScript programmers](11-narra-for-typescript-programmers.md): optional reading that compares Narra with TypeScript, topic by topic.

## Reading the samples

A program in the guide is a whole file that exports `run`, followed by its output:

```narra
export function run(args: Array<string>): int {
  print("hello")
  return 0
}
```

```text
hello
```

A sample marked as an error shows the message the compiler gives. Save a sample to a file and run it with `narra run file.nra` to try changes.

## Planned

The language is meant to have these features, and the compiler does not have them yet. The guide does not teach them.

- Generators and iterators written by users: `function*`, `yield`. They parse and type-check, but a program that declares one does not build.
- `async` and `await`, threads, channels and shared state.
- Regions, fixed-size arrays and SIMD vectors.
- `unsafe`, `extern` and C interop.
- Build conditions.
- `@inline`, `@const` functions and `lazy` values.
- C-style `for` loops and compound assignment (`+=`, `++`).
- Dependencies on other packages.
