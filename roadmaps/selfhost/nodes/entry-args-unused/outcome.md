# Outcome: entry-args-unused

## Built

- Not reproduced. The entry function's parameter is already exempt from the unused report (`entryParameter` in `selfhost/src/check/check.nra`, since 2026-11-07). No code change.

## Proof

- base: b373c7b4, `narrac check` and `narrac run` of a single-file `export function run(args: Array<string>): int` (plain and `async`), of packages with `entry = "main.run"` and `"app/main.run"`, of a workspace member, and of all 975 narra fences in docs, README.md and SPEC.md: no `args` report. Details in `log.md`.
- branch: not built, the problem did not reproduce.

## Deviations

- No fix, no new cases. The only report that can be forced is a package binary with `entry = "run"` (no module), which is not a valid entry.

## Gotchas

- If the maintainer saw the warning, the narrac was probably older than 2026-11-07, or the case is not covered above; ask for the exact sample and command. `snippet-warnings` should not wait for this node.
- The unused reports are errors (E0921, E0922), not warnings, and they print `parameter "args" is never read`.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point, seed current. Full `sh scripts/test.sh` not run: no source change.

## Size

- Estimated S, actual S.
