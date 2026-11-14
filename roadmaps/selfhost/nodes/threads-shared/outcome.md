# Outcome: threads-shared

## Built

- **`std/shared`** (`std/src/shared.nra`, Narra source; registry row in `THREAD_FUNCS`, `selfhost/src/stdthread.nra`). `Shared<T>` is an opaque struct `{ sync: Sync, slot: Box<Option<T>> }`, and the slot holds `Some(value)`. Each call takes the lock and releases it before it returns.
  - `create(value): Shared<T> uses alloc` and `set(shared, value)` mark the value shared (`Threads.share`) before they store it.
  - `get(shared): T from shared` reads the value with the new primitive `Threads.peek`, which retains it while the lock is held. It is `from shared`, like `Cells.get`.
  - `update(shared, change: (value: T) => T uses E) uses E` moves the value out with `Threads.take`, which turns a top count of -1 (one holder) into 1. It then calls `change`, marks the result shared and stores it.
  - `read(shared, look: (value: T) => U from value uses E): U from shared uses E` calls `look` on the value in the slot. The result of `look` holds its own reference, taken before the unlock.
- **`std/atomic`** (`std/src/atomic.nra`). `Atomic<T>` is an opaque struct `{ slot: Box<T> }`, with `create`, `load`, `store`, `exchange`, `add` (gives the old value) and `compareExchange(atomic, expected, replacement): bool`. Each function calls one new primitive of `std/internal/thread`: `Threads.atomicLoad`, `atomicStore`, `atomicExchange`, `atomicAdd` or `atomicCompareExchange`.
  - The emitter (`emitAtomicPrimitive`, `stdthread.nra`) writes each primitive inline as a `__atomic_*` builtin on `box->f_value`, with `__ATOMIC_SEQ_CST`. Signed adds wrap in two's complement, as D92 requires. `std_thread.c` is unchanged.
- **New primitives:**
  - `Threads.peek(box: Box<Option<T>>): T` gives an owned copy of the value, with the retain written at the read.
  - `Threads.syncChecked(): Sync uses alloc` gives a core whose mutex is `PTHREAD_MUTEX_ERRORCHECK` (`narra_sync_new_checked`).
  - `Threads.enter(sync): bool` locks such a core, or gives true without the lock when its own thread holds it (`narra_sync_enter`, `EDEADLK`).
- **Re-entry stop** (orchestrator ruling, after the D109 precedent). Every `Shared` call takes its lock through `enter`. A call that finds its own thread holding the lock, from a callback of `update` or `read`, panics with `Shared.<call>: this thread already holds the lock of this Shared; a callback of Shared.update or Shared.read reached the same Shared` (exit code 101). The message texts are constants, so no call allocates. It is documented in SPEC section 8 and in a new "Built" note under D111, which also says that `Atomic<T>` with a type parameter is E1411 by design.
- **`Threads.take` ordering** (review fix): it reads the count with `__ATOMIC_ACQUIRE` before it turns -1 into 1, so the read pairs with the acq_rel release of a copy dropped on another thread. The store stays relaxed: the taker is then the only holder. This also covers `Channel.receive`, which uses the same primitive.
- **Checks** (`selfhost/src/threadcheck.nra`, through `checkThreadCall` → `checkSharedCall`):
  - E1190 on `T` of `Shared.create` (site "the value of Shared.create").
  - E1410 for an async callback of `Shared.update` or `Shared.read`: an async arrow, an async function name, or any function whose result is an `Async`.
  - E1411 for `Atomic<T>` with `T` other than `bool`, `int`, `i8` to `i32` or `u8` to `u64`. It is checked at the result of `Atomic.create`, and at every written type through `checkAtomicType`, which `resolve.nra` calls at its two `noteTypeBounds` sites. A type parameter is E1411 too, outside the `std/atomic` scope.
  - E1412 for `Atomic.add` on an `Atomic<bool>`.
  - `Shared` and `Atomic` are rows of `THREAD_HANDLES` (`sendable: true`), so E1195 covers `==`, `!=` and map keys on them. `isThreadModule` includes both modules.
- **Checker fix**: an arrow passed to a generic function now takes the `from` of its slot's function type (`ClosureInfo.hintShared` in `check.nra`, set in `closures.nra`). Before, `Shared.read(team, (current) => current.scores)` was E0976, and so was the same call to any user generic function.
- **Tests:**
  - Samples, each with a `.out` file: `shared_counter`, `shared_array_in_place`, `shared_read_parts`, `shared_mixed` (four threads mixing update, read, get and set, the review's `c_mixed` probe), `atomic_counter`, `atomic_compare_exchange` and `atomic_wrap`.
  - Exit sample: `testdata/exits/shared_reentered` (a `Shared.get` inside `Shared.read` of the same Shared; prints `before`, exit code 101).
  - Leak test: `testdata/leaks/shared_cycle` (leaks 2 allocations, exit 102).
  - Error tests: `shared_async_callback` (E1410), `atomic_float` (E1411), `atomic_add_bool` (E1412), `shared_cell_value` (E1190) and `shared_equal` (E1195).
- **Docs:**
  - SPEC "Shared state" no longer has its "Not built yet" note. It now states the `from` of `get` and `read`, and E1412.
  - The SPEC sample checks, and its `spec-skip.txt` line is removed.
  - `compiler/README.md` has entries for `std/shared`, `std/atomic` and the `from` fix. "Not supported yet" no longer lists Shared and Atomic. The identity-type line under "Differences from the spec" is corrected.

## Proof

- base: cd734c3e (feat/batch-13), release compiler. `narrac run testdata/shared_counter.nra`: exit 1, `unknown module "std/shared"`.
- branch: the same command prints `8 threads, total 80000` and exits 0. It is clean with `-sanitize` and `-release`, both with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`.

## Deviations

- **The in-place `update` acceptance is met only to the letter.** The callback's parameter is borrowed, so it cannot be written without `copy(...)`. `copy` retains the value (count 2), so the first push copies the array once per update. The copy check reports nothing, because `copy(...)` explains the copy, but a copy is made. `Threads.take` still makes the top object plain, so no hidden copy comes from the sharing. A real in-place update needs the language to hand ownership to a callback parameter, which is outside this node.
- **Files outside the owned and shared lists:**
  - `selfhost/src/resolve.nra`: the import and two `checkAtomicType` calls.
  - `selfhost/src/check.nra` and `closures.nra`: the `hintShared` fix.
  - `selfhost/tests/spec-skip.txt`: the Shared line removed, as the brief asks.
  - Goldens: `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` now list `std/shared` and `std/atomic`.
  - Error test `testdata/errors/atomic_add_bool.*`, which is in the owned `atomic_*` pattern.
- **E1412** is used for `Atomic.add` on `bool`, from the reserved range. The brief named only E1410 and E1411.
- The atomics are inline builtins. `std_thread.c` (generated into `runtimestdthread.nra`) changes only for the checked core and `narra_sync_enter`, and now includes `errno.h`.
- `DECISIONS.md`: a "Built" note under D111, as the orchestrator asked.
- In `shared_async_callback` the type error "expected (int) => int, found (int) => Async<int>" and the D109 "Async handle is not used" are reported next to E1410.

## Stubs left

- None.

## Gotchas

- `Shared.get` and `Shared.read` are `from shared`. A write to their result while the `Shared` is in use is E0978, so take a `copy(...)`.
- A generic call does not adapt an integer literal to a bound `T`: `Atomic.add(small, 1)` with an `Atomic<u8>` is "expected u8, found int". Bind `const one: u8 = 1` first. This was not changed here.
- `Shared` and `Atomic` pass the contents-type check of a `mutable` parameter, because they are opaque structs. SPEC says that is an error; it is recorded in README "Differences from the spec".
- A `change` or `look` that calls the same `Shared` stops the program with exit code 101. Two threads that take two `Shared` locks in opposite orders still deadlock, and nothing reports it.
- Review notes, not changed:
  - (3) `update` copies the value once per update when the callback writes it. This is logged in the inbox.
  - (4) A type such as `Shared<Cell<int>>` written as a field type is caught only at `Shared.create` (E1190), not where the type is written.
- `Threads.result` still panics with "Thread.join found no result" when it finds `None`. `Shared.read` uses it too, but the slot is only empty while `update` holds the lock.

## Proposed learnings

- A std primitive that returns a value read under a lock must retain it in its own emitted code | a borrowed path is retained at the `return`, after the unlock | selfhost/src/std*.nra | yes (TSan on a get racing a set)

## Verification

- Samples (`/private/tmp/claude-501/threads-shared/samples.sh`): all 7 are ok with `-sanitize` and with `-release`, both with the leak and copy checks.
- `NARRAC=<branch> sh roadmaps/selfhost/nodes/threads/tsan.sh` on the 7 samples, plus the `threads_*` and `channel_*` samples: all ok.
- `tsan200.sh`: `shared_mixed` built with TSan and run 200 times. With the take load put back to relaxed in the C, 186 of 200 runs fail or report. With the fix, 0 of 200.
- `repeat.sh`: 5 samples in release, 30 runs each, 0 failures.
- `narrac run -sanitize testdata/leaks/shared_cycle.nra` with the leak check prints `linked`, `leak check: 2 allocations still live`, and exits 102.
- `--suite spec`: PASS, 65. `--suite generated`: PASS, 21. `--filter shared_`: PASS, 305. `--filter atomic_`: PASS, 25.
- Emit comparison (`compare.sh`): `narrac emit selfhost` from the base compiler (cd734c3e) and from the branch compiler, on the same source, is byte-identical (246348 lines after the review fixes), and so are the 7 `bench/*/*.nra`.
- After the review fixes, rebased on feat/batch-13 72e82ca9: `sh scripts/test.sh` gives `test.sh: PASS, 9604 cases in 132s`.
- `sh scripts/linux-check.sh`: `test.sh: PASS, 9604 cases in 1251s`, `linux-check: PASS in 1543s`.

## Size

- Estimated M, actual M.
