# Outcome: ptr-cast

## Built

- `Ptr.cast(pointer: Ptr<T>): Ptr<U>` and `Ptr.element(pointer: Ptr<FixedArray<T, N>>, index: int): Ptr<T>` in `std/ptr`, both with the `unsafe` effect. Rows, checks and emitters are in `selfhost/src/stdptr.nra` and `selfhost/src/emitpointers.nra`; the dispatch files needed no change.
- `Ptr.cast` takes `U` from the expected type (`const bytes: Ptr<u8> = Ptr.cast(words)`), like `Ptr.alloc`. Without one it reports E0998 "takes its element type from where the result goes". It emits `((U*)(p))`.
- `Ptr.element` proves its index with `checkFixedIndex`: a constant outside 0 to N - 1 is E0994, a non-constant index needs the same proof as a FixedArray read. A pointer to a non-FixedArray is E0998. It emits `(&(p)->items[i])`; there is no run-time check.
- fmt: `from /* x */ "c"` keeps the comment after `from`; a line comment or own-line comment after `from` is printed after the library name, and a second fmt pass leaves it unchanged. The parser notes nodes `k` (the `from` keyword) and `l` (the library name) and the printer wraps each with its comments (`parser.nra`, `fmt.nra`, key letters documented in `trivia.nra`).
- Samples: `testdata/unsafe_ptr_cast.nra` (+ `.out`), `unsafe_fixed_arrays.nra` rewritten to use `Ptr.element` (memset now takes `Ptr<u8>`, output unchanged). Errors: `testdata/errors/unsafe_ptr_cast_outside`, `unsafe_ptr_element_range`. Fmt: `testdata/fmt/comment_extern_from`.
- SPEC.md (std/ptr bullet) and the E0998 explanation in `diagnostics.nra` list both functions.

## Proof

- base: on the unchanged fmt, `printf 'extern {\n  function abs(value: i32): i32\n} from /* x */ "c"\n' > c.nra; narrac fmt c.nra` gave `} /* x */ from "c"`.
- branch: the same command gives `} from /* x */ "c"`; `comment_extern_brace` (comments before `from`) is unchanged.

## Deviations

- Guide: the guide has no chapter on pointers or `std/ptr`, so only SPEC.md lists the functions.
- `Ptr.cast<U>` is written with the expected type, not an explicit type argument, as `Ptr.alloc` does.
- `diagnostics.nra` and `trivia.nra` (doc comment) and `parser.nra`/`fmt.nra` were touched for text and the fmt fix. The review fixes also touched `selfhost/src/fmtprinter.nra` (new `moveLineCommentsToTrailing`) and `selfhost/tests/src/fmtstable.nra` (two MOVING_COMMENTS rows), outside the brief's paths.
- No `compiler/README.md` entry: no difference from the spec.

## Stubs left

- None.

## Gotchas

- A comment before an AST node on the same line becomes its leading comment, so a keyword that can carry a comment needs a node of its own.
- Machine load was 50 to 60 during the run; the full check took very long.

## Proposed learnings

- Give a keyword its own trivia node when comments can sit on either side of it | the nearest-node rule otherwise moves the comment | selfhost/src | yes (fmt goldens)

## Verification

- After the review fixes, `sh scripts/test.sh`: PASS in 265s. Before them: PASS in 331s (8 shards, 880 cases each, plus the bootstrap job; fixed point reached, stage 1 and stage 2 identical).
- Earlier full run failed only on the missing `comment_extern_from.fmt` golden, which I had deleted by mistake; restored.

## Size

- Estimated M, actual M.
