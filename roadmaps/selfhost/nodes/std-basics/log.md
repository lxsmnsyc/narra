# Log: std-basics

- Checked first: `Hash.of` did not exist (SPEC mentioned it; compiler/README said "does not exist yet"; only `Hashable.hash` worked). `Ranges.stepped` was in SPEC only. No math, time or random. So all of it was built.
- `-lm` is already on the cc line in `driver.nra`; no linking change needed.
- `of` and `from` are keywords. `Hash.of` needed the parser to accept `of` as a function name and after `.` (precedent: `parseMemberName` already took `from`).
- Primitives: `std/internal/math` (rows from `internalMathFuncs`, emitted by `emitFixedCall`), `std/internal/clock` and `std/internal/entropy` (C files in `selfhost/runtime`). All in `stdinternal.nra`.
- Range value got a third field `step`; `upTo` and `between` set 1. A loop over a Range value uses `narra_range_advance` so a step near INT64_MAX ends the loop instead of overflowing.
- Samples print only run-independent values; the system-seeded and time samples print booleans.
