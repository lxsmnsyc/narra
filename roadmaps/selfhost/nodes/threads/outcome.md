# Outcome: threads

## Built

- **Thread mode** (D111 item 2). `Emitter.threads` (`selfhost/src/emitstate.nra`) is set when `checker.used` holds `std/internal/thread`. `emit.nra` then writes `#define NARRA_THREADS 1` after the `FP_CONTRACT` pragma, and `driver.nra` adds `-pthread`. Programs without it emit the C they did before, apart from the runtime text and the immortal count (see Verification).
- **Counts** (`selfhost/runtime/runtime.h`). `NARRA_RC_RETAIN(rc)`, `NARRA_RC_RELEASE(rc)` (true for the last holder) and `NARRA_RC_UNSHARE(rc)` (true when a write must copy first) are the plain operations without `NARRA_THREADS`. In thread mode a positive count is plain (relaxed load and store), a negative count changes with `__atomic` operations (acq_rel on the release), and 0 is immortal. `narra_rc_share(&rc)` negates a positive count and returns true, so the caller goes on to the children. It returns false at 0 or a negative count, which stops the walk.
  - Runtime users: strings (`lstr_retain`/`lstr_release`; `lstr_share` in thread mode), generators (`narra_gen_*`, and the copy in `narra_gen_next`/`narra_co_next` now releases the old frame through `narra_gen_release`), maps (`runtime_map.h`: retain, release, unique), string builders (`std_string.c`; `narra_sb_share`).
  - Emitted users, in thread mode only (the plain text otherwise): array retain, release, unique and copy, box retain, unique and release, heap enum retain and release (`emitdecls.nra`), and closure environments (`ENV_RUNTIME_THREADS` in `emitclosures.nra`: the header gets a `share` field). In thread mode a copy on write releases the old object through its release helper instead of `rc--`, because another holder on another thread may release at the same time.
  - Immortal is 0 in both modes: string literals (`emitstate.nra` `literal`) and stack environments (`.h = {0, NULL}`). The `rc >= 0` guards became `rc != 0`, and `rc > 1` became `NARRA_RC_UNSHARE` (`rc != 1`). No other site read -1 as a marker (grep of `selfhost/src/emit*`, `std*` and `selfhost/runtime`).
- **Share walkers.** `shareCode(emitter, type, value)` (`emitstate.nra`) gives the share call of a type, or "" for types that cannot cross (generators, `Async`, `Handle`, `Dynamic`). The new helper operation `"share"` exists for every counted struct, inline enum and fixed array (`typeRcFunction`), heap enums (`buildShareFunction`), arrays and boxes (`arrayThreadFunction`, `boxThreadFunction`) and maps (`mapShareFunction`, a new helper slot for `TMap`), with prototypes. Nothing names a share helper outside thread mode, so none is emitted there. Each heap closure environment gets `S<owner>_<n>`, which shares its by-value captures; `narra_env_share` calls it.
- **Heaps** (D111 item 3, `runtime.h`). `narra_heap` gains `live`, `number`, `remote` and `next_idle` in thread mode. The main thread uses the static heap `narra_small` (number 0) and reads no thread-local while `narra_threads_running` is 0 (relaxed load). Started threads take a heap with `narra_heap_enter` (an idle one, or a new `calloc`, at most 4096 in `narra_heaps`) and give it back with `narra_heap_leave`. The header is `class | number << 8`. A free of another heap's block pushes onto `home->remote` with one CAS. `narra_block_new` collects the remote stack with one exchange when a class list is empty. `narra_live` is `narra_heap_here()->live` in thread mode, and `narra_live_total()` adds the heaps for the leak report. Under AddressSanitizer (`NARRA_PLAIN_MALLOC`) thread mode keeps `malloc` and only the per-heap live counts. Without `NARRA_THREADS` the block functions compile to the code they were before (`heap` is a macro for `&narra_small`).
- **Constants** (D61, `emitconsts.nra`). In thread mode the getter tests `__atomic_load_n(&G_ready, __ATOMIC_ACQUIRE)`, then takes `narra_const_lock()` (one mutex that the thread holding it may take again, since a constant can read another) and tests again. It builds the value, shares it and stores the flag with release.
- **Panics.** `narra_panic` takes `narra_panic_lock` (except inside a guarded coroutine resume), flushes, writes, and calls `_exit(101)` while `narra_threads_running > 0`. `narra_panic_jmp`, `narra_panic_text`, `narra_co_panicked` and the copy check's `narra_copy_granting` are `NARRA_PER_THREAD` (`_Thread_local` in thread mode only). The copy-check hooks take a per-thread-recursive lock (`narra_copy_lock`), which is a no-op outside thread mode.
- **std/thread** (`std/src/thread.nra`): `scope(body)`, `start(threadScope, escaping work)`, `join(thread)`, plus opaque `Scope` and `Thread<T>`. It sits over `std/internal/thread` (`selfhost/src/stdthread.nra`: `Threads.start(run: () => void): int`, `Threads.join(id: int)`) and `selfhost/runtime/std_thread.c` (`narra_thread_start`, `narra_thread_join`). Registered in `stdlib.nra`, `stdcalls.nra` (`checkThreads`), `emitstd.nra` (`Threads.`) and `runtimetext.nra`. `narra_thread_start` shares the environment, then creates a pthread with an 8 MiB stack. The thread takes a heap, runs the function and runs `narra_loop_exit_hook`, which `std_async_loop.c` sets to `narra_async_exit` when any task starts (the D109 exit rule). It then releases the environment, frees its drop worklist and gives the heap back.
- Samples: `testdata/threads_scope`, `threads_join_twice`, `threads_shared_array`, `threads_write_copies`, `threads_alloc_cross`, `threads_constant`, `threads_tasks`, `threads_deep`, and from the review probes `threads_copy_check`, `threads_copy_rounds`, `threads_share_long_list` and `threads_env_args` (each with `.out`), and `testdata/panics/threads_panic.nra` (row in `selfhost/tests/src/panics.nra`).
- `compiler/README.md` has a "Threads" section, and the "Not supported yet" line now names only what the later nodes build. In SPEC, the two "not built yet" notes for this node now name only threads-check. The `spec-skip.txt` reason for the Threads sample is now its undefined `sum`/`firstHalf`. The first line is unchanged.

## Proof

- base: c63ba12e (feat/batch-13). `narrac run testdata/threads_scope.nra` with the base release compiler: exit 1, `unknown module "std/thread"`.
- branch: the same command prints `total 4999950000` and exits 0. It is clean under `-sanitize` with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`, and in `-release`.

## Deviations

- **No `SyncSlot<T>`.** A builtin generic counted type touches `types.nra`, `cnames`, `traits`, `generics` and the dumps, none of them owned here, and `std/thread` does not need it. A thread result is written once before the thread ends and read only after `pthread_join`, which orders the two. So `Thread<T>` is `{ result: Box<Option<T>>, control: Box<ThreadControl> }` and `Scope` is `{ threads: Box<ThreadList> }`, a list newest first, so a start adds a cell instead of copying an array. The result box is shared at `Thread.start` (it is a capture), and the control boxes stay on the starting thread. threads-channels and threads-shared need a mutex and a condition variable next to a value. They will have to add `SyncSlot<T>`, or a non-generic C sync object beside a `Box<T>`.
- Files edited outside the owned and shared lists:
  - `selfhost/src/emitconsts.nra`: the D61 getter is there, not in `emitconstvalues.nra`.
  - `selfhost/src/emit.nra`: one `#define` line.
  - `selfhost/tests/src/panics.nra`: the `threads_panic` row.
  - `selfhost/tests/spec-skip.txt`: the reason text.
  - `testdata/panics/threads_panic.nra`.
  - Generated `runtimemap.nra`, `runtimestdstring.nra` and `runtimestdasyncloop.nra`, from their edited sources.
- `emitdynamic.nra` is unchanged: `Dynamic` is not `Sendable`, so its counts never go negative. Task frames (`std_async.c`) and socket handles are unchanged for the same reason.
- `runtime_map.h` is a per-map template, so its count lines differ in the program part of every C file with a map, not only in the header. The code is equivalent without `NARRA_THREADS`.
- `selfhost/runtime/std_env.c` (review nit): on Linux `narra_env_load` reads `/proc/self/cmdline` once under `pthread_once` in thread mode.
- `roadmaps/selfhost/nodes/threads/tsan.sh` builds the thread samples with ThreadSanitizer. It is not a suite: TSan cannot be combined with AddressSanitizer, and the build is slow.
- `selfhost/runtime/closure_env.c` (generated into `runtimeclosureenv.nra`) is not used by the emitter, which has its own `ENV_RUNTIME`. It still says -1 and was left alone.
- `threads_alloc_cross` takes an argument that doubles its rounds. The bounded memory check is a manual peak RSS comparison (Verification), not a case of the suite, because a program cannot read its live count.
- In `threads_tasks` the thread's task is cancelled by the exit rule (D109 item 6) when the thread function returns. Join then waits for the cancelled task to finish; a timer does not keep the thread alive.

## Stubs left

- None. E1190 to E1195 and `std/parallel` are threads-check's.

## Gotchas

- `Threads.start` takes any `() => void`. A capture by reference (a `let` or a `mutable` parameter) is not shared and would race; threads-check's E0612 and E1193 rules keep thread functions to `const` captures.
- A `Cell` that crosses a thread would become shared, and `CellSlot.store` writes in place whatever the count. That is a race, which E1190 (threads-check) rules out.
- At most 4096 heaps exist at once (`NARRA_HEAP_MAX`). Ended threads give theirs back, so the limit is on threads alive at the same time.
- Thread mode emits `__atomic` builtins: gcc and clang, not TCC.
- threads-process's gotcha (a signal during the last `detach`) is still open.

## Proposed learnings

- Thread mode text is chosen by `emitter.threads` at each emitted count site, never by a macro in non-thread output | keeps every program without threads byte-identical | selfhost/src/emit* | yes (emit comparison)

## Verification

- Samples: `/private/tmp/claude-501/threads/samples.sh` runs each `testdata/threads_*.nra` with `-sanitize` and with `-release`, both with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1`, and diffs against the `.out`. All 12 samples are ok in both builds.
- Review fixes, base 7832f8fd compiler against the branch. `threads_copy_check` (debug, copy check on) failed 1 of 30 runs on base and passes 30 of 30 on the branch. `threads_copy_rounds` failed 10 of 10 (exit 103) on base and passes 10 of 10. `threads_share_long_list` aborts under `-sanitize` on base (exit 134, stack overflow in the share walk) and passes on the branch. `NARRAC=... sh roadmaps/selfhost/nodes/threads/tsan.sh`: base reports a data race on `threads_shared_array` and `threads_copy_rounds`; the branch passes all 12 samples. `--suite panics`: PASS, 6 cases (threads_panic exits 101 with its message).
- Remote frees in a release build (block heap, not malloc): `threads_alloc_cross` built `-release` gives a peak RSS of 2359296 bytes for 40 rounds and 2359296 bytes for 80 rounds (`/usr/bin/time -l`). Both runs pass the leak check.
- Emit comparison (`/private/tmp/claude-501/threads/compare.sh`): `narrac emit selfhost` and the six `bench/*/*.nra`, base compiler against branch compiler. After mapping the immortal `{0,` / `.h = {0, NULL}` back to -1, the only differences are:
  - the runtime header text (`runtime.h`, before `// Types`);
  - the per-map `runtime_map.h` template: retain, release and unique through the macros, 9 lines per map, 41 maps in selfhost;
  - the four `ENV_RUNTIME` lines (`>= 0` to `!= 0`, and the comment).
  - Bench programs: none in the program part. The non-thread emitted helpers are the same text.
- Instructions retired by `narrac emit selfhost` (`/usr/bin/time -l`, 5 runs each, alternating, load average 35 to 80 from other workers): base median 3,311M, branch median 3,258M (-1.6%, within noise; the spread is about 270M).
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C (225278 lines). The seed is older than the source, as expected before the batch's seed update.
- `--suite spec`: PASS, 61. `--suite sources`: PASS. `--suite codes`: PASS. `--suite generated`: PASS, 21.
- After the review fixes and the rebase on feat/batch-13 (560ea4cb): `sh scripts/test.sh` PASS, 9110 cases. The `unknown_module` goldens now list `std/thread`. The share worklist variables are named `narra_shares*`, so the `_share_` count of the `const-generics/distinct-instances` emit check is unchanged.
- `sh scripts/linux-check.sh`: `test.sh: PASS, 9110 cases`, `linux-check: PASS in 1574s`.

## Size

- Estimated L, actual L.
