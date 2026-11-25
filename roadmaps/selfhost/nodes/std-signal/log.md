# std-signal log

- Design: self-pipe (socket pair on Windows). The handler writes one byte (1 SIGINT, 2 SIGTERM). The loop waits on the read end with narra_loop_wait_io, so std_async_loop.c needs no change.
- New files: selfhost/runtime/std_signal.c, selfhost/src/std/stdsignal.nra, std/src/signal.nra, testdata/signals/shutdown.*, testdata/signals/forward_run.*, testdata/wasi_cases/left_out_signal.*.
- Bootstrapped a seed narrac first (cc seed/narrac.c) to generate runtimestdsignal.nra, since bootstrap fails until that file exists.
