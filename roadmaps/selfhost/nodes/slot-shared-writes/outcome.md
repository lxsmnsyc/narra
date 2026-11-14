# Outcome: slot-shared-writes

## Built

- `selfhost/src/slotstores.nra`: a table of the std functions that keep an argument in a slot (`findSlotStore`, `isSlotHandle`): Cells.create and set, Shared.create and set, Channel.send, Arenas.insert and set.
- `selfhost/src/sharedwrites.nra`: `keepStoredValue` pairs the stored value with the holder (`holder[]`), like a field store. Dispatch rows only: `keepsShares` includes a slot store, `evaluateCall` calls `keepStoredValue`, and handles of the slot modules (Cell, Shared, Atomic, Sender, Receiver) stay live like a Cell. A `create` result shares the value.
- `selfhost/tests/src/slotstores.nra`: the `slot-stores` suite lists every exported function of std/cell, shared, channel, atomic and arena that takes a value of a type parameter and wants it in `SLOT_ROWS` or `NOT_STORED` (a missing entry fails the suite; checked by removing one).
- Tests: `testdata/errors/slot_shared_*` (10 error tests: stores, the three `update` callbacks, a `ends[0]` Sender), `testdata/slot_shared_copy.nra` (write after `copy(...)` works; runs under the debug copy check).

## Proof

- base: 6c616a46, `narrac check` on probes: Cells.set, Cells.create, Shared.set, Channel.send, Arenas.insert, Arenas.set each followed by `Arrays.push(mutable items, 4)` exit 0 with no error (`narrac run` of the Cells.set probe exits 3: the cell kept the old array through a hidden copy).
- branch: the same probes report the shared-write error (E0978); the 6 error tests pass; `slot_shared_copy` prints `3 4 / 4 5 / 5 6 / 2`.

## Deviations

- No `from`-style marker on std parameters: the seed rule forbids new syntax in std, so the rule is a table in the compiler. A later node can replace the table with a parameter marker.
- The E0978 message offers no fix for the stored value (origin text is the argument, but the fix builder declines); the `Cells.create` case does offer `copy(...)`.

- Atomic has no rows: `Atomic<T>` allows only bool and integers (an `Atomic<Array<int>>` is a type error), so no heap value can be stored.
- `update` of Cell, Shared and Arena: a callback that returns a value shared with a live name is already E0976 (the arrow returns a captured value), so it cannot escape; probes are the three `slot_shared_*_update` error tests.
- A Sender reached through a tuple (`ends[0]`) is caught: the pair holds the place `ends[0][]` until the block ends (`slot_shared_channel_tuple`).

## Stubs left

- None.

## Gotchas

- A new std function that stores its argument needs a row in `SLOT_ROWS` (`slotstores.nra`).

## Proposed learnings

- A std function that keeps an argument in a handle needs a row in `selfhost/src/slotstores.nra` | the shared-write check cannot see inside std bodies | selfhost/src | no

## Verification

- `narrac emit selfhost` from the base and the branch compiler: byte-identical C (`cmp` exit 0). Code that does not use the feature is unchanged.
- Debug build of the branch compiler, `NARRA_CHECK_COPIES=1 emit selfhost`: exit 0.
- `sh selfhost/bootstrap.sh`: fixed point reached (stage 1 and 2 identical).
- `sh scripts/test.sh`: see the final line in the report.

## Size

- Estimated M, actual S.
