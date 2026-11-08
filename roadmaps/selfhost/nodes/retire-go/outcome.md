# Outcome: retire-go

## Built

- The Go compiler and Go test harness are gone: `compiler/cmd`, `compiler/internal` (107 files, 96 tests) and `compiler/go.mod`. `git ls-files '*.go'` lists only the six files in `bench/`.
- `selfhost/runtime/`: the C runtime, one file per piece (`runtime.h`, `runtime_map.h`, `closure_env.c`, `test_runner.c`, and `std_<name>.c` for std/<name>: env, fs, int, io, path, process, string, test). Taken from the Go strings, byte for byte.
- Runner suite `generated` (`selfhost/tests/src/generated.nra`, 2 cases): `selfhost/src/runtimetext.nra` and `stdtext.nra` must equal what `selfhost/runtime/` and `std/src/` produce. `sh scripts/test.sh --suite generated --update` rewrites them (then run `sh scripts/seed-update.sh`). Replaces TestSelfhostRuntimeText and TestStdSources. Only the header comment of the two files changed.
- Runner suite `codes` (`selfhost/tests/src/codes.nra`, 1 case): Narra port of `check-codes.py`, reading `selfhost/src/*.nra`, no Python. 338 report sites, 0 problems, the same count as the script. Replaces TestEveryFormatHasCode. Checked by breaking two sites (a code that does not fit the message, and E0000): both were reported.
- `testdata/` at the repo root (was `compiler/testdata`); the runner default root follows.
- `scripts/linux-check.sh` and `scripts/linux/Dockerfile`: a Debian gcc image with no Go; runs `sh selfhost/bootstrap.sh` then `sh scripts/test.sh`. `PLATFORM=linux/amd64` selects the architecture and `JOBS` the number of shards.
- `bench/run.sh` builds narrac from the seed (`bootstrap.sh`) and needs only `cc`; Go is still timed when present. Checked with `bash bench/run.sh -n 1 fib`.
- `README.md`, `selfhost/CLAUDE.md`, `compiler/README.md` (usage, bootstrapping, tests, formatting, queries), `docs/guide/01-getting-started.md` and `.gitignore` drop Go. `DECISIONS.md` has D89.
- `seed/narrac.c` regenerated last with `scripts/seed-update.sh` (header: selfhost at f44ddc0).
- Commits: runtime move and new suites, deletion, testdata move, docs and scripts, seed.

## Deviations

- Outside the owned paths: `docs/guide/02..11` and `docs/guide/README.md`. The prose command `narra run|build|check|test` became `narrac ...` (mechanical, no code block changed), so the guide matches the install step of chapter 1.
- `selfhost/src/runtimetext.nra` and `stdtext.nra` (owned) changed only in the header comment. Comments in `selfhost/src` that name Go files are left, as the brief says.
- `compiler/README.md` stays, because GOAL.md points to it for "Differences from the spec". Its "Supported" and "Differences" lists still mention some Go-era file names in a few places.
- D89 is a new decision number, not reserved in the brief.
- Case count is 4389 sharded cases plus the bootstrap case (was 4386 plus bootstrap): +1 codes, +2 generated.

## Stubs left

- None.

## Gotchas

- Edit `selfhost/runtime/*` or `std/src`, never the generated `.nra` files. Then `--suite generated --update` and `sh scripts/seed-update.sh`.
- Docker on colima mounts only the home folder: run `scripts/linux-check.sh` from the worktree path.
- macOS bash 3.2 with `set -u`: expand optional arrays as `${a[@]+"${a[@]}"}`.

## Lines that still name Go (orchestrator updates)

`roadmaps/selfhost/CLAUDE.md`:
- line 18: `cd compiler && go vet ./... && go test ...` -> `sh selfhost/bootstrap.sh && sh scripts/test.sh`.
- line 19: names `TestSourcesFormatted` (now the `sources` suite).
- line 22: "New AST node types need a case in `clone.go`..." (Go files).

`roadmaps/selfhost/GOAL.md`:
- line 15 (Outcome): "The Go bootstrap compiler `narra` first grows..."
- line 19: `cd compiler && go test ./...` passes.
- line 20: output matches the Go compiler's on `compiler/testdata/`.
- line 21: bench Go times stay (Go is a benchmark language).
- lines 32, 33: `compiler/testdata/` and `compiler/testdata/errors/` -> `testdata/`.
- line 35: "Go standard library only in the bootstrap compiler."
- line 40: non-goal about keeping the Go compiler.
- line 44: the fast check command.
- line 58: "what the bootstrap compiler supports".
- Decision lines 97, 99, 102, 103 are history and may stay.

Open briefs that own Go paths (`compiler/internal/narra/...`) need the Go half dropped, as c-seed says.

## Proposed learnings

- Change `selfhost/runtime` or `std/src`, then rewrite the generated text and the seed | the generated files and the seed are checked, not edited | selfhost | yes (generated suite)

## Verification

- macOS arm64, `PATH=/usr/bin:/bin:/usr/sbin:/sbin` (no go): `sh selfhost/bootstrap.sh` ok, "seed is current"; `sh scripts/test.sh`: 4389 + 1 cases, 0 failed, 67 s.
- Linux aarch64 (gcc, colima, `sh scripts/linux-check.sh`): bootstrap ok, seed is current, 4389 + 1 cases pass, 183 s.
- Linux x86_64: COULD NOT RUN. The only x86_64 machine here is a linux/amd64 container emulated on an arm64 Mac, and AddressSanitizer is killed there (exit 137) even for `int main(void){return 0;}` built with `cc -fsanitize=address`. `PLATFORM=linux/amd64 sh scripts/linux-check.sh` ran twice (with 4 and with 2 shards, 28 and 35 min): the bootstrap (stage 0, 1, 2 fixed point) passed on x86_64 with gcc, and every case that builds without the sanitizer passed, but all sanitized cases (215 of 4390) failed with exit 137. That is the emulator, not the compiler. It needs a real x86_64 Linux host (for example a CI runner or the maintainer's machine).
- `git ls-files '*.go'`: the six `bench/*/*.go` files.
- Self-review of the diff (one round): no defects found beyond those fixed while building (bash 3.2 empty array in linux-check.sh, Narra narrowing in the two new suites). Each deleted Go test is replaced, dropped or newly covered per the goldens-programs table.

## Size

- Estimated M, actual M.
