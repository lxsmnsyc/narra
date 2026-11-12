# Premortem: batch-10

It is a week from now and batch-10 is late or reverted. Why?

1. **checker-emitter-copies was bigger than L.** About 1900 copies may come from many patterns, and a stricter E0978 can flag hundreds of sites in `selfhost/src`. Change: the worker records the patterns and their counts first. If the fix does not fit, it lands the rule and the fixes it has, and the orchestrator narrows the node and moves the rest to the inbox, as batch-9 did with hidden-copy-fix.
2. **A stricter E0978 broke user programs and samples.** Treating names as live until their scope ends rejects code that runs today. Change: D104 states the rule and its cost. Every rejected case gets an error test, and the guide and SPEC samples run in the full check.
3. **Per-holder credits slowed or changed the release build.** Change: the holder tracking compiles out under `-DNARRA_RELEASE`. The bench row at the end of the batch checks that `emit selfhost` instructions do not rise.
4. **checker-emitter-copies conflicted with the S nodes.** It edits code across `selfhost/src`. Change: the S nodes land first, and checker-emitter-copies rebases and reruns `narrac fmt` before review.
5. **The seed update failed at the end.** Runtime and std text changed, so the seed needs two updates. Change: the orchestrator runs both updates and `bootstrap.sh` before the fresh review of the branch, and reruns the debug-built `emit selfhost` on the result.

Accepted risk: destructuring-design ends at a gate. If the maintainer does not approve D103 in time, it moves out of the batch with its log, and the batch ships without it.
