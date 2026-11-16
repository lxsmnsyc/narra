# Wave 27 report: batch-19, workspaces

Batch-18 (PR #123) merged on 2026-11-16 and is set `done`. Batch-19 is the rest of p4-packages, on the local branch `feat/batch-19` in the worktree `../narra-batch-19`.

## Still to do

- [ ] workspace-loader and git-member-deps: workers running (wave 1). Land workspace-loader first; git-member-deps rebases on it.
- [ ] Fresh review of each, then squash onto `feat/batch-19`.
- [ ] Wave 2: workspace-commands and workspace-lsp, in parallel.
- [ ] Wave 3: repo-workspace.
- [ ] Seed update if `selfhost/src` changed, full `sh scripts/test.sh`, `pnpm --dir website build`, one fresh review of the whole branch, then the batch-19 pull request for the maintainer.
- [ ] Set `workspaces` done with `fulfilled_by: batch-19`.
