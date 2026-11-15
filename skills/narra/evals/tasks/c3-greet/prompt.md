---
name: c3-greet
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Give `greet` in `src/greeting.nra` a second parameter, `punctuation: string`, that ends the greeting: `greet("Ana", "!")` returns `Hello, Ana!`. Update every caller in the package. `welcome` keeps greeting with `!`, and `farewell` greets with `.`.
