# Log: async-generators

## Proof on the base

- Base: `feat/batch-13` at 7df2cd63. Compiler: `narrac0 build -release selfhost` from the committed seed.
- Sample: `async function* f(): AsyncGenerator<int> { yield 1 }` with `const _ = f()` in `main`.
- Command: `base-narrac run main.nra`.
- Result: exit 1, `main.nra:1:15: there is no async function*; an async function returns one value through Async<T>` (E1107).

## Work

- Built the branch compiler with the base compiler each time (`base-narrac build -release`), so the seed rule held: nothing in `selfhost/src` uses the feature.
- First full run (`NARRAC=branch sh scripts/test.sh`): one failure, `errors/async_generator.nra`, the E1107 sample of async-core. Removed it, since E1107 is retired; `async_generators_result` covers the `Generator<T>` case with E1112.
- Emit check: `base-narrac emit selfhost` and `narrac emit selfhost` (branch) on the branch source are byte-identical.
- Private stream effects: a reader of a private stream whose body prints must declare `io` (checked by hand: E0800 through `next`).
- Generic streams (`repeat<T>`) work: the Option<T> instance comes from each instance body.
