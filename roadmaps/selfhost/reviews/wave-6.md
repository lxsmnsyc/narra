# Wave 6

## Merged

- import-alias-rules: [#35](https://github.com/lxsmnsyc/narra/pull/35), merged by the maintainer. M, about 622k tokens, two review rounds plus one fix round after the last review.
- qualified-constructors: [#36](https://github.com/lxsmnsyc/narra/pull/36), merged by the maintainer. M, about 637k tokens, two review rounds (the second clean) plus a nit commit.

Both went to the maintainer because they changed files outside their briefs' paths in `compiler/internal/narra`.

## Plan changes

- `selfhost-check-bodies` now also ports E0421, E0423 and E0424 to `selfhost/src/scopes.nra` (the Narra checker reports only E0422 after #35). Its brief gained one acceptance line, a note and two shared paths.
- `bench/BASELINE.md` was re-recorded at 3fa2dd6 on a busy machine. `interp` is 0.393 s (was 0.237 s), as D55 expected until `perceus`.

## State

- Active: none.
- Next: `selfhost-check-bodies` (L, opus). Then `selfhost-check-calls` and `selfhost-check-patterns` can run in parallel.

## Not triaged

Two unclear messages from the #36 review, not added to the inbox yet: `Option.Some(...)` is read as the `std/option` namespace and asks for an import, and `const m: Maybe = None` says "cannot tell the type of None" instead of pointing to `Maybe.None`.

## Process lessons

- Both nodes' token counts were about twice the earlier M nodes. Each review round resumes the worker with its full context.
- `gh pr ready` is blocked by the permission classifier, so the maintainer marks PRs ready.
