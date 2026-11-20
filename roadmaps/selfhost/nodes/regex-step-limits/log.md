# Log: regex-step-limits

## Setup

- Branch `fix/regex-step-limits`, fast-forwarded to `feat/batch-20` (968d3d86).
- Compiler: `sh selfhost/bootstrap.sh <outdir>` from the batch source; stage 1 and stage 2 emit identical C. Test runs used `NARRAC=<outdir>/narrac`.
- Machine: 8 CPUs, ambient load 10 to 15 from other agents.
- Only `testdata/regex_linear` has a timing check. The "5,000 x case that expects over 1 s" in the brief is the fourth line of its output (it prints "under 1 s"; "over 1 s" is the failure text).

## Proof of the flake (base 968d3d86)

Load: 64 processes `perl -e 'alarm 150; exec "yes"'` (each stops itself by alarm; PIDs kept in a file and killed by PID at the end), load average 60 to 150.

- `NARRAC=<narrac> sh scripts/test.sh --filter regex_linear`: first run under that load, `FAIL programs/regex_linear.nra/debug`, lines `(a*)*b against 10,000 a: no match, over 1 s`, `(a+)+$ ...: over 1 s`, `(x+x+)+y ...: over 1 s`.
- Same program built with `narrac build -sanitize`, run 12 times with 64 load processes (load 90 to 220): 8 of 12 runs printed "over 1 s".
- Same, 40 load processes (load 55 to 70): 10 of 12 runs failed.
- Unloaded the case takes 0.6 s in all and each pattern 12 to 32 ms (release), so the 1 s limit fails only when the process is starved.

## Change

`testdata/regex_linear.nra` times each pattern on a text of n and one of 2n, alternating the sizes for 5 rounds and keeping the best time of each. It passes when the best time of 2n is at most 3 times the best time of n. A failing attempt repeats up to 4 times. The output reads "linear growth". No public std API is added.

## Proof of the fix (same load)

- 64 load processes, 8 runs: 1 of 8 failed (load 200 and above, three times the load of the first failure; the 2n run of one pattern was starved in every round of every attempt).
- 40 load processes (load 55 to 70), 12 runs: 0 of 12 failed. The old case failed 10 of 12 under the same load.

## Would it catch a backtracking engine

`backtrack.nra` (scratch, not committed) is a backtracking matcher for `(a+)+$` on `a` repeated n then `b`, timed with the same clock. n=11: 0.70 ms, n=22: 613 ms, a growth of about 876 times against the bound of 3. A polynomial regression of degree 2 grows by about 4 times and also fails the bound.

## Full run

See outcome.md.
