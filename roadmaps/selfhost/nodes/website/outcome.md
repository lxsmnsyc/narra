# Outcome: website

## Built

- `website/`: a brumate site (brumate 0.7.2, vite 8) that prerenders 18 static pages: `/`, `/guide`, the 13 chapters, `/install`, `/spec` and `404.html`. `pnpm --dir website build` writes `website/dist/client`.
- `website/plugins/markdown.ts`: renders `docs/guide/*.md`, `docs/install.md` and `SPEC.md` at build time, so the site holds no copies. Links between these files point at site pages, other repo links point at GitHub, heading ids follow GitHub, and the context comments of SPEC.md are dropped.
- `website/src/highlight/narra.ts`: the Narra language in `define_grammar` form, with the token classes of `editors/textmate/narra.tmLanguage.json` mapped to twinkleplop names. Highlighting is build-time HTML, so pages ship no highlighting script and nothing shifts when they load.
- `website/scripts/check-grammar.ts`, run first by `pnpm build`: runs `editors/textmate/check.cjs` and tokenizes every `narra` fence with the twinkleplop language. It fails on an `error` token, an unknown token class, text without a token, or a string, template or comment still open at the end of a sample.
- `website/vercel.json` (repo-setup template, `main` only, `dist/client` output), `website/README.md`, `website/.gitignore`, `website/pnpm-lock.yaml`.
- Layout: skip link, focus-visible rings, sidebar of chapters that becomes a "Chapters" disclosure under 800px, outline of each page, previous and next links, light and dark themes, system fonts.

## Proof

Not a fix node.

## Deviations

- `README.md` and the root `.gitignore` (shared paths) are not touched: `website/.gitignore` covers `node_modules` and `dist`.
- `brumate` and `@brumate/tsc` use the npm versions (0.7.2, 0.1.17), since brumate's own website uses workspace links.
- The site has no search, copy buttons or live samples.
- No browser extension was connected, so the layout was not looked at on a phone or in a browser. The CSS has a narrow-screen layout that is untested visually.
- The first build ran while the load average was above 30.
- Home page and footer text are mine and short; the maintainer may want to rewrite them. The repository has no LICENSE file, so the footer names none.

## Stubs left

- None.

## Gotchas

- `pnpm build` needs `docs/`, `SPEC.md` and `editors/textmate/check.cjs`, which sit above `website/`. On Vercel keep "Include source files outside of the Root Directory" on (the default).
- The twinkleplop grammar is checked for known token classes only. Its colors are approximate where TextMate is finer: `uses` effects such as `IO` show as constants, and types after `:` are colored by shape.
- `ignoreCommand` in `vercel.json` watches `website`, `docs/guide`, `docs/install.md`, `SPEC.md` and `editors/textmate`.

## Proposed learnings

- A change to `narra.tmLanguage.json` needs the same change in `website/src/highlight/narra.ts` | the site highlights with its own grammar | editors/textmate | no

## Verification

- `pnpm build` in `website/`: textmate check 285 samples, 0 problems; twinkleplop check 285 samples, 0 problems; client and server builds pass; "Prerendered 18 pages".
- A script over `dist/client`: 18 pages, 0 broken links, 0 missing anchors, 285 `narra` fences all highlighted, only the classes boolean, class_name, comment, constant, decorator, function, identifier, keyword, number, operator, punctuation, string, string_escape, template, type.
- `pnpm preview`: guide chapter 200, unknown path 404.
- `sh scripts/test.sh --suite guide`: PASS, 329 cases.

## Vercel steps for the maintainer

1. Import the repo `lxsmnsyc/narra` on Vercel.
2. Set Root Directory to `website`. Keep "Include source files outside of the Root Directory" on.
3. Framework preset: Other. Build command, output directory and install command come from `vercel.json` and `package.json` (`pnpm build`, `dist/client`); Node 22 or newer.
4. Production branch `main`; `vercel.json` disables other branches.
5. Add the domain, then set `SITE_URL` if it is not the Vercel production domain (used for canonical and preview links).

## Size

- Estimated M, actual M.
