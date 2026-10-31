# Baseline

Run of `bench/run.sh` on 2026-10-31 (7 runs per program, median wall-clock seconds).

## Machine

- Apple M2, 8 cores, 8 GB memory, macOS 27 (Darwin 27.0.0), arm64.
- Narra: bootstrap compiler at commit 3fa2dd6, `narra build -release`, C built with Apple clang 21.0.0.
- Go: go1.27.2 (`go build`).
- Rust: rustc 1.98.0 (`rustc -C opt-level=3`).

The machine was not quiet. No agent was building, but macOS background services (`mobileassetd`, `duetexpertd`, `logd`, `triald`) kept the load average between 13 and 8 during the run. A second run of `interp`, `sieve` and `particles` matched within 0.01 s. Record a new baseline on an idle machine (load average under 2) before comparing performance nodes.

## Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.110 | 0.125 | 0.111 | 39088169   |
| sieve     | 0.274 | 0.147 | 0.141 | 1270607    |
| matmul    | 0.111 | 0.201 | 0.115 | 3547800000 |
| strings   | 0.234 | 0.302 | 0.355 | 100000     |
| interp    | 0.393 | 0.244 | 0.169 | 972838     |
| particles | 0.188 | 0.074 | 0.093 | 6996184    |

## Notes

- Each run includes process start. `run.sh` checks that all languages print the same answer.
- `interp` went from 0.237 s at commit 587b654 to 0.393 s. D55 made every `Some`, `Ok` and `Err` allocate and expected `interp` to be about 2x slower until node `perceus` removes retains on borrowed values.
- `sieve` and `particles` are about 2x Go and Rust. A likely cause is reference counting and copy on write on array element updates; not profiled.
- The previous baseline (commit 587b654) was also recorded on a busy machine, so the small gains on the other programs are not evidence of a speedup.
