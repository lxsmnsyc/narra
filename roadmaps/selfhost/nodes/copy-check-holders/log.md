# copy-check-holders log

## Proof on the base (feat/batch-10, ececc85a)

Compiler built from the seed with `sh selfhost/bootstrap.sh <dir>`, debug build (no `-release`).
Sample: `testdata/copy_enum_original_first.nra` (two `copy(tree)` of a recursive enum, the original `tree` released at the end of `makeCopies`, then one write to an array of each copy's payload).

    narrac build -o c <sample> && NARRA_CHECK_COPIES=1 ./c
    list 9 8
    copy check: copies on write that no copy(...) explains: array 1
    exit 103

The same program with the original `tree` kept alive in `run` (passed into `makeCopies`) exits 0. So the false report comes from the original holder releasing first: `NARRA_COPY_ENDED` assumed the original was one of the holders and ended a credit.

Only the heap enum release had the flaw. Arrays, boxes, maps, builders and generators forget their credits only when the count falls to 2 or less, which leaves no shared holder.

## Fix

- Each heap enum value carries a debug-only `copy_held` id (`NARRA_COPY_ONLY`, in the padding after `tag`, so the struct size is unchanged).
- `copy(x)` brands the holder it makes with a new id (`NARRA_COPY_BRAND`). A retain of a branded value counts a holder of the id (`NARRA_COPY_HELD`). A release counts it down, and the credit ends when the last holder of the id is released. Releasing the original (id 0) ends nothing. When the count falls back to one, all credits still end, as before.
- The emitter change is in `selfhost/src/emitdecls.nra` (struct field, retain hook, brand at the credit code).
