# Log: lexer-punctuation

- Branch feat/lexer-punctuation reset to feat/batch-5 (48823a59). No node work existed before.
- Proof, macOS, `/usr/bin/time -l <narrac> emit selfhost`, instructions retired, 5 runs each side interleaved, load average about 5 (shared machine):
  - base (48823a59, narrac from `sh selfhost/bootstrap.sh`): 2626583582, 2628638856, 2628747558, 2629553259, 2674455446. Median 2628747558.
  - branch: 2544306695, 2548384093, 2548668842, 2554173572, 2557173765. Median 2548668842.
  - Gain 3.0%; base spread without the one outlier is 0.1%.
- Change: `PUNCTS` became `PUNCTS_BY_FIRST_BYTE`, a `Map<int, Array<string>>` const, longest first per byte. `lexPunctuation` looks up `peek(lexer, 0)` and tries only those candidates.
- Formatted with `narrac fmt`. Seed not regenerated; selfhost builds with the committed seed and the bootstrap fixed point holds.
