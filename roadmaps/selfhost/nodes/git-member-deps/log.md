# Log: git-member-deps

- Branch feat/member-deps reset to feat/batch-19 (8a5d7abc).
- Cache tag bug proof on the base: case `resolve-member-cache` (fixture `testdata/deps/workspace_member_cache`, repository with tags `v1.2.0` and `json/v1.2.0`; the plain tag is cached first with `narrac pkg --fetch`). Command: `NARRAC=out/narrac sh scripts/test.sh --suite deps --filter resolve-member-cache` on 8a5d7abc. Result: FAIL, the lock has `tag = "v1.2.0"` where a clean machine picks `json/v1.2.0`.
- Review fix, item 1: case `resolve-member-offline` (plain-tag repos a and b; a 1.0.0 needs b 1.0.0, the app needs a 1.0.0 and b 1.1.0; the second run has PATH=/nonexistent). On commit 1e90e7f3 (before the fix): FAIL, `error E0440: git cannot run`. After the fix (a cached plain tag is taken when the lock has an entry of the same name and URL with a plain tag): PASS. `resolve-member-cache` still passes.
- Added `resolve-member-sibling` (E0486 for a sibling the checkout lacks), a `break` in the sibling search, and a comment on the `workspace` skip in `pushRequirements`.
