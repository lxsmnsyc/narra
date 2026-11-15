# Outcome: narra-skills

## Built

- `skills/narra/SKILL.md`: one skill, about 850 words (under 2,000 tokens). It has the check loop (`narrac query`, `check --json`, `explain`, `apply-fix`, `fmt`, `test`), eight rules with a wrong and a right example each (shadowing E0103, nested function E0930, unused name E0921, dead code E1500, `let` never reassigned E1510, missing `uses` E0800, missing `mutable` E0315, `throw` E0927), the global functions, tests, and links to the guide chapters. The rules come from `selfhost/src/diagnostics.nra` and from running the compiler, not from memory. The text does not cite SPEC or D-numbers and assumes no TypeScript.
- `.claude-plugin/marketplace.json`: marketplace `narra` with one plugin, `narra`, whose source is the repository root. Claude Code finds the skill in the default `skills/` folder. Install: `/plugin marketplace add lxsmnsyc/narra`, then `/plugin install narra@narra`.
- `skills/narra/evals/`: 10 tasks (`tasks/<id>/` with `prompt.md`, `starter/`, `checks.nra`, `solution/`): 4 write-a-function (w1 to w4), 3 fix-the-compile-error (f1 to f3), 3 change-an-existing-program (c1 to c3). Each starter is a package with a `[[binary]]`. `grade.sh` copies `checks.nra` into the package and runs `narrac check` and `narrac test`. `verify-tasks.sh` checks the tasks without a model: each starter must fail and each starter with its solution must pass.
- `skills/narra/evals/check-examples.py`: extracts every `narra` fence of a markdown file and runs `narrac check` on it. A plain `narra` fence must compile. A `narra error CODE` fence must fail with that code.
- README: a line that links the skill and gives the install commands.

## Proof

Not a fix node. Checks run with a compiler built by `sh selfhost/bootstrap.sh` from this branch:

- `python3 -I skills/narra/evals/check-examples.py <narrac> skills/narra/SKILL.md`: 13 of 13 examples as tagged.
- `sh skills/narra/evals/verify-tasks.sh <narrac>`: 10 tasks, 0 problems.
- `claude plugin validate .`: passed.

## Deviations

- (Superseded by the Eval section below: the eval ran.) The first worker did not run it. Reasons: the brief says to run it alone at the end of the batch after every code child has landed, and the sandbox of this worker refused the `claude plugin` eval command, so no trial run gave a cost estimate. Decision taken in place of the gate: write the tasks and tests before any run (done), keep the cost cap and the threshold of the brief, and leave the 40 runs to the orchestrator at the end of the batch. The threshold is unchanged: at least 15 points more passes, or a third fewer median `narrac check` calls with an equal pass share; report the counts.
- (Superseded: the `case.yaml` files now exist, see Eval.)
- Acceptance item 1 is partly checked. `claude plugin validate .` passes on the marketplace file. The worker was refused `claude plugin marketplace add` (it changes the user's settings), so the load from a local path was not tried. Run `/plugin marketplace add <path to this repository>` once.
- Plugin source: the repository root, not `skills/narra`, because `skills/narra` has no `.claude-plugin` folder and the default `skills/` lookup is the layout Claude Code documents. Installing copies the whole repository into the plugin cache.
- The marketplace file names `lxsmnsyc` as owner and author, taken from the repository path in the brief. There is no LICENSE file in the repository, so `claude plugin validate --profile directory` would fail on it. Not needed for a marketplace of your own.
- Edited outside the owned paths: `README.md` (shared).

## Stubs left

- None for the eval. The README link and the local `/plugin marketplace add` load are as above.

## Gotchas

- In a package whose `narra.toml` has no `[[binary]]`, `run` is an ordinary exported function: an unused `args` is E0921, and it needs `uses io, alloc` when it prints. With a `[[binary]]` entry neither applies. The eval packages have a `[[binary]]`.
- `refs` and `callers` skip import lines and calls through function values. The skill says so.
- `Array.length` is a function: `items.length` is an error. The skill does not repeat this, the guide covers it.
- `narrac check` of a single file needs an exported `run`; `check-examples.py` adds a stub when a fence has none.
- The eval `solution/` folders ship inside the plugin, because the plugin source is the repository root. An agent that browses the plugin cache can read them. Run the eval from a copy of `starter/` outside the cache.
- A reviewer saw an exported function with `uses alloc` that prints and still compiles with no `uses io`; with no `uses` it gives E0800. Possible compiler gap, not checked further, not a skill error.
- E0800 has several messages. `narrac explain E0800` prints the trait one, not the missing-`uses` one.

## Eval

Ran 2026-11-16 with `claude plugin eval` 2.1.325, model claude-opus-5-5 in both arms, `--ablation with-without -j 2`, 10 tasks x 2 runs x 2 arms = 40 runs, alone on the machine. Full results (JSON, HTML report, per-run traces, kept run folders, `scored.json`) are in `/private/tmp/claude-501/narra-eval/` and `/private/tmp/e-*`; nothing is written into the repository.

Threshold (fixed in the brief): pass share up by at least 15 points, or median check calls down by a third with an equal pass share.

| Arm | Passed (check and task test) | Share | Median `narrac check` calls |
| --- | --- | --- | --- |
| with the skill | 19 of 20 | 95% | 2 |
| without the skill | 15 of 20 | 75% | 2.5 |

Verdict: threshold MET. The pass share rose by 20 points (4 more passes out of 20). The median calls fell by 20%, not a third, so the second branch alone would have missed.

Cost: 585 s wall time (9 min 45 s), reported cost $6.40 (with $3.42, without $2.98). Tokens from the run results: 304k fresh (input, cache write, output) with the skill and 264k without, 568k together; cache reads add 5.3M more. The 1.5M token cap was applied to fresh tokens. Counting cache reads the run would have passed it by a wide margin. The trial runs (task f3-queue, one run per arm) cost $0.37 for 2 runs, which gave an estimate of about $7.5 and under 15 minutes for 40.

### How it was run

- `skills/narra/evals/tasks/<id>/` holds `case.yaml` (scaffold script and one grader), `prompt.md` (frontmatter: `runs: 2`, `max_turns: 30`, `timeout_seconds: 900`, read-only `allowed_tools`, and an `append_system_prompt` that says `narrac` is on PATH and the working folder is the package; the same text goes to both arms) and `scaffold.sh` (copies `starter/` into the run's empty working folder; needs `--scaffold`).
- `claude plugin eval` has no script grader (only regex, tool_used, tool_order, file_exists, llm, baseline). The only grader in each case is a `tool_used` count of Bash calls matching `narrac check` with `min: 0`, so it always passes and only records the call count. The pass or fail result comes from running `grade.sh` on each run's kept working folder afterwards (`--keep-temp`; the folders are sealed and need `chmod 700` on the run folder and its `sealed/` subfolder).
- Check calls come from each run's `trace.jsonl`: the Bash tool calls whose command contains `narrac check`. The number is per run over the whole run, not "before the first pass"; failed runs with no call count as 0.
- Solutions hidden: `sh skills/narra/evals/prepare-eval.sh DEST` builds a plugin copy with `SKILL.md`, the marketplace file and the eval folders but without every `solution/` and `checks.nra`. The eval ran on that copy (`DEST/skills/narra` is the plugin). The harness has no exclude option of its own.
- Command: `PATH=<bin>:$PATH claude plugin eval DEST --scaffold --trust-plugin --allow-tools Bash Write Edit --ablation with-without -j 2 --keep-temp --no-publish --json ... --output-dir ... --report ...`. The compiler came from one `sh selfhost/bootstrap.sh <bin>` run. `Bash`, `Write` and `Edit` can only be granted on the command line.

### Caveats

- `narrac test` fails inside the harness sandbox in both arms: it runs `mktemp -d` with no TMPDIR, which falls under `/var/folders`, and the sandbox refuses the write. `narrac check` and `narrac fmt` work. So no agent could run the tests; the task tests only ran in grading.
- The system prompt line that names `narrac` was added after the first trial: without it the without-skill arm did not know the compiler's name, which would have inflated the gap.
- With 20 runs per arm a gap of 4 passes is not far from noise. Both skill misses and baseline misses are few.

### Failure kinds (the 6 failed runs)

- Without the skill, 5 of 6 failures: Narra syntax from other languages. `+` to join strings (w3, both runs), `int` plus `void` from a call used as a value (w3), a multi-line conditional without parentheses and an `if` branch without braces (c2, both runs), a function signature written wrongly (w4). None of the five failed runs called `narrac check`.
- With the skill, 1 of 20: w3 run 2 called `isTrue` from `std/test`, which does not exist; the skill does not list the test helpers.

## Proposed learnings

- A skill that teaches a language keeps each example in a fence that a script compiles | the example cannot drift from the compiler | skills/ | yes (check-examples.py)

## Verification

- See Proof. `sh scripts/test.sh`: see the final lines in the worker report.

## Size

- Estimated M, actual M without the eval run.
