---
id: remove-inline
title: "Remove @inline"
status: done
priority: should
type: feature
story: batch-15
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 170000
owns:
  - roadmaps/selfhost/nodes/remove-inline/**
  - selfhost/src/inlining.nra
  - selfhost/src/emitinline.nra
  - testdata/inline*
  - testdata/errors/inline_*
  - testdata/panics/inline_bounds*
  - testdata/packages/inline/**
shared:
  - selfhost/src/*.nra
  - std/src/option.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - docs/guide/*.md
  - compiler/README.md
reserves: [decision D120]
provides: []
depends_on:
  - on: inline-benchmark
    kind: hard
    why: the benchmark found no gain from @inline (outcome.md); the maintainer approved removal on 2026-11-14
---

# Remove @inline

## Goal

The `@inline` attribute no longer exists. `std/option` has no attribute, `Options.map`, `andThen` and `orElse` behave as before, and the compiler is about 830 lines smaller. A program that writes `@inline` gets the parser's unknown-attribute error.

## Notes

The maintainer approved the removal on 2026-11-14 after reading `nodes/inline-benchmark/outcome.md`. Write the decision as D120 (superseding the `@inline` parts of D119 and of the record that introduced it). Proof first: `bench/inline` and `testdata/std_option.nra` outputs and instruction counts before the change, recorded in `log.md`. `selfhost/src` and seed-compiled std may use only features the committed seed compiles. Scratch folder: /private/tmp/claude-501/remove-inline/. Follow `roadmaps/selfhost/CLAUDE.md`.

What removal touches:

- Attribute: the `inline` row in the attribute table (`selfhost/src/parser.nra`, near line 1377) and the extern-block check near line 1219.
- Error codes: E1030 and E1031 in `selfhost/src/diagnostics.nra` (table rows and long texts). Keep E0990 (attributes in `implement` blocks, used by other attributes). The `codes` suite and the `testdata/errors/inline_*` samples go with them.
- Inliner: `selfhost/src/inlining.nra` (371 lines) and `selfhost/src/emitinline.nra` (458 lines), the call to `checkInlineFunctions` in `body.nra`, the branch in `emitexprs.nra` (`isInlineFunction`, the `Ref.Value` case), and the emitter state in `emitstate.nra` (`inlineSites`, `inlineCount`, `labelSuffix`, `InlineSite`, `InlineCaller`, `InlineArrow`).
- Std: remove `@inline` from `map`, `andThen`, `orElse` in `std/src/option.nra`, then regenerate `selfhost/src/stdtext.nra` (`sh scripts/test.sh --suite generated --update`). The orchestrator regenerates the seed once per batch.
- Samples: `testdata/inline*.nra` (about ten, with goldens), `testdata/panics/inline_bounds.nra`, `testdata/packages/inline`, `testdata/astdump/inline_attribute.nra`, `testdata/fmt/comment_inline_attribute.*`, and the `@inline` use in `testdata/param_defaults.nra` (keep the sample, drop the attribute). Keep `testdata/std_option.nra`, which shows the three functions with capturing and non-capturing arrows.
- SPEC: section 2 (the `@inline` paragraph and its list), the closures sentence in section 14, the generics sentence in section 1, and the default-parameter list. Add a decision record that supersedes the `@inline` parts of D119 and the record that introduced it, and says why (the benchmark). Do not rewrite old records.
- Guide: `docs/guide` chapters that mention `@inline` (grep first).
- Fmt and astdump: the attribute printing is generic; check `fmt` has no special case for `inline`.

What `@inline` did besides speed, and what removal keeps or drops:

- Arrow copies: an arrow passed to an `@inline` function was copied into each place the body calls the parameter, with no environment. Without it the call uses the non-escaping closure path. Keep that path; confirm `std_option` output is unchanged.
- Effects: `uses` rows and `from` markers are checked on the declaration, not on the copy. Nothing to keep.
- Labels: `narra_inline_N` jump labels and the `labelSuffix` rule for loop labels exist only for copies. Drop them, and check no other emitter path sets `labelSuffix`.
- Value uses: an `@inline` function used as a value already got an ordinary C function. No change.
- Guarantee: copying was promised ("a guarantee, not a hint"). No program relies on it; the only observable difference is the C text.

Acceptance outline: `sh scripts/test.sh` passes; the output of `testdata/std_option.nra` is unchanged; `bench/inline` runs at the benchmark's without figure (1911.5M instructions); no `@inline` is left in `selfhost/src`, `std/src`, `SPEC.md` or `docs` except in superseded records.
