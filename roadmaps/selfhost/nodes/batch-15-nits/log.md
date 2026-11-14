# Log: batch-15-nits

- Base: feat/batch-15 at c4884fe1. Base compiler built from a `git archive` of that commit with `sh selfhost/bootstrap.sh`.
- Proof on base (`base/narrac fmt` on `testdata/fmt/uses_list_comments.nra`): every comment inside a `uses` list moves, for example `Generator<int> uses (io, /* c */ alloc)` becomes `Generator<int /* c */> uses (io, alloc)`, `uses /* f */ io, /* g */ alloc` becomes `uses io, alloc /* f */ /* g */`, and `() => int uses (io, /* h */ alloc)` becomes `() => int /* h */ uses (io, alloc)`.
- Proof on base (`base/narrac check testdata/errors/simd_load_fixed.nra`): `simd_load_fixed.nra:6:43: expected Array<i32>, found FixedArray<i32, 4>`, no hint.
- Fix: the parser notes a trivia node for the `uses` keyword (`u<pos>`) and one per effect (`u<pos>.<index>`), keyed by the position of the declaration keyword or of the type. fmt prints both through `printUses` in `fmtprinter.nra`.
- Fix: `Simd.load` reports a FixedArray first argument with a hint naming `Simd.fromFixed` (code E0119, the code of the plain mismatch; no new code).
- Iterating: a release build through the seed takes 25 minutes under load. Faster: build with the base compiler, `base/narrac emit selfhost`, then `cc -O0`.
