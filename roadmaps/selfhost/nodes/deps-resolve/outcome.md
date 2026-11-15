# Outcome: deps-resolve

## Built

- `selfhost/src/depresolve.nra`:
  - `resolveDependencies(roots, overrides)`: minimal version selection over several roots; returns the lock entries in discovery order (`formatLock` sorts them).
  - `checkAgainstLock(resolved, locked)`: E0452 when an entry with the same name, URL and version has another commit or tree.
  - `runResolveCommand`: the test-only command `narrac pkg --resolve <root dir> [<more root dirs>]`. It reads the first directory's `narra.lock`, resolves, checks, writes the lock when the text differs, and prints it. Only the first directory's `[overrides]` apply.
- `selfhost/src/manifest.nra`: `Manifest.overrides` and the `[overrides]` table, read like `[dependencies]`.
- Codes in `diagnostics.nra`: E0450 one name with two URLs, E0451 version without a tag, E0452 lock mismatch, E0453 manifest names another package, E0454 `[overrides]` outside the root, E0455 override nobody needs, E0456 override tag missing, E0457 dependency without narra.toml, E0458 a file the command cannot read or write.
- Cases `resolve-*` in the `deps` suite (`selfhost/tests/src/depsresolve.nra`), fixtures in `testdata/deps/resolve/`.

## Acceptance

- Diamond (higher minimum wins, the unused 1.1.0 is not fetched, commit and tree equal `git rev-parse`): `resolve-diamond`. Chain three deep over one repository with member tags: `resolve-chain`. Cycle between two packages: `resolve-cycle`.
- Errors: `resolve-error-two-urls` E0450, `-no-tag` E0451, `-lock-mismatch` E0452, `-wrong-name` E0453, `-overrides-in-dependency` and `-overrides-in-second-root` E0454, `-override-unused` E0455, `-override-tag` E0456.
- Override pins the shared package to a fork URL at a lower version, and the lock shows the fork; the override also replaces a second URL for the same name: `resolve-override`.
- Two roots with different minimums select the higher: `resolve-roots`.
- Lock written twice is byte-identical, a stale entry is dropped, and the second run starts no process (PATH empty): `resolve-lock-twice`.

## Deviations

- No fake fetch: the resolver calls `depcache` directly and the cases use real `file://` git repositories, which proves the same graph logic and the real tag lookup.
- Touched outside the owned and shared lists: `selfhost/src/main.nra` (one dispatch row), `selfhost/src/depcache.nra` (`isVersionBefore` exported), `selfhost/tests/src/deps.nra` (calls the new cases) and `selfhost/tests/src/depsresolve.nra` (the new cases).
- The caller of `resolveDependencies` enforces that only the build root has `[overrides]` (E0454); the library trusts the `overrides` it is given. The command does this for its extra roots.
- A version cached under its plain tag is used without a fetch even if the repository also has a member tag for it; a clean machine would pick the member tag. This matters only in a repository that has both tag forms.
- `Dependency.workspace` is not added; workspace-manifest owns it.
- The library takes the build root's overrides as a parameter. The command applies the rule "first directory only".
- Requirements of a version that loses stay in the graph (Go-style), so a lower version's dependencies can raise other packages.
- A dependency's manifest is read at the root of its checkout; finding a package inside a workspace repository is for package-dependencies.
- A cached version is found by its member tag, then its plain tag, without git. Only a cache miss runs `ls-remote`.
- Names of version errors carry no line numbers; the LoadError has line 0.

## Stubs left

- `package-dependencies` and the workspace loader call `resolveDependencies`; nothing in `narrac build` does yet.

## Gotchas

- `pushRequirements(mutable queue, ...)` needs `copy(queue[position])` for the requirement read afterwards.
- A shell variable named `d` in a fixture generator was clobbered by a `for d` loop; fixtures are plain files now.

## Contract

`dependencies` version 2 holds. `Manifest.overrides` is added as the contract says.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 9955 cases in 246s` (after the final rebase onto feat/batch-16; no timing case failed).
- `sh scripts/test.sh --suite deps` (separate build): `test.sh: PASS, 24 cases in 5s`.
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C.

## Size

- Estimated M, actual M.
