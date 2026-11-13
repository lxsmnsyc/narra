---
id: threads
title: Scoped threads, channels and shared state
status: todo
priority: must
type: feature
outline: true
milestone: p1-syntax
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: closures
    kind: hard
    why: thread functions are escaping closures
  - on: identity-cycles
    kind: hard
    why: Shared, channels and tasks follow the cycle decision
---

# Scoped threads, channels and shared state

## Goal

`Thread.scope`, `Thread.start`, channels, `Shared<T>`, `Atomic<T>`, and atomic reference counts for values shared across threads.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

Cycles: the identity-cycles node decides how reference cycles through `Shared`, channels and tasks are handled. Implement its answer here.
- 2026-11-12, from the process-signals review: `Process.run` (`selfhost/runtime/std_process.c`) and the test runner keep one process-global child slot and save and restore the signal handlers on each call, with `sigprocmask`. Two threads in `Process.run` at once can leave the forwarding handler installed with no child, so the process then ignores SIGTERM, SIGINT and SIGHUP. Before threads ship: a mutex-guarded count for installing and restoring the handlers, a set of child pids, and `pthread_sigmask`.
