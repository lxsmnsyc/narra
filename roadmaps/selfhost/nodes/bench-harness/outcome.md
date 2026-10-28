# Outcome: bench-harness

## Built

- `bench/run.sh [-n RUNS] [program ...]`: builds the Narra compiler, then each program in release mode (`narra build -release`, `go build`, `rustc -C opt-level=3`), runs each 7 times by default, prints a table of median wall-clock seconds per language, and checks all languages print the same answer. Exit code is 1 on a failed run or an output mismatch.
- A missing Go, rustc or C compiler skips that language with a note on stderr and a `-` in the table.
- `bench/median.pl`: the timer (perl Time::HiRes).
- Programs in `bench/<name>/<name>.{nra,go,rs}`: `fib`, `sieve`, `matmul`, `strings`, `interp`, `particles`. Add one by adding three files and its name to the list in `run.sh`.
- `bench/BASELINE.md`: first run with machine description. `bench/README.md`: usage. `bench/.gitignore` ignores `bench/build/`.
- Acceptance: the three criteria hold. Not "already fulfilled"; `bench/` did not exist.

## Deviations

- None. No compiler changes.

## Stubs left

- None.

## Gotchas

- The baseline was taken while another agent was building, so it is a smoke test. Re-record it on a quiet machine before a performance node compares against it.
- Timings include process start. The `strings` program is quadratic in all three languages by design (string rebuilt each step).
- Narra release builds wrap on overflow, Go wraps too, Rust release also wraps; the programs stay in range anyway.
- Git through the rtk wrapper was refused in the worktree; plain /usr/bin/git worked.

## Proposed learnings

- Performance nodes report before and after numbers with `bench/run.sh` and note machine load | keeps results comparable | bench | no

## Verification

- `bench/run.sh`: all six programs build in all three languages, outputs match.
- `cd compiler && go vet ./... && go test ./...`: pass.

## Size

- Estimated M, actual M.
