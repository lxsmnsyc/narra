# Outcome: guide-numbers

## Built

A section "Numbers of other sizes" in `docs/guide/02-values-and-types.md`, before "Group a few values". Each subsection has a runnable example:

- "Sized integers": the seven types, literal typing and E0294, wrapping at the width, `toString`, `parse`, `addChecked` of the sized modules (row 81).
- "No mixing, and the conversions": E0295, `from`, `tryFrom`, `wrap` (81).
- "The f32 float": `f32` literals, `F32.from`, `Float.from`, `F32.truncate` (82).
- "How a float prints": shortest text, `.0`, exponents, `nan`, `inf`, `-0.0` (118).
- "Bits": `&`, `|`, `^`, `~`, `<<`, `>>`, and the shift-count proof with E0980 (44, 116).
- "The math module", "Random numbers", "Hashing": `std/math`, `std/random`, `std/hash` (66, 68, 69).

## Deviations

- The website grammar check fails on the new samples: the TextMate grammar reads module aliases that have digits and capitals (`I32`, `U8`, `F32`, `U64`) as `class_name`, the site as `constant` (24 mismatches, all on these four names). `pnpm --dir website build` therefore stops at `check:grammar`. `vite build` alone passes. Owned by node grammar-fixes; `check-grammar.ts` is untouched.
- Left out: `u16`, `u32`, `i8`, `i16` get no sample of their own (the pattern is the same as `u8` and `i32`); `Show.show` on numbers (chapter 7); the `std/simd` vectors (a later chapter); `Math.trunc`, `tan`, `atan2`, `exp`, `log` are named but have no sample.

## Stubs left

- None.

## Gotchas

- Compiler bug: a program that calls `U8.parse` (any sized `parse`) without importing `std/int` fails in the C compiler with an undeclared `narra_sized_parse`. The sample imports `std/int` and uses `Int.from`, which hides it. Needs a compiler fix.
- `Hash.of` is not printed in the guide because its value may change between builds.
- In a `narra error` sample the reported line counts the `// error:` comment line.

## Proposed learnings

- None.

## Verification

- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 442 cases.
- `pnpm --dir website build`: fails at `check:grammar` (see Deviations); `pnpm --dir website exec vite build` passes.
- `check_prose.py` on chapter 2: two long-sentence warnings left, both in text this node did not write.

## Size

- Estimated M, actual M.
