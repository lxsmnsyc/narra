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

- The eval was not run. Acceptance item 2 is open. Reasons: the brief says to run it alone at the end of the batch after every code child has landed, and the sandbox of this worker refused the `claude plugin` eval command, so no trial run gave a cost estimate. Decision taken in place of the gate: write the tasks and tests before any run (done), keep the cost cap and the threshold of the brief, and leave the 40 runs to the orchestrator at the end of the batch. The threshold is unchanged: at least 15 points more passes, or a third fewer median `narrac check` calls with an equal pass share; report the counts.
- The eval cases are not in the `case.yaml` form of `claude plugin eval`: the worker could not read that command's help. The `prompt.md`, `starter/` and `grade.sh` files hold everything a case needs. Whoever runs the eval writes the `case.yaml` files from them.
- Acceptance item 1 is partly checked. `claude plugin validate .` passes on the marketplace file. The worker was refused `claude plugin marketplace add` (it changes the user's settings), so the load from a local path was not tried. Run `/plugin marketplace add <path to this repository>` once.
- Plugin source: the repository root, not `skills/narra`, because `skills/narra` has no `.claude-plugin` folder and the default `skills/` lookup is the layout Claude Code documents. Installing copies the whole repository into the plugin cache.
- The marketplace file names `lxsmnsyc` as owner and author, taken from the repository path in the brief. There is no LICENSE file in the repository, so `claude plugin validate --profile directory` would fail on it. Not needed for a marketplace of your own.
- Edited outside the owned paths: `README.md` (shared).

## Stubs left

- The eval run and the `case.yaml` files.

## Gotchas

- In a package whose `narra.toml` has no `[[binary]]`, `run` is an ordinary exported function: an unused `args` is E0921, and it needs `uses io, alloc` when it prints. With a `[[binary]]` entry neither applies. The eval packages have a `[[binary]]`.
- `refs` and `callers` skip import lines and calls through function values. The skill says so.
- `Array.length` is a function: `items.length` is an error. The skill does not repeat this, the guide covers it.
- `narrac check` of a single file needs an exported `run`; `check-examples.py` adds a stub when a fence has none.
- The eval `solution/` folders ship inside the plugin, because the plugin source is the repository root. An agent that browses the plugin cache can read them. Run the eval from a copy of `starter/` outside the cache.
- A reviewer saw an exported function with `uses alloc` that prints and still compiles with no `uses io`; with no `uses` it gives E0800. Possible compiler gap, not checked further, not a skill error.
- E0800 has several messages. `narrac explain E0800` prints the trait one, not the missing-`uses` one.

## Proposed learnings

- A skill that teaches a language keeps each example in a fence that a script compiles | the example cannot drift from the compiler | skills/ | yes (check-examples.py)

## Verification

- See Proof. `sh scripts/test.sh`: see the final lines in the worker report.

## Size

- Estimated M, actual M without the eval run.
