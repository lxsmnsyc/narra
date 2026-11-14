# Outcome: threads-channels

## Built

- **`std/channel`** (`std/src/channel.nra`, Narra source; registry row in `THREAD_FUNCS`, `selfhost/src/stdthread.nra`): `Channel.create<T>(capacity): #[Sender<T>, Receiver<T>] uses alloc`, `send(sender, value): bool uses io`, `receive(receiver): Option<T> uses io`, `values(receiver): Generator<T> uses io` and `close(sender)`. `Sender<T>` and `Receiver<T>` are opaque structs.
  - The queue is a ring of `capacity` boxes (`Array<Box<Option<T>>>`, a capacity below 1 is 1) and a `Box<QueueState>` with the head, the count and the closed flag. Under the lock, a send writes one slot and the state box in place with `CellSlot.store`, so the ring is never copied.
  - `receive` moves the value out of its slot with `Threads.take`, so the result is fresh (no `from`), and a write to it does not hold the receiver.
  - Each end holds the queue and an end object of its own (`Threads.end`). Copies of a `Sender` share one end object, and copies of a `Receiver` share another. When the last copy of an end goes, its end count drops and every waiter wakes up. With no sender left, `receive` gives `None` once the queue is empty. With no receiver left, `send` gives `false`.
  - `send` marks the value shared (`Threads.share`) before it enters the queue.
- **`Sync`**, a builtin non-generic type of `std/internal/thread` (`Type.TSync`, named `Sync` in modules that import `std/internal/thread`, C `narra_sync*`). It is Sendable and has no `==`. Its count starts at -1 (shared), so copies on any thread retain and release it atomically, and the share walk stops at it (`narra_sync_share` does nothing). Rows are in `INTERNAL_THREAD_FUNCS`, and the C is in `selfhost/runtime/std_thread.c`. A Sync is one of two kinds:
  - A core (`narra_sync_core`) has a mutex, two condition variables (0 and 1) and two end counts.
  - An end is only the header (count, core, kind).
- **Primitives of `std/internal/thread`**, for threads-shared to reuse:
  - `Threads.sync(): Sync uses alloc` gives a new core.
  - `Threads.end(sync, kind: int): Sync uses alloc` gives an end of kind 0 or 1. It adds one to the core's count of that kind. When its last copy is dropped, it takes the lock, lowers the count, wakes both conditions and releases the core.
  - The lock calls take a core: `Threads.lock(sync)`, `Threads.unlock(sync)`, `Threads.wait(sync, condition) uses io`, `Threads.wakeOne(sync, condition)`, `Threads.wakeAll(sync, condition)`, and `Threads.ends(sync, kind): int` (read it with the lock held).
  - `Threads.share(value)` marks any value shared in thread mode (`shareCode`).
  - `Threads.take(box: Box<Option<T>>): Option<T>` moves the value out and leaves `None`, with no copy. When the value is an array, a box or a map with a count of -1 (one holder, the result), its count becomes 1, so a write to the top object needs no copy; its parts stay shared. `Shared.update` can use it in place of a copy.
  - `Threads.result(box: Box<Option<T>>): T` gives the value of a `Some` (it panics on `None`), and nothing when `T` is void. `Thread.join` uses it.
- **Void thread functions.** `Thread.start(scope, () => { ... })` with a function that returns nothing gives a `Thread<void>`, and `Thread.join` on it returns nothing, as for `Async<void>`.
  - Inference: `bindThreadVoid` in `asyncvoid.nra`, called from `checkGenericCall` in `calls.nra`. It binds `T` of a `std/thread` function to void from `() => T` given `() => void`, and from `Thread<T>` given `Thread<void>`.
  - Instances of `std/thread` with a void `T`:
    - An enum field of type void has no C member, and constructing it runs its value as a statement (`buildVariantMembers` and `enumAllocation` in `emitdecls.nra`), which covers `Some(work())`.
    - A `return` of a void value is allowed in a void instance (`isVoidInstance`, in `stmts.nra`). `emitReturn` in `emitstmts.nra` writes it as the expression followed by `return;`.
  - `void` still cannot be written as a type argument, so `Thread.Thread<void>` cannot be named.
- **Checks** (`selfhost/src/threadcheck.nra`):
  - A call of `Channel.create` checks its type argument with `checkSendable` (E1190, site "the values of Channel.create").
  - `Sender` and `Receiver` are rows of `THREAD_HANDLES` with `sendable: true`, so they get E1195 on `==`, `!=` and map keys. Each row now carries its own "compare" hint (`describeCompare`), so a Sender says "compare the values sent through the channel".
  - `isThreadModule` includes `std/channel`.
- **Type sites of `TSync`**:
  - `types.nra`: the variant and its key.
  - `check.nra`: `typeString` and `isCountedType`.
  - `cnames.nra`.
  - `emitstate.nra`: `refCountCode` writes `narra_sync_retain`, `narra_sync_release` and `narra_sync_share`.
  - `exprs.nra`: `hasNoFields`.
  - `traits.nra`: no std trait.
  - `stdtypes.nra`: `"Sync"` in row texts.
  - `scopes.nra`: the name, as for `Handle`.
- **Tests:**
  - Samples, each with a `.out` file: `testdata/channel_basic` (the SPEC sample, with its void thread function), `channel_full_blocks`, `channel_close_drop`, `channel_receivers_gone`, `channel_many_receivers`, `channel_reply`, `channel_capacity_zero`, `channel_receive_write` (a received array pushed in place, with no copy reported), and `threads_void` (join of a `Thread<void>`, twice, and an expression body).
  - Leak test: `testdata/leaks/channel_cycle`.
  - Error tests: `testdata/errors/channel_cell_element` (E1190) and `channel_equal` (E1195).
- **Docs.** SPEC section 8 "Channels" keeps its original sample and loses its "Not built yet" note. Its `spec-skip.txt` line stays, with a new reason: the sample uses `scope`, a `Thread.scope` parameter that it does not define, like the Threads sample. `compiler/README.md`, Threads section:
  - New lines for `std/channel` and for void thread functions.
  - `std/channel` is listed among the modules that turn on the thread checks.
  - E1195 lists `Sender` and `Receiver`.
  - "Not supported yet" names only `Shared` and `Atomic`.

## Proof

- base: cb39b8b6 (feat/batch-13), release compiler from `sh selfhost/bootstrap.sh`. `narrac run testdata/channel_basic.nra`: exit 1, `unknown module "std/channel"`.
- branch: the same command prints 0 to 99 and exits 0. It is clean with `-sanitize` and with `-release`, both with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`.
- Review probe `p_mutate3` (`/private/tmp/claude-501/rv-probes`), run with `-sanitize` and the leak and copy checks:
  - Before: `receive` was `from receiver`, so the received array could not be written while the receiver was still used.
  - After the review fixes: it prints `6` and `1` and exits 0, with no copy reported.

## Deviations

- **No builtin `SyncSlot<T>`** (accepted at review). The builtin is the non-generic `Sync`, and a guarded value lives in a `Box<T>` beside it. For threads-shared, `Shared<T>` can be `{ sync: Sync, slot: Box<Option<T>> }`, using `Threads.take` and `CellSlot.store`.
- **Files outside the owned and shared lists:**
  - The `TSync` sites: `types.nra`, `check.nra`, `cnames.nra`, `emitstate.nra`, `exprs.nra`, `traits.nra`, `stdtypes.nra` and `scopes.nra`.
  - `calls.nra`: `checkThreadCall` takes the call position and the result type, and calls `bindThreadVoid`.
  - The void thread functions: `asyncvoid.nra`, `stmts.nra`, `emitdecls.nra` and `emitstmts.nra`.
  - `std/src/thread.nra`: `join` now uses `Threads.result`, plus a comment on `start`.
  - `testdata/threads_void.nra`.
- **Goldens changed**: `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` now list `std/channel`.
- The SPEC Channels sample still does not check, because of its undefined `scope`; its skip line has a new reason.
- No codes of E1400-E1409 are used.

## Stubs left

- None.

## Gotchas

- `Sync` only works in thread mode: its count starts at -1, and without `NARRA_THREADS` a release would never reach 0. Only modules that import `std/internal/thread` can name it, and importing it turns thread mode on.
- `Threads.share` on a borrowed parameter marks the caller's value shared in place. Later writes to it copy first, as with any shared value.
- A void `T` reaches instance bodies only through `std/thread`. Another std module that wants void instances needs its own case in `bindThreadVoid`, and its body may hold the void value only in an enum field, a function result or a `return`. A void local, a pattern binding of a void field and a void struct field are not written.
- From the review, not fixed here:
  - (5) A send on a full channel that only the same thread could drain, or a receive on an empty one that only it could fill, blocks forever, and nothing reports it.
  - (6) Every thread program carries the `narra_sync` functions of `std_thread.c` (about 80 lines of statics), whether or not it uses them.
- A cycle through a queued `Sender` leaks the core, the ends and the boxes. The leak check reports it.
- E1190 at `Channel.create` is reported at the call's argument list, not at the type argument.

## Proposed learnings

- Keep a builtin that only guards memory non-generic, and keep the value in a `Box<T>` beside it | a generic builtin type touches every type switch, while a leaf type needs about ten match arms | selfhost/src | no

## Verification

- Samples (`/private/tmp/claude-501/threads-channels/try.sh`): all 9 (the 8 `channel_*` and `threads_void`) are ok with `-release` and with `-sanitize`, both with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`. Each release binary was also run 50 times (`repeat.sh`), with 0 failures.
- `NARRAC=<branch> sh roadmaps/selfhost/nodes/threads/tsan.sh` on the 9 samples: all ok, with no report.
- Leak test: `narrac run -sanitize testdata/leaks/channel_cycle.nra` with the leak check prints `sent`, `leak check: 6 allocations still live`, and exits 102.
- Emit comparison (`compare2.sh`): the base compiler was built from feat/batch-13 dc5547ae (`git archive` plus `bootstrap.sh`). `narrac emit selfhost` from it and from the branch compiler is byte-identical, and so are the seven `bench/*/*.nra`.
- Full check, rebased on feat/batch-13 dc5547ae: `sh scripts/test.sh` gives `test.sh: PASS, 9536 cases in 294s`.
- `sh scripts/linux-check.sh`: `test.sh: PASS, 9536 cases in 167s`, `linux-check: PASS in 642s`.

## Size

- Estimated M, actual M+. The review added the take primitive and void thread functions.
