# Log: async-timing-flake

- Base: feat/batch-13 at 11d255b0, compiler from `sh selfhost/bootstrap.sh`, sample built with `narrac build -release`.
- Load: 8 parallel loops of 25 runs each (200 runs), output compared with the `.out` file.
- Base runs: 2, 18 and 8 failures of 200 (three loads; the first with no extra load).
- Fix: first, raced and late sleep 100 ms (was 5), quick sleeps 200 ms (was 10). The 1 ms sleeps stay; each is compared with a 100 ms one.
- Branch runs: 0 and 0 of 200 under 8 loops; 0 of 200 with 16 extra busy loops.
- Mutation: removing `|| awaited->observers != 0` from the emitted C of the sample makes the output differ (shared cancelled, all gives [1, 0]).
