---
name: w2-sum-evens
runs: 2
max_turns: 30
timeout_seconds: 900
allowed_tools: [Read, Glob, Grep, Skill]
append_system_prompt: "The Narra compiler `narrac` is on PATH. Use it to check your work before you finish. The working folder is the Narra package to edit."
---

Add a module `task/numbers` in `src/numbers.nra` that exports `sumEvens(numbers: Array<int>): int`, the sum of the even numbers in the array. An empty array gives 0. Negative numbers count.
