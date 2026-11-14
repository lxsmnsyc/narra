# Log: std-iterator

Machine: Apple M2, macOS 27. Base: `feat/batch-14` at 4403befe. Compilers built with `sh selfhost/bootstrap.sh <dir>` (`-O2 -DNARRA_RELEASE`): base in `/private/tmp/claude-501/std-iterator/base`, branch in `/private/tmp/claude-501/std-iterator/br`.

## Check: what exists

No `std/src/iterator.nra`, no `std/iterator` row, no `testdata/iterator*`. Nothing to reuse.

## Proof first

- base (4403befe): `base/narrac run testdata/iterator.nra` fails: `iterator.nra:3:1: unknown module "std/iterator" (this compiler supports ...)`, then one `unknown module "Iterators"` per call.
- branch: `br/narrac run testdata/iterator.nra` prints every operator (`testdata/iterator.out`); `NARRA_CHECK_LEAKS=1` exit 0.

## Findings while building

- `uses E,` before `)` in a parameter list does not parse: the comma continues the effect list. No trailing comma after a parameter whose type ends in `uses ...`.
- `test` is a keyword; the eager callbacks are named `predicate`.
- An exported generator function's callback that makes a generator allocates, and in a function type `Generator<U> uses E` takes the `uses` list for the generator. A second list is the function's: `(value: T) => Generator<U> uses E uses E, alloc` (flatMap).
- An effect variable inside an `Option` of a result is not bound at the call. `bindCallResult` (`exprs.nra`) and `bindResultEffects` (`effects.nra`) only look at a top-level `TGenerator`; generic enum instances are `TEnum(index)`, so binding a nested one means a new enum instance per call. `chunk(source: Generator<T> uses E)` returning `Option<Generator<Array<T>> uses E, alloc>` left `E` in the caller's type (`expected Option<Generator<Array<int>> uses alloc>, found Option<Generator<Array<int>> uses E, alloc>`). With an unannotated binding, a private caller declared `uses alloc` ran an io source without an error (`/private/tmp/claude-501/std-iterator/p2/main.nra`). The same hole without std/iterator, on the base compiler: `/private/tmp/claude-501/std-iterator/p4/main.nra` gives `Function pure uses E (through Generator<int> uses E) but its signature does not declare it`, an effect name that should never leave the callee. Not fixed here (compiler change, shared hot files); chunk and window take `Generator<T> uses alloc`.

## Emit comparison

`base/narrac emit selfhost` and `br/narrac emit selfhost` on the branch tree: identical (`cmp`), 223,897 lines.

## Checks

- `NARRAC=br/narrac sh scripts/test.sh --filter iterator`: PASS, 28 cases.
- `--suite guide`: PASS, 301 cases.
- `--suite generated --update` rewrote `stdtext.nra`; `--filter unknown_module --update` rewrote the two module-list goldens.
