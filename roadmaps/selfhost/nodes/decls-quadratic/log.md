# Log: decls-quadratic

## Setup

- Base: origin/feat/batch-3 (dca242d). Compiler built once with `sh selfhost/bootstrap.sh <out>` (fixed point holds, seed current).
- Scratch directory outside the repo. Single-module packages with N declarations in `src/lib.nra`.
  - Shape A: cycle of struct, function, const, enum (2 fields struct, one-statement function, `const C: int`, enum with one payload variant). Private.
  - Shape B: cycle of 5, all exported: struct with `Array<string>` and `Option<int>` fields, function with a local and an `if`, const, enum with three variants, generic struct `P<A, B> derives Show`.
- Command: `narrac decls <pkg>` and `narrac check <pkg>`, stdout and stderr to /dev/null, wall time, 5 runs each. Machine loaded by two other agents.

## Timings (seconds, 5 runs each)

| shape | command | N | runs |
|---|---|---|---|
| A | decls | 2000 | 0.04 0.04 0.04 0.04 0.04 |
| A | decls | 4000 | 0.06 0.06 0.06 0.06 0.06 |
| A | decls | 8000 | 0.10 0.10 0.10 0.10 0.10 |
| A | check | 2000 | 0.04 0.04 0.04 0.04 0.04 |
| A | check | 4000 | 0.05 0.06 0.05 0.06 0.06 |
| A | check | 8000 | 0.09 0.09 0.10 0.09 0.09 |
| A (second pass) | decls | 2000 / 4000 / 8000 | 0.04-0.06 / 0.06-0.07 / 0.11-0.13 |
| B | decls | 2000 | 0.06 0.06 0.07 0.06 0.06 |
| B | decls | 4000 | 0.14 0.11 0.11 0.11 0.11 |
| B | decls | 8000 | 0.26 0.24 0.25 0.24 0.24 |
| B | check | 2000 | 0.05 0.05 0.05 0.05 0.05 |
| B | check | 4000 | 0.10 0.11 0.11 0.10 0.10 |
| B | check | 8000 | 0.23 0.24 0.24 0.23 0.23 |

Doubling N roughly doubles the time (about 2.0x to 2.3x per step, a quadratic step would give 4x). At 8000 declarations the slowest run is 0.26 s, against 4.9 s reported on the Go-era compiler.

## Decision

Growth is linear. Not reproduced. No profile, no code change.
