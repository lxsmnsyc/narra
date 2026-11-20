# Outcome: guide-generators

## Built

`docs/guide/12-generators.md` teaches generators in nine sections. Each has a runnable sample, and each rule has a `narra error` sample with the compiler's message:

- **Write a generator.** `function*`, `yield`, `Generator<T>`. Prints in the body show that the call runs nothing and the body stops at each `yield`.
- **Stop early.** An endless generator left with `break`. `return` ends a generator, and `return value` is an error.
- **What a loop can read.** A table of `for ... of` sources: arrays, `Arrays.entries`, `Maps.keys`, `Maps.entries`, `Str.chars`, ranges and generators. A generic `each<T>` generator passes an array where a `Generator<T>` is wanted.
- **`yield*`.** A recursive tree walk.
- **A generator is a value.** A loop reads its own copy, so two loops over one `const` run the body twice. No `==`, map keys or `Show`.
- **Rules for the body.** No `mutable` parameter, function-type parameters must be `escaping` (error and fix), and no `yield` in an arrow function.
- **Effects of a generator** (audit row 144). Effects on the generator type, `uses (io, alloc)` on an exported generator, making a generator costs only `alloc`, an effect variable over `Generator<int> uses E`, and the error for passing a generator with effects where `Generator<int>` is expected.
- **Resume a generator by hand** (row 147). `Co.resume`, `Step.Yielded`, `Step.Done`, and when `Step.Panicked` appears.
- **Coroutines** (row 147). `Coroutine<In, Out>`, `const input = yield output`, `Co.send`, the dropped first input, no `yield*`, and the error for `for ... of` over a coroutine. A pointer to chapter 14 for `async function*`.

`docs/guide/13-standard-library.md`: the one-sentence generator explanation that opened the `std/iterator` section is now a link to chapter 12. The `Co.resume` sentence links to the chapter 12 section.

## Deviations

- **Array and map loops are not generators.** The brief asks how array and map loops are generators. The compiler does not make them generators. `Maps.keys`, `Maps.entries` and `Str.chars` return arrays, `Arrays.values` and `Arrays.entries` are allowed only at the head of a loop (E0920), a range is the type `Range`, and passing an array or a `Range` where a `Generator<int>` is expected is a type error. The chapter teaches what the compiler does, in "What a loop can read". `SPEC.md` section 2 ("An iterator is a `Generator<T>`") and the audit's chapter-order note say these return generators. The orchestrator should fix the SPEC line or record the difference in `compiler/README.md`. This node owns neither file.
- `Step.Panicked` is described but not shown. The guide suite runs samples with `narrac run`, and `Panicked` only appears under `narrac test`.

## Stubs left

- None.

## Gotchas

- The line of an `// error:` position counts the comment itself as line 1.
- `print` needs `io`, so an exported generator that prints needs `uses (io, alloc)` on its result type.

## Proposed learnings

- Check every library claim of `SPEC.md` against the compiler before teaching it | SPEC section 2 says `Maps.keys` returns a generator, the compiler returns an array | docs/guide | no

## Verification

- `NARRAC=<main-narrac> sh scripts/test.sh --suite guide`: PASS, 417 cases.
- `pnpm --dir website build`: passes (27 pages prerendered).
- Every sample checked with `narrac check` and `narrac run` in a scratch folder outside the repo.
- `check_prose.py docs/guide/12-generators.md`: ok. The warnings left in `13-standard-library.md` are on lines this node did not change.

## Size

- Estimated M, actual M.
