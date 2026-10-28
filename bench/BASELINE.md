# Baseline

First recorded run of `bench/run.sh` (7 runs per program, median wall-clock seconds).

## Machine

- Apple M2, 8 cores, 8 GB memory, macOS 27 (Darwin 27.0.0), arm64.
- Narra: bootstrap compiler at commit 587b654 plus this node, `narra build -release`, C built with Apple clang 21.0.0.
- Go: go1.27.2 (`go build`).
- Rust: rustc 1.98.0 (`rustc -C opt-level=3`).

The machine was not quiet: another agent was building in parallel. Treat the numbers as a smoke test. Record a new baseline on an idle machine before comparing performance nodes.

## Results (seconds)

| program   | narra | go    | rust  | output     |
| --------- | ----- | ----- | ----- | ---------- |
| fib       | 0.141 | 0.132 | 0.123 | 39088169   |
| sieve     | 0.321 | 0.186 | 0.232 | 1270607    |
| matmul    | 0.130 | 0.262 | 0.183 | 3547800000 |
| strings   | 0.300 | 0.735 | 0.368 | 100000     |
| interp    | 0.237 | 0.246 | 0.173 | 972838     |
| particles | 0.191 | 0.075 | 0.095 | 6996184    |

## Notes

- Each run includes process start. `run.sh` checks that all languages print the same answer.
- `sieve` and `particles` are where Narra is slowest against the others. A likely cause is reference counting and copy on write on array element updates; not profiled.
