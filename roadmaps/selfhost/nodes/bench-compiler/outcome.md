# Outcome: bench-compiler

## Built

- `bench/run.sh` has a seventh default program, `compiler`: the median wall time of `narrac emit selfhost` (C output discarded), then one more run under `/usr/bin/time -l` (macOS) or `-v` (Linux) for peak memory. `bench/run.sh compiler` runs it alone. Go and Rust print `-`.
- `bench/README.md` describes the row.
- `bench/BASELINE.md` has a new first section, "2026-11-10, start of batch-4", with all seven rows at commit 4518a91, recorded with `-n 9` at a 1-minute load of 1.92.

## Baseline

| program   | narra | go    | rust  |
| --------- | ----- | ----- | ----- |
| fib       | 0.114 | 0.128 | 0.116 |
| sieve     | 0.281 | 0.155 | 0.143 |
| matmul    | 0.114 | 0.207 | 0.122 |
| strings   | 0.242 | 0.287 | 0.371 |
| interp    | 0.398 | 0.248 | 0.173 |
| particles | 0.196 | 0.076 | 0.095 |
| compiler  | 0.616 (129 MB peak) | - | - |

## Deviations

- `bench/CLAUDE.md` (not an owned path): the "recorded on a busy machine" rule now says to compare against the latest section and to record only under a 1-minute load of 2.

## Gotchas

- The benchmarked compiler is the one `selfhost/bootstrap.sh` builds into `bench/build/narrac-out` (stage 2, built from the seed). A node that changes the emitter measures its own compiler only after the bench build picks up its source, which `bootstrap.sh` does from `selfhost/src`.
- The load average stays high for several minutes after `sh scripts/test.sh` runs in parallel worktrees. Wait for the 1-minute average to fall under 2 before recording.
- The worker stopped before recording because the load never fell under 2 within its turns; the orchestrator recorded the baseline and wrote this outcome.
