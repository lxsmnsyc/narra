# Log: threads-check

## Proof first

- Base: afaaff2e (feat/batch-13), compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/threads-check/base` (release flags).
- `base/narrac check testdata/errors/threads_cell_capture.nra`: exit 0. `base/narrac run` on it prints `1`, exit 0: a `Cell` captured by a thread function compiles and runs.
- Branch: `branch/narrac check testdata/errors/threads_cell_capture.nra`: exit 1, `threads_cell_capture.nra:9:40: capture "counter" of the thread function: Cell<int> is not Sendable; keep state that threads share in a Shared from std/shared` (E1190).

## Notes

- Nothing of the node existed on the base (no `Sendable`, no E1190-E1195 rows, no `std/parallel`).
- `TraitInfo.std` means "derivable" (`isDerivable`), so the prelude's `Sendable` has `std: false`. The checks key on module "" and the name instead (`isSendableTrait`); `stdTrait("Sendable")` would give -1.
- A `uses E` function type as the last parameter must not have a trailing comma: the parser reads `, )` as another effect name (D117 fixes this). `std/src/parallel.nra` leaves the comma out, as `std/src/thread.nra` does.
- `sysconf(_SC_NPROCESSORS_ONLN)` is hidden on macOS under `_POSIX_C_SOURCE`; `std_thread.c` defines it as 58 there (the Darwin value).
- Adding `trait Sendable {}` to the prelude shifts node ids, but `narrac emit selfhost` from the base and the branch compiler is byte-identical (cmp), so ids do not reach the C.
