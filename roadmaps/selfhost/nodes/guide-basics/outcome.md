# Outcome: guide-basics

## Built

- `compiler/internal/narra/guide_test.go`: `TestGuide`, unchanged by the rewrite. Reads `../../../docs/guide/*.md`, extracts fenced blocks (`narra`, `narra error`, `narra fragment`, `text`) by `contracts/guide-samples.md`. One subtest per block, named `<file>:<line>` (line of the opening fence). A `narra` block runs through `runProgram` in every entry of `modes` (ASan/UBSan debug with leak check, and release), needs exit 0 and stdout equal to the `text` block that directly follows it (only blank lines between), and fails with "needs a `text` block" otherwise. A `narra error` block must start with `// error: <text>`; the test compiles it with `CompileToC` and requires the message to contain the text after stripping `sample.nra:` (so positions are `line:col` counted inside the sample, comment line included). `narra fragment` is skipped. Other fence types (`sh`, `toml`, `text` alone) are ignored.
- `docs/guide/README.md`: all eleven chapters with links (6 to 11 are dead until guide-advanced and guide-typescript land), how to read a program and an error sample, and the "Planned" list.
- `docs/guide/01-getting-started.md` to `05-pattern-matching.md`, rewritten 2026-11-05 as a guided tutorial after the maintainer's review of #87. Each chapter follows one thread, and each section starts from a task, builds a program, shows its output, then explains. Restrictions appear as the `narra error` the reader would hit, with the fix. 71 samples in total (programs, error samples, fragments), all pass.
  - Ch1: build the compiler, say hello (each part of `run` explained), print a number (the `expected string, found int` error leads to the first import), build an executable (debug and release), split into a package `shop` with `narra.toml`, the command table.
  - Ch2: a shop receipt. `const` (reassign error), `let` and `while` (saving toward a goal; `+=` and shadowing errors), `int` and `f64` (splitting a bill; mixing error; `+%`), text (padded receipt line, `+` error, `Str.split`, bytes and chars), `if` and `bool` (discount; `if (count)` error), arrays (prices, `push` with `mutable`, method-call error, ranges and entries, copies, `const` array error), maps (stock counts; `Option<int>` error leading to `is Some`), tuples.
  - Ch3: receipt code into functions (argument-count and nested-function errors), recursion taught with counting the ways to make change (two recursive calls, stopping cases, stack depth), then a short practical note that a call in last position runs like a loop (`addUpTo(10_000_000, 0)`), functions as values (`applyToAll` with a named function and an arrow function; untyped `const` arrow error), generics (`mapAll<T, U>`), closures (capturing `const`; `let` capture error and fix), returning closures and `escaping` (error, fix, cost).
  - Ch4: `Item` struct (missing-field error, name left out from context), copies, spread and `==`, `Payment` enum with `match` (`_` for unused fields; non-exhaustive error after adding `GiftCard`), gift baskets as a recursive enum walked by recursion, `Order` that contains itself (error) fixed with `Box`, `Option` and the bare `Some`/`None`/`Ok`/`Err`, distinct ids (mismatch error), opaque `Span` (fragment).
  - Ch5: a till that reads typed commands. Guards and statement `match`, literal and tuple patterns (`int` without `_` error), struct patterns, destructuring (refutable pattern error), parsing lines into `Option<Command>` with `is` and `Int.parse`, narrowing with an early return (`.value` error), `is` with `&&` (error and two fixes), block expressions, value `if` and `? :`, and the whole till as a closing program.

## Deviations

- The brief's chapter 3 list says "tail calls". The chapter teaches recursion first with a natural example (making change), then mentions tail calls in two short paragraphs and one sample, without the word "tail call". The maintainer called the old "Tail calls are loops" section and "Recursion needs no keyword" pointless; both are gone.
- Chapter 3 no longer shows a function stored in a struct field (the old sample used a struct before chapter 4). Chapter 4 does not show it either; a later chapter may.
- The contract says "a first line `// error: 3:9: ...`" without the file name; the test strips the sample file name, as described above.

## Stubs left

- None.

## For chapter 11

Comparisons left out of chapters 1 to 5, for guide-typescript. Each has the chapter that teaches the Narra form.

- Ch1: Narra reads like TypeScript but has value semantics, no `null`, no exceptions and no garbage collector.
- Ch1: no semicolons, and a `;` is an error with a fix.
- Ch1: `print` takes only a `string`; there is no `console.log` of any value, and no implicit `toString`.
- Ch1: named functions always write parameter and result types.
- Ch2: `const` is deep: a TypeScript `const` array can still be pushed to; a Narra one cannot change at all.
- Ch2: no `+=`, `-=`, `++` or `--` (also in Planned).
- Ch2: one `number` in TypeScript versus `int` and `f64` here, with no implicit conversion and integer `/` that truncates.
- Ch2: both quote styles make strings in TypeScript; here `'a'` is a `char`.
- Ch2: no `+` for strings; templates only, and every `${...}` part must already be a string.
- Ch2: no truthiness (`if (count)`, `if (name)`), and `===` is an error; `==` compares by structure.
- Ch2: `T[]` is not a type, write `Array<T>`; no methods (`items.push`), use `Arrays.push(mutable items, x)`.
- Ch2: assigning an array or struct copies it (TypeScript shares the reference).
- Ch2: `map[key]` gives an `Option`, not `V | undefined`.
- Ch2: no `for ... in`; C-style `for` is not there yet.
- Ch3: no optional parameters, default parameters, varargs, overloads or currying; no nested function declarations.
- Ch3: no explicit type arguments at a call (`f<int>(x)`).
- Ch3: a closure cannot capture a `let` (a TypeScript closure sees later assignments).
- Ch3: function parameters are non-escaping unless marked `escaping`; TypeScript has no such marker.
- Ch4: no `class`, `this`, `new` or `interface`; behavior is functions over data and traits.
- Ch4: enums with fields replace TypeScript `enum`, tagged unions with a `kind` field, and class hierarchies.
- Ch4: distinct types replace branded types (`number & { __brand: "UserId" }`).
- Ch4: opaque structs replace private class fields.
- Ch5: `match` replaces `switch`; no fallthrough, and exhaustiveness is checked.
- Ch5: `{ ... }` in value position is a block, not an object literal; `{ name }` shorthand builds no object.
- Ch5: `is` with narrowing replaces `x !== null` checks and type guards.
- Ch5: the ternary has the same form, with a `bool` condition only.

## Gotchas

The first item is for the roadmap only (it names SPEC and D-numbers); do not repeat it in the chapters.

- The compiler lacks things SPEC and DECISIONS describe: sized integers and `f32` (`i32`, `u8` are "unknown type"; the old chapter 2 had a table of them), compound assignment and C-style `for` (D76), `Arrays.map` and other array helpers (std/array has only `length`, `values`, `entries`, `push`, `filled`), `Float.PI`, and explicit type arguments in calls. The chapters avoid them, and sized numbers are now in the Planned list.
- An unused local is not an error in the compiler (the old chapter 2 said it was). The rewrite does not claim either way.
- An exported function other than the entry `run` needs `uses alloc` when it builds a string (effects, chapter 10). That is why the chapter 1 package exports `withTax`, a pure function, and single-file samples keep helpers private.
- Inside a package folder, `narra build .` writes the executable named after the `[[binary]]`; `narra build shop` from the parent folder fails because the output name equals the folder name.
- `print` and `${...}` take only strings, so every number goes through `Int.toString` or `Float.toString`.
- An `is` pattern that binds names must be the whole `if` condition; an `is` without bindings combines with `&&` fine (`removed && stock["tea"] is None`).
- Error samples count the `// error:` comment as line 1. Columns point at the operator or argument the compiler names (`1 + 2` reports the `+`), so take positions from the compiler, not by eye.
- `TestGuide` also fails on an info string that starts with `narra` (any case) but is not `narra`, `narra error` or `narra fragment`, on a fence that is never closed, and when no guide file matches. `TestGuideExtract` covers the first two.
- A `sh` or other block between a `narra` block and its `text` block makes the test fail with "needs a `text` block"; put commands in prose or after the output.
- Chapters 6 to 10 samples should use `Options`/`Results` aliases (plural) as other std aliases are.
- `.nra` samples in the guide are not formatted by `narrac fmt` and `TestSourcesFormatted` does not cover `docs/`.

## Proposed learnings

- Guide samples are checked by `TestGuide`; run `go test -count=1 ./internal/narra -run TestGuide` after editing any `docs/guide/*.md` | the test is the only link between the prose and the compiler | docs/guide | yes
- Guide chapters teach from a task: build a program, show its output, then explain; show a rule as the `narra error` the reader hits, with the fix | maintainer review of #87 | docs/guide | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestGuide'`: pass after the rewrite (71 sample subtests, each program in debug with ASan, UBSan and the leak check, and in release).
- During writing, each chapter was also checked with a local script that runs every `narra` block with `narra run` and every `narra error` block with `narra check`; it is not committed.
- Every claim about compiler behavior not covered by a sample was checked with a scratch program: `args` holds one element (the program path), `narra check .`, `narra test .`, `narra run .` and `narra build .` inside a package, `narra explain E0003`, tuple exhaustiveness (`#[false, false]` missing is an error), `let` narrowing after an early return, and the block-body arrow fragment in chapter 3.
- No Go code changed in the rewrite, so the full check was not rerun. Earlier runs: `go vet ./... && go test -timeout 30m ./...` passed on the review-fix commit (about 24 min, loaded machine).
- Earlier hand check: a broken sample made TestGuide fail with the file and line (`01-getting-started.md:16/debug`).

## Size

- Estimated M, actual M. The rewrite replaced the prose of all six files.
