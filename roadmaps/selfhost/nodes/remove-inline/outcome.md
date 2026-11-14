# Outcome: remove-inline

## Built

- `@inline` is gone. Deleted: `selfhost/src/inlining.nra` (371 lines), `selfhost/src/emitinline.nra` (458 lines), the `inline` row of the attribute table and the extern-member check in `parser.nra`, E1030 and E1031 (table rows and long texts in `diagnostics.nra`), the `checkInlineFunctions` call in `body.nra`, the `isInlineFunction` branch in `emitexprs.nra`, and `inlineSites`, `inlineCount`, `labelSuffix`, `InlineSite`, `InlineCaller`, `InlineArrow`, `LoopInfo.suffix` in `emitstate.nra`.
- `Ref.Value` calls now go straight to `Closures.emitValueCall` (the old `emitInlineValueCall` fell through to it when no arrow was bound).
- `std/src/option.nra`: `map`, `andThen`, `orElse` have no attribute; `selfhost/src/stdtext.nra` regenerated.
- Samples removed: `testdata/inline*.nra` (nine, with goldens), `testdata/packages/inline`, `testdata/panics/inline_bounds.nra`, `testdata/errors/inline_*` (eight) and `value_prelude_inline_uncalled`, `testdata/astdump/inline_attribute`, `testdata/fmt/comment_inline_attribute`. `testdata/param_defaults.nra` keeps its sample without the attribute. Added `testdata/errors/inline_removed.nra` (the unknown-attribute error for `@inline`). `target_attribute_unknown.err` lost `@inline` from the known list.
- Test runner: removed the `inline/no-closure` emit check (`emitchecks.nra`), the `inline_bounds` panic sample, and the `@inline` line of `spec-skip.txt`.
- Docs: D120 in DECISIONS.md (supersedes the `@inline` parts of D119 and D95 with its batch-11 amendments); SPEC section 2 paragraph and list removed, the section 14 closures sentence trimmed; `compiler/README.md` `@inline` section removed; `docs/guide/README.md` planned list drops `@inline`.

## Proof

- base (feat/batch-15, 76f911ff): `bench/inline` `-release` retires 1926.2M instructions; `std_option` output matches its golden; `narrac check` accepts `@inline function`.
- branch: 1929.9M instructions (+0.2%, noise; same code path, ordinary generic functions and closures that clang -O2 inlines); `std_option` output unchanged; `@inline` gives `unknown attribute @inline; the known attributes are @target, @trusted, @extern, @const`.
- Emit: `narrac emit selfhost` from the base compiler and from the branch compiler, both on the branch source, are byte-identical (246,442 lines). The branch source calls none of `map`, `andThen`, `orElse`.

## Deviations

- Added `testdata/errors/inline_removed.nra` (under the owned `testdata/errors/inline_*`), which the brief did not list.
- Removed test-runner entries in `selfhost/tests/src/emitchecks.nra`, `panics.nra` and `spec-skip.txt` (shared path `testdata/**` does not cover them; they reference deleted samples).
- The brief's SPEC list named a generics sentence in section 1 and a default-parameter list; neither mentions `@inline` in SPEC.md, so only section 2 and the section 14 closures sentence changed. D115's record text still says `@inline` (old record, not rewritten).
- Seed not regenerated; `selfhost/src` never used `@inline`, so the seed still compiles the source.

## Stubs left

- None.

## Gotchas

- The seed (`seed/narrac.c`) still contains the inliner until the orchestrator regenerates it; bootstrap goes through stage 3 meanwhile.
- A user program that writes `@inline` now fails at the parser, like any unknown attribute.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point (stage 2 and 3 identical, 246,442 lines).
- `narrac fmt` on every edited `.nra` file; `--suite generated --update` and `--suite errors --update` run.
- Full `sh scripts/test.sh`: see the report line below.

## Size

- Estimated M, actual M.
