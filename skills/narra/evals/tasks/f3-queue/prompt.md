---
name: f3-queue
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

`narrac check .` reports errors in `src/queue.nra`. Fix them. `addAll` must add the extra numbers to the caller's array, and `firstOr` must keep returning the first number or the fallback.
