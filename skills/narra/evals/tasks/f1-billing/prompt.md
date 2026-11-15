---
name: f1-billing
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

`narrac check .` reports errors in `src/billing.nra`. Fix them without changing what `total` and `discounted` return.
