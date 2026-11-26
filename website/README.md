# Narra website

The guide, the installation guide and the specification, built with brumate and highlighted with twinkleplop at build time.

- The pages are `docs/guide/*.md`, `docs/install.md` and `SPEC.md`, read at build time. The site holds no copies. [`plugins/markdown.ts`](./plugins/markdown.ts) renders them and points links between them at the pages of the site.
- The packages section is `docs/packages/README.md` and one page for each `packages/*/README.md`, at `/packages/<name>`. A new package folder with a README appears in the navigation, the index and the search with no edit to the site.
- [`scripts/check-links.ts`](./scripts/check-links.ts) runs at the end of the build. It checks each link and anchor of the package pages.
- [`src/highlight/narra.ts`](./src/highlight/narra.ts) is the Narra language of twinkleplop. It has the token classes of `editors/textmate/narra.tmLanguage.json`.
- [`scripts/check-grammar.ts`](./scripts/check-grammar.ts) runs in the build. It runs the TextMate check and tokenizes every `narra` fence with the twinkleplop language, and fails on an unknown token, an unterminated string or an open comment.

```bash
pnpm install
pnpm build     # writes the static site to dist/client
pnpm preview   # serves it at http://localhost:4173
pnpm dev       # serves the pages with live updates
```

Vercel deploys `main` only (`vercel.json`), with `website` as the root directory.
