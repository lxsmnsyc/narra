# changesets

A tool that releases the members of a Narra workspace. A contributor adds a small file under `.changeset/` that names the members a change touches and how big the change is. When it is time to release, one command turns the pending files into new versions, changelog entries and git tags.

It is a command line program with a library behind it. It has no dependency other than `std`.

## Running it

Run it from the workspace root, the folder whose `narra.toml` has the `[workspace]` table. The package builds one program, `changesets`:

```sh
narrac build -bin changesets -o changesets packages/changesets
./changesets status
```

Without a build, `narrac run -bin changesets packages/changesets <command>` does the same.

## Members

The members are the packages listed in `members` of the root `narra.toml`, with `*` as the last segment of an entry expanded to every folder that has a `narra.toml`, and `!` entries removed. A member takes part only when its `[package]` table has both a `name` and a `version`. A package with no version, such as the compiler itself, is left out.

A member depends on another member when its `[dependencies]` table has `name = { workspace = true }`.

## Commands

### add

```sh
changesets add <member>:<level> ... -m "<summary>"
```

Writes `.changeset/<three-random-words>.md` and prints its path. It takes arguments only and asks nothing. The summary comes from `-m` or `--message`, which may be given once. `changesets --help` and `changesets <command> --help` print the usage; a wrong argument prints the problem and the usage and exits with 2. The level is `patch`, `minor` or `major`. An unknown member, an unknown level, a member named twice, no member, or an empty summary is an error, and no file is written.

### status

```sh
changesets status [--since <ref>]
```

Lists the pending releases, one per line, such as `core 0.3.1 -> 0.4.0 (minor)`. It exits with 1 when a member changed since `<ref>` (default `main`) and no pending changeset names it, and prints `missing changeset: <member> changed since <ref>` for each. It also exits with 1 when a changeset file has a problem, or when git cannot compare with the ref.

A member counts as changed when a file under its folder differs from the point where the current commit left `<ref>`: committed changes, changes in the working folder and files git does not track yet. A file moved out of a member counts for that member. A ref that starts with `-` is refused. Its `CHANGELOG.md` does not count. A member that is released only because it depends on another member does not need a changeset of its own.

### version

```sh
changesets version
```

Applies every pending changeset file:

- Each released member gets its new `version` in its `narra.toml`. Only the `version = "..."` line of the `[package]` table changes. Comments, order and spacing stay as they were.
- Each released member gets a new section at the top of its `CHANGELOG.md`, which is created when missing.
- The applied files are deleted.

Every file is checked and every new text is built before anything is written, so a bad changeset changes nothing on disk. The new texts are then written next to their targets under a temporary name and renamed into place, so a file that cannot be written leaves the workspace as it was. The changeset files are deleted last. A `version` line may use double or single quotes, and a comment after it may hold quotes.

### tag

```sh
changesets tag
```

Creates the git tag `<name>/v<version>` at the current commit for each member whose version has no such tag yet, and prints the tags it made. It never pushes. A second run creates nothing.

## Changeset files

A file has a frontmatter block with one line for each member, then the summary:

```md
---
"core": minor
"docs": patch
---

Add streaming reads.
```

The frontmatter is only these lines, not YAML. Blank lines in it are ignored. The summary is the rest of the file and may have several lines. A problem is reported as `<file>:<line>: <message>` for any of: no `---` line at the start, a frontmatter that never closes or names no member, a line that is not `"member": level`, an unknown member, an unknown level, a member named twice in one file, or an empty summary.

## Bump rules

- Several changesets for one member: the highest level wins.
- A version below 1.0.0 follows the changesets rule for `0.x`: `minor` moves 0.3.1 to 0.4.0, `patch` moves it to 0.3.2, and `major` moves it to 1.0.0. From 1.0.0 on, `major`, `minor` and `patch` move that part and reset the lower ones.
- A member that depends on a released member through `{ workspace = true }` gets a `patch` release, and so does a member that depends on that one, along the whole chain. Its changelog has the line `Updated dependency <name> to <version>`. A member that already has a higher level from a changeset keeps that level and gets the line too.
- A version that does not read as `MAJOR.MINOR.PATCH` is an error. A prerelease suffix is dropped by a bump.

## Changelog

A release adds this at the top of the member's `CHANGELOG.md`, below its `# <name>` title:

```md
## 0.4.0

### Minor changes

- Add streaming reads.

### Patch changes

- Fix a crash on empty input.
```

The headings `### Major changes`, `### Minor changes` and `### Patch changes` appear only when they have a bullet. Each summary is one bullet, and later lines of a summary are indented under it. A bullet sits under the level its changeset gave, and the dependency line sits under patch.

## Library

The modules a program can import are `changesets/changeset` (parse and write changeset files), `changesets/workspace` (read members, replace a version), `changesets/release` (plan the releases), `changesets/changelog` (write sections), `changesets/git` (the git calls) and `changesets/commands` (the four commands, each returning the exit code and the lines to print).

## Not included

A config file with fixed or linked groups, and prereleases.

## Tests

`narrac run -sanitize -bin changesets-tests packages/changesets` runs the checks from the repository root. They cover the file parser (with a fuzz of random and damaged text), the version rules, a fixture workspace compared with golden files, and `tag` and `status` on temporary git repositories that use their own identity and ignore the user's git configuration. `scripts/test.sh` runs them in a full run.
