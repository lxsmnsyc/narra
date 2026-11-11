# Log: copy-check-macros

Base: feat/batch-7 (7f0e0534). Compiler built with `sh selfhost/bootstrap.sh <dir>`.

## Before (feat/batch-7)

- `narrac emit selfhost > base.c`: 186150 lines, 8618951 bytes.
- `unifdef -DNARRA_RELEASE base.c`: 183626 lines, 8369889 bytes.
- `cc -E -DNARRA_RELEASE base.c`: 185818 lines, 8383405 bytes. Debug `cc -E`: 187272 lines, 8673535 bytes.
- Counts in base.c: `#ifndef NARRA_RELEASE` 487, `NARRA_COPY_GRANT(` 1035, `credit_` 599.

## Work

- The runtime source of truth is `selfhost/runtime/runtime.h`; `selfhost/src/runtimecore.nra` is generated from it (suite `generated`, `--update`). Edited the header and regenerated, so `selfhost/runtime/runtime.h` is a deviation from the brief's shared list.
- New macros: `NARRA_COPY_ONLY(...)` (credit and uncredit definitions and prototypes), `NARRA_COPY_CREDIT(...)`, `NARRA_COPY_GRANTED(...)` (wraps the retain), `NARRA_COPY_ENDED(v, uncredit)` (the enum release block).
- First try made CREDIT a do/while. It sits inside GRANTED's arguments without a `;`, so the C failed. It is now a plain `if (narra_copy_on) { ... }`.
- Bootstrap with the committed seed: stage 1 C differs from stage 2 C, stage 3 C equals stage 2 C (fixed point reached through stage 3, as the script allows).
- The shared scratchpad is clobbered by other agents; I used unique directory names.
- `takeOwnCopy` keeps its two `NARRA_COPY_GRANT` lines: `consume` writes a declaration and the retain over several lines, which cannot sit inside a macro argument. They already expand to `((void)0)` in release.

## After

- `narrac emit selfhost`: 183657 lines, 8574436 bytes.
- `unifdef -DNARRA_RELEASE`: 183534 lines, 8569378 bytes (grows: unifdef no longer sees the debug code, it is inside macros).
- `cc -E -DNARRA_RELEASE`: 183807 lines, 8366661 bytes. Debug `cc -E`: 184061 lines, 8678534 bytes.
- Counts: `#ifndef NARRA_RELEASE` 2, `NARRA_COPY_GRANT(` 5 (string literals and `takeOwnCopy`), `NARRA_COPY_GRANTED` 521, `NARRA_COPY_CREDIT` 333, `NARRA_COPY_ONLY` 119, `NARRA_COPY_ENDED` 42.
