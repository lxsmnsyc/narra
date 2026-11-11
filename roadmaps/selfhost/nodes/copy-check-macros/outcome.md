# Outcome: copy-check-macros

## Built

- Four macros in `selfhost/runtime/runtime.h` (and the generated `selfhost/src/runtimecore.nra`): `NARRA_COPY_ONLY(...)`, `NARRA_COPY_CREDIT(...)`, `NARRA_COPY_GRANTED(...)`, `NARRA_COPY_ENDED(v, uncredit)`. In release builds ONLY, CREDIT and ENDED expand to nothing and GRANTED to the plain retain.
- `stdcopy.nra` (`emitCopyCall`) writes one line per `copy(...)` site: `NARRA_COPY_GRANTED(<retain> NARRA_COPY_CREDIT(<credits>));` (or without CREDIT).
- `emitdecls.nra`: credit and uncredit prototypes and definitions are wrapped in `NARRA_COPY_ONLY(...)`; the enum release block is `NARRA_COPY_ENDED(v, uncredit_X);`.
- The grant macro wraps the retain (one statement), as the brief asked to be told. Exception: `takeOwnCopy` in `emitgenerators.nra` keeps two `NARRA_COPY_GRANT` statements, because `consume` writes a declaration plus the retain over several lines, which cannot sit in a macro argument. Both already expand to `((void)0)` in release (that is its guard). Unchanged.

## Proof

Sizes of `narrac emit selfhost` (lines / bytes):

| | base (feat/batch-7) | branch |
|---|---|---|
| as emitted | 186150 / 8618951 | 183657 / 8574436 |
| `unifdef -DNARRA_RELEASE` | 183626 / 8369889 | 183534 / 8569378 |
| `cc -E -DNARRA_RELEASE` | 185818 / 8383405 | 183807 / 8366661 |
| `cc -E` (debug) | 187272 / 8673535 | 184061 / 8678534 |

- Counts: `#ifndef NARRA_RELEASE` 487 -> 2; `NARRA_COPY_GRANT(` 1035 -> 5; `credit_` 599 -> 561 (names only; definitions are macro-wrapped).
- The `unifdef` row grows because the debug code now sits in macros that unifdef cannot remove; the compiler's own release build (`cc -E`, row 3) is the real release size: -2011 lines, -16.7 KB. The emitted C shrinks by 2493 lines and 44.5 KB. Debug preprocessed size grows by 5 KB (macro wrappers).
- Commands: `sh selfhost/bootstrap.sh <dir>`, `<dir>/narrac emit selfhost`, `unifdef -DNARRA_RELEASE`, `cc -E [-DNARRA_RELEASE]`. Details in `log.md`.

## Deviations

- `selfhost/runtime/runtime.h` edited (not in the brief's `shared:`): it is the source of `runtimecore.nra`, which the `generated` suite checks. `runtimecore.nra` regenerated with `--suite generated --update`.

## Stubs left

- None.

## Gotchas

- Review fixes: release `NARRA_COPY_GRANTED` is `do { ... } while (0)` like debug, and `NARRA_COPY_CREDIT` is `if (!narra_copy_on) {} else { ... }` so a following `else` cannot attach to it.
- Seed: `sh selfhost/bootstrap.sh` reaches the fixed point with the committed seed, through stage 3 (stage 1 C differs from stage 2 C, stage 2 equals stage 3). One seed update is needed for "seed is current"; run `sh scripts/seed-update.sh` once at the end of the batch. Not regenerated here. The emitter and runtime change together, so no second update is needed.
- Functions inside `NARRA_COPY_ONLY(...)` collapse to one line after preprocessing, so debug builds lose per-line debug info and C diagnostics for them.
- The credit text must not contain `#` directives or an unparenthesised issue with macro commas: the macros are variadic, so commas are fine.

## Proposed learnings

- Debug-only generated C goes through a variadic macro that is empty in release, not `#ifndef` blocks | one line per site, smaller emitted C | selfhost/src/emit*.nra | no

## Verification

- `sh selfhost/bootstrap.sh`: fixed point (stage 3), seed older than source.
- `sh scripts/test.sh`: PASS in 396s (all shards, 805 cases each, copy samples included).

## Size

- Estimated M, actual M.
