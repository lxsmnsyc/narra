# Log: threads

## Proof first

- Base: feat/batch-13 at c63ba12e. Compiler: `sh selfhost/bootstrap.sh /private/tmp/claude-501/threads/base` (release, `-DNARRA_RELEASE -O2`).
- Command: `/private/tmp/claude-501/threads/base/narrac run testdata/threads_scope.nra`
- Result: exit 1, `threads_scope.nra:3:1: unknown module "std/thread"` and the follow-on errors.

## Branch

- After the change, the same command on the branch compiler prints `total 4999950000` (exit 0).
- No SyncSlot<T>: Thread and Scope are boxes in Narra source; see the outcome's Deviations.
- Machine load average 35 to 80 during the measurements (other workers).
