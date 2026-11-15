---
name: w3-parse-port
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Add a module `task/port` in `src/port.nra` that exports `parsePort(text: string): Result<int, string>`. It returns `Ok(port)` when the text is a whole number from 1 to 65535, and `Err(message)` for any other text. Do not panic.
