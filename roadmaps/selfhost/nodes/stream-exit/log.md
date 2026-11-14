# Log: stream-exit

- Base: feat/batch-15 at c4884fe1. Proof: `narrac build testdata/async_exit_ignored_cancel.nra` with a compiler bootstrapped from the base, then the binary run with a kill after 8 s: still running (HUNG). The sample has a detached task and a detached stream reader that ignore the cancel of their sleep.
- Maintainer corrected the rule: no extra Err returns; the first new suspending call after the exit cancel drops the task. Counters removed.
- Branch: same command on the branch compiler exits at once; output stable over 5 runs.
- `emit selfhost` from the base and branch compilers is byte-identical (cmp).
