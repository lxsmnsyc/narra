# Outcome: copy-check-holders

## Built

- The debug copy check counts the end of a heap enum copy per holder. A heap enum value carries a debug-only `copy_held` id (in the padding after `tag`). `copy(x)` brands the holder it makes (`NARRA_COPY_BRAND`), a retain of a branded value counts a holder (`NARRA_COPY_HELD`), and `NARRA_COPY_ENDED` ends a credit when the last holder of an id is released. Releasing the original ends nothing.
- `selfhost/runtime/runtime.h` (macros, `narra_copy_brand`, holder table), `selfhost/src/runtimecore.nra` (regenerated), `selfhost/src/emitdecls.nra` (struct field, retain hook, brand in the credit code).
- Also `copy_enum_one_alone` and `copy_enum_first_dies` (review probes: a surviving copy after the original, or after another copy, is released; exit 0). The credits of the other counted types still end at a count of 2 or less; the heap enum credits end when the box is freed.
- Samples: `testdata/copy_enum_original_first.nra` (exit 0) and `copy_enum_original_first_hidden.nra` (exit 103, `array 1`). Existing `copy_enum_credit_end` and `copy_enum_credit_twice` still exit 103.
- Release build: the field, hooks and table are inside `NARRA_COPY_ONLY` or `#ifndef NARRA_RELEASE`; the `copy_` samples' release-hook check passes.

## Proof

- base: feat/batch-10 (ececc85a), `narrac build -o c testdata/copy_enum_original_first.nra && NARRA_CHECK_COPIES=1 ./c` prints `copy check: copies on write that no copy(...) explains: array 1`, exit 103 (debug build).
- branch: the same command exits 0, output `list 9 8`.

## Deviations

- `selfhost/src/emitdecls.nra` is outside the owned paths. The per-holder id needs a field in the enum struct and two emitted hooks, which only the emitter can add.
- The runtime text changed and so did the emitter output that uses it, so the seed needs two seed updates (the first with the runtime, the second to use the emitter output). The seed is not regenerated here. Until then, `stage 3` of `bootstrap.sh` reports the seed as older than the source.

## Stubs left

- None.

## Gotchas

- Plain shares of a branded value share its id, so the credit ends only when the last of them is released. A value copied by bits without a retain would unbalance the count.
- The copy ids are not reused: 4 bytes per `copy(...)` with the check on, and a panic after 2^31 - 2 copies.
- Arrays, boxes, maps, builders and generators are unchanged; they were not affected.
- Running `--suite generated --update` with a compiler built from the new `emitdecls.nra` fails to bootstrap until `runtimecore.nra` is regenerated; regenerate it with a compiler built from the old source (`NARRAC=<old narrac>`).

## Proposed learnings

- None.

## Verification

- `NARRAC=<built from branch> sh scripts/test.sh --filter copy`: PASS, 106 cases.
- Full `sh scripts/test.sh` on the rebased tree: `test.sh: PASS, 7131 cases in 184s`.

## Size

- Estimated S, actual S.
