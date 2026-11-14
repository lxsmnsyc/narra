# Outcome: index-assign-order

## Built

- `selfhost/src/emitassignorder.nra`: `runsCode`, `holdTargetParts`, `holdIndexes`. An assignment, a compound assignment and a map write evaluate the index and key expressions of the target (left to right, through fields and nested indexes) before the right side, then store. This applies only when the right side and the index expression both run code (a call, an await, a match or a block); other assignments emit as before.
- Dispatch rows: `emitAssignment` and `emitCompoundAssignment` in `emitstmts.nra` call `holdTargetParts`; `emitMapSet` in `emitmaps.nra` holds the map path and the key, then writes the right side.
- The held values are frame temporaries in async bodies, so they survive an await. The store still takes the array pointer after the right side runs (`_at` makes the array unique there), so copy-on-write and the hoisted-array path are unchanged.
- `testdata/index_assign_order.nra` and `.out`: plain, compound, nested, map, field-then-index, and async targets with `await` on the right.

- Held targets: an identifier index or map key is held too when the right side runs code and the variable is a pointer variable or a `let`/`mutable` local that an arrow of the body captures. Plain locals that nothing can write emit as before.
- Compound assignment (added at the coordinator request) to a field or array element whose right side runs code reads the old value before the right side (`readsBeforeValue` in `emitassignorder.nra`, used by `emitCompoundAssignment`); a string old value is retained until the statement ends. The store recomputes the pointer after the right side from the held index values. Map targets have no compound assignment in the language (the checker rejects them).
- The sample's `readsFirst` function writes the same element through a captured `let` (array, field, string element); on the base it printed `101 20`, `51 72` and `long!`, on the branch `11 20`, `11 4` and `a!`.

## Proof

- base: 6c616a46, `narrac run testdata/index_assign_order.nra` (base compiler): `value p` printed before `index p`, and the same for r, h, k; with await, `end s` before `index a`.
- branch: same command prints every `index` and `key` line before its `value` or `start` line; `sh scripts/test.sh --filter index_assign` passes (10 cases, sanitizer, leak and copy checks included).

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- `emit selfhost` is not byte-identical: 201 hunks after normalizing temp names: `map[key(...)] = value(...)` sites (key call first) and compound assignments to a field or element with a right side that runs code (old value read into a temp first). The rest of the output is identical after normalizing temp names.
- The compiler source does not rely on the new order; the seed was not regenerated.

## Proposed learnings

- Reorder evaluation only when both sides run code, so the emitted C of pure code stays identical | keeps the emit diff small | selfhost/src | no

## Verification

- `NARRAC=<release build> sh scripts/test.sh --filter index_assign`: PASS, 10 cases.
- `sh scripts/test.sh`: PASS (final line in the report).

## Size

- Estimated M, actual S. The held-argument mechanism of async argument ordering already did the work.
