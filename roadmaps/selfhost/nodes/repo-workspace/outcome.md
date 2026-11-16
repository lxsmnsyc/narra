# Outcome: repo-workspace

## Built

- `narra.toml` at the repository root: `[workspace] members = ["selfhost", "selfhost/tests"]`, no `[package]`, no `[fmt]`.
- `selfhost/tests/src/packages.nra`: case `packages/repository/workspace`. It runs `check .` and two `query --package . def` steps (one name in `selfhost`, one in `selfhost/tests`) from the repository root.
- `selfhost/CLAUDE.md`: one added bullet about the root workspace.

## Deviations

- `selfhost/tests/src/packages.nra` is in neither the owned nor the shared paths of the brief; the guard case needed it, since the brief puts the test in the `packages` suite.
- The criterion "reports the seed current" does not hold, and did not before this node: bootstrap prints "seed/narrac.c is older than the source" on feat/batch-19 itself. This node changes no seed input.

## Stubs left

- None.

## Gotchas

- `check .` prints nothing on success, so the guard proves member coverage through `query def`, which at a root answers from every member.
- `fmt .` at the root now formats the members only (not `testdata/`, `std/`, `bench/`). The sources suite names its paths, so it is unchanged.

## Proposed learnings

- None.

## Verification

- Before (feat/batch-19): `sh scripts/test.sh` PASS, 10434 cases. After: PASS, 10435 cases (the one new guard case).
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- `narrac emit selfhost` sha1 83c4e9c0 before and after: identical.
- `narrac check .` exit 0; `narrac fmt --check .` exit 0.
- The guard case fails with the root manifest removed (`check .` exit 1) and with `selfhost/tests` dropped (`unknown name "runCommandsCase"`), and passes with the manifest.

## Size

- Estimated S, actual S.
