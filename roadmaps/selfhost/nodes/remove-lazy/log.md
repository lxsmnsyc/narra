# remove-lazy log

## Proof first

Base: feat/batch-15 at d64eee66, seed compiler built with `cc -O2 -DNARRA_RELEASE seed/narrac.c`.

- `narrac run testdata/lazy_values.nra` on the base prints `1`, `3`, ... (the lazy sample compiles and runs).
- `git grep -n -i lazy feat/batch-15 -- selfhost/src std/src SPEC.md docs/guide` finds 163 lines (the acceptance grep fails on the base).

## Notes

- Scratch folder: /private/tmp/claude-501/remove-lazy/. The base compiler is built from `git archive feat/batch-15`, so edits in the worktree cannot leak into it.
- `lazy` was a contextual word, not a keyword. Removing `startsLazyValue` makes `const x = lazy 1` an ordinary parse error, with no special message.
- `ConstDecl.lazyPos` is gone, so the AST and declarations dumps lose the `lazy` flag.
- The emitter's "lazy constants" are the D61 getter constants (built on first use); renamed to getter constants, same C.
- E1032-E1035 rows stay in diagnostics.nra with the retired format text, as E0022 does.
