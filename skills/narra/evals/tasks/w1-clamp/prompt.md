---
name: w1-clamp
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Add a module `task/math` in `src/math.nra` that exports `clamp(value: int, low: int, high: int): int`. It returns `low` when `value` is below `low`, `high` when `value` is above `high`, and `value` otherwise.
