# Log: proven-operands

- Branch renamed to feat/proven-operands, reset to feat/batch-6.
- Facts about integer locals (low, high, nonzero) live in checker.localNames under `fact|<local>` keys, next to the variant narrowing. endNarrowing ends both, so assignment, `mutable` pass and loops end a fact the same way.
- Choice: a field path (`size.count`) is not narrowed. The hint says to copy it into a local.
- Choice: a `const` local set from a constant is that constant (`const items = 7`); a top-level `const` built from literals with + - * << is one too.
- Choice: a name an arrow function captures by reference is not narrowed inside the arrow (isCapturedByArrow).
- Sites rejected: compiler 3 (hash.rotateLeft32 twice, check.alignUp), tests runner 1 (shard split), std 1 (Random.nextInt span); samples 4 files (coroutine_tests, std_test, patterns_tuples, wrapping), guide 3 snippets, SPEC 1 snippet.
- Runtime: narra_div/rem/shl/shr (and sized forms) lost the check and the `where` argument. testdata/panics/divzero.nra removed.
