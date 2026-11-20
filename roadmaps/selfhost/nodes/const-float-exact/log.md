# Log: const-float-exact

- Base 94107ca4 (feat/batch-20). Case `testdata/const_float_literals.nra`, run with `NARRAC=<base narrac> sh scripts/test.sh --filter const_float_literals`. Result: FAIL, "constant HALF_BELOW cannot be computed while compiling: it uses the float literal 0.49999999999999994, which has more digits or a larger exponent than the compiler reads exactly".
- Cause: `parseFloatText` in `selfhost/src/check/consteval.nra` only accepts at most 18 digits with a power of ten up to 22 (Clinger fast path), and returns None otherwise.
- Fix: keep that path (renamed `parseShortFloatText`); on None, rewrite the literal as `<digits>e<power>` and read it with `std/json`, which uses the C library reader (correctly rounded). An infinite result is still rejected.
- Branch: same filter and `--filter const_` pass (459 cases).
- F32.from(int) beyond 2^53 (E1052 in @const): skipped. It needs a direct int-to-f32 conversion in the evaluator and the emitter must agree; not a small change.
- A first bootstrap.sh run was stopped because I edited the source while it ran; the base compiler was rebuilt from a clean tree.
