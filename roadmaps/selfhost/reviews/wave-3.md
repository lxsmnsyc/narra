# Wave 3

## Merged

- poly-recursion: [#7](https://github.com/lxsmnsyc/narra/pull/7)
- hashed-cnames: [#8](https://github.com/lxsmnsyc/narra/pull/8)
- box-recursive: [#10](https://github.com/lxsmnsyc/narra/pull/10)
- io-std: [#9](https://github.com/lxsmnsyc/narra/pull/9), plus hotfix [#11](https://github.com/lxsmnsyc/narra/pull/11)

## Process lesson

#9 merged after #10 without a rebase. Each passed review alone, but #9 built enum values by hand and #10 had made payloads heap pointers, so main broke until #11. New rule in GOAL.md Decisions: rebase and rerun the fast check before merging a second pull request of the same wave. Nodes that share an emitter layout change should also be listed in merge order in the wave plan.

## Contracts

- `module-graph` and `trait-model` are v3, frozen (hashed C names; `__` and edge `_` are allowed). Change requests are recorded in `contracts/`. Affected nodes: opaque-distinct, c-interop, query-cli and the selfhost nodes use the new naming.

## Briefs changed

maps, strings-std, opaque-distinct, c-interop, effects, json-diagnostics, selfhost-spike. Each now says what changed: C names in `cnames.go`, heap enum payloads, `unknown_module.nra` prefix match, and new std modules. perceus was already updated.

## Promoted

- `compiler/internal/narra/CLAUDE.md`: kind letters in `cnames.go`, operation before type in helper names, `enumAlloc` and `.u->`, `narra_drop` worklist, polymorphic recursion check, `Result` builders, prefix-matched module-list tests.
- `DECISIONS.md`: D54 (hashed C names), D55 (heap enum payloads, `Box<T>`), D56 (polymorphic recursion), D57 (`Result` I/O).

## Enforcement ideas (no node written)

- A test that fails if `.u.` or a hand-written `{.tag=` appears in `emit*.go` or `std_*.go`. Small, but a grep test is crude; worth a node `lint-emit-rules` only if the mistake repeats.
- The C name table already panics on clashes, so that rule is enforced.

## Open follow-ups

- `interp` benchmark is about 2x slower (0.20 s to 0.41-0.49 s) until node perceus.
- `std/env` `args` reads `/proc/self/cmdline` off macOS; Linux is untested.
- User code recursing about 20000 deep overflows the C stack; only drops are iterative.
- `Result<void, string>` is not possible; `writeText` returns the byte count.

## Ready next

`plan.py ready`: maps, strings-std, effects, json-diagnostics, opaque-distinct.
