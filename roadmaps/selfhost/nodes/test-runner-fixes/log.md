# Log: test-runner-fixes

Base: feat/batch-9 f62634c1.

- Item 1, not reproduced as a total 0-case PASS. Commands on base (compiler built from the seed, runner built by scripts/test.sh):
  `NARRAC=<narrac> sh scripts/test.sh --suite bench|errors|packages|emit` ran 6, 75, 4 and 2 cases per shard and printed PASS. `--repository /nonexistent` and `--filter zzzz` exit 1 with "no test case matches". What does show on base: with 8 shards, a suite of fewer than 8 cases prints `shard 7/8: 0 passed, 0 failed` lines, and the message for an empty plan does not name the filters. Fixes: the message names the flags, and the final line of test.sh prints the case total. Likely cause of the batch-7/8 sightings: `| tail -2` showed `shard 8/8: 0 passed, 0 failed` then `PASS in 0s`. A guard that failed a zero total was dropped: it broke `--list`, and an empty plan already fails every shard.
- Item 2, shows on base: the old testdata/io_files.nra compiled and run 6 times at once: output differed from io_files.out in 2, 1 and 0 of 3 rounds. On the branch: 0 differences in 3 rounds of 6; two `scripts/test.sh --filter io_files` at once both pass.
- Item 3: dead returns after panic in textops.nra removed (the checker accepts panic as the last statement); cutText panics with a message, function comment removed; `--suite generated --update` regenerated selfhost/src/stdtext.nra.
- New case runner/no-match-filters builds the runner and checks exit 1 and that stderr names `--suite bench --filter no-case-has-this-text`. On base the message is "no test case matches", so the case fails there.
