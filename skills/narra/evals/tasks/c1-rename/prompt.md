---
name: c1-rename
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Rename the function `area` in `src/shapes.nra` to `rectangleArea`, and update every use of it in the package. Behavior stays the same.
