# Log: website

- Renamed the branch to feat/website and reset it to feat/batch-16. No website/ existed, so the work was still to do.
- Read the brumate website and the twinkleplop json and typescript packages. Versions as in brumate's website; brumate 0.7.2 and @brumate/tsc 0.1.17 come from npm, since the website of brumate uses workspace links.
- Wrote website/src/highlight/narra.ts in define_grammar form. Names are told apart with the core promote_* passes.
- Wrote scripts/check-grammar.ts: runs editors/textmate/check.cjs (NODE_PATH to website/node_modules) and a twinkleplop pass over every narra fence.
- First build ran while the load average was above 30; later builds waited.
- Built: 18 pages. verify script found no broken link or anchor and all 285 narra fences highlighted.
- No browser extension was connected, so no visual check of the layout.
