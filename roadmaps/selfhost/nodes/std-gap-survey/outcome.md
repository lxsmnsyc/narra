# Outcome: std-gap-survey

## Built

- `roadmaps/selfhost/nodes/std-gap-survey/survey.md`: 33 missing std pieces, ranked by need. Each row has the std libraries that have it, the evidence in this repo and where the repo promises it. It also has the module, Narra or runtime, a size and a before or after `v0.1.0` mark.
- Before the release (8 items):
  - sorting, `std/set`, and array `pop`, `truncate`, `contains`, `indexOf`, `reverse`
  - `Float.parse`, `Float.toFixed` and character classes in `std/char`
  - string `trimStart`, `trimEnd`, `padStart`, `padEnd`, `lastIndexOf`
  - `Fs.appendText`, `rename`, `copyFile`, `removeDirAll` and `Process.cwd`

## Deviations

- The coordinator widened the scope during the run. The table has a "Promised" column for pieces that SPEC, DECISIONS, compiler/README or the inbox already name. The comparison adds Web APIs, Bun and Crystal.

## Stubs left

- None.

## Gotchas

- `Float.parse` needs no new C: `std/internal/float` already has `FloatText.parse`, which `std/json` and `std/toml` use.
- `Arrays.pop` and `truncate` must shrink the array in place, so they are compiler builtins. A Narra version would copy the array, as `popLast` in `driver/modules.nra` does.
- New fs and process functions that WASI cannot back need a row in `selfhost/src/buildconditions.nra` (inbox 2026-11-15).
- `Process.run` options are marked after the release on purpose. Parameter defaults let them arrive without breaking a call.

## Proposed learnings

- Look for hand-written copies in `selfhost/src` before adding a std function | its copies show real need | `std/src`, `selfhost/src` | no

## Verification

- Read-only survey: grep counts over `selfhost/`, `scripts/`, `testdata/`, `bench/` and `docs/guide/` on `feat/batch-22` at b373c7b4. No compiler run.
- `check_prose.py` on `survey.md`: ok.

## Size

- Estimated M, actual M.
