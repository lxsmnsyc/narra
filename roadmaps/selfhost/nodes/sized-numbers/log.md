# Log: sized-numbers

- Read the brief, selfhost/CLAUDE.md, types/exprs/emit sites. Go retired, so only selfhost/src and the runtime change.
- Added seven `Type` variants (no sized type used in the compiler source, per the seed rule).
- Literal typing by `want`; negated literal flag on the Checker; u64 range by digit count.
- Runtime: one macro per type instead of 8 hand-written families.
- First stage 2 build ran the sample clean under ASan on the first try. `d << 1 >> 1` on u8 200 gives 72 (the shift truncates to the width, no overflow check, as for int).
- `sh selfhost/bootstrap.sh` fails the stage 1 vs stage 2 comparison until the seed is regenerated (runtime text differs). Stage 2 equals stage 3.
- Error goldens and the sample's `.out` written with the stage 2 compiler.
