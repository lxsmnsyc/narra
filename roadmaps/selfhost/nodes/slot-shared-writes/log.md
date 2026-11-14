# Log: slot-shared-writes

## Proof on the base (feat/batch-15, commit 6c616a46)

Base compiler: `sh selfhost/bootstrap.sh <dir>` at 6c616a46. Command `narrac check <file>` on each probe:

- Cells.set(cell, items) then Arrays.push(mutable items, 4): exit 0, no error. `narrac run` exits 3 (the cell kept the old array: a hidden copy).
- Channel.send(sender, items) then a push: exit 0, no error.
- Shared.set, Cells.create(items), Arenas.insert(mutable arena, items), Arenas.set(mutable arena, handle, items): all exit 0, no error.
- Control: a struct literal field store (`Holder { items: items }`) then a push reports the shared-write error.

## Stores found

Std functions that keep an argument in a slot, none tracked before:
Cells.create, Cells.set, Shared.create, Shared.set, Atomic.create, Atomic.store, Channel.send, Arenas.insert, Arenas.set.
Already handled: Arrays.push and other builtin `mutable` calls (keepArguments), struct and constructor fields.
Not stored by value: Cells.update, Shared.update, Arenas.update (the closure result is stored; its argument is the old value, no new sharing from the call site).

## Approach

The seed rule forbids new syntax in std, so a `from`-style marker on parameters is not possible yet. The rule is a table in `selfhost/src/slotstores.nra` (module, function, holder argument, value argument); `sharedwrites.nra` has three dispatch lines: `keepsShares`, the call to `keepStoredValue`, and `findLiveReason` (handles of the slot modules stay live like Cell).
