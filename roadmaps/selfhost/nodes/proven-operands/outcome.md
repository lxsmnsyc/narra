# Outcome: proven-operands

## Built

- E0979 (integer `/`, `%`, `/=`, `%=` with a divisor that is not a non-zero constant or a narrowed name) and E0980 (a shift count that is not a constant in `[0, width)` or a name checked at both ends; an unsigned count needs only the upper bound). Code: `selfhost/src/operands.nra` (`narrowByCondition`, `recordConstantLocal`, `checkDivisor`, `checkShiftCount`), wired into `exprs.nra` (`checkBinary`), `stmts.nra` (`checkIf`, `checkWhile`, compound assignment, `const` declarations) and `patterns.nra` (`endNarrowing` also ends facts).
- Review fix: a folded constant is reduced to the operand type's width and signedness (`wrapConstant`) before it is judged, for divisors, shift counts, `const` locals and comparison bounds. `const D: i32 = 65536 * 65536` and `const D: u8 = 255 + 1` are zero, so E0979; `const C: u8 = 255 + 3` is 2, so a shift by it is accepted. Samples: `testdata/errors/divisor_wraps_i32`, `divisor_wraps_u8`, `shift_wraps_i32`, and `wrappedConstants` in `proven_operands.nra`.
- Second review fix: a u64 constant from 2^63 up is negative as an int, so for a u64 local such a bound proves nothing (`refineFact` skips it) and a `const` u64 local holding one records only "non-zero" (`recordConstantLocal`). As a divisor it is still accepted. Samples: `testdata/errors/divisor_u64_high_bound`, `shift_u64_high_bound`, and `highDivisor` in `proven_operands.nra`.
- Narrowing forms: comparisons of a local integer with a constant (either side) joined by `&&`, `||`, `!`; the then and else branches of `if`, the rest of the block after a diverging `then`, `while` bodies, the right side of `&&` and `||`.
- The emitted C and `selfhost/runtime/runtime.h` have no division or shift panic (`narra_div`, `narra_rem`, `narra_shl`, `narra_shr` and the sized forms take no position).
- D92 amendment in DECISIONS.md; SPEC.md operator table, compound assignment and section 6 bullets; compiler/README.md line; guide chapters 5, 6 and 10 adjusted.
- Tests: `testdata/errors/divisor_*`, `shift_*` (10 samples), `testdata/proven_operands.nra`, an emit check `proven-operands/no-panic-path` in `selfhost/tests/src/emitchecks.nra`.

## Deviations

- Edited shared paths: `testdata/{coroutine_tests,std_test,patterns_tuples,wrapping}.nra` (they relied on the division panic or a `const` divisor), `selfhost/tests/src/{main,panics,emitchecks}.nra`, `docs/guide/{05,06,10}`. Nothing outside the owned and shared paths.
- `testdata/panics/divzero.nra` and its entry in the panics suite were removed: the panic no longer exists.
- The generated `runtimecore.nra` and `stdtext.nra` were regenerated.

## Stubs left

- None.

## Gotchas

- Sites rejected by the new rule: 3 in the compiler (`hash.nra` rotateLeft32 for both shifts, `check.nra` alignUp), 1 in the test runner (`main.nra` shard split), 1 in std (`Random.nextInt`). All fixed by an added check.
- The seed is not regenerated. Stage 1 (built by the old seed) does not know the rule; stage 2 does, so the check on `selfhost/src` runs in stage 2.
- `std-no-panic` will turn `Random.nextInt` into an `Option`; the `span == 0` branch added here returns `low` and can go then.
- A divisor like `n + 1` or `Arrays.length(x)` is never proven; the user copies it into a local and checks it.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: pass (8 shards plus bootstrap, 109 s).
- `sh selfhost/bootstrap.sh <out>`: stage 2 and stage 3 emit identical C; the seed is older than the source, as expected.

## Size

- Estimated M, actual M.
