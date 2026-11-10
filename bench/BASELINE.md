# Baseline

The latest run is first. Compare performance nodes against it.

## 2026-11-10, start of batch-4

Run of `bench/run.sh -n 9` (median of 9 runs, wall-clock seconds) at commit 4518a91 on `feat/batch-4`.

### Machine

- Apple M2, 8 cores, 8 GB memory, macOS 27 (Darwin 27.0.0), arm64.
- Narra: compiler built from the seed by `bench/run.sh`, `-release`, C built with Apple clang 21.0.0.
- Go: go1.27.3. Rust: rustc 1.98.0 (`-C opt-level=3`).

The machine was quiet: no agent was building or testing. The load average was 1.92 (1 minute) at the start, with the CPU 80% idle, and 2.99 at the end, mostly from the benchmark itself. The 5 and 15 minute averages (7.4 and 9.1) were still falling from earlier test runs.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.114 | 0.128 | 0.116 | 39088169   |
| sieve     | 0.281 | 0.155 | 0.143 | 1270607    |
| matmul    | 0.114 | 0.207 | 0.122 | 3547800000 |
| strings   | 0.242 | 0.287 | 0.371 | 100000     |
| interp    | 0.398 | 0.248 | 0.173 | 972838     |
| particles | 0.196 | 0.076 | 0.095 | 6996184    |
| compiler  | 0.616 | -     | -     | peak memory 129 MB |

The `compiler` row is the release-built compiler running `narrac emit selfhost` (its own source, about 184k lines of C out). Go and Rust have no compiler row.

These numbers are within about 3% of the 2026-10-31 run, so the busy machine then did not distort it much. `interp` is still about 1.7x its time before D55 (0.237 s at 587b654).

## 2026-10-31

Run of `bench/run.sh` on 2026-10-31 (7 runs per program, median wall-clock seconds).

### Machine

- Apple M2, 8 cores, 8 GB memory, macOS 27 (Darwin 27.0.0), arm64.
- Narra: bootstrap compiler at commit 3fa2dd6, `narra build -release`, C built with Apple clang 21.0.0.
- Go: go1.27.2 (`go build`).
- Rust: rustc 1.98.0 (`rustc -C opt-level=3`).

The machine was not quiet. No agent was building, but macOS background services (`mobileassetd`, `duetexpertd`, `logd`, `triald`) kept the load average between 13 and 8 during the run. A second run of `interp`, `sieve` and `particles` matched within 0.01 s. Record a new baseline on an idle machine (load average under 2) before comparing performance nodes.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.110 | 0.125 | 0.111 | 39088169   |
| sieve     | 0.274 | 0.147 | 0.141 | 1270607    |
| matmul    | 0.111 | 0.201 | 0.115 | 3547800000 |
| strings   | 0.234 | 0.302 | 0.355 | 100000     |
| interp    | 0.393 | 0.244 | 0.169 | 972838     |
| particles | 0.188 | 0.074 | 0.093 | 6996184    |

### Notes

- Each run includes process start. `run.sh` checks that all languages print the same answer.
- `interp` went from 0.237 s at commit 587b654 to 0.393 s. D55 made every `Some`, `Ok` and `Err` allocate and expected `interp` to be about 2x slower until node `perceus` removes retains on borrowed values.
- `sieve` and `particles` are about 2x Go and Rust. A likely cause is reference counting and copy on write on array element updates; not profiled.
- The previous baseline (commit 587b654) was also recorded on a busy machine, so the small gains on the other programs are not evidence of a speedup.
