# Log: effect-export-gap

Compilers: the seed (`cc -O1 -w -o narrac seed/narrac.c -lm -lpthread`) and the stage 1 compiler from `sh selfhost/bootstrap.sh <dir>`, both built on `feat/batch-20` (0fd1e813). Each shape was run with `narrac check` on a single file or `narrac check .` and `narrac build .` on a package.

Single file, exported function with `uses alloc` that prints:

- print in the body: error "uses `io` (through `print`)". Reported.
- print through a local arrow function that is called: reported.
- print of a local string, with a return value: reported.
- print through a private helper (with and without `uses alloc` on it): reported on the helper and on the exported caller.
- print inside a nested arrow that returns a value: reported.
- `export function run(args: Array<string>): int uses alloc` that prints: no error. `run` as the entry is exempt, as `skills/narra/SKILL.md` says.

Package with `[[binary]] entry = "main.run"` and `src/lib.nra`:

- exported `noisy(): string uses alloc` that prints: reported by `check` and `build`, seed and stage 1.
- exported function that calls a local arrow that prints: reported.
- exported function that takes a callback and calls it, with the caller passing a printing arrow: reported at the call ("the argument for `callback` ... uses `io`").

Likely source of the narra-skills report: the function was the entry (`run`), which needs no `uses`. Not a compiler gap.

Not tried: trait implementations (out of scope, separate inbox item).
