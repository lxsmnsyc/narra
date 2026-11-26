# Log: docs-packages

- Checked first: no `docs/packages`, no packages route, no `[c]` in `narra-toml.md`. Nothing built yet.
- `pnpm build` failed on the base (`feat/batch-26`) before any change: `check:grammar` rejected the fence `narra file describer wrap.c` in guide chapter 16 (no language for `.c`). Fixed in `website/src/highlight/fences.ts`: `.c` and `.h` files are `text`.
- Pages come from a second `import.meta.glob` in `website/src/pages.ts` (`packages/*/README.md`). Search lists the folders of `packages/` at build time. Static pages come from the links on the index, which lists every package from the same glob.
- `rewriteLink` in `website/plugins/markdown.ts`: a link to `packages/<name>/README.md` goes to `/packages/<name>`, one to `docs/packages/README.md` to `/packages`. Other repository files link to GitHub, as before.
- Link check: `website/scripts/check-links.ts`, run at the end of `pnpm build`. It reads the HTML of the package pages and checks pages and anchors.
- `vercel-ignore.sh`: added `:/docs/packages` and `':(top,glob)packages/*/README.md'`. The glob form was checked with `git diff --name-only` on the commit that last changed `packages/http/README.md`.
- `docs/README.md` does not exist, so nothing to edit there.
