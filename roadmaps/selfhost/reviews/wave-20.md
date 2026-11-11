# Wave 20 report: batch-7

Batch-7 ran on `feat/batch-7` as one story: take back batch-6's compiler cost, shrink the C the compiler writes, and build two D95 syntax features. Children landed as direct commits after a worker check and a fresh review; two reviewers did differential testing (old and new compiler on tens of thousands of generated programs).

## Landed

"Compiler" is `narrac emit selfhost`, instructions retired.

- place-ids: shared-write places as integer ids, and bodies with no write that can hit a pair are skipped. -0.37G. Review: 32,538 differential inputs, one wrong constant name in a message (aliased imports), fixed by the orchestrator after the worker stopped.
- shared-results-walk: the `from` check from 3.8% to 1.3% of the compiler. Review: 4,391 differential inputs, identical.
- return-trailer: 1,878 trailers to 141. Review found that a shadowed `panic` counted as ending a body (silent fall-off in C), and that the same name-only test let a shadowed `panic` narrow an `Option` on main. Fixed for every caller (E0969/E0970, narrowing).
- copy-check-macros: debug copy check as one-line macros, -2.5k lines of C. One seed update was enough.
- tree-shake and tree-shake-types: whole-program tree shaking of functions, constants, types, helpers, map functions and std runtime sections. `bench/fib` is 975 lines (1,096 before). Reviews: 21 and 228 probe programs identical in debug, release, sanitizers and the copy and leak checks.
- contracts: `requires` on integer parameters, proven at each call (E0984 to E0986). Three review rounds: `>` result types, LSP, the prover's sentinel bounds, parameter shadowing, wrapped bounds.
- contracts-ensures: `ensures <pattern>`, `!=` added by the maintainer (E0987, E0988). Review: no false fact; u64 `!= 0` fixed.
- build-conditions: `std/build`, folded `if`, host from `/usr/bin/uname` or `NARRA_HOST_TARGET` (E0989).
- build-targets: `@target` and `check --all-targets` in debug and release (E0990, E0991).
- bench-check-suite: the `bench` suite checks every benchmark program.
- End of batch: seed regenerated; docs for D95 made consistent.

Whole batch: compiler 2.47G to 2.18G on its own source, and 2.44G to 2.05G on the same input (-16%). The brief's target of 2.1G on the end-of-batch source is missed by about 0.08G; the source grew 7.6% with the four syntax features (about +0.13G). Peak memory 132 MB to 130 MB. Seed 186,158 to 177,598 lines.

## Whole-branch review

No defects where the children interact (17 probes: contracts in dropped branches, `@target` through generics, tree shaking with folded branches, trailers next to folded returns, the shared-write and `from` checks in dropped code, copy-check helpers). Three stale texts in `compiler/README.md` and D95 were fixed.

## Checks

- `sh scripts/test.sh` passes on the story branch, and `sh scripts/linux-check.sh` passes (Debian, gcc, sanitizers: 6,677 cases, 158 s), including the Linux golden of `build_os`.

## Process lessons

- Workers shared one scratchpad folder and two test runs wrote the same log. Give every worker a folder named after its node.
- `NARRAC=... sh scripts/test.sh --suite <name>` ran 0 cases and printed PASS (inbox). Do not trust a targeted suite that finishes in seconds.
- A rebase that stops on a conflict must not be followed by a test run in the same command: one run tested a half-merged file and its `tail` hid the failure. Check the rebase exit code first.
- Four parallel workers plus reviewers on an 8 GB machine pushed the load past 100; each test run took 2 to 3 times longer.
- Differential testing (old against new compiler on generated programs) is the right review for analyses that gain speed by skipping work.
