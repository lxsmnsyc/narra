# Wave 8

## Merged

- selfhost-check-calls: [#41](https://github.com/lxsmnsyc/narra/pull/41), merged by the maintainer first. L, about 383k tokens. One fresh review (one should-fix: a call case that stopped loading was skipped, not failed) and one fix round.
- selfhost-check-patterns: [#40](https://github.com/lxsmnsyc/narra/pull/40), merged by the maintainer second. M, about 431k tokens. One fresh review (one should-fix: no golden for leaking narrowing or the 10-case limit, two nits), one fix round, then a merge of main after #41 with conflicts in check.nra, exprs.nra, body.nra and typesdump.go. The conflict resolution was not reviewed again; the full check passed on it.

## Results

TestSelfhostBodies compares 245 and skips 122 (wave 7: 143 and 202). No compared input differs. Skips by node: builtins 54, closures 27, does not load 30, constant declared twice 6, calls 1 (E0425, needs ImportName positions in the parser).

## Plan changes

- `Ref.Variant` from patterns was folded into `Ref.Constructor` from calls.
- `func=<Type>.<Trait>.<function>` replaces the C name with hash in the types dump for trait calls into implement blocks (Go and Narra).
- Edits outside briefs, accepted: scopes.nra exports `Namespace` and `findNamespace` (calls); ctors.nra exports `namesEnum` and `findEnumName` (patterns).
- E0425 stays a stub. No node owns it yet; it needs a parser change.

## State

- Active: selfhost-check-closures (M, sonnet).

## Process lessons

- Two parallel nodes on about ten shared files cost one conflict-heavy merge (about 50k tokens) but saved a wave. The review on the earlier branch flagged the conflict points ahead, which made the merge brief precise.
- The plain-writing hook rejects commit subjects over 72 characters; the roadmap's `<id>: <title>` format overflows for long titles. Workers shorten the subject.
- A finished worker can keep re-notifying from leftover timers; stop it with TaskStop.
