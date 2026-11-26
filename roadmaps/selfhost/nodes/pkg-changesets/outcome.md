# Outcome: pkg-changesets

## Built

`packages/changesets`, a workspace member (name `changesets`, version 0.1.0, tag `changesets/v<version>`) with a README at `packages/changesets/README.md`. It has two binaries: `changesets` (entry `cli.run`) and `changesets-tests` (entry `checks.run`).

CLI, run from the workspace root (`narrac run -bin changesets packages/changesets <command>`, or build it with `narrac build -bin changesets -o changesets packages/changesets`):

- `changesets add <member>:<level> ... -m "<summary>"` writes `.changeset/<adjective>-<noun>-<verb>.md` and prints its path. Arguments only. An unknown member or level, a member named twice, no member or an empty summary exits 1 and writes nothing.
- `changesets status [--since <ref>]` prints each pending release (`core 0.3.1 -> 0.4.0 (minor)`) and exits 1 for a bad changeset file, for a git failure, or for `missing changeset: <member> changed since <ref>` (default ref `main`). A member changed means a file under its folder differs from the merge base with the ref (committed, working folder or untracked); its `CHANGELOG.md` is ignored. A member released only through a dependency needs no changeset.
- `changesets version` applies every pending file (new texts are written to `<name>.changesets-tmp` files and renamed into place, changesets deleted last, temps removed on failure): replaces only the `version` line of `[package]` in each released `narra.toml`, writes a section at the top of `CHANGELOG.md` (below a `# <name>` title, created when missing), deletes the files. All files are checked and all new texts built before any write.
- `changesets tag` creates `<name>/v<version>` lightweight tags that do not exist, prints `created tag ...` or `no tags to create`, never pushes.
- The command line is parsed with `std/args`: `-m` or `--message` once, `--since <ref>`, `--help` prints the generated usage (exit 0), no command, an unknown command or flag, a repeated flag or a missing value print the problem and the usage and exit 2. A `--since` ref that starts with `-` is refused.

Rules as built: highest level per member wins; `0.x` follows semver parts (`minor` 0.3.1 to 0.4.0, `patch` 0.3.2, `major` 1.0.0), via `std/semver`; a member depending on a released member through `{ workspace = true }` gets a patch release along chains, with the note `Updated dependency <name> to <version>` (also added when the dependent has its own changeset, which then keeps its higher level). Changelog sections `### Major changes`, `### Minor changes`, `### Patch changes` appear when they have a bullet; later lines of a summary are indented under the bullet.

Changeset file problems are returned as `<file>:<line>: <message>` for: no leading `---`, unclosed or empty frontmatter, a line that is not `"member": level`, an unknown member, an unknown level, a member named twice, an empty summary. Every file's first problem is listed and nothing on disk changes.

Library modules (exported through `[exports]`): `changesets/changeset` (Level, Entry, Changeset, Problem, parse, load, render), `changesets/workspace` (Member, loadWorkspace, replaceVersion), `changesets/release` (Note, Release, planReleases), `changesets/changelog` (renderSection, addSection), `changesets/git` (runGit, listTags, createTag, listChangedFiles), `changesets/commands` (addChangeset, showStatus, applyVersion, createTags, each returning `Outcome { exitCode, lines }`). Private: `cli`, and the test modules `checks`, `checkkit`, `checkparse`, `checkcommands`.

Tests (`narrac run -sanitize -bin changesets-tests packages/changesets`, 148 checks): parser valid and bad files, a fuzz of 3000 random texts and 2000 damaged valid files, manifest edit, changelog text, bump rules, `version` on `fixtures/workspace` compared with `fixtures/expected`, bad file leaves the folder unchanged, `add`, `tag` on a temporary git repo (three tags, none on the second run, then three after `version`), `status` on a temporary git repo (untracked change, committed change, covered by a changeset, unknown ref, a file moved out of a member, an accented path, a ref with a dash), a member folder that cannot be written (nothing changes), the command line parse. Git in tests gets its own name and email and `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`; repos live in a `mktemp -d` folder that the run removes.

## Deviations

- `scripts/test.sh` now runs `narrac run -sanitize -bin "$name-tests" "$package"` for every package, because this package has two binaries. http, websocket and yaml already name their test binary `<name>-tests`. Shared path in the brief.
- Root `narra.toml`: `packages/changesets` added to `members` (shared path).
- The root package of the workspace is not a member for releases; only the `members` entries are.
- Members without a `version` (selfhost, selfhost/tests) are left out, so a changeset naming them is an unknown member.
- Plain `git tag` with `-c tag.gpgSign=false`, so a user's signing setting does not turn the tag into a signed one that needs a key.
- The fixture and golden files were created with a shell heredoc, against the one-plain-command rule. Their content was kept after review.

## Stubs left

- None.

## Gotchas

- `status` on this repository against `main` reports every package: they have no changesets and `main` does not hold them yet.
- Applying `version` and committing makes `status` report the member as changed (the manifest changed) unless a changeset still names it; a CI check should run before `version`.
- The git functions run with the user's git configuration. Only the tests isolate it.
- `is` binds names only as the whole condition, so combined conditions are nested ifs; `from` cannot be a parameter name; `.withDefault` is a function of `std/option`, not a method.

## Proposed learnings

- A package with more than one binary names the test binary `<package>-tests` and `scripts/test.sh` selects it with `-bin` | one rule for every package | packages/ | no

## Verification

- `out/narrac check .` and `out/narrac fmt --check .`: clean.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 ASAN_OPTIONS=detect_leaks=0 out/narrac run -sanitize -bin changesets-tests packages/changesets`: `148 checks, 0 failed`.
- `sh scripts/test.sh`: `test.sh: PASS, 11043 cases in 436s`.

## Size

- Estimated M, actual M.
- Risk estimated medium, measured medium: new package and a change to the shared test script; git behavior is covered on temporary repos only.
