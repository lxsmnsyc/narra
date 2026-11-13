# Log: threads-process

- Base: feat/batch-13 at 7df2cd63. Branch renamed fix/threads-process.
- Proof, base: scratch C program (`/private/tmp/claude-501/threads-process/two.c`) includes `runtime.h` and the base `std_process.c`, starts two pthreads that each call `narra_proc_run` on a `sh` child that traps SIGTERM and exits 7, waits for both children, then sends SIGTERM to the process. Built with `cc -std=c11 -g -fsanitize=address,undefined -pthread`. Result: the second child never gets the signal, the program hangs until its 15 s alarm, exit 142.
- Proof, branch: same program, same flags, 5 runs: `codes 7 7; SIGTERM handler restored: yes; mask clear: yes`, exit 0 each time.
- Same pattern in all three files, with prefixes `narra_proc`, `narra_runner`, `narra_panics`: mutex-guarded count of active calls, handlers saved and installed by the first call and restored by the last, 64 slots of `_Atomic int` child pids claimed with compare-and-swap, `pthread_sigmask`.
- A slot holds -1 between the claim and the fork. The handler stores the signal in a per-slot pending cell before it reads the pid, and `attach` stores the pid before it reads the pending cell, so a signal that lands on another thread in that window is sent to the child after attach.
- Generated files rewritten with `--suite generated --update` using a release compiler built from the base source.
