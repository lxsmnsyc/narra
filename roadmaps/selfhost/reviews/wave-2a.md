# Wave 2, first half

Nodes: modules (PR #4), traits (PR #3). Both were merged by the user. The user chose to run only these two first. patterns, closures, strings-std and io-std are still to come.

## Changes to the plan

- Contract `module-graph` is frozen at v2 from its change request: `SourceModule`, `Load`, `Program.Entries` and `Program.Files`, the injective C name rule (no `__` in names or path segments), and both entry forms.
- Contract `trait-model` is frozen at v2 from its change request: `__eq`-style names, emitter `use*` methods, prelude traits and `Ordering`, the implementation naming rule, the `FuncType` mangle, and `implement` only in the type's module.
- New node `poly-recursion` (must, S, type fix). A generic type with polymorphic recursion makes the compiler loop forever; the bug was already on main.
- maps brief: it must handle a user `Equatable` that has no `Hashable`, and must use `e.useEq` and `e.useHash`.

## Learnings promoted

`compiler/internal/narra/CLAUDE.md` now covers:
- C names are built only through the helpers.
- Types compare by module and name.
- `enter` the module before resolving its types.
- Where package tests go.
- The `use*` trait functions.
- The trait name rule.

## Notes for later nodes

- closures must build on the function-type parsing and unification that traits added (`funcTypeExpr`, `FuncType` in `sameType`, `unify` and `mangle`).
- The constructor lookup order is: expected type, then this module, then imported by name, then prelude. Ambiguity inside one of those groups is an error.
- Open nits:
  - A failed constraint reports twice.
  - `narra.toml` errors point at `0:0`.
  - An extra error line follows "cannot tell the type".
  - `cmp` and `eq` disagree on NaN.
- On this machine, the fast check took 70 to 150 s under load, against about 10 s in wave 1. The base commit before traits was just as slow, so traits did not cause it.

## Process

- Running two L nodes that edit the same files at once worked. traits kept its logic in its own files, so the rebase only conflicted in `check.go` and the README.
- Every node needed a second review round. In each case, round 1 found a real bug that the worker's own tests had missed.
