# Outcome: docs-packages

## Built

- Packages section on the site: `/packages` (index, from `docs/packages/README.md` plus a generated list of every package with its first paragraph) and `/packages/<name>` for each `packages/*/README.md` (changesets, http, websocket, yaml). Pages are read by a glob in `website/src/pages.ts` (`packagePages`, `findPackage`); routes are `website/src/routes/packages/index.brm` and `[page].brm`. A new `packages/<name>/README.md` appears in the sidebar, the index and the search with no site edit.
- Navigation: a "Packages" link in `website/src/routes/_layout.brm`, a packages sidebar and kicker in `website/src/components/DocPage.brm`.
- Search: `website/plugins/search.ts` lists the folders of `packages/` that have a README.
- Links: `rewriteLink` in `website/plugins/markdown.ts` maps `packages/<name>/README.md` and `docs/packages/README.md` to site routes. Other repository files link to GitHub, as before. `website/scripts/check-links.ts` runs at the end of `pnpm build` and checks pages and anchors of the package pages.
- `website/scripts/vercel-ignore.sh` deploys on a change to `docs/packages` or to `packages/*/README.md`.
- `docs/reference/narra-toml.md`: new `[c]` section (`sources`, `include`, `defines`, `[c.os.<Os>]`, errors), "C packages" (`git`, `tag`, lock keys `c_*`, E1663, E1664), and "The from of an extern block" (library name, `./file.c`, `<package>/<path>`).
- Guide chapter 8: one paragraph pointing to the packages section.

## Deviations

- `website/src/highlight/fences.ts`: `.c` and `.h` file fences are `text`. `pnpm build` failed on the base because chapter 16 has a `narra file describer wrap.c` fence (from package-c-sources).
- `docs/README.md` does not exist, so it was not edited. `docs/reference/README.md` and the package READMEs were not edited either.
- Added `website/scripts/check-links.ts` and a `check:links` script in `website/package.json`; `website/README.md` mentions both.
- No `docs` suite checks links, so no change under `selfhost/tests`.

## Stubs left

- None.

## Gotchas

- The site has no C highlighter; C samples render as plain text.
- The static pages of a package exist because the index links to them. A package without a README gets no page.
- `website` has no linter or formatter script; none ran.

## Proposed learnings

- None.

## Verification

- `pnpm build` in `website/` with `NARRAC` set: grammar check 0 problems, `Prerendered 94 pages`, `links: 5 package pages, 0 problems`.
- `pnpm typecheck` in `website/`: clean.
- `dist/client/packages/{index,changesets,http,websocket,yaml}`; the search index holds `/packages/yaml`.
- `git diff --name-only` with `:(top,glob)packages/*/README.md` on the last commit that changed `packages/http/README.md` lists it.
- `sh scripts/test.sh`: `PASS, 11201 cases in 345s`.

## Size

- Estimated S, actual S.
- Risk estimated low, measured low.
