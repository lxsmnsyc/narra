# Outcome: async-timing-flake

## Built

- `testdata/async_api_all_observed.nra`: the timers that decided the print order now differ by 100 ms or more (first, raced, late 100 ms; quick 200 ms; shared and early stay 1 ms). The expected output is unchanged.

## Proof

- base: feat/batch-13 11d255b0, 8 parallel loops of 25 runs of the built sample: 2, 18 and 8 failures of 200 in three tries.
- branch: the same command: 0 and 0 failures of 200; also 0 of 200 with 16 extra busy loops.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- The sample now takes about 0.5 s of wall time.
- By hand, removing the observer check in `std_async_loop.c` (in the emitted C) still fails the sample.

## Proposed learnings

- Order a sample's output by awaited results, or put 100 ms between timers whose order is printed | close timers swap under load | testdata | no

## Verification

- Stress runs above; `narrac emit selfhost` is byte-identical by construction (only a testdata file changed).
- `--filter async` run: see the report.

## Size

- Estimated S, actual S.
