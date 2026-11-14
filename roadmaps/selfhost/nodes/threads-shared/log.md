# Log: threads-shared

## Proof first

- base: cd734c3e (feat/batch-13), release compiler from `sh selfhost/bootstrap.sh /private/tmp/claude-501/threads-shared/base`.
- `narrac run testdata/shared_counter.nra`: exit 1, `unknown module "std/shared"`.

## Design notes

- `Shared<T>` is `{ sync: Sync, slot: Box<Option<T>> }`, as the threads-channels outcome proposed. `get` needed a primitive that retains under the lock: a plain `Threads.result` path is borrowed, and the retain would come at the `return`, after the unlock. `Threads.peek` writes the retain next to the read.
- `Atomic<T>` is `{ slot: Box<T> }`. The five primitives are emitted inline as `__atomic_*_n` builtins on `box->f_value`; no C was needed in `std_thread.c`.
- In place `update`: the callback's parameter is borrowed, so a callback cannot write it without `copy(...)`, and `copy` retains it (count 2), so the push copies once. The copy is explained, so the copy check reports nothing, but it is a real copy. Recorded as a gotcha and a question for the orchestrator.
- `Shared.read(team, (current) => current.scores)` failed with E0976 ("which its function type does not allow"): an arrow passed to a generic function did not take the `from` of its slot. Reproduced with a user generic function (`/private/tmp/claude-501/threads-shared/probe/p2.nra`), so it is a checker gap, not std/shared. Fixed with `ClosureInfo.hintShared` (check.nra, closures.nra).
- E1411 inside `std/atomic` instances: the instance body of `create<f64>` resolved `Atomic<T>` with T bound, and reported at `std/src/atomic.nra`. The type check now skips the `std/atomic` scope; the call site reports instead.
- Codes: E1410, E1411, and E1412 for `Atomic.add` on `Atomic<bool>` (an `__atomic_fetch_add` on `bool` is a C error).

## Runs

- `/private/tmp/claude-501/threads-shared/samples.sh`: 6 samples, `-sanitize` and `-release`, leak and copy checks on: all ok.
- `tsan.sh` on the 6 samples: all ok.
- `repeat.sh`: 5 samples, release, 30 runs each: 0 failures.
- `compare.sh`: `emit selfhost` and the 7 bench programs, base compiler against branch compiler: identical.
