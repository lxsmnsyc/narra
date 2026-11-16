# Outcome: docs-search

Search palette over the guide, the install page and the spec. One minisearch entry per section (h1 to h3), built at build time into `virtual:search-index` by `website/plugins/search.ts`.

## Sizes

- Index entries chunk: 358.6 kB raw, 106.7 kB gzip (3000 characters of text at most per section).
- minisearch chunk: 17.5 kB raw, 5.7 kB gzip.
- Loaded only when search opens: both chunks, 124 kB gzip together.
- Extra weight before search opens: 5.2 kB of CSS inlined in every page (1.4 kB gzip), plus the palette code in the layout chunk (7.4 kB raw, 3.4 kB gzip in total with the layout). The entries and minisearch are not in any preload or script tag of a page.

## Checked

`pnpm --dir website build` and `pnpm --dir website typecheck` pass. Served the build on port 4500 and drove it in Chrome:

- Before opening, the page requested none of the search chunks. Cmd+K loaded both.
- Cmd+K and `/` open the palette, the input has focus. Esc and a click on the overlay close it, and focus returns to the trigger.
- `uses alloc` lists the effects sections, `mutable` lists chapter 9 first, `install` lists the install page sections.
- Arrow keys move the active result (`aria-activedescendant`). Enter opens it as a client navigation (the page did not reload) and scrolls to the anchor.
- Dark and light theme looked right.

Not checked: phone width and reduced motion by eye (CSS only: icon-only trigger under 40rem, near-instant transitions with `prefers-reduced-motion`).
