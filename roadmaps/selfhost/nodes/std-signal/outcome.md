# Outcome: std-signal

## Built

`std/signal`: a program waits for SIGINT or SIGTERM as an async event, so a server can stop accepting work and exit cleanly. The default action of both signals stays until the first call of `wait`.

Design: a self-pipe. The first `wait` installs one `sigaction` handler for SIGINT and SIGTERM and opens a nonblocking pipe. The handler only writes one byte (1 for SIGINT, 2 for SIGTERM) and stays installed. The loop waits on the read end with the existing `narra_loop_wait_io`, so `std_async_loop.c` did not change (kqueue, epoll and `WSAPoll` all watch it). A signal that comes while no task waits stays in the pipe, so a second signal during the shutdown reaches the next `wait`.

Windows: `SetConsoleCtrlHandler` maps Ctrl-C to Interrupt and Ctrl-Break to Terminate. The handler thread sends one byte to a loopback socket pair, because `WSAPoll` cannot watch a pipe. Written to match the Windows code of `std_async_io.c`; not compiled or run on Windows.

Wasi: `std/signal` is in `WASI_LEFT_OUT_MODULES`, so a use is E0991.

## Public API

```
import * as Signals from "std/signal"
import { Signal } from "std/signal"

export enum Signal { Interrupt, Terminate }
export async function wait(): Result<Signal, AsyncError> uses io, alloc
```

A cancel gives `Err(AsyncError.Cancelled)`.

## Files

- `std/src/signal.nra`: the module.
- `selfhost/runtime/std_signal.c`, generated `selfhost/src/runtime/runtimestdsignal.nra`, `runtimetext.nra` (one import, one case).
- `selfhost/src/std/stdsignal.nra`: rows for `std/signal` and `std/internal/signal` (`Signals.start`, `take`, `wait`, `message`) and the emitter; wired in `stdlib.nra` and `emitstd.nra`.
- `selfhost/src/check/buildconditions.nra`: `std/signal` in the Wasi left-out list.
- Tests: `testdata/signals/shutdown.nra` and `shutdown.sh` (a `signal_shutdown` case of the exits suite in `selfhost/tests/src/exits.nra`: the harness sends SIGTERM then SIGINT with `kill`, under the sanitizers and the leak check); `testdata/wasi_cases/left_out_signal.*` (case in `emitchecks.nra`).
- Generated and regenerated: `selfhost/src/std/stdtext.nra`, `docs/reference/std/*.md`; the module lists in the error goldens. `scripts/std/generate.nra` has the module row (chapter 14).
- `SPEC.md` (section 8 and the Wasi list), `compiler/README.md`.

## Review fixes

- A failed install restores the saved SIGINT action and closes the pipe; on Windows a failed `SetConsoleCtrlHandler` closes the socket pair.
- `Process.run` and `std/signal` cooperate through a small block both C files define (guarded by `NARRA_SIGNAL_COMMON`): the pipe's write end and the handler in use. The forwarder of `std_process.c` records a SIGINT or SIGTERM in the pipe as well as passing it to the child, so a program that shells out keeps its shutdown signal. When the last `Process.run` call ends, it puts back the handler of `std/signal` instead of the defaults it saved, so a first `wait` on another thread during a call is not undone. Chosen over a shared lock because the two files stay independent runtime parts.
- No deterministic test for the thread race (a first `wait` while another thread's `Process.run` is active): it needs two threads timed against each other. The forward case covers the recording path. A first `wait` that lands in the middle of a call can still replace the forwarder for that one call, so the child misses that signal.
- `pipe2` is hidden under `_POSIX_C_SOURCE 200809L` in the runtime, so the pipe is made with `pipe` and `fcntl`.
- A second `wait` while one is pending on the same loop is `Err(AsyncError.Failed(...))` (the loop allows one reader per descriptor); documented in the module comment and SPEC.
- New case `signal_forward_run` (`testdata/signals/forward_run.nra`): SIGTERM sent to the parent during `Process.run`; the child reports it and the earlier `wait` gives Terminate.

## Deviations

- Outside the brief's paths: `selfhost/src/std/stdsignal.nra`, `stdlib.nra`, `emitstd.nra`, `check/buildconditions.nra`, `scripts/std/generate.nra` and the exits and emitchecks suites. A new std module needs each of them.
- No change to `std_async_loop.c`.
- `wait` takes no argument and catches both signals. A program that wants only one cannot ask for it.
- A signal sent while `Process.run` runs goes to the child and is also kept for the wait (see Review fixes). The harness sends from outside.
- The testdata sample `testdata/signals/shutdown.sh` now takes the expected output file and the signals to send.

## Gotchas

- Signals that arrive before the first `wait` get the default action. Start the wait before anything that announces readiness (the sample starts its first wait, then writes its pid).
- Windows path untested.
- The seed is not regenerated; the orchestrator does that once per batch.
