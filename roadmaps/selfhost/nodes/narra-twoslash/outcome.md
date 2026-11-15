# Outcome: narra-twoslash

## What was built

Hover popups and inline errors in the `narra` and `narra error` fences of the docs, from `narrac query --json hovers` (contract query-hovers v1), computed at build time.

- `website/plugins/twoslash.ts`: picks the fences (`narra` and `narra error`; `narra fragment` and `narra file ...` are skipped), writes each to a temp package (`narra.toml` plus `src/main.nra`), runs one process per distinct fence with a pool of 4, caches by content hash in `website/node_modules/.cache/narra-twoslash` (the key includes the narrac path, size and mtime), and writes the answer into the highlighted HTML. Columns are code points. A token is cut where a hover starts or ends, so a merged token still gets the right range.
- `website/plugins/markdown.ts`: the transform is async; it collects the fences, queries, then renders. A fence with no answer stays plain.
- `website/scripts/narrac.ts`: a stub with the shape of the contract. It guesses hovers from `const`, `let`, `function` and parameters, and takes the diagnostic from the `// error: L:C: message` line.
- `website/src/twoslash/client.ts` and `twoslash.css`: one popup element for the page (a scrolling `pre` would clip a popup inside it). It opens on pointer hover and on focus (no delay), closes on Escape, on leave (120ms grace so the pointer can cross to it) and on blur. Left and Right arrows in a focused code block move between tokens, so the page has no tab stop per name. Touch taps toggle. Placement is below the token, above when there is no room, clamped to the viewport, and follows scroll. Motion is opacity and transform only, 200ms in with `--ease`, 150ms out with `--ease-in`, driven by `data-open`, so it reverses mid-way. Reduced motion removes the transitions. Colors are tokens only. The inline error uses `--danger` and `--danger-soft`, under the line, with a wavy underline on the token.
- `DocPage.brm` calls `startTwoslash()`; `enhance.ts` copies the code without the error lines (`readCodeText`); `theme.ts` inlines `twoslash.css`.

## Switching to the real command

Set `NARRAC` to the built compiler (`NARRAC=/path/to/narrac pnpm --dir website build`), or put `narrac` on the PATH. With neither, the build prints one line (`twoslash: no narrac ...`) and renders plain highlighted code. The stub is `NARRAC=scripts/narrac.ts` (from `website`). Delete the cache folder if a compiler change keeps the same size and mtime.

## Checks

- `pnpm --dir website build` and `pnpm --dir website typecheck` pass, with the stub and without narrac.
- Stub: 135 hovers on the values-and-types chapter; 9 inline error lines there. The text of all 383 code blocks of the guide equals the plain build (error lines excluded).
- Checked in Chrome on a static preview: hover on `totalCents` opened the popup below the token, the right arrow focused the first token and showed its popup, Escape closed it.
- Not checked: the real compiler output, phone width, light theme, reduced motion in a browser.

## Build time and page weight (stub, 4 processes of node)

| | Build | Guide HTML (13 pages) | gzip |
| --- | --- | --- | --- |
| before (plain) | 3.6 s | 1,786 KB | 263 KB |
| after, cold cache | 8.9 s | 1,899 KB | 269 KB |
| after, warm cache | 3.4 s | same | same |

The cold time is the stub (node start per fence), so the real narrac should be faster. Page weight is +6% raw and +2% gzip. The client script adds a few hundred bytes of CSS and about 1.5 KB of JS.

## Vercel

`amazonlinux:2023` has no `cc`, `gcc` or `make`. `dnf install -y gcc` works as root and took 2 minutes in docker (not measured on Vercel). Plan, not done: the build runs `dnf install -y gcc` only when `cc` is missing, then `sh selfhost/bootstrap.sh` and sets `NARRAC`. If that is too slow or not allowed, use the wasm narrac from `scripts/wasm-build.sh`, run through node, or commit the cache folder's answers. `website/vercel.json` and `package.json` are unchanged. No deploy.

## Deviations

- `narra file ...` multi-file fences are skipped; only whole-file fences are queried.
- A plain `narra` fence whose answer has diagnostics shows no error (only `narra error` does).
- The Vercel plan is a finding only; the build does not install a compiler yet.
