# Wave 10

## Landed

- selfhost-names-exported: [#45](https://github.com/lxsmnsyc/narra/pull/45), merged by the maintainer. M, about 144k tokens (worker 95k, fresh review 49k). Clean review, four nits: `std` and `ast` field names kept, `pkgIds` in `main.nra`, private `enumOf` in `match.nra`, mixed old and new names in comments. `pos` was kept; a rename to `position` stays open as a possible follow-up.
- effect-error-impl-name: direct commit d49ad08, landed by the orchestrator after `can-merge` passed. S, about 122k tokens (worker 76k, fresh review 46k). Both checkers print ``(through `Cat.Named.name`)``. The error text has no module prefix; the types dump `func=` keeps it. `narra emit` output is unchanged on all 44 samples.

## Plan changes

- selfhost-emit is broken into ten nodes plus the draft `contracts/emitter.md` (planner about 222k tokens). The planned first node was XL, so it became four M nodes: core, program, decls, bodies.
- The orchestrator changed two edges from behavior to hard. The wave-3 feature nodes wait for bodies, so program runs in wave 2. options waits for closures, which frees a wave-3 slot for traits.
- Must line: wave 6. Parallelism 3.1x.

## State

- Active: none. The plan is approved; the maintainer paused before selfhost-emit-core.
- Next: selfhost-emit-core. Stop for the maintainer's review of the contract when its PR is up.

## Process lessons

- Self-review workers skip the review step; spawn a fresh reviewer for direct-commit nodes, since `can-merge` needs it anyway.
- The auto-mode classifier blocks `gh pr merge` even when `can-merge` passes; the maintainer merges PRs.
