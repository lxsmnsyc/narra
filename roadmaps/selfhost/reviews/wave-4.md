# Wave 4

## Merged

- strings-std: [#12](https://github.com/lxsmnsyc/narra/pull/12)
- maps: [#13](https://github.com/lxsmnsyc/narra/pull/13)
- effects: [#14](https://github.com/lxsmnsyc/narra/pull/14)
- json-diagnostics: [#15](https://github.com/lxsmnsyc/narra/pull/15)
- selfhost-spike: [#16](https://github.com/lxsmnsyc/narra/pull/16) (gate; user decisions in GOAL.md, 2026-10-30)
- std-convert: [#17](https://github.com/lxsmnsyc/narra/pull/17)
- opaque-distinct: [#18](https://github.com/lxsmnsyc/narra/pull/18)
- top-level-const: [#19](https://github.com/lxsmnsyc/narra/pull/19)

8 nodes, all estimated M and finished M. 3.2M tokens in total.

## Process lessons

- Every pull request edited files outside its brief's owned paths, a few lines each. The user approved accepting these: the orchestrator adds the files to the brief's `shared` list, then `plan.py can-merge` decides. Briefs for nodes that touch the checker should list `check.go`, `check_call.go`, `modules_check.go`, `cnames.go`, `diagnostic_codes.go` and `compiler/README.md` up front.
- Incident: while merging #13, a shell chain deleted the remote `feat/maps` branch while the pull request was open, which closed it. The orchestrator restored the branch at the same commit, reopened the pull request and merged it. No code was lost. Rule: run merge steps one command at a time, and never chain a branch delete after a check that can fail.
- Fresh-context review caught real bugs in five of eight nodes: `StringBuilder` nameable without its runtime emitted (#12), the map key Equatable/Hashable check skipped in signatures (#13), private higher-order functions wrongly strict on effects (#14), `clone` dropping `StringLit.End`, which broke a fix in generic code (#15), and a constant/type name clash not detected (#19).
- The fast check took 27 to 180 s. It took 180 s once with no workers running, and 35 to 56 s later. Not investigated; if it recurs, profile the tests (the sanitizer builds and `TestSelfhostLexer` are the likely cost). GOAL.md still says "about 10 s".
- Decision 5 of the spike gate (port soon): every m4-m7 node except opaque-distinct is parked as `later`.

## Contracts

No change requests this wave. `module-graph` and `trait-model` stay at v3. The format table in `diagnostic_codes.go` now acts as the registry for diagnostic codes; the `diagnostic` contract still describes the `Diagnostic` shape.

## Briefs changed

- selfhost-front was split in three, because the spike estimated 3500-4000 Narra lines:
  - selfhost-ast (new, M, full brief): `narra parse` AST dump on the Go side, Narra AST types with node ids, and the Narra printer.
  - selfhost-front (L, full brief): the parser, `TestSelfhostParser`, and replacing the lexer stubs.
  - selfhost-modules (new, M, outline): `narra.toml` and package loading.
- selfhost-check-decls: now also depends on selfhost-modules; a note says to refine it after selfhost-front fixes the node id scheme.
- GOAL.md Coverage lists the new node ids.

## Promoted

- `compiler/internal/narra/CLAUDE.md`, 12 rules: codes from the format table and `TestEveryFormatHasCode`, how to add a code, exact fixes for TypeScript habits, std type names only where their runtime is emitted, `Builtin.Effects` on every builtin, `Mutable` argument check order, statement calls through `e.temp`, queued map key checks, `underlying` before type switches, constants through `e.constRef`, name resolution order, `clone.go` copies every field.
- `selfhost/CLAUDE.md` (new), 10 rules from the spike: dump comparison, lexer quirks, one `mutable` context struct, node ids and side tables, `failed` error flow, loops over recursion, `uses` on exports, closures in struct fields, top-level `const` and std conversions, package location.
- `DECISIONS.md`: D58 (effect checking), D59 (diagnostic codes by format), D60 (byte offsets, copying slices), D61 (top-level constant emission), D62 (distinct types), D63 (self-hosted compiler design, port soon), D64 (map insertion order).
- `CALIBRATION.md`: 8 rows and two notes (token range for M nodes, files that checker nodes always touch).

## Enforcement ideas (no node written)

- Diagnostic codes are enforced by `TestEveryFormatHasCode`.
- A test that every registered `Builtin` has a non-nil `Effects` (or an explicit empty set). Small; worth adding to the next node that registers builtins.
- A test that `cloneExpr` copies every field, by reflection over the AST structs. Would have caught the #15 bug. Worth a node `lint-clone` if another clone bug appears.

## Open follow-ups

- Strings: `ByteIndex` is `int`, slices copy; `Str.find`, unicode functions and char iterators are not built. `StringBuilder` has no equality or `Show`.
- Maps: no `Equatable`, `Hashable`, `Ordered` or `Show`, so no `==` on maps and no map in a `derives` struct (spike gap 13). No `Map.values`, `has` or `get`. `Map.keys` and `Map.entries` are arrays outside a loop until generators.
- Effects: function types carry no effects, so calls through struct fields, call results and arrays of functions count as pure. User `implement` bodies are not followed. `uses` before a `mutable` parameter is not parsed. No `unsafe` source until c-interop.
- Diagnostics: no extents (range end equals start); the lexer and parser stop at the first error; only 7 codes have `explain` texts; no fix for `'abc'` strings.
- Distinct types: no `implements`, `derives`, `Show` or `Ordered`; no generic distinct types.
- Constants: built on first use, not thread safe (revisit in threads); `1 / 0` panics at first use; no `@const` functions or `lazy`.
- Self-hosting gaps the user did not take: no `Result<void, E>`, no recover, no 32-bit unsigned type, no sort, no `eprint` (use `/dev/stderr`), single-expression `match` arms, no file embedding for the runtime C (selfhost-emit needs a generator), contextual keywords `from`/`as`/`of` cannot be names.
- Carried over: the `interp` benchmark is about 2x slower until perceus (parked); user code overflows the C stack near 20000 frames; `std/env` `args` untested on Linux.
- Fast check time: see Process lessons.

## Milestones

- m3-core: done (14 nodes).
- m8-selfhost: spike done; 7 nodes open, all on one chain. Must line: all must work placed by wave 7.
- m4-m7: parked, except opaque-distinct (done).

## Ready next

`plan.py ready`: selfhost-ast. The self-hosting chain is serial (ast, front, modules, check-decls, check-bodies, emit, bootstrap), so one worker runs at a time unless a parked node is unparked.
