# Outcome: c-packages

## Built

A C library is a package whose manifest names its upstream by git URL and tag. Other packages depend on it and bind its files with `from "<package>/<path>"`.

### Manifest

```toml
[package]
name = "c_zlib"
version = "1.3.1"

[c]
git = "https://github.com/madler/zlib"
tag = "v1.3.1"
sources = ["*.c", "!example.c", "!minigzip.c"]
include = ["."]
defines = ["Z_HAVE_UNISTD_H"]
```

- `[c]` takes `git` and `tag` (strings, both or neither; one alone is E1000, `[c.os.<name>]` with either is E1000). `CTable.git` and `CTable.tag` in `selfhost/src/driver/manifest.nra`; `findUpstream(tables)` returns the pair (`Upstream`).
- With them, the `[c]` paths (`sources`, `include`) are relative to the upstream checkout. A `from "./file.c"` in the package's own modules stays relative to the package. `PackageC.upstream` holds the checkout folder (empty otherwise), and `findCRoot` in `csources.nra` picks the root.
- A C package still needs a `[package]` name and one `.nra` module in `src/` (a dependency with none is E0462).
- Only a package fetched as a dependency can have `git` and `tag`; the package being built and a workspace member that has them get E1665 (`createUpstreamError` in `modules.nra`).

### Lock

A lock entry of a C package has four more keys after `content`, all or none (E0438 for some):

```toml
[[package]]
name = "c_zlib"
git = "https://github.com/example/narra-c-zlib"
version = "1.3.1"
tag = "v1.3.1"
commit = "<package commit>"
content = "<package tree>"
c_git = "https://github.com/madler/zlib"
c_tag = "v1.3.1"
c_commit = "<upstream commit>"
c_content = "<upstream tree>"
```

- Decision: the upstream is a field group of the package entry, not an entry of its own. Version selection picks one entry per package name, so the upstream follows the selected version and a lock stays a list of packages. `LockEntry.cGit`, `cTag`, `cCommit`, `cContent` (default empty) in `lockfile.nra`.
- `resolveDependencies` (`depresolve.nra`) fetches the upstream of each version it reads, with `fetchVersion` and the same cache folder as the git dependencies (`<cache>/git/...`). `checkAgainstLock` compares the upstream commit and tree with the lock when URL and tag agree: a different tree is E0452 and names the C package (`the C source of <name> <version> ...`). A missing tag is E0442 (`the C source of <name> <version>: the repository ... has no tag ...`).
- The tree check runs on a fetch, like a git dependency: with the checkout in the cache nothing is fetched and a moved tag goes unseen until the cache is gone.
- `listPackages` (`depload.nra`, `findUpstreamFolder`) loads the upstream from the lock and the cache for the language service and for a build after the resolve: the entry must name the manifest's `git` and `tag` (E0461), the checkout must be in the cache (E0460), its recorded tree must equal `c_content` (E0443). `DependencyPackage.upstream` is the folder.

### `from "<package>/<path>"`

- `isPackageFilePath` (`syntax/ast.nra`) accepts a package name, a `/`, then a path; the parser takes it where it takes a library name. `from "c_zlib/zlib.h"` binds the functions of a header; a path ending in `.c` that `[c] sources` does not list is compiled too (`collectUnits`).
- Load-time checks (`checkPackageFiles` in `csources.nra`, called from `modules.nra` after every C package is known): E1663 the binding package does not list the package in its `[dependencies]` (a package in the build only through another one does not count; neither does the package itself), E1664 the named package has no `[c]` table, E1660 the file does not exist in the C root, E1661 the path leaves it (`..`, symbolic links).
- A bare name is still a system library (`-l`); `emitexterns.nra` skips the `<package>/<path>` form in `collectLibraries` and reports `.c` uses in `collectCFiles`.
- The objects of the whole C package compile and link whether or not a block names a file of it (`[c] sources` of every package in `Package.cPackages`).

### Include propagation

`createUnit` builds the include folders of a unit from its own package and from each package in `PackageC.declared` (the names in the `[dependencies]` of its manifest), by `findIncludeFolders`. Defines come from the unit's own package only. It is one level: a package that depends on a dependent of the C package gets nothing. Proven by `testdata/packages/c-packages/basic`: a C file of `app` includes `<mini.h>`, and sees no `MINI_SECRET` while `c_mini` itself does.

### Tests

`testdata/packages/c-packages/<case>/` (cases `basic`, `two_versions`, `moved_tag`, `missing_tag`, `missing_header`, `outside`, `not_c_package`, `undeclared`, `upstream_in_root`, `upstream_without_tag`), run by the `packages` suite as `commands` cases. A case with a `repos/` folder builds git repositories from its folders (fixed author and date, `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`, `file://` URLs through `@REPOS@`), and `NARRA_C_LOG=1` shows what compiled. New steps in `selfhost/tests/src/packages.nra`: `delete` (remove a path, read-only checkouts included), `has` (a file has or lacks texts), `retag` (commit a folder and move a tag), `locked` (the lock holds the commit and tree of a tag). `testdata/lockfile/lock_c_upstream` and `lock_c_upstream_partial` cover the lock reader.

New codes: E1663, E1664, E1665 in `selfhost/src/driver/diagnostics.nra`, with explanations. Documented in SPEC sections 9 and 15, `compiler/README.md`, `docs/reference/errors.md` (generated) and the guide chapter 16 ("C libraries as packages").

## Proof

Not a fix node.

## Deviations

- Added the lock keys `c_git`, `c_tag`, `c_commit`, `c_content` and the three codes; the brief left both to the worker.
- Edited `selfhost/tests/src/packages.nra` and `gitfixtures.nra` (shared) for the new steps and for repositories in `commands` cases; `removeScratch` is replaced there by `Fixtures.removeFixtures` for every `commands` case, since a cache with checkouts is read-only.
- Added `testdata/lockfile/lock_c_upstream*` (not in the owned paths, new folders only).
- `isPackageFilePath` lives in `syntax/ast.nra` (the parser needs it), outside the brief's shared list for `selfhost/src/syntax`; `emit/emitexterns.nra` and `syntax/parser.nra` changed by a few lines for the same reason.
- `docs/reference/errors.md` regenerated with `--suite generated --update`.
- The text of the E1000 message for an unknown `[c]` key still lists `sources, include, defines`; `git` and `tag` are not in it, to keep its goldens.

## Stubs left

- None. Seed update: not needed (the compiler source uses nothing the seed lacks).

## Gotchas

- What a first real C package needs (zlib for `compress`):
  - A repository for the package, `narra.toml` as in the example above, `src/<name>.nra` with one export, tagged `v<version>` (or `<name>/v<version>` in a workspace repository).
  - The `sources` list must leave out every upstream file with a `main` (`example.c`, `minigzip.c`) and any test.
  - zlib builds from its tree with `-I.` and no generated header, so it fits. A library that needs a generated `config.h` does not: `[c] include` paths are all relative to the upstream, so a header kept in the C package's own folder is not reachable. Such a library needs `defines` only, or a way to add a package-relative include folder, which this node does not have.
  - The binding package (`compress`) lists `c_zlib` in `[dependencies]` and writes `from "c_zlib/zlib.h"`.
  - Source files are compiled with warnings off and the flags of the program, so a library that needs flags beyond `-D` and `-I` (`-fPIC`, `-std=`, `-pthread`) has no place for them yet.
  - `[c.os.<name>]` has no `git` or `tag`: one upstream per package.
- Relative header paths in `.deps` files are relative to the working directory of the build (from package-c-sources); an upstream checkout path is absolute, so a C package from the cache is unaffected.
- The checkout is read-only. A C compiler that writes next to its source would fail; object output goes to the object cache, so none of the compiles here does.
- `resolveDependencies` fetches the upstream of every version it reads, including versions a later minimum replaces; that costs one cached fetch each.
- `Process.run` has no timeout: a git fetch of a large upstream blocks the build until it ends.
- A moved tag with the checkout still in the cache is not noticed (the fetcher never asks git then). That is the rule for git dependencies too.

## Proposed learnings

- A test that needs a git repository builds it in the harness from folders, with `GIT_CONFIG_GLOBAL=/dev/null` and a fixed author, and names it by `file://` through `@REPOS@` | no test reaches the network and the hashes are stable | selfhost/tests | yes (the `packages` cases do it)

## Verification

- `sh selfhost/bootstrap.sh out`: stage 1 and stage 2 emit identical C (280858 lines).
- `NARRAC=out/narrac sh scripts/test.sh --suite packages --filter c-packages`: 10 cases pass.
- `sh scripts/test.sh`: `PASS, 11201 cases in 229s` (first run, no timing sample failed).

## Size

- Estimated M, actual M.
- Risk estimated medium, measured medium: the lock reader and the dependency loader changed for every build with dependencies, and the `commands` harness removes its scratch folder another way.
