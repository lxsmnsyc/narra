# Log: docs-concept-audit

## 2026-11-20

- Worktree check: toplevel is under `.claude/worktrees/`. Branch renamed to `spike/docs-concept-audit`, fast-forwarded to `feat/batch-21` (`501f19f9`).
- Built `narrac` from `seed/narrac.c` in the session scratchpad (`cc -O1 -w ... -lm -lpthread`).
- Read `SPEC.md` sections 1 to 16 in full, every guide chapter and `docs/install.md` in full, `skills/narra/SKILL.md`, and the feature sections of `compiler/README.md`.
- Coordinator message mid-task (maintainer input): use `SPEC.md` as the inventory baseline, one row or more per subsection of sections 1 to 15, and check async and FFI first. The inventory follows that.
- Keyword search of the guide (`std/fs`, `Io.`, `Env.`, `async`, `extern`, `Cell`, `Thread` and about 60 more) to confirm what no chapter names. Script: scratchpad `map.sh`.
- Spot checks with the seed build:
  - `pair[0]` on a tuple compiles and prints 1. Chapter 2 says tuples cannot be indexed.
  - A local `const` array passed as `mutable` to `Arrays.push` compiles. Chapter 2 says only a `let` can be passed.
  - `total / Arrays.length(items)` fails with E0979 at the divisor.
  - A `Cell` counter returned from a function, `Fs.writeText` and `Fs.readText`, `Json.decode<User>` and `Json.encode` all compile and run.
  - `export async function run` with `Async.all` over two `AsyncTime.sleep` tasks, `Thread.scope` with two threads, and `extern { function cbrt(...) } from "m"` called in `unsafe` inside `@trusted` all compile and run (output `6`, `42`, `3.0`).
  - The unknown-module error lists 54 importable std modules.
- Counts in `audit.md` come from parsing the inventory table (awk over the `| N |` rows), not from estimates.
- The plain-writing checker flags every use of the word "crucial" as a stock phrase. It is the rank name the brief defines, so it stays, for one term per meaning. The other remaining findings are long-sentence warnings in lists.
- `DECISIONS.md` read only through the D numbers that `SPEC.md` cites.
