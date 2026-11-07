# Outcome: linux-check

## Built

- `scripts/linux-check.sh`: builds the image, mounts the repo read-only at `/repo`, keeps Go caches in the named volume `narra-linux-check-cache`, runs `go vet ./... && go test -timeout 45m ./...` (extra arguments go to `go test`, for example `-run TestGuide`). Exits non-zero on failure. Prints how to start colima when Docker is not running (exit 2).
- `scripts/linux/Dockerfile`: `golang:1.23-bookworm` plus gcc. The container runs with `SYS_PTRACE` and unconfined seccomp so LeakSanitizer works.
- `compiler/README.md` names the script under Tests.

## Deviations

- Two files outside the owned paths changed because the Linux run needed them: `compiler/internal/narra/std_string.go` (the `Str.split` temp type) and `selfhost/src/emitstd.nra` (same fix, needed for the byte-for-byte emit tests). `selfhost/src/runtimetext.nra` was regenerated from `runtime.h` with `NARRA_UPDATE=1`.

## Gotchas

- Linux failures found by the first full run (33 minutes):
  1. `CLOCK_MONOTONIC` undeclared in the `narra test` runner C: `-std=c11` hides POSIX names on glibc. Fixed with `#define _POSIX_C_SOURCE 200809L` at the top of `runtime.h`.
  2. `Str.split` crashed in release builds with gcc -O2: the emitter typed the temp `lstr***` but `narra_str_split` returns `lstr**`. The wrong type let gcc reorder the read ahead of the stores (strict aliasing). Fixed to `lstr**` in the Go emitter and the selfhost emitter. This was undefined behavior on macOS too, hidden by clang.
  3. Selfhost emit, bootstrap and runtime-text tests failed after fix 1 until `runtimetext.nra` was regenerated.
- The `fgetc` stdin loop from PR #91 needed no change. LeakSanitizer found no leaks.
- Container tests run slower than macOS under colima's 4 CPUs and 5 GiB: 33 minutes when cold with failures, 10.7 minutes for the final full run.
- colima mounts only the home directory, so scratch mounts must live under it.
- Only arm64 gcc was tried. clang and x86_64 were not.

## Proposed learnings

- Generated C temps must use the exact C type the helper returns | a wrong pointer depth is hidden by `-w` and breaks under gcc -O2 | compiler/internal/narra | no

## Verification

- `scripts/linux-check.sh`: PASS in 640s (final run).
- macOS `cd compiler && go vet ./... && go test -timeout 45m ./...`: PASS in 522s (after the last change).

## Size

- Estimated S, actual S. Three real portability bugs, all small.
