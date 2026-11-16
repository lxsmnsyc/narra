# Wave 27 report: batch-19, workspaces

Batch-18 (PR #123) merged on 2026-11-16 and is set `done`. Batch-19 builds the rest of p4-packages on the local branch `feat/batch-19`, in the worktree `../narra-batch-19`.

## Landed

Each child landed as one commit after a worker check and a fresh review:

- workspace-loader: member dependencies, one lock at the root, E0484 and E0485. The review found that a git package deeper in the graph could share a member's name; E0450 now covers it.
- git-member-deps: member tags, sibling requirements, E0486 and E0487, and the cache tag fix from the inbox. The review found that the fix made a locked rebuild call git. A cached plain tag is now used when the lock shows the name tagged plainly.
- workspace-lsp: the language service over members. The review found load errors published on a file that does not exist; that is fixed.
- workspace-commands: root commands, E0488 to E0490. The member cycle got its own code, E0490, instead of sharing E0489. Root `test` follows the order of the `members` entries.
- repo-workspace: the root `narra.toml`. The review was clean.

The seed was regenerated after the last child. The umbrella node `workspaces` is `done` with `fulfilled_by: batch-19`.

## Orchestrator rulings

- E0490 for a `workspace = true` cycle. `fmt` and `query` skip the cycle check.
- Root `test` follows the order of the `members` entries.
- Root `query` locations carry the member folder.
- workspace-lsp maps member files through `overrides` in `lspnavigation.nra`. Go to definition into git dependencies stays an inbox entry.
- The double body check in the language service (a body-only edit in a member checks its bodies again in each dependent) is deferred and recorded in the workspace-lsp outcome.

## Still to do

- [ ] Land `fix/batch-19-review`, the fixes from the whole-branch review (docs, query and selection error files, E0450 URL, the shared walk in the language service, duplicated helpers).
- [ ] Regenerate the seed again. Run the full checks: `sh scripts/test.sh`, bootstrap, `pnpm --dir website build`, `sh scripts/linux-check.sh`.
- [ ] Push `feat/batch-19` and open the batch-19 pull request for the maintainer.
