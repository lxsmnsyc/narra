# Outcome: std-semver

## Built

`std/semver` in `std/src/semver.nra`. Import it as `import * as Semver from "std/semver"`.

Public API:

- `struct Version { major: int, minor: int, patch: int, prerelease: Array<string>, build: Array<string> }` (derives Show, Equatable)
- `enum Part { Major, Minor, Patch }` (derives Show, Equatable)
- `parse(text: string): Option<Version> uses alloc`
- `toString(version: Version): string uses alloc`
- `compare(left: Version, right: Version): int` returns -1, 0 or 1; build metadata is ignored
- `bump(version: Version, part: Part): Version uses alloc` resets lower parts and drops prerelease and build
- `satisfies(version: Version, requirement: string): bool uses alloc`

`satisfies` reads the requirement as SPEC section 14 defines a dependency `version`: the minimum, written MAJOR.MINOR.PATCH. It is true when the version is at least the minimum. A requirement of any other form (`1.2`, `^1.2.0`, a prerelease) gives `false`. A prerelease of the minimum (`1.4.2-rc.1` against `1.4.2`) is below it.

Numeric prerelease identifiers compare by length then text, so they have no int limit. The three numbers must fit an `int`, as in `isVersion` of the driver.

## Files

- `std/src/semver.nra`, `testdata/std_semver.nra` and `.out`
- `selfhost/src/std/stdsemver.nra` (registry row) and one import and `registerAll` line in `selfhost/src/std/stdlib.nra`
- `scripts/std/generate.nra`: a ModuleInfo for the reference page
- generated: `selfhost/src/std/stdtext.nra`, `docs/reference/std/semver.md`, `docs/reference/std/README.md`
- goldens that list the std modules (`testdata/errors/*.err` and similar), rewritten by `sh scripts/test.sh --update`

## Deviations

- Touched `stdsemver.nra`, `stdlib.nra`, `scripts/std/generate.nra` and the module-list goldens. A std module needs them, as std/set did. Other std nodes touch the same lines and goldens, so the orchestrator should expect conflicts there; the order of the module list in the error text follows the `registerAll` order.
- The seed is not regenerated. The compiler's own `isVersionBefore` and `isVersion` stay as they are.

## Gotchas

- Std text is embedded in the compiler, so a change to `std/src` needs the generated update and a bootstrap before a test sees it.
- Assigning a value bound by `is Some(x)` to a local array needs `copy(x)`.
