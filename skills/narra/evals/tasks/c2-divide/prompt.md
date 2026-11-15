---
name: c2-divide
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

`divide` in `src/calc.nra` panics when the divisor is 0. Change it to return `Result<int, string>`: `Ok(quotient)`, or `Err(message)` for a zero divisor. Update `ratio` so that it returns the quotient as text, or `undefined` when the divisor is 0, and update every other caller.
