# Log: parser-token-kinds

- Base cc7baebe (feat/batch-5). Machine loaded (load average 22 to 33).
- Design: Token gets `symbol: int`, set in `emit` from `SYMBOLS_BY_TEXT` (keywords and punctuation only, else 0). `isAt`, `accept`, `expect` and `parsePatternList` take a `SYMBOL_` constant. `BINARY_LEVELS` holds symbols. The `expect` error text is looked up from the table.
- Proof: `/usr/bin/time -l <narrac> emit selfhost` on the same input, base and branch narrac both built with `sh selfhost/bootstrap.sh`. A first attempt was discarded: the base build raced with my edits and the branch dir held a stale binary.
- 8 alternating pairs, instructions retired:
  - base: 1828195332 1811304433 1804034337 1810970627 1814367669 1807641554 1793606585 1795520503 (median 1.8095e9, range 1.9%)
  - branch: 1782611245 1757423707 1749964279 1763868280 1748487883 1753777622 1754866645 1755481001 (median 1.7545e9, range 2.0%)
  - median gain 3.0%; branch won 8 of 8 pairs; the smallest pair gain is 2.2%.
- The base and branch compilers emit identical C for `selfhost` (cmp).
