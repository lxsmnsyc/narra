# Outcome: deps-fetch

## Built

- `selfhost/src/depcache.nra`:
  - `checkoutPath(git, tag)`: `<cache>/git/<name>/<name>`, cache = `$NARRA_CACHE`, else `$HOME/.cache/narra`.
  - `fetchVersion(git, tag)`: returns `Fetched { path, commit, tree }`.
  - `listVersions(git, name)`: versions of `<name>/vX.Y.Z` tags when any exist, else of `vX.Y.Z` tags, sorted by number.
  - `checkTree(fetched, expectedTree)`: E0443 when the tree differs from the expected one (the lock's `content`).
  - `runDepsCommand`: the test-only commands `narrac pkg --fetch <git> <tag> [<tree>]` and `narrac pkg --versions <git> <name>`.
- Codes in `diagnostics.nra`: E0440 git not found, E0441 repository unreachable, E0442 tag missing, E0443 tree hash mismatch, E0444 no cache folder, E0445 checkout failed.
- Suite `deps` (`selfhost/tests/src/deps.nra`) with the helper `selfhost/tests/src/gitfixtures.nra`; fixture sources in `testdata/deps/members/`.
- `scripts/linux/Dockerfile` installs git.

## Acceptance

- `fetchVersion` returns path, commit and tree; a second call runs no git: case `fetch-twice` repeats the call with `PATH=/nonexistent`, where no process can start, and gets the same output. The commit and tree equal `git rev-parse <tag>^{commit}` and `^{tree}` of the fixture.
- `listVersions` member tags first, plain tags otherwise, sorted, other tags ignored: case `versions` (tags `a/v1.0`, `a/v2.0.0-rc1`, `release`, an annotated `a/v1.2.0`, and `v1.9.0`, `v1.10.0` give `1.0.0 1.2.0 1.10.0` for `a`, `1.0.0` for `b`, `0.5.0 1.9.0 1.10.0` for `c`).
- Two members on different commits go to two folders: case `two-members`.
- Errors: `error-git-missing` E0440, `error-unreachable` E0441, `error-tag-missing` E0442, `error-tree-mismatch` E0443, `error-no-cache` E0444, `error-checkout` E0445 (a tag on a tree object).
- No network: every repository is a `file://` URL, every cache a scratch folder.
- `sh scripts/test.sh --suite deps`: PASS, 9 cases.

## Deviations

- Touched outside the owned and shared lists: `selfhost/src/main.nra` (one dispatch row for the two test commands, as deps-manifest-lock did), `selfhost/tests/src/deps.nra` (the suite module; `registry.nra` imports it) and `scripts/linux/Dockerfile` (git in the Linux image).
- The tree check on a cached version compares the recorded hash with an expected one. Re-hashing the files would need git, and the brief says a cached version runs no git. A fetch writes `<checkout>.hash` (commit and tree) next to the folder; a missing or damaged file is E0443. The caller passes the lock's `content` to `checkTree`.
- Folder names carry the xxh32 of the original text (`a_v1.0.0-1a2b3c4d`) so two tags that sanitize alike do not share a folder.
- E0444 and E0445 are extra codes inside the reserved range.
- A checkout is read-only (`chmod -R a-w`), so removing the cache needs `chmod -R u+w` first. The suite does that.
- The tag is fetched by its full name (`git init`, `git fetch --depth 1 <url> refs/tags/<tag>`, `git checkout FETCH_HEAD`), not with `clone --branch`, which prefers a branch of the same name. Case `tag-beats-branch`.
- Two fetches of one version at once: the loser finds the folder in place and returns the recorded hash instead of an error.
- Only POSIX tools are used (`sh`, `mv`, `chmod`, `rm`, `mktemp`); Windows is not covered.
- A fetch killed by SIGKILL leaves `<cache>/tmp/fetch.*`; nothing sweeps it yet.
- The contract text says `<sanitized url>/<sanitized tag>`; the names now end in `-<xxh32>`. Contract change for the orchestrator: say so in version 2.
- `checkoutPath` keeps its contract name; it is not renamed to a verb.
- The suite passes `GIT_CONFIG_GLOBAL=/dev/null` to narrac so a user's git config cannot change a case.
- Read-only check in the suite uses the mode bits (`find -perm -u+w`), so it also holds as root in the Linux image.

## Stubs left

- `deps-resolve` and `package-dependencies` call these functions; nothing does yet.

## Gotchas

- Process.run cannot set environment variables or keep git from prompting, so git runs under `sh -c 'GIT_TERMINAL_PROMPT=0 exec git "$@"'`.
- Narrow `Result` values with `is` before reading `.value`; a function that returns a parameter inside `Ok(...)` needs `from <parameter>`.
- The seed is not regenerated (batch rule).

## Contract

`dependencies` version 2 holds as written: `checkoutPath(git, tag)` is in `depcache.nra`. Additions to record: `fetchVersion`, `listVersions`, `checkTree` and `Fetched`.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` after the rebase: every shard passed except `programs/async_api_all_cancel.nra/debug`, a timing case that failed at load average 80; `--suite programs` then passed (713 cases).
- `--suite deps`: PASS, 10 cases. `--suite generated`: PASS, 23 cases. `--suite guide`: PASS, 330 cases.

## Size

- Estimated M, actual M.
