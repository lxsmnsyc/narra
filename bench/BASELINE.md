# Baseline

The latest run is first. Compare performance nodes against it.

## 2026-11-11, end of batch-7

Run of `bench/run.sh -n 9` at commit 7706132c on `feat/batch-7`, same machine and toolchains as the batch-6 run below. No agent was running. The load average was 1.84 (1 minute) at the start, with the 5 and 15 minute averages (3.45 and 4.83) still falling after the batch's test runs, and 1.89 at the end.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.111 | 0.124 | 0.112 | 39088169   |
| sieve     | 0.134 | 0.144 | 0.133 | 1270607    |
| matmul    | 0.106 | 0.203 | 0.118 | 3547800000 |
| strings   | 0.213 | 0.297 | 0.355 | 100000     |
| interp    | 0.185 | 0.242 | 0.169 | 972838     |
| particles | 0.068 | 0.074 | 0.093 | 6996184    |
| compiler  | 0.235 | -     | -     | peak memory 130 MB |

The benchmark programs are within noise of batch-6. Instructions retired by `narrac emit` (median of 5, measured right after the run): the batch-7 compiler on its own source 2.18G (batch-6: 2.47G on its source); on the same input, the batch-6 source, 2.44G with the batch-6 compiler and 2.05G with the batch-7 compiler (-16%). The source grew 7.6% (56,903 to 61,201 lines) with contracts, ensures, build conditions and `@target`, which costs about 0.13G. The seed shrank from 186,158 to 177,598 lines (release, preprocessed: 180,497 to 173,624).

## 2026-11-11, end of batch-6

Run of `bench/run.sh -n 9` at commit d52185b2 on `feat/batch-6`, same machine and toolchains as the batch-5 run below. No agent was running. The load average was 1.93 (1 minute) at the start, with the 5 and 15 minute averages (7.1 and 21.4) still falling after the test runs, and 2.32 at the end. `interp` no longer compiled: its `acc % n` had no proof that `n` is not 0 (proven-operands), and no test suite builds `bench/`. It now calls a `remainder` function that checks the divisor, as Go and Rust do implicitly, and was rerun alone at a load average of 2.29.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.112 | 0.125 | 0.111 | 39088169   |
| sieve     | 0.128 | 0.138 | 0.163 | 1270607    |
| matmul    | 0.109 | 0.217 | 0.119 | 3547800000 |
| strings   | 0.228 | 0.327 | 0.356 | 100000     |
| interp    | 0.185 | 0.242 | 0.169 | 972838     |
| particles | 0.073 | 0.074 | 0.093 | 6996184    |
| compiler  | 0.245 | -     | -     | peak memory 132 MB |

The benchmark programs are within noise of batch-5. The compiler row emits the batch-6 source: 0.175 s to 0.245 s, and 1.74G to 2.47G instructions, mostly the shared-write analysis that now runs in every command (`roadmaps/selfhost/reviews/wave-19.md`). The text helpers' temporary +32% is gone.

## 2026-11-11, end of batch-5

Run of `bench/run.sh -n 9` at commit cb6cef4b on `feat/batch-5`, same machine and toolchains as the batch-4 runs below. No agent was running. The machine had booted 10 minutes earlier: the load average was 1.99 (1 minute) at the start, with the 5 and 15 minute averages (14.5 and 13.4) still falling from startup indexing, and 2.50 at the end. The Go and Rust times are 2% to 5% lower than in the end-of-batch-4 run, so Narra changes of that size are noise.

The start of batch-5 is the end-of-batch-4 run below: no compiler, runtime or seed change landed between afb1eac3 and the batch-5 base.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.112 | 0.124 | 0.112 | 39088169   |
| sieve     | 0.132 | 0.140 | 0.129 | 1270607    |
| matmul    | 0.106 | 0.202 | 0.118 | 3547800000 |
| strings   | 0.217 | 0.290 | 0.353 | 100000     |
| interp    | 0.184 | 0.241 | 0.168 | 972838     |
| particles | 0.067 | 0.074 | 0.092 | 6996184    |
| compiler  | 0.175 | -     | -     | peak memory 121 MB |

`sieve` (0.265 s to 0.132 s) and `particles` (0.108 s to 0.067 s) come from array-write-cold-paths and loop-unique-hoist. The compiler row emits the batch-5 source, which differs from the batch-4 input; wave 18 (`roadmaps/selfhost/reviews/wave-18.md`) has the same-input instruction counts.

## 2026-11-10, end of batch-4

Run of `bench/run.sh -n 9` at commit afb1eac3 on `feat/batch-4`, same machine and toolchains as the start-of-batch run below. No agent was running. The load average was 1.99 (1 minute) at the start and 3.57 at the end; the benchmark itself adds about 1.

### Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.115 | 0.127 | 0.123 | 39088169   |
| sieve     | 0.265 | 0.177 | 0.148 | 1270607    |
| matmul    | 0.112 | 0.204 | 0.126 | 3547800000 |
| strings   | 0.238 | 0.370 | 0.369 | 100000     |
| interp    | 0.190 | 0.274 | 0.179 | 972838     |
| particles | 0.108 | 0.074 | 0.099 | 6996184    |
| compiler  | 0.272 | -     | -     | peak memory 125 MB |

The compiler row emits the final source, which is larger than at the start of the batch. On the same input (the start-of-batch `selfhost`, 9 alternating runs each), the start-of-batch compiler took 0.61 s, 9.30G instructions and 129 MB, and the final one 0.23 s, 2.59G instructions and 121 MB.

Instructions retired for programs built by each compiler (start of batch to end): interp 4.88G to 1.97G, particles 4.21G to 2.61G, strings 1.24G to 1.10G, sieve 2.41G to 2.28G, fib and matmul unchanged.

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
