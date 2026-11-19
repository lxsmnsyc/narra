# Outcome: docs-structure-review

## Built

- `roadmaps/selfhost/nodes/docs-structure-review/proposal.md`: the current pages, the findings with evidence, the proposed structure (Install, Guide, Reference), 17 changes marked before or after the release, and 10 follow-up nodes with sizes. It waits for the maintainer's approval (gate).
- Before the release: one node, `docs-release-pass` (M). It fixes chapter 1, the guide index, three TypeScript comparisons and the install page. It also fixes the command name in `SPEC.md`, the README status and the home page evidence link.

## Proof

Not a fix node.

## Deviations

- The brief's lead that `narrac query interface` can generate the std reference does not hold yet: it answers `unknown module` for every std module, and builtin modules have no source. The proposal adds node `std-interface-query` before `std-reference`.

## Stubs left

- None.

## Gotchas

- `website/scripts/vercel-build.sh` builds the site without `narrac` when `cc` is missing. Generated reference pages should be committed and checked by a test, not built on Vercel.
- Node `narrac-version` already changes `docs/install.md` to `narrac --version`. `docs-release-pass` should start after it lands.
- `vercel.json`'s `ignoreCommand` lists the paths that trigger a deploy. A new `docs/reference/` folder must be added there.

## Proposed learnings

- Docs claims about a `narrac` command are checked by running it on the seed build | the `query interface` lead was wrong when checked | `docs/`, `website/` | no

## Verification

- `narrac` built from `seed/narrac.c`; `query interface`, `explain`, `builtins`, `--version` and bare `narrac` run by hand (see `log.md`).
- `python3 <plain-writing>/scripts/check_prose.py` on `proposal.md`, `outcome.md` and `log.md`.
- `pnpm --dir website build` not run: no docs or site file changed.

## Size

- Estimated M, actual M.
