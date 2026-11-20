# Outcome: regex-step-limits

## Built

- `testdata/regex_linear.nra` no longer compares against a number of seconds. It times each pattern on a text of n and one of 2n (best of 5 alternating rounds, up to 4 attempts) and passes when the 2n time is at most 3 times the n time. `testdata/regex_linear.out` prints "linear growth".
- No std API was added and no hook. `std/regex` is unchanged.

## Proof

- base: 968d3d86, `NARRAC=<narrac> sh scripts/test.sh --filter regex_linear` under 64 `yes` load processes: `FAIL programs/regex_linear.nra/debug` ("over 1 s" on three patterns). The sanitized binary run 12 times under 64 load processes failed 8 times, and under 40 load processes (load 55 to 70) failed 10 times.
- branch: same sanitized binary, 40 load processes: 0 of 12 failed. Under 64 load processes (load 200 and above): 1 of 8 failed.
- A backtracking matcher for `(a+)+$` timed the same way grew 876 times from n=11 to n=22 (0.70 ms to 613 ms), far above the bound of 3, so the check still fails for a backtracking engine.

## Deviations

- The brief says the regex sample has a 5,000 x case that expects "over 1 s". No such case exists: all four lines of `regex_linear` expect "under 1 s". Only that file was changed.
- Text sizes shrank to 3,000 and 6,000 so the sanitized run stays near 2 s. The check is a time ratio, not a step count, so it stays load dependent at extreme load (above about 200).

## Stubs left

- None.

## Gotchas

- A step counter in `std/regex` would be exact, but it needs an exported function. That is a public API decision for the maintainer.
- The bound of 3 catches growth of degree 2 (about 4 times), not growth by n log n.

## Proposed learnings

- A test must not compare wall-clock time with a fixed limit; compare two sizes in the same process | load slows both runs | testdata | no

## Verification

- `sh scripts/test.sh` without load: `test.sh: PASS, 10478 cases in 161s`.
- `narrac fmt --check testdata/regex_linear.nra`: clean.

## Size

- Estimated S, actual S.
