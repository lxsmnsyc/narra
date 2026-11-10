# Log: wrapping-arithmetic

- Branch reset to origin/feat/batch-3 first.
- Runtime: `narra_add/sub/mul/neg` and the sized forms are now `narra_wrap_*` with no position and no check. `NARRA_CHECKED` is gone. `narra_div` and `narra_rem` keep the zero check; `MIN / -1` gives `MIN`, `MIN % -1` gives 0.
- Emitter: `emitArithmetic` sends `+ - *` to the wrap functions; unary `-` too. The call stays in a temp so the order of evaluation matches the old checked calls.
- Parser: `+%`, `-%`, `*%` leave the precedence table and the compound list. The lexer still reads them as one token, and `parseBinary` reports E0971 with a fix that replaces the two characters by the plain operator. `x +%= 1` hits the same check, so the fix gives `x += 1`. The checker and `fmt` tables lost the three operators.
- Checked functions: `addChecked`, `subChecked`, `mulChecked` are rows in `conversionFuncs` of `stdsized.nra`, so `std/int` and `std/i8` to `std/u64` all get them. The emitter uses `__builtin_*_overflow` on a temp of the result type.
- Bootstrap order: the generated runtime text had to be rewritten with the seed-built compiler (`NARRAC=<seed build> --suite generated --update`), because a compiler built from the new source emits `narra_wrap_neg` and the old runtime text lacks it.
- Seed rule: the source of `selfhost/src` only lost `+%` uses, so the seed still compiles it. The seed-built compiler is a release build, so the hash and lexer code still wraps. The seed was not regenerated.
- Tests: `testdata/wrapping.nra` (debug with sanitizers and release give the same output), six error samples `testdata/errors/wrapping_*`, added to `FIXED_ERRORS` in `habits.nra` so each fix is applied and re-checked. The two overflow panic samples are removed.
- Goldens changed by the node: `testdata/astdump/exprs.ast`, `testdata/parse/export_const.ast` (columns, plain operators), `testdata/typesdump/bodies/basics.dump`, `testdata/compound_assign.out`, and `selfhost/src/runtimecore.nra`, `stdtext.nra` (generated).
