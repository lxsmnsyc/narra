# Website review fixes

Branch fix/batch-17-website. Checks: `pnpm --dir website typecheck` passes; `NARRAC=/tmp/claude-501/fix-narrac pnpm --dir website build` passes with no "gave no answer" line.

## Should fix

1. twoslash race (`website/plugins/twoslash.ts`). Fixed: the real promise (a deferred) is stored per key before queuing; a worker settles it, also when the query throws. Proof: `Promise.all([queryHovers(8 samples), queryHovers(the 8th)])`, cold cache. Before: A=8, B=0 and "narrac gave no answer for 1 sample(s)". After: A=8, B=1.
2. Radiogroups in InstallCommand and BenchChart. Fixed: toggle buttons in a `role="group"` with `aria-pressed` (simpler than roving radios). CSS selectors follow. Chrome: pressed states read true/false as expected.
3. ToolDemo tabs. Fixed: ArrowLeft/Right (wrap), Home, End, roving tabindex, tab ids, panel `aria-labelledby` the selected tab. Chrome: ArrowRight moved focus to apply-fix with tabindex [-1,0,-1,-1] and the panel label followed; End gave explain, Home gave check. The `fail` class was dead (no line starts with FAIL) and is removed.
4. BenchChart bars. Fixed: each bar `role="img"` with `aria-label` (for example "Narra 0.111 s"), tooltip kept; the hidden chart is `inert`. Chrome: 24 bars labelled.
5. Motion. Fixed: color, background-color, border-color and text-decoration-color transitions removed from styles.css, twoslash.css and search.css; `.progress` animation sits under `prefers-reduced-motion: no-preference`.

## Nits

- Search Enter opens `results[active]`, the list on screen.
- Snippets cut by code point at spaces, and skip a leading `<heading>. `.
- Copy failure shows an alert icon, "Copy failed" label and a danger color, in enhance.ts and InstallCommand.
- Outline indicator is placed again on resize and when fonts are ready.
- `website/scripts/narrac.ts` deleted; nothing referenced it.
- BenchChart comment and caption now say Apple M2.

Not run: snippet and Enter behavior were not exercised in a browser.
