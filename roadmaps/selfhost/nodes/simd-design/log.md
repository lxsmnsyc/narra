# Log: simd-design

- 2026-11-13: D112 did not exist; DECISIONS.md ended at D109 on `feat/batch-13`.
- Read D12, D47, D92, D95 (fixed arrays), the FixedArray amendments, SPEC sections 2, 4, 9 and 11, `fixedarrays.nra` (`checkFixedIndex`), `operands.nra` (`isProvenCount`), `stdptr.nra` (type visible on import), `driver.nra` (release flags `-O2 -std=c11 -w`, no `-march`), `runtime.h` (`narra_block_new`: blocks start at 8 mod 16, so vector types need `aligned(8)`).
- Measured C prototypes in `/private/tmp/claude-501/simd-design/` (`dot.c`, `optload.c`), Apple arm64, clang 21, `-O2`, three runs each:
  - scalar unchecked 0.197-0.199 s, F32x4 0.049-0.050 s, F32x8 0.025 s, Array<F32x4> read in place 0.049 s.
  - scalar bounds checked 0.288 s, F32x8 through an Option load 0.036 s.
  - `_Alignof` of both typedefs with `aligned(8)` is 8; ASan and UBSan build of `dot.c` ran clean.
  - x86-64 assembly (`cc -arch x86_64 -S`): `movups`/`mulps`/`addps` for F32x8, `movdqu`/`paddd` for the wrapping I32x4 add through the unsigned twin, scalar loop stays `mulss`/`addss`.
- Not measured: gcc (no Linux host; docker daemon not running), x86-64 run time (no Rosetta).
- No SPEC `ts` samples added (they would need `selfhost/tests/spec-skip.txt`, outside this node); samples live in the build briefs.
- `git branch -m` failed once on a race with other spikes creating `logs/refs/heads/spike/`; retry worked.
