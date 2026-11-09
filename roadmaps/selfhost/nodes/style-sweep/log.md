# style-sweep log

## Before (origin/feat/batch-2 at 058f368)

- Go-form grep (`^\s*// [a-z][A-Za-z]* (is|returns|reports|checks) `): 574 lines.
- D-number grep (`D[0-9]{2}`): 88 lines. Some are hex constants (`0xD800`, `0x94D049BB...`) and text generated from `selfhost/runtime/`.
- Function names ending in `Of`: 15 (`queryTypeOf`, `checkValueCallOf`, `keyOf` twice, `keysOf`, `emitIndexOf`, `enumOf`, `prefixOf`, `collectUsesOf`, `typeOf`, `refOf`, `localOf`, `stepOf`, `lastIndexOf`, `lineStartOf`), plus four with `Of` inside (`keyOfFunction`, `rowOfType`, `charOfByte`, `endOfName`).
- Comment blocks directly above a function: 1323 blocks, 2546 lines (script count).

## Method

- Baseline compiler built from the base tree. After each group: the compiler's own C (stage 2) is compared with the baseline's after applying the renames and dropping source positions and string lengths; the C of all 111 top-level samples in `testdata/` is compared byte for byte.
- std/src positions appear in the emitted C of samples (bounds-check messages, for example `"std/src/json.nra:L:C"`). Removing lines in `std/src` would change sample C, so std/src edits keep every line in place.
- `narrac fmt` with no paths formats the whole repo; pass paths. std/src is not formatted (the sources suite does not cover it, and formatting moves lines).
