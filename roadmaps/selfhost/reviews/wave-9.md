# Wave 9

## Merged

- selfhost-check-closures: [#42](https://github.com/lxsmnsyc/narra/pull/42), merged by the maintainer. M, about 316k tokens. Review found committed build output (`compiler/narrac`, `narrac.dSYM`); the branch was rewritten so the blobs never reached main, and `.gitignore` covers both now. Two Array-builtin samples were deferred to builtins.
- selfhost-check-builtins: [#43](https://github.com/lxsmnsyc/narra/pull/43), merged by the maintainer. M, about 274k tokens. One should-fix (never-skip the deferred closures samples), two nits.
- typesdump-targs-spacing: direct commit 8957591, landed by the orchestrator after `can-merge` passed (fresh review clean, fails on main and passes on the branch). S, about 90k tokens.
- selfhost-check-effects: [#44](https://github.com/lxsmnsyc/narra/pull/44), merged by the maintainer. M, about 291k tokens. Two should-fixes; one became the effect-error-impl-name node.

## Results

The self-hosted checker is complete. TestSelfhostBodies compares 352 and skips 30, all "does not load"; it fails on any `unsupported:` line. E0425 and duplicate constants were ported in #44 (ImportName gained `pos`).

## Plan changes

- New nodes: typesdump-targs-spacing (done), selfhost-names-exported (must, M, before selfhost-emit), effect-error-impl-name (should, S).
- selfhost-emit now depends on selfhost-names-exported.
- Both types dumps end each body with an `effects module=... <set>` line.
- Inbox: the targs, exported-names and E0425 items are resolved or turned into nodes.

## State

- Active: none. Paused by the maintainer.
- Next: selfhost-names-exported and effect-error-impl-name (merge the rename first). The effect-error text proposed is `Cat.Named.name` (`<Type>.<Trait>.<function>`); confirm before starting.

## Process lessons

- Workers must check `git status` before committing; the closures worker committed a binary and its debug symbols.
- `can-merge` for a direct commit needs a fresh review and a fails-on-base proof; the orchestrator can produce the proof cheaply by running the new golden on main.
