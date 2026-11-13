# Outcome: std-iterator

## Built

- `std/src/iterator.nra`, module `std/iterator` (alias `Iterators`), pure Narra source.
  - Lazy, each a `function*` with `escaping` callbacks and one effect variable `E` over the source and callback: `map`, `filter`, `flatMap`, `take`, `drop`, `takeWhile`, `dropWhile`, `enumerate` (`#[int, T]`), `zip` (`#[A, B]`, ends with the shorter; resumes its own `copy` of the second source with `Co.resume`), `concat`. `flatMap`, `zip` add `alloc` to their result.
  - `chunk(source, size)` and `window(source, size)`: `Option<Generator<Array<T>> uses alloc>`, `None` when `size < 1`. The last chunk may be short; a source shorter than the window yields none.
  - `take` and `drop` treat a negative count as 0.
  - Eager, non-escaping callbacks with `E`: `reduce(source, combine, initial)`, `toArray`, `forEach`, `some`, `every`, `find` (`Option<T>`), `count`.
- `selfhost/src/stditerator.nra`: one `ITERATOR_FUNCS` row (`check: "source"`), registered last in `stdlib.nra`. `stdtext.nra` regenerated. `std/iterator` is now last in the unknown-module list (`testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump`).
- Samples: `testdata/iterator.nra` (every operator, an early `break` out of a `map` over an endless generator), `iterator_consume.nra` (`take(lines, 2)` leaves `lines` at its start; `Co.resume(mutable lines)` goes on), `iterator_effects.nra` (`map` over `uses io` is `uses io`; pure stays pure). Errors: `testdata/errors/iterator_capture_let.nra` (E0612), `iterator_chunk_unwrapped.nra` (`for ... of` over the `Option`).
- SPEC section 3: a built `std/iterator` bullet before the D118 "Not built yet" list (its iterator line removed), and `std/iterator` in "Standard modules". `compiler/README.md`: one line. Guide: new chapter `docs/guide/12-standard-library.md` with a `std/iterator` section, listed in `docs/guide/README.md`.

## Proof

- base: 4403befe, `narrac run testdata/iterator.nra` with a bootstrap-built `-release` compiler: `unknown module "std/iterator"`.
- branch: same command prints `testdata/iterator.out`; clean under the leak check; `--filter iterator` PASS, 28 cases.

## Deviations

- `chunk` and `window` take `source: Generator<T> uses alloc`, not an effect variable as D118's other operators do. The checker binds an effect variable in a result only when the result is the generator itself (`bindCallResult`, `exprs.nra`; `bindResultEffects`, `effects.nra`). Inside `Option<...>` the name `E` leaks into the caller's type, and in one case (private caller, unannotated binding) an `io` source ran in a function declared `uses alloc` without an error. So a source with `io` cannot be chunked yet. SPEC and README say so. The hole exists on base without std/iterator (repro in `log.md`, `p4/main.nra`): a bug for its own fix node.
- `flatMap`'s callback type is `(value: T) => Generator<U> uses E uses E, alloc`: the first `uses` list is the generator's, the second the function's (making a generator allocates, and an exported function's callback must declare it).
- The guide chapter number 12 was free; the chapter is new and holds only the iterator section.
- No codes from E1300-E1309 used.

## Stubs left

- None.

## Gotchas

- No trailing comma after a parameter whose type ends in a `uses` list: `uses E,` then `)` fails to parse (the comma continues the list).
- `test` is a keyword; it cannot name a parameter.
- An operator takes its source by value, so `Iterators.take(lines, 1)` after two `Co.resume(mutable lines)` starts at the third item, and `lines` stays there.
- A generator from an operator with `alloc` in its type does not fit a parameter typed `Generator<T>`; helpers that print generators take `Generator<T> uses alloc`.

## Proposed learnings

- An effect variable in a result type must be the generator's own list, not nested in `Option` or another type argument | it is not bound at the call and can leak or hide an effect | selfhost/src (checker) | yes (a compile error for a nested effect variable until binding works)

## Verification

- `sh scripts/test.sh` (fast check, seed-built): see the final line in the reply to the orchestrator.
- `narrac emit selfhost` from the base and the branch compilers: byte-identical (223,897 lines).
- `--suite guide` PASS, 301 cases; `--filter iterator` PASS, 28 cases.

## Size

- Estimated M, actual M (about 270 lines of Narra in the module, three samples, two error tests, one guide chapter).
