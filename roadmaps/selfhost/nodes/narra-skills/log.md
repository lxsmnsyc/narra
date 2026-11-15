# narra-skills log

- Branch feat/narra-skills from feat/batch-16. Built a compiler with `sh selfhost/bootstrap.sh` into the scratch folder (stage 3 ran, the seed is older than the source).
- Rules mined from `selfhost/src/diagnostics.nra` and `narrac explain`. Guessed codes were wrong twice (shadowing is E0103, a missing `mutable` is E0315); `check-examples.py` caught both.
- `Arrays.push` needs `uses alloc` on an exported function; the first mutable example failed on it.
- Single-file `narrac check` demands an exported `run`; the checker adds a stub.
- Eval tasks: first versions failed on `args` unused and on `.length` in a package. Fixed with a `[[binary]]` in the starter `narra.toml` and `Arrays.length`.
- `claude plugin marketplace add` and the plugin eval command were refused by the sandbox. Not retried. `claude plugin validate .` passes.
- Plugin layout: first tried source `./skills/narra` with `skills: ["./"]`; `validate` lists no contents either way, so I used the documented default (root source, `skills/<name>/SKILL.md`).
- 2026-11-16: wrote case.yaml and scaffold.sh per task, prepare-eval.sh; trial on f3-queue; ran all 40 runs (19/20 with, 15/20 without). See outcome.md.
