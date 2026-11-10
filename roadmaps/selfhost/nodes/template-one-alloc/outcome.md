# Outcome: template-one-alloc

## Built

- `selfhost/runtime/runtime.h`: `lstr_concat_n(count, parts)` sums the lengths and copies each part once into one `lstr`. Like `lstr_concat`, it only borrows its parts.
- `selfhost/src/emitexprs.nra`: `emitTemplate` holds earlier parts when a later one may write, then emits `lstr_concat_n(n, (lstr*[]){a, b, ...})` for two or more parts, as one owned temp. A single part is returned as is (same ownership as before). An empty template is still the empty literal.
- `selfhost/src/runtimecore.nra` regenerated.
- `testdata/template_one_alloc.nra` and `.out`: 1, 2 and many parts, empty parts, non-ASCII text, interpolated ints and calls. The `.out` was captured from the compiler before the change.

## Proof

Apple M2, `cc -O2 -DNARRA_RELEASE`, `/usr/bin/time -l <narrac> emit selfhost`, 3 alternating runs each, base feat/batch-4 at the rebase point.

- base: 3.625, 3.622, 3.611 G instructions retired.
- branch: 3.464, 3.467, 3.442 G.
- Gain 4.6%; the spread on each side is under 0.4%.
- The sample output is identical between the base and branch compilers, including the two probes where a later part assigns a variable that earlier parts read.

## Deviations

- None. The seed is not regenerated. String `+` chains are unchanged.

## Stubs left

- String `+` chains and `+=` still use `lstr_concat`.

## Gotchas

- Template parts must be strings, so ints are written with `Int.toString`.
- The array of parts is built after the statements of every part. When a later part may assign a variable (`mayWrite`), `emitTemplate` holds each earlier part in a retained temp first (`holdArgument`, as calls do), so the written left-to-right order stays. A plain pointer copy would not do: the assignment releases the old string.

## Verification

- `sh scripts/test.sh`: PASS in 161 s, 8 shards, 0 failed.
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C.

## Size

- Estimated M, actual S.
