# Outcome: package-exports

## Built

- `[exports]` in `narra.toml` (`manifest.nra`), the import check and rewrite (`selfhost/src/exports.nra`, called from `modules.nra`), the std rule in `scopes.nra` (E0474, E0426 removed), codes E0470 to E0474.
- SPEC section 15, guide chapter 8, and D97 in DECISIONS.md ("Built" replaces "Not built yet").

## Acceptance

- Errors in E0470-E0479: E0470 key without the package name, E0471 value that is not a module file or names no file, E0472 two entries with one key, E0473 import of an unlisted path with the exposed keys in the message. Cases: `testdata/manifest/exports_{bad_key,bad_value,duplicate_key,not_string}`, `build-error-export-file`, `build-error-unexported`.
- Fixture dependency `lib` with `"lib" = "src/index.nra"` and `"lib/stream" = "src/streaming/reader.nra"` (real file:// git repository): `build-exports` imports both paths and runs with `-sanitize`; `build-error-unexported` imports `lib/streaming/reader` and gets E0473 listing `"lib", "lib/stream"`.
- `std/internal/unicode` (`unicode_internal_import`) and `std/ext/int` (`exports_std_ext`) from a user package fail with E0474; E0426 is gone from the table, source and docs.
- SPEC section 15 documents `[exports]`; D97 no longer says "Not built yet" for it.

## Deviations

- `narra move` does not exist: updating the table on a move is out of scope.
- Fixtures are under `testdata/deps/build/` and `testdata/manifest/`, not `testdata/packages/exports_*`.
- Touched outside the shared list: `depload.nra` (the table in `DependencyPackage`), `unused.nra` (remove-import fix, see log), `diagnostics.nra` rows, `selfhost/tests/src/depsbuild.nra`, `compiler/README.md` (E0426 to E0474).
- The std error message text is unchanged, and it does not list exposed std paths (about 80 modules).
- A package imports its own modules by real paths even when they are not in its table.

## Stubs left

- `narra move` (command does not exist).

## Gotchas

- An import of another package is rewritten to the module path before the checker runs, so diagnostics, `narrac pkg` dumps and find-references show the real module path, not the key.

## Contract

`dependencies` unchanged. Additions: `Manifest.exports`, `DependencyPackage.exports`.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 10063 cases in 211s`.
- `--suite deps`: `test.sh: PASS, 35 cases in 6s`. `--suite manifest`: `test.sh: PASS, 32 cases in 0s`. The generated and guide suites are inside the full run (no generated file changed).
- `emit selfhost` by a compiler built from the base commit and by the new one: cmp equal (12244484 bytes).
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C.

## Size

- Estimated M, actual M.
