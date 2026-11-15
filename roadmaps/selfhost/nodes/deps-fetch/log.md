# Log: deps-fetch

- New `selfhost/src/depcache.nra`: `checkoutPath`, `fetchVersion`, `listVersions`, `checkTree`, and the test-only command `runDepsCommand` behind `narrac pkg --fetch` and `narrac pkg --versions`.
- Codes E0440-E0445 added to `diagnostics.nra`; `main.nra` has one dispatch row.
- Git runs as `sh -c 'GIT_TERMINAL_PROMPT=0 exec git "$@"'`: Process.run cannot set variables, and this keeps git from asking for a password. Exit code 127, or a failed start of sh, is E0440.
- A fetch clones with `--depth 1 --branch <tag>` into `<cache>/tmp/fetch.XXXXXX/work`, reads HEAD and HEAD^{tree}, deletes `.git`, writes the two hashes to `<tmp>/hash`, then one shell step moves the hash file to `<checkout>.hash` and the folder to `<checkout>`, then makes it read-only.
- The folder name is the sanitized text plus the xxh32 of the original, so `a/b` and `a_b` do not share a folder.
- Suite `deps` (`selfhost/tests/src/deps.nra`, helper `gitfixtures.nra`): nine cases against repositories built in a scratch folder. Fixture sources: `testdata/deps/members/*`.
- `scripts/linux/Dockerfile` installs git.
