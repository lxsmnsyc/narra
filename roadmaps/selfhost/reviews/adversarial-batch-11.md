# Adversarial review of batch-11 (fresh-context agent), with dispositions

Nine findings; all applied as frontmatter changes and "Plan review" notes in the briefs.

1. The trait order in the batch brief was not encoded, and plan.py scheduled the opposite. Now generic-implement goes first; distinct-traits and dynamic-traits depend on it (hard). The batch brief's wave lists defer to `plan.py ready`.
2. generic-implement: the emitter does not work unchanged. `writeDefinition` (emittraits.nra:249-251) looks up user blocks by key, so generic members would fall back to derived functions. emittraits.nra added; the brief names the checker and emitter changes and records `findImplementation`.
3. generic-implement: body.nra, query.nra and declsdump.nra read `checker.implementations` by block. Added to shared.
4. dynamic-traits: kept `hard` on generic-implement; it builds tables on `findImplementation`, whose result type generic-implement's outcome records.
5. contract-logic: missing edge to array-patterns (both change the Pattern enum and contracts.nra). Added.
6. inline and std-function-values both change `emitIdent` and `readNodeType`. inline now depends on std-function-values, which names one emitter context helper.
7. inline and lazy-values both change `emitReachableBodies`. The second to land keeps both rules and reruns the tree-shake checks.
8. lazy-values: emitchecks.nra added; the lazy name's type is `T` with a flag, not a new `Lazy<T>` type (orchestrator decision).
9. inline: the async case of E1031 cannot be tested before async parses; moved to the async node.
