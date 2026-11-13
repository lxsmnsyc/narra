# Log: std-arena

## Proof (base)

- Base: feat/batch-14 at cb834b6e, compiler built with `sh selfhost/bootstrap.sh <dir>` (release).
- Command: `narrac run testdata/arena.nra` (the sample as first written).
- Result on base: `arena.nra:3:1: module "std/arena" has no export "Arena"` (and the same for `Handle`); no `std/arena` exists.

## Notes

- Generic opaque structs and `derives Equatable, Hashable, Show` on a phantom type parameter work; no checker change was needed.
- `std/array` has no `pop`, so the free list is a stack kept as an array plus a `freeCount` field.
- `remove` returns `copy(old)`: the checker treats the value taken from a slot as shared with the `mutable arena` parameter (E0976 otherwise). The slot is already `None`, so the copy is a reference count increment.
- A string template part must be a string: the samples print through `Show.show(...)`.
- Quick loop used while the std text was being edited: copy `std/src/arena.nra` into a scratch package as `q/arena` and run the sample with `narrac run .` (no rebuild of the compiler).
- Callbacks stored in a node must `copy(...)` a captured handle (`edges: [copy(b)]`), and a `let` that holds a value read from the arena blocks `remove` until its block ends (D104), so the sample walks in a helper function.
