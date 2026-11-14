# Outcome: std-arena

## Built

- `std/src/arena.nra`, module `std/arena` (alias `Arenas`), pure Narra over `Array`.
  - `export opaque struct Handle<T> derives Equatable, Hashable, Show { index, generation }` and `export opaque struct Arena<T>` (slots as `Array<Option<T>>`, a generation per slot, a free list kept as a stack with a `freeCount`, a live count).
  - `create<T>()`, `insert(mutable arena, value): Handle<T>`, `get(arena, handle): Option<T> from arena`, `set(mutable arena, handle, value): bool`, `update(mutable arena, handle, change): bool` (one effect variable for the callback; false for a stale handle), `remove(mutable arena, handle): Option<T>` (drops the value at once, raises the slot's generation, frees the slot), `contains`, `size`, `clear(mutable arena)` (keeps the slots, raises every generation) and `entries(arena): Generator<#[Handle<T>, T]>` in slot order.
- `selfhost/src/stdarena.nra`: one `ARENA_FUNCS` module row (`check: "source"`), registered after `std/iterator` in `stdlib.nra`. `stdtext.nra` regenerated. `std/arena` is now last in the unknown-module list (`testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump`).
- Samples: `testdata/arena.nra` (every function, slot reuse, `clear`, `entries` order, handle as a map key), `testdata/arena_graph.nra` (a cycle a -> b -> c -> a, then a removal). Errors: `testdata/errors/arena_handle_fields.nra` (opaque field read), `arena_wrong_type.nra` (`Handle<int>` to an `Arena<string>`).
- SPEC: a built `std/arena` bullet in section 3 (the "Not built yet" line is gone), the module in "Standard modules", the "Graphs" bullet of "Cycles" without "not built yet" and with the other-arena caveat. `compiler/README.md`: one line. Guide: a "Graphs: std/arena" section in `docs/guide/12-standard-library.md`.

## Proof

- base: cb834b6e, `narrac run testdata/arena.nra` with a bootstrap-built release compiler: `module "std/arena" has no export "Arena"` (no such module).
- branch: same command prints `testdata/arena.out`; clean under the leak check; `--filter arena` PASS.

## Deviations

- `remove` returns `copy(old)`: the checker counts the value read from the slot as shared with the `mutable arena` parameter (E0976 otherwise). The slot already holds `None`, so the copy only adds a reference.
- No codes from E1310-E1319 used.
- Generated and shared files touched: `stdtext.nra`, the two unknown-module goldens only.

## Stubs left

- None.

## Gotchas

- `std/array` has no `pop`; a stack over an array needs a count field.
- A handle from another arena is not detected (D118 cost; stated in SPEC).
- A `let` holding a value read from the arena blocks `remove` until its block ends (D104); the graph sample walks in a helper. A closure that stores a captured handle must `copy` it.
- A template part must be a string: print handles, bools and ints through `Show.show`.
- Programs in `testdata/` export `run(args)`; an unused private `main` is an error.

## Proposed learnings

- None.

## Verification

- `narrac emit selfhost` from the base and branch compilers: byte-identical (223,942 lines).
- `sh selfhost/bootstrap.sh`: fixed point holds. `--filter arena`: PASS, 23 cases. Full check: see the reply to the orchestrator.

## Size

- Estimated S, actual S (about 130 lines of Narra, two samples, two error tests).
