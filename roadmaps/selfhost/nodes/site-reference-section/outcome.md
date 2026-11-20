# Outcome: site-reference-section

## Built

- `/reference` index from `docs/reference/README.md`, route `website/src/routes/reference/index.brm`.
- `website/src/routes/reference/[...page].brm`: serves every other Markdown file under `docs/reference/`, at any depth, at `/reference/<path without .md>`. An unknown path answers 404.
- `website/src/pages.ts`: `referenceIndex`, `referencePages`, `findReference`, from `import.meta.glob('../../docs/reference/**/*.md')`.
- `website/plugins/markdown.ts`: links to `docs/reference/*.md` become site links (`README.md` becomes `/reference`).
- `website/plugins/search.ts`: indexes `docs/reference/README.md` and every other `.md` under `docs/reference/`.
- `website/vercel.json`: `:/docs/reference` in `ignoreCommand`.
- Header: Install, Guide, Reference. `Reference` is also current on `/spec`, whose URL is unchanged. The search hint names the reference.
- README lists Std, Commands, `narra.toml` and Errors as plain text marked "Coming" (no link, since the page does not exist), and links Spec.

## Deviations

- None from the brief. The site has no front matter support; a page's title is its first `# ` heading.

## Gotchas

How a later node adds a reference page:

1. Create `docs/reference/<name>.md` (or `docs/reference/<dir>/<name>.md`). No other file changes.
2. Start it with one `# Title` heading. The title, the first paragraph (the page description) and the `##`/`###` headings (outline) are read from the Markdown. No front matter.
3. Link it from `docs/reference/README.md` with a relative link such as `[Commands](commands.md)`, replacing the "Coming" line. The build only writes a page that something links to, so an unlinked page is not prerendered.
4. Links between reference pages and to the guide use relative `.md` paths; they are rewritten to site URLs.
- Pages get no sidebar and no pager. A generated page tree (`std/<module>`) works, since the route accepts nested paths.
- Pages with `narra` fences get hovers from `narrac` when `NARRAC` is set, as in the guide.

## Verification

- `pnpm --dir website build`: passes. `pnpm --dir website typecheck`: passes.
- A temporary `docs/reference/std/option.md` linked from the README was prerendered at `/reference/std/option` and found in the search index; removed afterwards.

## Size

- Estimated S, actual S.
