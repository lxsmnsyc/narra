# Wave 18 report: batch-5

Batch-5 ran on `feat/batch-5` as one story. Children landed as direct commits after a worker check and, for M and L nodes, a fresh review.

## Landed

Numbers are instructions retired, measured in alternating base and branch pairs on a loaded machine. "Compiler" is `narrac emit selfhost`.

- lexer-punctuation: compiler -3.0%.
- scope-map: compiler -1.4% (7 of 7 pairs).
- array-write-cold-paths: sieve 2.27G to 0.72G; compiler -2 to -3%. Box and map splits dropped for no gain.
- tail-loop-borrow: compiler -19.6%. The fresh review probed borrowed parameters under AddressSanitizer and found them sound.
- effect-values-skip: compiler -2.0%.
- map-helpers-split: compiler -5.4%.
- copy-builtin: `copy(x)`, E0975, the debug copy check. Two review rounds (generic instances, the remove-the-call fix, the E0290 hint).
- cli-skip-teardown: emit -2.2%, build -1.6%. The maintainer accepted the self-holding `Cell` as a stopgap (inbox).
- loop-unique-hoist: particles -42%, sieve a further -25%.
- parser-token-kinds: compiler -3.0%. Whole-tree parse, fmt and diagnostic diffs and 4000 mutated inputs were identical.
- cache-type-names-rest: compiler -3.4%.
- from-marker: the `from` result marker, E0976, E0977. Two review rounds. The maintainer accepted `shared-results = "unchecked"` in `selfhost/narra.toml` until shared-write-sweep; the compiler has about 185 shared returns, not the 30 to 50 the debate estimated.

Dropped: body-id-ranges (about 1% against a 4% spread; code kept on the local branch `feat/body-id-ranges`).

Whole batch: compiler 2.65G to 1.64G instructions (-38%), peak memory about 130 MB to 119 MB. Benchmark programs keep the same peak memory.

## Whole-branch review

The fresh review of the story branch found what the per-node reviews could not see:

- from-marker walked recursive enums again on every path, exponential in the number of fields. Fixed: each type is walked once.
- loop-unique-hoist copied a shared array before a loop that runs zero times, or whose write never runs, and the copy check reported it. Fixed in two steps: the uniqueness test waits for the first iteration, and the hoist applies only to writes at the top level of the body before any `break`, `continue`, `return` or `panic` (maintainer's choice).

## Process lessons

- Run the full check once per change, not once per step. Workers run the suites they touch plus one full run; reviewers reuse the worker's result unless the commit changed; the orchestrator runs `sh scripts/test.sh` and `sh scripts/linux-check.sh` once at the end, after the seed update. This batch ran the full suite after nearly every landing, mostly on trees a worker had already tested.
- Review the whole story branch before the pull request. Both blocking problems of this batch came from nodes that were each reviewed clean on their own.
- Measure instructions retired, not wall-clock, when workers share the machine; record the batch benchmark rows only on an idle machine.
- Workers that report and then wait on their own background jobs keep their worktrees locked. Stop them once their node lands.
