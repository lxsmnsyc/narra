# Outcome: wrapping-arithmetic

## Built

- `+`, `-`, `*` and unary `-` wrap at the type's width in every build (`narra_wrap_*` in `selfhost/runtime/runtime.h`, `emitArithmetic` and `emitUnary` in `selfhost/src/emitexprs.nra`). `MIN / -1` gives `MIN`, `MIN % -1` gives 0.
- `+%`, `-%`, `*%` and the compound forms do not exist. They are gone from the checker and `fmt` tables.
- `Int.addChecked`, `subChecked`, `mulChecked` and the same in `std/i8` to `std/u64` return `Option` (`selfhost/src/stdsized.nra`).
- The 28 uses in `selfhost/src` and `std/src` are plain operators now.
- Tests: `testdata/wrapping.nra`, The two overflow panic samples are removed.
- SPEC.md, D92, compiler/README.md and the guide chapters 1, 2 and 6 updated.

## Proof

Not a fix node.

## Result

Not an experiment node.

## Deviations

- The E0971 deprecation path for `+%`, `-%`, `*%` and the compound forms is removed (diagnostic, parser check, fix, lexer tokens, six `wrapping_*` samples). Removed syntax needs no deprecation path (maintainer, 2026-11-10); the source now lexes as `+` then `%`.
- D92 stays marked as partly built: the division, shift, missing-return and std `Result` parts are not.

## Stubs left

- None.

## Gotchas

- A compiler built from this source emits `narra_wrap_neg`, which only the new generated runtime text has. After changing `runtime.h`, rewrite the generated files with the seed-built compiler (`NARRAC=<seed build> sh scripts/test.sh --suite generated --update`) and then bootstrap.
- `seed/narrac.c` is older than the source (bootstrap prints stage 3); the orchestrator regenerates it for the batch.

## Proposed learnings

- Change a runtime helper and its generated text with the seed-built compiler first | the new compiler emits calls the old text lacks | selfhost | no

## Verification

- `sh scripts/test.sh`: PASS (8 shards of 708 cases and the bootstrap job). No golden changed after that run.
- `NARRAC=<built> sh scripts/test.sh --suite dumps --update` and `--suite edge --update` rewrote the goldens listed in log.md.
- `testdata/wrapping.nra` gives identical output in `-sanitize` and `-release`.

## Size

- Estimated M, actual M.
