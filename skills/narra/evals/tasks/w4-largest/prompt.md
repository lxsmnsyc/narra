---
name: w4-largest
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Add a module `task/stats` in `src/stats.nra` that exports `largest(numbers: Array<int>): Option<int>`. It returns `Some(biggest)`, or `None` for an empty array.
